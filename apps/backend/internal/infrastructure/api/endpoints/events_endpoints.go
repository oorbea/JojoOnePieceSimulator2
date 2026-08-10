package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// heartbeatInterval keeps the SSE connection from looking idle to any
// intermediary (proxy, load balancer) that might otherwise time it out.
const heartbeatInterval = 20 * time.Second

// pictureEventDTO is PictureEvent's wire shape - camelCase to match every
// other response DTO in this API.
type pictureEventDTO struct {
	Kind      string `json:"kind"`
	SubjectID string `json:"subjectId"`
	Status    string `json:"status"`
}

// EventsEndpoints serves the admin-only SSE stream of picture pipeline
// events (PENDING -> READY/FAILED), replacing client-side polling.
type EventsEndpoints struct {
	hub    *services.PictureEventHub
	issuer ports.ITokenIssuer
	// ctx is the application's root context (cancelled on SIGINT/SIGTERM),
	// not a per-request one - watched so the stream loop exits promptly on
	// shutdown instead of blocking main's graceful-shutdown window.
	ctx context.Context
}

func NewEventsEndpoints(hub *services.PictureEventHub, issuer ports.ITokenIssuer, ctx context.Context) *EventsEndpoints {
	return &EventsEndpoints{hub: hub, issuer: issuer, ctx: ctx}
}

// Routes mounts the stream. Deliberately not behind RequireAuth: EventSource
// cannot set custom headers, so authentication here also accepts the token
// as a query param (see authenticate) - kept local to this handler rather
// than changing RequireAuth, so no other route gains query-param auth.
func (e *EventsEndpoints) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", e.stream)
	return r
}

// authenticate validates the caller the same way RequireAuth does (same
// ports.ITokenIssuer, same Bearer header support) but additionally accepts
// ?token=<jwt> for browsers' EventSource, which cannot set request headers.
// The token appearing in a URL (and therefore in access logs) is an
// accepted tradeoff for this one endpoint - see the SSE design note.
func (e *EventsEndpoints) authenticate(r *http.Request) (ports.Claims, error) {
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, bearerPrefix) {
		return e.issuer.Parse(strings.TrimSpace(header[len(bearerPrefix):]))
	}
	if token := r.URL.Query().Get("token"); token != "" {
		return e.issuer.Parse(token)
	}
	return ports.Claims{}, ports.ErrUnauthenticated
}

func (e *EventsEndpoints) stream(w http.ResponseWriter, r *http.Request) {
	claims, err := e.authenticate(r)
	if err != nil {
		handleError(w, ports.ErrUnauthenticated)
		return
	}
	if claims.Role != enums.Admin {
		handleError(w, ports.ErrForbidden)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Hints any buffering proxy in front of this (e.g. Nginx Proxy Manager)
	// to flush immediately - doesn't replace the manual `proxy_buffering
	// off` config NPM still needs, but costs nothing to also send.
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe := e.hub.Subscribe()
	defer unsubscribe()

	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-e.ctx.Done():
			return
		case <-heartbeat.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case evt, ok := <-events:
			if !ok {
				return
			}
			data, err := json.Marshal(pictureEventDTO{
				Kind:      evt.Kind.String(),
				SubjectID: evt.SubjectID,
				Status:    evt.Status.String(),
			})
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: picture\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}
