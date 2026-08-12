package services_test

import (
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

func TestGameEventHub_SubscribeReceivesPublishedEvent(t *testing.T) {
	hub := services.NewGameEventHub()
	var id game.GameID
	id[0] = 1

	ch, unsub := hub.Subscribe(id)
	defer unsub()

	hub.Publish(services.GameEvent{GameID: id, Name: "GAME_STARTED"})

	select {
	case evt := <-ch:
		if evt.Name != "GAME_STARTED" {
			t.Fatalf("evt.Name = %q, want GAME_STARTED", evt.Name)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for published event")
	}
}

func TestGameEventHub_NoCrossTalkBetweenGames(t *testing.T) {
	hub := services.NewGameEventHub()
	var idA, idB game.GameID
	idA[0] = 1
	idB[0] = 2

	chA, unsubA := hub.Subscribe(idA)
	defer unsubA()
	chB, unsubB := hub.Subscribe(idB)
	defer unsubB()

	hub.Publish(services.GameEvent{GameID: idA, Name: "GAME_STARTED"})

	select {
	case <-chA:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for game A's event")
	}

	select {
	case evt := <-chB:
		t.Fatalf("game B received an event meant for game A: %+v", evt)
	case <-time.After(50 * time.Millisecond):
		// expected: nothing arrives for game B
	}
}

func TestGameEventHub_UnsubscribeStopsDelivery(t *testing.T) {
	hub := services.NewGameEventHub()
	var id game.GameID
	id[0] = 1

	ch, unsub := hub.Subscribe(id)
	unsub()

	hub.Publish(services.GameEvent{GameID: id, Name: "GAME_STARTED"})

	if _, ok := <-ch; ok {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

func TestGameEventHub_SlowSubscriberDropsWithoutBlocking(t *testing.T) {
	hub := services.NewGameEventHub()
	var id game.GameID
	id[0] = 1

	_, unsub := hub.Subscribe(id) // never drained
	defer unsub()

	done := make(chan struct{})
	go func() {
		// Publish far more events than the subscriber buffer can hold -
		// this must never block regardless.
		for i := 0; i < 100; i++ {
			hub.Publish(services.GameEvent{GameID: id, Name: "VOTE_CAST"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on a full, undrained subscriber")
	}
}
