package services

import (
	"sync"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// PictureEvent notifies that a subject's picture pipeline reached a
// terminal state (READY or FAILED) - the SSE handler serializes this
// directly to subscribed clients.
type PictureEvent struct {
	Kind      enums.PictureSubjectKind
	SubjectID string
	Status    enums.PictureStatus
}

// PictureEventHub is an in-process, single-instance pub/sub for
// PictureEvents. There is exactly one backend process (no multi-instance
// fan-out need), so a plain in-memory hub is enough - no Redis pub/sub.
// A dropped event (slow/stuck subscriber) is only ever a missed nudge: the
// SSE client re-syncs its whole picture-status view via a normal query
// refetch on every reconnect, so nothing durable is lost.
type PictureEventHub struct {
	mu   sync.Mutex
	subs map[chan PictureEvent]struct{}
}

// subscriberBuffer bounds how many events a slow subscriber can lag behind
// before Publish starts dropping for it - small on purpose, since a client
// that's this far behind will resync on its next reconnect anyway.
const subscriberBuffer = 8

func NewPictureEventHub() *PictureEventHub {
	return &PictureEventHub{subs: make(map[chan PictureEvent]struct{})}
}

// Subscribe registers a new listener and returns its event channel plus an
// unsubscribe function the caller must call exactly once (typically via
// defer) to deregister and close the channel.
func (h *PictureEventHub) Subscribe() (<-chan PictureEvent, func()) {
	ch := make(chan PictureEvent, subscriberBuffer)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if _, ok := h.subs[ch]; ok {
			delete(h.subs, ch)
			close(ch)
		}
	}
	return ch, unsubscribe
}

// Publish fans evt out to every current subscriber without blocking - a
// subscriber whose buffer is full simply misses this event rather than
// stalling the picture worker.
func (h *PictureEventHub) Publish(evt PictureEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.subs {
		select {
		case ch <- evt:
		default:
		}
	}
}
