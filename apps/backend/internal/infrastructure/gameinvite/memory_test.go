package gameinvite

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

func testInvite(seed byte) ports.GameInvite {
	return ports.GameInvite{
		GameID:    game.GameID{seed, 1},
		Code:      "ABC123",
		CreatedBy: user.UserID{seed, 2},
	}
}

func TestMemoryStore_IssueThenLookup_ReturnsExactInvite(t *testing.T) {
	s := NewMemoryStore(Config{TTL: 15 * time.Minute})
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

// TestMemoryStore_Lookup_RepeatedLookupStillSucceeds is the headline
// assertion for this package: unlike streamticket.MemoryStore.Redeem, an
// invite is multi-use, so looking it up twice must not consume it.
func TestMemoryStore_Lookup_RepeatedLookupStillSucceeds(t *testing.T) {
	s := NewMemoryStore(Config{TTL: 15 * time.Minute})
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

func TestMemoryStore_Lookup_ExpiredFails(t *testing.T) {
	s := NewMemoryStore(Config{TTL: 15 * time.Minute})
	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testInvite(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(16 * time.Minute)
	if _, err := s.Lookup(ctx, token); !errors.Is(err, ports.ErrInviteInvalid) {
		t.Fatalf("Lookup error = %v, want ErrInviteInvalid", err)
	}
}

func TestMemoryStore_Lookup_UnknownOrEmptyFails(t *testing.T) {
	s := NewMemoryStore(Config{TTL: 15 * time.Minute})
	ctx := context.Background()

	for _, token := range []string{"", "does-not-exist"} {
		if _, err := s.Lookup(ctx, token); !errors.Is(err, ports.ErrInviteInvalid) {
			t.Fatalf("Lookup(%q) error = %v, want ErrInviteInvalid", token, err)
		}
	}
}

func TestMemoryStore_Issue_TokensAreDistinctAndWellFormed(t *testing.T) {
	s := NewMemoryStore(Config{TTL: 15 * time.Minute})
	ctx := context.Background()

	const n = 10_000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
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
		if _, dup := seen[token]; dup {
			t.Fatalf("token %q issued twice", token)
		}
		seen[token] = struct{}{}
	}
}

func TestMemoryStore_Issue_RandReadFailure(t *testing.T) {
	s := NewMemoryStore(Config{TTL: 15 * time.Minute})
	wantErr := errors.New("boom")
	s.randRead = func([]byte) (int, error) { return 0, wantErr }

	token, _, err := s.Issue(context.Background(), testInvite(1))
	if err == nil {
		t.Fatal("Issue succeeded, want an error")
	}
	if token != "" {
		t.Fatalf("Issue returned token %q on error, want empty", token)
	}
	if len(s.invites) != 0 {
		t.Fatalf("store has %d invites after a failed Issue, want 0", len(s.invites))
	}
}

func TestMemoryStore_DeleteExpired_RemovesOnlyExpired(t *testing.T) {
	s := NewMemoryStore(Config{TTL: 15 * time.Minute})
	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }
	ctx := context.Background()

	expiredToken, _, err := s.Issue(ctx, testInvite(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(16 * time.Minute)
	liveToken, _, err := s.Issue(ctx, testInvite(2))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if n := s.DeleteExpired(); n != 1 {
		t.Fatalf("DeleteExpired = %d, want 1", n)
	}

	if _, err := s.Lookup(ctx, expiredToken); !errors.Is(err, ports.ErrInviteInvalid) {
		t.Fatalf("Lookup(expired) error = %v, want ErrInviteInvalid", err)
	}
	if _, err := s.Lookup(ctx, liveToken); err != nil {
		t.Fatalf("Lookup(live) after reap: %v", err)
	}
}
