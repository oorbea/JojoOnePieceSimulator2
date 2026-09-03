package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// wsReadLimit bounds a single command frame. Most commands (VOTE/
// REMOVE_BOT) are a few dozen bytes, but UPDATE_CONFIG can carry a
// PoolFilter with a sizeable banned-PowerID list, so this is generous
// headroom for that one command without letting a client send unbounded
// frames.
const wsReadLimit = 16384

// wsWriteTimeout bounds each individual write (state frame or ping) so a
// stalled TCP connection can't hang the write pump forever.
const wsWriteTimeout = 10 * time.Second

// wsOutboundBuffer bounds how many frames can queue for a connection before
// it is considered too slow and closed - see send's doc.
const wsOutboundBuffer = 32

// errUnknownCommand is returned by dispatch when a ClientCommand's Type
// doesn't match any known command.
var errUnknownCommand = errors.New("unknown command")

// connKey identifies one participant's presence in one game, used by
// connRegistry to make Reconnect/Disconnect correct across multiple
// simultaneous sockets for the same participant (e.g. two browser tabs).
type connKey struct {
	game        game.GameID
	participant game.ParticipantID
}

// connRegistry reference-counts open sockets per connKey, so Reconnect only
// fires on the first socket and Disconnect only on the last - closing one
// of several tabs must not mark the participant unreachable.
type connRegistry struct {
	mu sync.Mutex
	n  map[connKey]int
}

func newConnRegistry() connRegistry {
	return connRegistry{n: make(map[connKey]int)}
}

func (r *connRegistry) acquire(k connKey) (first bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	first = r.n[k] == 0
	r.n[k]++
	return first
}

func (r *connRegistry) release(k connKey) (last bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.n[k]--
	if r.n[k] <= 0 {
		delete(r.n, k)
		return true
	}
	return false
}

// originPatterns strips the scheme off each configured CORS origin, since
// websocket.AcceptOptions.OriginPatterns wants host[:port] patterns, not
// full origin URLs. An empty result leaves OriginPatterns nil, meaning only
// same-origin upgrades are allowed - the same deny-by-default posture as an
// unconfigured CORSConfig.
func originPatterns(origins []string) []string {
	if len(origins) == 0 {
		return nil
	}
	out := make([]string, 0, len(origins))
	for _, o := range origins {
		if idx := strings.Index(o, "://"); idx >= 0 {
			o = o[idx+len("://"):]
		}
		out = append(out, o)
	}
	return out
}

// serveWS upgrades to a WebSocket for one participant's view of one game.
// See router.go for why this route is mounted outside the normal
// Timeout+RequireAuth group.
func (e *GameEndpoints) serveWS(w http.ResponseWriter, r *http.Request) {
	// ParseGameID runs before authentication because the ticket must be
	// checked against this specific game's id (its Resource) - a malformed
	// id 400s before a ticket is even redeemed/burned.
	gameID, err := game.ParseGameID(chi.URLParam(r, "id"))
	if err != nil {
		handleError(w, err)
		return
	}
	claims, err := authenticateStream(r, e.issuer, e.tickets, ports.TicketPurposeGameWS, gameID.String())
	if err != nil {
		handleError(w, err)
		return
	}
	g, err := e.svc.GetGame(r.Context(), gameID)
	if err != nil {
		handleError(w, err)
		return
	}
	self, err := resolveParticipant(g, claims.UserID)
	if err != nil {
		handleError(w, err)
		return
	}

	// The ticket is already burned by authenticateStream above, before this
	// Accept - so a rejected origin here costs the ticket. Harmless: the
	// frontend always re-mints on its next connect attempt.
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  originPatterns(e.cfg.AllowedOrigins),
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		// Accept has already written its own response on failure.
		return
	}
	conn.SetReadLimit(wsReadLimit)

	// ctx is cancelled either when the request's own context ends or when
	// the app's root context (e.ctx) does, so an open socket cannot outlive
	// a graceful shutdown - same reasoning as EventsEndpoints.
	ctx, cancel := context.WithCancel(r.Context())
	go func() {
		select {
		case <-e.ctx.Done():
			cancel()
		case <-ctx.Done():
		}
	}()

	// Subscribe before anything else that could race a concurrent mutation,
	// so no domain event can slip between the initial STATE snapshot and
	// this connection's subscription.
	events, unsubscribe := e.hub.Subscribe(gameID)

	key := connKey{game: gameID, participant: self}
	first := e.conns.acquire(key)
	if first {
		if _, err := e.svc.Reconnect(ctx, gameID, self); err != nil {
			log.Printf("game ws: reconnect %s/%s: %v", gameID, self, err)
		}
	}

	outbound := make(chan []byte, wsOutboundBuffer)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		e.writePump(ctx, conn, outbound)
	}()
	go func() {
		defer wg.Done()
		e.forwardEvents(ctx, conn, outbound, gameID, self, events)
	}()

	e.pushCurrentState(ctx, conn, outbound, gameID, self)

	// Blocks until the client disconnects, the read limit is exceeded, or
	// ctx is cancelled (Read then errors).
	e.readPump(ctx, conn, outbound, gameID, self)

	cancel()
	wg.Wait()
	unsubscribe()

	if last := e.conns.release(key); last {
		// The request's ctx is already dead; Disconnect still must run - it
		// can reassign the host, abort the game, or early-close a voting
		// window, none of which should wait on a client that already left.
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if _, err := e.svc.Disconnect(bgCtx, gameID, self); err != nil &&
			!errors.Is(err, ports.ErrGameNotFound) && !errors.Is(err, game.ErrParticipantNotFound) {
			log.Printf("game ws: disconnect %s/%s: %v", gameID, self, err)
		}
		bgCancel()
	}
	_ = conn.CloseNow()
}

// writePump is the sole writer on conn (coder/websocket allows only one),
// draining outbound and sending a protocol-level ping every
// heartbeatInterval so intermediaries don't consider the connection idle.
func (e *GameEndpoints) writePump(ctx context.Context, conn *websocket.Conn, outbound <-chan []byte) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, wsWriteTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		case data := <-outbound:
			writeCtx, cancel := context.WithTimeout(ctx, wsWriteTimeout)
			err := conn.Write(writeCtx, websocket.MessageText, data)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

// readPump is the sole reader on conn, dispatching each command and
// answering with a fresh STATE on success or an ERROR frame on failure. A
// malformed command doesn't close the connection - only a protocol
// violation (handled inside conn.Read itself, e.g. the read limit) does.
func (e *GameEndpoints) readPump(ctx context.Context, conn *websocket.Conn, outbound chan<- []byte, gameID game.GameID, self game.ParticipantID) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var cmd dto.ClientCommand
		if err := json.Unmarshal(data, &cmd); err != nil {
			e.sendError(conn, outbound, "", &dto.ValidationError{Errors: []string{err.Error()}})
			continue
		}

		g, err := e.dispatch(ctx, gameID, self, cmd)
		if err != nil {
			e.sendError(conn, outbound, cmd.RequestID, err)
			continue
		}
		// A nil Game means the command changed nothing about *this* game and
		// has nothing to resend (REMATCH: it creates a different game and
		// announces it over the hub instead).
		if g == nil {
			continue
		}
		e.pushState(ctx, conn, outbound, g, self)
	}
}

// dispatch maps one ClientCommand onto exactly one GameService method. See
// dto/game_ws.go's doc on the command constants for what is deliberately
// NOT reachable this way (CreateGame/JoinByCode/GetGame* - plain HTTP only;
// Disconnect/Reconnect - socket lifecycle only; CloseVotingWindow -
// server-internal only).
func (e *GameEndpoints) dispatch(ctx context.Context, gameID game.GameID, self game.ParticipantID, cmd dto.ClientCommand) (*game.Game, error) {
	switch cmd.Type {
	case dto.CommandLeave:
		return e.svc.LeaveGame(ctx, gameID, self)
	case dto.CommandAddBot:
		var p dto.AddBotPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		teamID, err := game.ParseTeamID(p.TeamID)
		if err != nil {
			return nil, err
		}
		return e.svc.AddBot(ctx, gameID, self, teamID)
	case dto.CommandRemoveBot:
		var p dto.RemoveBotPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		botID, err := game.ParseParticipantID(p.BotID)
		if err != nil {
			return nil, err
		}
		return e.svc.RemoveBot(ctx, gameID, self, botID)
	case dto.CommandStart:
		return e.svc.StartGame(ctx, gameID, self)
	case dto.CommandAbort:
		return e.svc.AbortGame(ctx, gameID, self)
	case dto.CommandVote:
		var p dto.VotePayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		return e.svc.CastVote(ctx, gameID, self, game.OptionID(p.Option))
	case dto.CommandResync:
		return e.svc.GetGame(ctx, gameID)
	case dto.CommandRevealReady:
		return e.svc.MarkRevealReady(ctx, gameID, self)
	case dto.CommandSummaryReady:
		return e.svc.MarkSummaryReady(ctx, gameID, self)
	case dto.CommandRematch:
		// Rematch returns the NEW game (and its code), which is deliberately
		// discarded here: this socket belongs to the OLD game, and every
		// client on it - not just the requester - learns the new id from the
		// REMATCH_READY frame Rematch publishes on the old game's hub. A nil
		// Game means dispatch skips its usual STATE resend, which is right:
		// the old game hasn't changed.
		if _, _, err := e.svc.Rematch(ctx, gameID, self); err != nil {
			return nil, err
		}
		return nil, nil
	case dto.CommandSwitchTeam:
		var p dto.SwitchTeamPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		teamID, err := game.ParseTeamID(p.TeamID)
		if err != nil {
			return nil, err
		}
		target := self
		if p.ParticipantID != "" {
			target, err = game.ParseParticipantID(p.ParticipantID)
			if err != nil {
				return nil, err
			}
		}
		return e.svc.SwitchTeam(ctx, gameID, self, target, teamID)
	case dto.CommandMovePlayer:
		var p dto.MovePlayerPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		target, err := game.ParseParticipantID(p.ParticipantID)
		if err != nil {
			return nil, err
		}
		teamID, err := game.ParseTeamID(p.TeamID)
		if err != nil {
			return nil, err
		}
		return e.svc.SwitchTeam(ctx, gameID, self, target, teamID)
	case dto.CommandKick:
		var p dto.KickPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		target, err := game.ParseParticipantID(p.ParticipantID)
		if err != nil {
			return nil, err
		}
		return e.svc.KickParticipant(ctx, gameID, self, target)
	case dto.CommandTransferHost:
		var p dto.TransferHostPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		target, err := game.ParseParticipantID(p.ParticipantID)
		if err != nil {
			return nil, err
		}
		return e.svc.TransferHost(ctx, gameID, self, target)
	case dto.CommandSetLock:
		var p dto.SetLockPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		return e.svc.SetLobbyLocked(ctx, gameID, self, p.Locked)
	case dto.CommandUpdateConfig:
		var p dto.UpdateConfigPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return nil, &dto.ValidationError{Errors: []string{err.Error()}}
		}
		input, err := p.Validate()
		if err != nil {
			return nil, err
		}
		return e.svc.EditLobbyConfig(ctx, gameID, self, input)
	default:
		return nil, errUnknownCommand
	}
}

// forwardEvents relays gameID's domain events to this connection, following
// the "event + snapshot" rule: after any state-changing event, a fresh
// STATE follows so the client never has to reconstruct state from deltas
// and a dropped event (the hub silently drops for a full subscriber) is
// self-healing. VOTE_CAST, GAME_FINISHED and GAME_ABORTED are the
// exceptions - see buildEventFrame's doc.
func (e *GameEndpoints) forwardEvents(ctx context.Context, conn *websocket.Conn, outbound chan<- []byte, gameID game.GameID, self game.ParticipantID, events <-chan services.GameEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-events:
			if !ok {
				return
			}
			votingWindow := evt.VotingWindow
			if votingWindow == 0 {
				votingWindow = e.cfg.VotingWindow
			}
			frameType, payload, resendState := buildEventFrame(evt.Event, votingWindow, evt.RevealMs, evt.SummaryWindow, evt.ClosesAt)
			if frameType == "" {
				continue
			}
			data, err := json.Marshal(dto.ServerFrame{Type: frameType, Payload: payload})
			if err == nil {
				send(conn, outbound, data)
			}
			// A kicked participant's own connection is closed outright
			// instead of getting the usual resend: resolveParticipant would
			// 403 their next RESYNC anyway (they're no longer seated), and
			// an explicit close gives the client a deterministic signal
			// instead of a dangling socket waiting on a timeout.
			if kicked, ok := evt.Event.(game.PlayerKicked); ok && kicked.ParticipantID == self {
				_ = conn.Close(websocket.StatusNormalClosure, "kicked")
				return
			}
			if resendState {
				e.pushCurrentState(ctx, conn, outbound, gameID, self)
			}
		}
	}
}

// buildEventFrame maps one domain event onto its wire frame, reusing
// events.go's wire names verbatim. VOTE_CAST carries no participant/option
// (votes stay hidden until a round resolves - see GameRoundResponse), only
// an anonymous human-vote count, and never triggers a resend, since it's
// high-frequency and self-describing.
// GAME_FINISHED/GAME_ABORTED DO resend. They used to be exceptions, on the
// reasoning that finalizeLocked had already deleted the game so a fresh
// STATE fetch would race ErrGameNotFound - and on the (wrong) claim that the
// event alone "already carries everything a client needs to render the
// terminal screen". It never did: the terminal frames carry no rounds, no
// per-round votes and no stage list, so the final result screen has nothing
// to recap from. finalizeLocked now saves the game under a short TTL
// (services.FinishedGameTTL) instead of deleting it, and holds the game's
// lock for the whole call - which the resend's own read path re-acquires -
// so the STATE that follows is guaranteed to be the fully-finalized one,
// never a half-finalized snapshot.
//
// closesAt is the authoritative deadline the service stamped on the event
// at publish time (services.GameEvent.ClosesAt); it is used verbatim for
// the three timed frames. See frameDeadline for the fallback.
func buildEventFrame(evt game.DomainEvent, votingWindow time.Duration, revealMs time.Duration, summaryWindow time.Duration, closesAt time.Time) (frameType string, payload any, resendState bool) {
	switch e := evt.(type) {
	case game.PlayerJoined:
		return dto.FramePlayerJoined, dto.PlayerJoinedPayload{ParticipantID: e.ParticipantID.String()}, true
	case game.PlayerLeft:
		return dto.FramePlayerLeft, dto.PlayerLeftPayload{ParticipantID: e.ParticipantID.String()}, true
	case game.HostReassigned:
		return dto.FrameHostReassigned, dto.HostReassignedPayload{NewHostID: e.NewHostID.String()}, true
	case game.GameStarted:
		return dto.FrameGameStarted, struct{}{}, true
	case game.LoadoutsAssigned:
		return dto.FrameLoadoutsAssigned, dto.LoadoutsAssignedPayload{
			RoundIndex: e.RoundIndex, RevealMs: revealMs.Milliseconds(),
		}, true
	case game.VotingOpened:
		return dto.FrameVotingOpened, dto.VotingOpenedPayload{
			RoundIndex: e.RoundIndex, ClosesAt: frameDeadline(closesAt, votingWindow),
		}, true
	case game.VoteCast:
		return dto.FrameVoteCast, dto.VoteCastPayload{
			RoundIndex: e.RoundIndex, VotesCast: e.HumanVotesCast, Voters: e.HumanVoters,
		}, false
	case game.RevealReadyChanged:
		return dto.FrameRevealReadyChanged, dto.RevealReadyChangedPayload{
			Ready: e.ReadyCount, Total: e.TotalHumans,
		}, false
	case game.SummaryOpened:
		return dto.FrameSummaryOpened, dto.SummaryOpenedPayload{
			RoundIndex: e.RoundIndex, ClosesAt: frameDeadline(closesAt, summaryWindow),
		}, true
	case game.SummaryReadyChanged:
		return dto.FrameSummaryReadyChanged, dto.SummaryReadyChangedPayload{
			Ready: e.ReadyCount, Total: e.TotalHumans,
		}, false
	case game.TiebreakOpened:
		return dto.FrameTiebreakOpened, dto.TiebreakOpenedPayload{
			RoundIndex: e.RoundIndex, ClosesAt: frameDeadline(closesAt, votingWindow),
		}, true
	case game.RoundResolved:
		return dto.FrameRoundResolved, dto.RoundResolvedPayload{
			RoundIndex: e.RoundIndex, Winner: string(e.Winner), DecidedByCoinFlip: e.DecidedByCoinFlip,
		}, true
	case game.GameFinished:
		return dto.FrameGameFinished, dto.GameFinishedPayload{
			Result: dto.NewGameResultResponse(e.Result),
		}, true
	case game.GameAborted:
		return dto.FrameGameAborted, dto.GameAbortedPayload{Reason: e.Reason}, true
	case game.RematchReady:
		// No resend: this frame is about a *different* game, and the old one
		// is unchanged - the client navigates to the new id instead.
		return dto.FrameRematchReady, dto.RematchReadyPayload{GameID: e.GameID.String()}, false
	case game.TeamChanged:
		return dto.FrameTeamChanged, dto.TeamChangedPayload{
			ParticipantID: e.ParticipantID.String(), FromTeamID: e.FromTeamID.String(), ToTeamID: e.ToTeamID.String(),
		}, true
	case game.PlayerKicked:
		return dto.FramePlayerKicked, dto.PlayerKickedPayload{ParticipantID: e.ParticipantID.String()}, true
	case game.LobbyLockChanged:
		return dto.FrameLobbyLockChanged, dto.LobbyLockChangedPayload{Locked: e.Locked}, true
	case game.ConfigUpdated:
		return dto.FrameConfigUpdated, struct{}{}, true
	default:
		return "", nil, false
	}
}

// frameDeadline renders a timed frame's closesAt. It prefers the
// authoritative deadline the service stamped on the event when it armed the
// phase's own timer, so the client's countdown matches the server's to the
// millisecond no matter how long hub delivery took.
//
// The time.Now()+window fallback only applies to a zero closesAt, which
// should not happen for any of the three timed frames: it would mean the
// event was published without its phase's deadline recorded - i.e. a frame
// delivered after its own window already closed, or a new emit path that
// publishes before arming its timer (see GameService.publish's doc). Kept
// rather than emitting an empty string so a client always gets *some*
// countdown to render.
func frameDeadline(closesAt time.Time, window time.Duration) string {
	if closesAt.IsZero() {
		return time.Now().Add(window).Format(time.RFC3339)
	}
	return closesAt.Format(time.RFC3339)
}

// pushCurrentState fetches the freshest state for gameID and sends it. Used
// both for the connection's initial snapshot and after every
// state-changing event/command.
func (e *GameEndpoints) pushCurrentState(ctx context.Context, conn *websocket.Conn, outbound chan<- []byte, gameID game.GameID, self game.ParticipantID) {
	g, err := e.svc.GetGame(ctx, gameID)
	if err != nil {
		return
	}
	e.pushState(ctx, conn, outbound, g, self)
}

func (e *GameEndpoints) pushState(ctx context.Context, conn *websocket.Conn, outbound chan<- []byte, g *game.Game, self game.ParticipantID) {
	code, err := e.svc.GameCode(ctx, g.ID())
	if err != nil {
		return
	}
	locale := e.viewerLocale(ctx, g, self)
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
	resp, err := dto.NewGameStateResponse(ctx, g, code, self,
		e.cfg.ResolveStandPicture, e.cfg.ResolveDevilFruitPicture, e.cfg.ResolveStagePicture, e.cfg.ResolveAvatarPicture,
		e.stageTextResolver(locale), e.standTextResolver(locale), e.devilFruitTextResolver(locale), deadlines)
	if err != nil {
		log.Printf("game ws: building state for %s: %v", g.ID(), err)
		return
	}
	data, err := json.Marshal(dto.ServerFrame{Type: dto.FrameState, Payload: resp})
	if err != nil {
		return
	}
	send(conn, outbound, data)
}

// sendError builds and sends an ERROR frame, reusing errorCode(err) so the
// frontend's errors.<CODE> i18n lookup works over the socket with zero new
// client-side mapping.
func (e *GameEndpoints) sendError(conn *websocket.Conn, outbound chan<- []byte, requestID string, err error) {
	data, marshalErr := json.Marshal(dto.ServerFrame{
		Type:      dto.FrameError,
		RequestID: requestID,
		Payload:   dto.ErrorResponse{Error: err.Error(), Code: errorCode(err)},
	})
	if marshalErr != nil {
		return
	}
	send(conn, outbound, data)
}

// send enqueues data for the write pump without blocking. If outbound is
// already full, the client is genuinely too slow to keep up - closing with
// StatusPolicyViolation and letting it reconnect for a fresh snapshot is
// safer than serving a client whose view has silently fallen behind.
func send(conn *websocket.Conn, outbound chan<- []byte, data []byte) {
	select {
	case outbound <- data:
	default:
		_ = conn.Close(websocket.StatusPolicyViolation, "too slow")
	}
}
