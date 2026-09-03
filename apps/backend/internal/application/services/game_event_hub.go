package services

import (
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

// GameEvent wraps one of game.Game's DomainEvents with the GameID it came
// from, so a single hub can multiplex every live Game's events to whichever
// transport (websockets, later) ends up subscribing.
type GameEvent struct {
	GameID game.GameID
	// Name mirrors Event.Name(), spelled out here so the transport layer can
	// filter/log without a type-switch over every concrete event.
	Name  string
	Event game.DomainEvent
	// VotingWindow is the lobby's own configured voting-window duration at
	// the moment this event was published - the transport uses it to
	// compute VOTING_OPENED/TIEBREAK_OPENED's closesAt instead of the
	// service-wide default, since each lobby may now configure its own.
	VotingWindow time.Duration
	// RevealWindow is revealDurationFor(g) at the moment this event was
	// published - the third member of the VotingWindow/SummaryWindow
	// trio. It is never sent over the wire; it exists purely as the
	// fallback window the transport synthesizes LOADOUTS_ASSIGNED's
	// closesAt from if ClosesAt ever arrives zero (see
	// endpoints.frameDeadline). The authoritative value clients actually
	// pace against is ClosesAt below.
	RevealWindow time.Duration
	// SummaryWindow is the lobby's own configured summary-screen duration
	// (Config.SummaryDurationSeconds) at the moment this event was
	// published - the transport uses it to compute SUMMARY_OPENED's
	// closesAt, same pattern as VotingWindow.
	SummaryWindow time.Duration
	// ClosesAt is the authoritative wall-clock deadline of the timed phase
	// this event opened, read straight out of GameService's own
	// votingEnds/revealEnds/resultEnds/summaryEnds maps at publish time
	// (see GameService.publish). All five timed frames carry one -
	// VOTING_OPENED, TIEBREAK_OPENED, SUMMARY_OPENED, LOADOUTS_ASSIGNED,
	// ROUND_RESOLVED; for every other event - and for a timed event
	// published from a path that somehow has no recorded deadline - the
	// zero value means "not a timed frame", and the transport falls back
	// to synthesizing one from the window duration. Stamping it here
	// rather than in the transport is what keeps the client's countdown
	// from drifting by however long hub delivery took.
	ClosesAt time.Time
}

// GameEventHub is an in-process, single-instance pub/sub for GameEvents,
// scoped per GameID - mirrors PictureEventHub's role and non-blocking
// drop-on-full-subscriber behavior, but keyed so a slow subscriber on one
// Game can never affect another Game's subscribers.
type GameEventHub struct {
	mu   sync.Mutex
	subs map[game.GameID]map[chan GameEvent]struct{}
}

// gameSubscriberBuffer bounds how many events a slow subscriber can lag
// behind before Publish starts dropping for it, mirroring
// PictureEventHub's subscriberBuffer.
const gameSubscriberBuffer = 8

func NewGameEventHub() *GameEventHub {
	return &GameEventHub{subs: make(map[game.GameID]map[chan GameEvent]struct{})}
}

// Subscribe registers a new listener for id's events and returns its
// channel plus an unsubscribe function the caller must call exactly once
// (typically via defer) to deregister and close the channel.
func (h *GameEventHub) Subscribe(id game.GameID) (<-chan GameEvent, func()) {
	ch := make(chan GameEvent, gameSubscriberBuffer)

	h.mu.Lock()
	if h.subs[id] == nil {
		h.subs[id] = make(map[chan GameEvent]struct{})
	}
	h.subs[id][ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if subs, ok := h.subs[id]; ok {
			if _, ok := subs[ch]; ok {
				delete(subs, ch)
				close(ch)
			}
			if len(subs) == 0 {
				delete(h.subs, id)
			}
		}
	}
	return ch, unsubscribe
}

// Publish fans evt out to every current subscriber of evt.GameID without
// blocking - a subscriber whose buffer is full simply misses this event.
func (h *GameEventHub) Publish(evt GameEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.subs[evt.GameID] {
		select {
		case ch <- evt:
		default:
		}
	}
}
