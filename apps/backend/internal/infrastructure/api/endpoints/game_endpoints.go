package endpoints

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// GameWSConfig configures the parts of GameEndpoints that only the
// WebSocket route needs.
type GameWSConfig struct {
	// VotingWindow mirrors services.VotingPolicy.Window and is now only a
	// fallback: the authoritative per-lobby window travels on
	// services.GameEvent.VotingWindow (each lobby can configure its own),
	// and this is only consulted if that value comes back zero (e.g. a
	// pre-this-tanda event source in a test double).
	VotingWindow time.Duration
	// AllowedOrigins mirrors CORSConfig.AllowedOrigins - reused to build the
	// WebSocket upgrade's origin allowlist (see originPatterns).
	AllowedOrigins []string
	// ResolveStandPicture/ResolveDevilFruitPicture/ResolveStagePicture
	// resolve a loadout's/stage's picture keys into URLs - same functions
	// StandEndpoints/DevilFruitEndpoints/StageEndpoints already use
	// (StandService.PictureURL etc).
	ResolveStandPicture      dto.PictureURLResolver
	ResolveDevilFruitPicture dto.PictureURLResolver
	ResolveStagePicture      dto.PictureURLResolver
	// ResolveAvatarPicture resolves a participant's avatar thumb key into a
	// URL - same function UserEndpoints already uses (UserService.AvatarURL).
	ResolveAvatarPicture dto.PictureURLResolver
}

// GameEndpoints wires the game feature's HTTP surface (creation, discovery,
// resume) and its WebSocket surface (everything that needs live fan-out) to
// GameService. See router.go for why /games/{id}/ws is mounted outside the
// normal Timeout+RequireAuth group.
type GameEndpoints struct {
	svc         *services.GameService
	hub         *services.GameEventHub
	stages      ports.IStageRepository
	stands      ports.IStandRepository
	devilFruits ports.IDevilFruitRepository
	users       ports.IUserRepository
	issuer      ports.ITokenIssuer
	tickets     ports.IStreamTicketStore
	// ctx is the application's root context (cancelled on SIGINT/SIGTERM),
	// watched by every open WebSocket so it can exit promptly on shutdown -
	// same reasoning as EventsEndpoints.ctx.
	ctx context.Context
	cfg GameWSConfig

	conns connRegistry
	media dto.MediaURLBuilder
}

// SetMediaURLBuilder - see StandEndpoints.SetMediaURLBuilder's doc.
func (e *GameEndpoints) SetMediaURLBuilder(media dto.MediaURLBuilder) {
	e.media = media
}

// NewGameEndpoints builds a GameEndpoints. stages/users back the per-viewer
// Stage description resolution (see stageTextResolver): a live Game is one
// instance shared by every participant, so a Stage frozen into a Round
// can't carry an already-resolved description for each viewer - each
// response instead looks up the viewer's own configured user.Language() and
// re-resolves the description at serialize time. stands/devilFruits back the
// same re-resolution for a loadout's Stand/DevilFruit text (see
// powerTextResolver).
func NewGameEndpoints(svc *services.GameService, hub *services.GameEventHub, stages ports.IStageRepository, stands ports.IStandRepository, devilFruits ports.IDevilFruitRepository, users ports.IUserRepository, issuer ports.ITokenIssuer, tickets ports.IStreamTicketStore, ctx context.Context, cfg GameWSConfig) *GameEndpoints {
	return &GameEndpoints{svc: svc, hub: hub, stages: stages, stands: stands, devilFruits: devilFruits, users: users, issuer: issuer, tickets: tickets, ctx: ctx, cfg: cfg, conns: newConnRegistry()}
}

// viewerLocale resolves self's preferred locale from their user record,
// defaulting to enums.EnGB for a bot (no UserID) or if the lookup fails -
// this must never block rendering a game state.
func (e *GameEndpoints) viewerLocale(ctx context.Context, g *game.Game, self game.ParticipantID) enums.Locale {
	p, ok := g.Participant(self)
	if !ok || p.UserID() == nil {
		return enums.EnGB
	}
	u, err := e.users.FindByID(ctx, *p.UserID())
	if err != nil {
		return enums.EnGB
	}
	return u.Language()
}

// stageTextResolver builds a dto.StageTextResolver bound to locale - each
// call resolves the Stage's translations and picks the first available
// locale in enums.FallbackChain(locale) (defense in depth: writes always
// populate all three, per the owner's decision, so the fallback should
// never actually trigger in practice).
func (e *GameEndpoints) stageTextResolver(locale enums.Locale) dto.StageTextResolver {
	return func(ctx context.Context, id game.StageID) (string, error) {
		translations, err := e.stages.Translations(ctx, id)
		if err != nil {
			return "", err
		}
		for _, l := range enums.FallbackChain(locale) {
			if description, ok := translations[l]; ok {
				return description, nil
			}
		}
		return "", nil
	}
}

// standTextResolver/devilFruitTextResolver are stageTextResolver's analogue
// for a loadout's Stand/DevilFruit (see dto.PowerTextResolver) - same
// enums.FallbackChain walk, over ports.PowerTranslations instead of a plain
// description string.
func (e *GameEndpoints) standTextResolver(locale enums.Locale) dto.PowerTextResolver {
	return func(ctx context.Context, id powers.PowerID) (ports.PowerContent, error) {
		translations, err := e.stands.Translations(ctx, id)
		if err != nil {
			return ports.PowerContent{}, err
		}
		for _, l := range enums.FallbackChain(locale) {
			if content, ok := translations[l]; ok {
				return content, nil
			}
		}
		return ports.PowerContent{}, nil
	}
}

func (e *GameEndpoints) devilFruitTextResolver(locale enums.Locale) dto.PowerTextResolver {
	return func(ctx context.Context, id powers.PowerID) (ports.PowerContent, error) {
		translations, err := e.devilFruits.Translations(ctx, id)
		if err != nil {
			return ports.PowerContent{}, err
		}
		for _, l := range enums.FallbackChain(locale) {
			if content, ok := translations[l]; ok {
				return content, nil
			}
		}
		return ports.PowerContent{}, nil
	}
}

// Routes returns the /games sub-router. Unlike every other Routes method in
// this package, this one applies its own Timeout+RequireAuth to the REST
// sub-group instead of relying on the caller's group, precisely so the bare
// /{id}/ws route can be mounted alongside it without inheriting either -
// see router.go's doc comment.
func (e *GameEndpoints) Routes(rateCfg RateLimitConfig) chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}/ws", e.serveWS)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))
		r.Use(RequireAuth(e.issuer))
		write := writeRateLimit(rateCfg)
		read := readRateLimit(rateCfg)
		ticket := ticketRateLimit(rateCfg)
		r.With(write).Post("/", Wrap(e.create))
		r.With(write).Post("/join", Wrap(e.join))
		r.With(read).Get("/public", Wrap(e.listPublic))
		r.With(read).Get("/preview", Wrap(e.preview))
		r.With(read).Get("/me", Wrap(e.getMine))
		r.With(read).Get("/{id}", Wrap(e.get))
		r.With(read).Get("/by-code/{code}", Wrap(e.getByCode))
		r.With(write).Post("/{id}/join", Wrap(e.joinByID))
		r.With(write).Post("/{id}/leave", Wrap(e.leave))
		r.With(write).Patch("/{id}/config", Wrap(e.editConfig))
		r.With(ticket).Post("/{id}/ws-ticket", Wrap(e.mintWSTicket))
		r.With(ticket).Post("/{id}/invite", Wrap(e.createInvite))
		r.With(read).Get("/invite/{token}", Wrap(e.invitePreview))
		r.With(write).Post("/join-invite", Wrap(e.joinByInvite))
	})
	// PUBLIC - deliberately outside the Group above (and its RequireAuth),
	// same reasoning as /{id}/ws: an invite link may be opened by a visitor
	// who has never logged in, and the whole point of this route is to
	// answer "is this link still good" before asking them to. See
	// GameService.InviteStatus's doc for what it withholds precisely
	// because this route has no bearer token to gate it.
	r.With(inviteStatusRateLimit(rateCfg)).Get("/invite/{token}/status", Wrap(e.inviteStatus))
	return r
}

// mintWSTicket godoc
//
//	@Summary		Mint a game WebSocket connection ticket
//	@Description	Issues a single-use ticket, scoped to this game, to present as `?ticket=...` to `GET /games/{id}/ws` (browsers can't set headers on a WebSocket upgrade). Requires the same authorization the socket itself requires (a seated participant).
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Game ID"
//	@Success		200	{object}	dto.StreamTicketResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/games/{id}/ws-ticket [post]
func (e *GameEndpoints) mintWSTicket(w http.ResponseWriter, r *http.Request) error {
	claims, ok := claimsFromContext(r.Context())
	if !ok {
		return ports.ErrUnauthenticated
	}
	gameID, err := game.ParseGameID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	g, err := e.svc.GetGame(r.Context(), gameID)
	if err != nil {
		return err
	}
	if _, err := resolveParticipant(g, claims.UserID); err != nil {
		return err
	}
	token, expiresAt, err := e.tickets.Issue(r.Context(), ports.StreamTicket{
		UserID:   claims.UserID,
		Role:     claims.Role,
		Purpose:  ports.TicketPurposeGameWS,
		Resource: gameID.String(),
	})
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.StreamTicketResponse{Ticket: token, ExpiresAt: expiresAt})
	return nil
}

const defaultPublicListLimit = 20
const maxPublicListLimit = 50

// resolveParticipant finds the ParticipantID seated in g for userID, the
// same scan GameService.JoinByCode itself does. A caller who isn't seated
// gets ports.ErrForbidden - a guessed GameID or a leaked join code must not
// dump a lobby's roster and loadouts to a stranger.
func resolveParticipant(g *game.Game, userID user.UserID) (game.ParticipantID, error) {
	for _, p := range g.Participants() {
		if uid := p.UserID(); uid != nil && *uid == userID {
			return p.ID(), nil
		}
	}
	return game.NilParticipantID, ports.ErrForbidden
}

func (e *GameEndpoints) respondState(w http.ResponseWriter, r *http.Request, g *game.Game, self game.ParticipantID, status int) error {
	code, err := e.svc.GameCode(r.Context(), g.ID())
	if err != nil {
		return err
	}
	locale := e.viewerLocale(r.Context(), g, self)
	var deadlines dto.GameStateDeadlines
	if t, ok := e.svc.RevealEndsAt(g.ID()); ok {
		deadlines.RevealEndsAt = &t
	}
	if t, ok := e.svc.VotingEndsAt(g.ID()); ok {
		deadlines.VotingEndsAt = &t
	}
	if t, ok := e.svc.ResultEndsAt(g.ID()); ok {
		deadlines.ResultEndsAt = &t
	}
	if t, ok := e.svc.SummaryEndsAt(g.ID()); ok {
		deadlines.SummaryEndsAt = &t
	}
	resp, err := dto.NewGameStateResponse(r.Context(), g, code, self,
		e.cfg.ResolveStandPicture, e.cfg.ResolveDevilFruitPicture, e.cfg.ResolveStagePicture, e.cfg.ResolveAvatarPicture,
		e.media,
		e.stageTextResolver(locale), e.standTextResolver(locale), e.devilFruitTextResolver(locale), deadlines)
	if err != nil {
		return err
	}
	writeJSON(w, status, resp)
	return nil
}

// create godoc
//
//	@Summary		Create a game
//	@Description	Seats the caller as host in a new LOBBY. The response's `game.code` is the 6-character join code.
//	@Tags			games
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateGameRequest	true	"Game to create"
//	@Success		201		{object}	dto.GameStateResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Failure		501		{object}	dto.ErrorResponse
//	@Failure		503		{object}	dto.ErrorResponse
//	@Router			/games [post]
func (e *GameEndpoints) create(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	var req dto.CreateGameRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	input, err := req.Validate()
	if err != nil {
		return err
	}

	g, _, err := e.svc.CreateGame(r.Context(), claims.UserID, input)
	if err != nil {
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	w.Header().Set("Location", fmt.Sprintf("/api/v1/games/%s", g.ID()))
	return e.respondState(w, r, g, self, http.StatusCreated)
}

// join godoc
//
//	@Summary		Join a game by its code
//	@Tags			games
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.JoinGameRequest	true	"Join code"
//	@Success		200		{object}	dto.GameStateResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/games/join [post]
func (e *GameEndpoints) join(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	var req dto.JoinGameRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	code, err := req.Validate()
	if err != nil {
		return err
	}

	g, err := e.svc.JoinByCode(r.Context(), code, claims.UserID)
	if err != nil {
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	return e.respondState(w, r, g, self, http.StatusOK)
}

// get godoc
//
//	@Summary		Get a game by id
//	@Description	Only a seated participant may resume a game this way.
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Game id (UUID)"
//	@Success		200	{object}	dto.GameStateResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/games/{id} [get]
func (e *GameEndpoints) get(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	id, err := game.ParseGameID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	g, err := e.svc.GetGame(r.Context(), id)
	if err != nil {
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	return e.respondState(w, r, g, self, http.StatusOK)
}

// getMine godoc
//
//	@Summary		Resume the caller's active game, if any
//	@Description	Returns whichever Game the caller is currently seated in as a human participant (excluding FINISHED/ABORTED ones, which keep their own short-lived result screen instead), so reopening the app can route straight back to a lobby/match instead of leaving the player stranded with no way back in. 204 if there is none.
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GameStateResponse
//	@Success		204
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/games/me [get]
func (e *GameEndpoints) getMine(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	g, err := e.svc.ActiveGameForUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, ports.ErrGameNotFound) {
			writeJSON(w, http.StatusNoContent, nil)
			return nil
		}
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	return e.respondState(w, r, g, self, http.StatusOK)
}

// getByCode godoc
//
//	@Summary		Get a game by its join code
//	@Description	Only a seated participant may resume a game this way - this is a resume route, not a pre-join preview.
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Param			code	path		string	true	"Join code"
//	@Success		200		{object}	dto.GameStateResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/games/by-code/{code} [get]
func (e *GameEndpoints) getByCode(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	code := chi.URLParam(r, "code")
	g, err := e.svc.GetGameByCode(r.Context(), code)
	if err != nil {
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	return e.respondState(w, r, g, self, http.StatusOK)
}

// listPublic godoc
//
//	@Summary		Browse public lobbies
//	@Description	Roster-free, join-code-free summaries of lobbies currently joinable through the public browser. Works for any authenticated caller, not just participants.
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"Max results (default 20, capped at 50)"
//	@Success		200		{object}	dto.PublicLobbyListResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/games/public [get]
func (e *GameEndpoints) listPublic(w http.ResponseWriter, r *http.Request) error {
	if _, ok := ClaimsFromRequest(r); !ok {
		return ports.ErrUnauthenticated
	}
	limit := defaultPublicListLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxPublicListLimit {
		limit = maxPublicListLimit
	}
	listings, err := e.svc.ListPublicLobbies(r.Context(), limit)
	if err != nil {
		return err
	}
	items := make([]dto.PublicLobbyResponse, len(listings))
	for i, l := range listings {
		items[i] = dto.NewPublicLobbyResponse(l)
	}
	writeJSON(w, http.StatusOK, dto.PublicLobbyListResponse{Items: items})
	return nil
}

// preview godoc
//
//	@Summary		Preview a lobby by its join code
//	@Description	Roster-free summary reachable without being a participant - the code itself is the credential, so this also works for PRIVATE lobbies. Not roster/loadout-bearing, unlike GET /games/by-code/{code}.
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Param			code	query		string	true	"Join code"
//	@Success		200		{object}	dto.LobbyPreviewResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/games/preview [get]
func (e *GameEndpoints) preview(w http.ResponseWriter, r *http.Request) error {
	if _, ok := ClaimsFromRequest(r); !ok {
		return ports.ErrUnauthenticated
	}
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("code")))
	if code == "" {
		return &dto.ValidationError{Errors: []dto.FieldError{{Field: "code", Code: dto.ValRequired, Message: "code is required"}}}
	}
	listing, err := e.svc.PreviewByCode(r.Context(), code)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewLobbyPreviewResponse(code, listing))
	return nil
}

// joinByID godoc
//
//	@Summary		Join a public lobby by id
//	@Description	Only reachable for a PUBLIC, unlocked lobby - see POST /games/join for the code-based path, which also works for PRIVATE lobbies.
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Game id (UUID)"
//	@Success		200	{object}	dto.GameStateResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/games/{id}/join [post]
func (e *GameEndpoints) joinByID(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	id, err := game.ParseGameID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	g, err := e.svc.JoinByID(r.Context(), id, claims.UserID)
	if err != nil {
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	return e.respondState(w, r, g, self, http.StatusOK)
}

// editConfig godoc
//
//	@Summary		Edit a lobby's configuration
//	@Description	Host-only, LOBBY-only. Replaces the whole Config, exactly like POST /games's body.
//	@Tags			games
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Game id (UUID)"
//	@Param			request	body		dto.UpdateConfigPayload	true	"New configuration"
//	@Success		200		{object}	dto.GameStateResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/games/{id}/config [patch]
func (e *GameEndpoints) editConfig(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	id, err := game.ParseGameID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	var req dto.UpdateConfigPayload
	if err := decode(w, r, &req); err != nil {
		return err
	}
	input, err := req.Validate()
	if err != nil {
		return err
	}
	g, err := e.svc.GetGame(r.Context(), id)
	if err != nil {
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	g, err = e.svc.EditLobbyConfig(r.Context(), id, self, input)
	if err != nil {
		return err
	}
	return e.respondState(w, r, g, self, http.StatusOK)
}

// leave godoc
//
//	@Summary		Leave a game
//	@Description	Removes the caller from the game entirely. Exists as a REST route (mirroring the WS LEAVE command GameService.LeaveGame already backs) specifically for a visitor who has opened a lobby share link while already seated somewhere else and confirmed they want to switch - that flow has no open socket to the game being left.
//	@Tags			games
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Game id (UUID)"
//	@Success		204
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/games/{id}/leave [post]
func (e *GameEndpoints) leave(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	id, err := game.ParseGameID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	g, err := e.svc.GetGame(r.Context(), id)
	if err != nil {
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	if _, err := e.svc.LeaveGame(r.Context(), id, self); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}

// createInvite godoc
//
//	@Summary		Mint a lobby share-link invite token
//	@Description	Any seated human participant may mint one, not only the host - see POST /games/{id}/ws-ticket for the same authorization shape. Multi-use and short-lived (see GameService.CreateInvite); dies early if the host rotates the join code before it expires on its own.
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Game id (UUID)"
//	@Success		200	{object}	dto.GameInviteResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/games/{id}/invite [post]
func (e *GameEndpoints) createInvite(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	id, err := game.ParseGameID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	token, expiresAt, err := e.svc.CreateInvite(r.Context(), id, claims.UserID)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.GameInviteResponse{Token: token, ExpiresAt: expiresAt})
	return nil
}

// inviteStatus godoc
//
//	@Summary		Check whether a lobby invite link is still usable
//	@Description	PUBLIC - no bearer token required, reachable by a visitor who has not logged in yet (or an unfurl bot). Answers only VALID/EXPIRED, deliberately nothing else - see dto.InviteStatusResponse's doc. Always 200; Cache-Control: no-store so neither a cache nor the HTTP status itself becomes an existence oracle.
//	@Tags			games
//	@Produce		json
//	@Param			token	path		string	true	"Invite token"
//	@Success		200		{object}	dto.InviteStatusResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/games/invite/{token}/status [get]
func (e *GameEndpoints) inviteStatus(w http.ResponseWriter, r *http.Request) error {
	token := chi.URLParam(r, "token")
	status := e.svc.InviteStatus(r.Context(), token)
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, dto.NewInviteStatusResponse(status))
	return nil
}

// invitePreview godoc
//
//	@Summary		Preview a lobby by its invite token
//	@Description	Authenticated counterpart of GET /games/preview - the token is the credential, so this also works for PRIVATE lobbies. Roster-free, join-code-free, exactly like every other lobby preview shape.
//	@Tags			games
//	@Produce		json
//	@Security		BearerAuth
//	@Param			token	path		string	true	"Invite token"
//	@Success		200		{object}	dto.InvitePreviewResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/games/invite/{token} [get]
func (e *GameEndpoints) invitePreview(w http.ResponseWriter, r *http.Request) error {
	if _, ok := ClaimsFromRequest(r); !ok {
		return ports.ErrUnauthenticated
	}
	token := chi.URLParam(r, "token")
	listing, err := e.svc.InvitePreview(r.Context(), token)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewInvitePreviewResponse(listing))
	return nil
}

// joinByInvite godoc
//
//	@Summary		Join a lobby by its invite token
//	@Description	Works against PRIVATE lobbies too - see GameService.JoinByInvite's doc for exactly how this differs from POST /games/join (idempotent if already seated, a dedicated 409 if the game already started, revoked if the host has since rotated the join code).
//	@Tags			games
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.JoinByInviteRequest	true	"Invite token"
//	@Success		200		{object}	dto.GameStateResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/games/join-invite [post]
func (e *GameEndpoints) joinByInvite(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return ports.ErrUnauthenticated
	}
	var req dto.JoinByInviteRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	token, err := req.Validate()
	if err != nil {
		return err
	}
	g, err := e.svc.JoinByInvite(r.Context(), token, claims.UserID)
	if err != nil {
		return err
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		return err
	}
	return e.respondState(w, r, g, self, http.StatusOK)
}
