package services

import (
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestPictureEventHub_PublishDeliversToSubscriber(t *testing.T) {
	hub := NewPictureEventHub()
	events, unsubscribe := hub.Subscribe()
	defer unsubscribe()

	want := PictureEvent{Kind: enums.StandSubject, SubjectID: "abc", Status: enums.PictureReady}
	hub.Publish(want)

	select {
	case got := <-events:
		if got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestPictureEventHub_PublishFansOutToEverySubscriber(t *testing.T) {
	hub := NewPictureEventHub()
	events1, unsubscribe1 := hub.Subscribe()
	defer unsubscribe1()
	events2, unsubscribe2 := hub.Subscribe()
	defer unsubscribe2()

	want := PictureEvent{Kind: enums.DevilFruitSubject, SubjectID: "xyz", Status: enums.PictureFailed}
	hub.Publish(want)

	for _, ch := range []<-chan PictureEvent{events1, events2} {
		select {
		case got := <-ch:
			if got != want {
				t.Fatalf("got %+v, want %+v", got, want)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for event")
		}
	}
}

func TestPictureEventHub_PublishAfterUnsubscribeIsANoop(t *testing.T) {
	hub := NewPictureEventHub()
	events, unsubscribe := hub.Subscribe()
	unsubscribe()

	// Must not panic (sending on/closing a channel already removed).
	hub.Publish(PictureEvent{Kind: enums.UserSubject, SubjectID: "u1", Status: enums.PictureReady})

	if _, ok := <-events; ok {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

func TestPictureEventHub_PublishDropsWhenSubscriberBufferFull(t *testing.T) {
	hub := NewPictureEventHub()
	events, unsubscribe := hub.Subscribe()
	defer unsubscribe()

	// Fill the buffer without ever draining it - Publish must not block.
	for i := 0; i < subscriberBuffer+5; i++ {
		hub.Publish(PictureEvent{Kind: enums.StandSubject, SubjectID: "flood", Status: enums.PictureReady})
	}

	if len(events) != subscriberBuffer {
		t.Fatalf("buffered events = %d, want %d (excess should have been dropped)", len(events), subscriberBuffer)
	}
}

func TestPictureEventHub_UnsubscribeIsIdempotent(t *testing.T) {
	hub := NewPictureEventHub()
	_, unsubscribe := hub.Subscribe()

	unsubscribe()
	unsubscribe() // must not panic (double-close)
}
