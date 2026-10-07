//go:build integration

package repositories_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

func newTestPush(t *testing.T) (*repositories.PushSubscriptionRepository, *pgxpool.Pool) {
	t.Helper()
	_, pool := newTestGameHistory(t)
	return repositories.NewPushSubscriptionRepository(pool), pool
}

func uniqueEndpoint() string {
	return "https://fcm.googleapis.com/fcm/send/" + uuid.NewString()
}

func TestPushSubscriptions_UpsertAndList(t *testing.T) {
	repo, pool := newTestPush(t)
	owner := insertTestUser(t, pool)
	ctx := context.Background()

	endpoint := uniqueEndpoint()
	if err := repo.Upsert(ctx, ports.PushSubscription{UserID: owner, Endpoint: endpoint, P256dh: "key-1", Auth: "auth-1"}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	// Same endpoint again refreshes the keys instead of duplicating.
	if err := repo.Upsert(ctx, ports.PushSubscription{UserID: owner, Endpoint: endpoint, P256dh: "key-2", Auth: "auth-2"}); err != nil {
		t.Fatalf("Upsert again: %v", err)
	}

	subs, err := repo.ListByUser(ctx, owner)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("len(subs) = %d, want 1", len(subs))
	}
	if subs[0].P256dh != "key-2" || subs[0].Auth != "auth-2" || subs[0].Endpoint != endpoint {
		t.Errorf("subscription = %+v, want refreshed keys on %s", subs[0], endpoint)
	}
}

func TestPushSubscriptions_EndpointMovesToNewOwner(t *testing.T) {
	repo, pool := newTestPush(t)
	first := insertTestUser(t, pool)
	second := insertTestUser(t, pool)
	ctx := context.Background()

	endpoint := uniqueEndpoint()
	if err := repo.Upsert(ctx, ports.PushSubscription{UserID: first, Endpoint: endpoint, P256dh: "k", Auth: "a"}); err != nil {
		t.Fatalf("Upsert first: %v", err)
	}
	// Another account logs in on the same device.
	if err := repo.Upsert(ctx, ports.PushSubscription{UserID: second, Endpoint: endpoint, P256dh: "k", Auth: "a"}); err != nil {
		t.Fatalf("Upsert second: %v", err)
	}

	if subs, _ := repo.ListByUser(ctx, first); len(subs) != 0 {
		t.Errorf("first owner still has %d subscriptions, want 0", len(subs))
	}
	if subs, _ := repo.ListByUser(ctx, second); len(subs) != 1 {
		t.Errorf("second owner has %d subscriptions, want 1", len(subs))
	}
}

func TestPushSubscriptions_TrimsToPerUserCap(t *testing.T) {
	repo, pool := newTestPush(t)
	owner := insertTestUser(t, pool)
	ctx := context.Background()

	const registered = 13
	for i := 0; i < registered; i++ {
		sub := ports.PushSubscription{UserID: owner, Endpoint: uniqueEndpoint(), P256dh: fmt.Sprintf("k%d", i), Auth: "a"}
		if err := repo.Upsert(ctx, sub); err != nil {
			t.Fatalf("Upsert %d: %v", i, err)
		}
	}

	subs, err := repo.ListByUser(ctx, owner)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(subs) != 10 {
		t.Errorf("len(subs) = %d, want the cap of 10", len(subs))
	}
}

func TestPushSubscriptions_DeleteIsScopedToTheOwner(t *testing.T) {
	repo, pool := newTestPush(t)
	owner := insertTestUser(t, pool)
	stranger := insertTestUser(t, pool)
	ctx := context.Background()

	endpoint := uniqueEndpoint()
	if err := repo.Upsert(ctx, ports.PushSubscription{UserID: owner, Endpoint: endpoint, P256dh: "k", Auth: "a"}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	// Someone else naming the endpoint must not remove it.
	if err := repo.DeleteForUser(ctx, stranger, endpoint); err != nil {
		t.Fatalf("DeleteForUser(stranger): %v", err)
	}
	if subs, _ := repo.ListByUser(ctx, owner); len(subs) != 1 {
		t.Fatalf("subscription removed by a non-owner, got %d", len(subs))
	}

	if err := repo.DeleteForUser(ctx, owner, endpoint); err != nil {
		t.Fatalf("DeleteForUser(owner): %v", err)
	}
	if subs, _ := repo.ListByUser(ctx, owner); len(subs) != 0 {
		t.Errorf("subscription survived its owner's delete, got %d", len(subs))
	}
}

func TestPushSubscriptions_DeleteByEndpoint(t *testing.T) {
	repo, pool := newTestPush(t)
	owner := insertTestUser(t, pool)
	ctx := context.Background()

	endpoint := uniqueEndpoint()
	if err := repo.Upsert(ctx, ports.PushSubscription{UserID: owner, Endpoint: endpoint, P256dh: "k", Auth: "a"}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if err := repo.DeleteByEndpoint(ctx, endpoint); err != nil {
		t.Fatalf("DeleteByEndpoint: %v", err)
	}
	if subs, _ := repo.ListByUser(ctx, owner); len(subs) != 0 {
		t.Errorf("got %d subscriptions after DeleteByEndpoint, want 0", len(subs))
	}
	// Deleting an unknown endpoint is not an error.
	if err := repo.DeleteByEndpoint(ctx, endpoint); err != nil {
		t.Errorf("DeleteByEndpoint(unknown): %v", err)
	}
}
