package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// heartbeatInterval keeps the SSE connection from looking idle to any
// intermediary (proxy, load balancer) that might otherwise time it out.
const heartbeatInterval = 20 * time.Second

// EventsEndpoints serves the admin-only SSE stream of picture pipeline
// events (PENDING -> READY/FAILED), replacing client-side polling.
type EventsEndpoints struct {
	hub     *services.PictureEventHub
	issuer  ports.ITokenIssuer
	tickets ports.IStreamTicketStore
	// ctx is the application's root context (cancelled on SIGINT/SIGTERM),
	// not a per-request one - watched so the stream loop exits promptly on
	// shutdown instead of blocking main's graceful-shutdown window.
	ctx context.Context
}

func NewEventsEndpoints(hub *services.PictureEventHub, issuer ports.ITokenIssuer, tickets ports.IStreamTicketStore, ctx context.Context) *EventsEndpoints {
	return &EventsEndpoints{hub: hub, issuer: issuer, tickets: tickets, ctx: ctx}
}

// Routes mounts the stream bare (EventSource cannot set custom headers, so
// its own authentication - see authenticateStream - accepts a ?ticket=
// query param instead) alongside a normal RequireAuth+RequireAdmin group for
// minting one. Two handlers on one mount with different middleware mirrors
// GameEndpoints.Routes's /{id}/ws vs the REST sub-group.
func (e *EventsEndpoints) Routes(rateCfg RateLimitConfig) chi.Router {
	r := chi.NewRouter()
	r.Get("/", e.stream)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))
		r.Use(RequireAuth(e.issuer))
		r.Use(RequireAdmin)
		r.With(ticketRateLimit(rateCfg)).Post("/ticket", Wrap(e.mintTicket))
	})
	return r
}

// mintTicket godoc
//
//	@Summary		Mint an SSE connection ticket
//	@Description	Issues a single-use ticket to present as `?ticket=...` to `GET /events` (EventSource can't set headers). Admin-only, same as the stream itself.
//	@Tags			events
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.StreamTicketResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/events/ticket [post]
func (e *EventsEndpoints) mintTicket(w http.ResponseWriter, r *http.Request) error {
	claims, ok := claimsFromContext(r.Context())
	if !ok {
		return ports.ErrUnauthenticated
	}
	token, expiresAt, err := e.tickets.Issue(r.Context(), ports.StreamTicket{
		UserID:  claims.UserID,
		Role:    claims.Role,
		Purpose: ports.TicketPurposeEvents,
	})
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.StreamTicketResponse{Ticket: token, ExpiresAt: expiresAt})
	return nil
}

func (e *EventsEndpoints) stream(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateStream(r, e.issuer, e.tickets, ports.TicketPurposeEvents, "")
	if err != nil {
		handleError(w, err)
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
			data, err := json.Marshal(dto.PictureEventPayload{
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
