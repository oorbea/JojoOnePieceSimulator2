package redis_test

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	inviteredis "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/gameinvite/redis"
)

// newTestStore connects to TEST_REDIS_URL, skipping the test entirely when
// it is unset - same convention as streamticket/redis's tests.
func newTestStore(t *testing.T, ttl time.Duration) *inviteredis.Store {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set, skipping redis-backed game invite store test")
	}
	s, err := inviteredis.New(context.Background(), inviteredis.Config{
		URL: url, DialTimeout: 2 * time.Second, OpTimeout: 2 * time.Second, TTL: ttl,
	})
	if err != nil {
		t.Fatalf("connecting to redis: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func testInvite(seed byte) ports.GameInvite {
	return ports.GameInvite{
		GameID:    game.GameID{seed, 1},
		Code:      "ABC123",
		CreatedBy: user.UserID{seed, 2},
	}
}

func TestStore_IssueThenLookup_ReturnsExactInvite(t *testing.T) {
	s := newTestStore(t, 15*time.Minute)
	ctx := context.Background()
	want := testInvite(1)

	token, expiresAt, err := s.Issue(ctx, want)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if token == "" {
		t.Fatal("Issue returned an empty token")
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("expiresAt %v is not in the future", expiresAt)
	}

	got, err := s.Lookup(ctx, token)
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got != want {
		t.Fatalf("Lookup = %+v, want %+v", got, want)
	}
}

// TestStore_Lookup_RepeatedLookupStillSucceeds proves Lookup is a plain
// read, never a burn - the multi-use property the port requires.
func TestStore_Lookup_RepeatedLookupStillSucceeds(t *testing.T) {
	s := newTestStore(t, 15*time.Minute)
	ctx := context.Background()
	want := testInvite(1)

	token, _, err := s.Issue(ctx, want)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	for i := 0; i < 5; i++ {
		got, err := s.Lookup(ctx, token)
		if err != nil {
			t.Fatalf("Lookup #%d: %v", i, err)
		}
		if got != want {
			t.Fatalf("Lookup #%d = %+v, want %+v", i, got, want)
		}
	}
}

func TestStore_Lookup_ExpiredFails(t *testing.T) {
	s := newTestStore(t, 100*time.Millisecond)
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testInvite(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	if _, err := s.Lookup(ctx, token); !errors.Is(err, ports.ErrInviteInvalid) {
		t.Fatalf("Lookup error = %v, want ErrInviteInvalid", err)
	}
}

func TestStore_Lookup_UnknownOrEmptyFails(t *testing.T) {
	s := newTestStore(t, 15*time.Minute)
	ctx := context.Background()

	for _, token := range []string{"", "does-not-exist"} {
		if _, err := s.Lookup(ctx, token); !errors.Is(err, ports.ErrInviteInvalid) {
			t.Fatalf("Lookup(%q) error = %v, want ErrInviteInvalid", token, err)
		}
	}
}

func TestStore_Issue_TokensAreWellFormed(t *testing.T) {
	s := newTestStore(t, 15*time.Minute)
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testInvite(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if len(token) != 43 {
		t.Fatalf("token %q has length %d, want 43", token, len(token))
	}
	if _, err := base64.RawURLEncoding.DecodeString(token); err != nil {
		t.Fatalf("token %q is not valid base64url: %v", token, err)
	}
}
