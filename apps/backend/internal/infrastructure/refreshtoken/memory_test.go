package refreshtoken

import (
	"context"
	"encoding/base64"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

func testToken(seed byte, familyID string) ports.RefreshToken {
	return ports.RefreshToken{
		UserID:   user.UserID{seed, 1},
		Role:     enums.Regular,
		FamilyID: familyID,
	}
}

func TestMemoryStore_IssueThenRedeem_ReturnsExactToken(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()
	want := testToken(1, "family-1")

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

	got, err := s.Redeem(ctx, token)
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}
	if got != want {
		t.Fatalf("Redeem = %+v, want %+v", got, want)
	}
}

func TestMemoryStore_Redeem_UnknownFails(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()

	for _, token := range []string{"", "does-not-exist"} {
		if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrRefreshInvalid) {
			t.Fatalf("Redeem(%q) error = %v, want ErrRefreshInvalid", token, err)
		}
	}
}

func TestMemoryStore_Redeem_ExpiredFails(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testToken(1, "family-1"))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(2 * time.Minute)
	if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrRefreshInvalid) {
		t.Fatalf("Redeem error = %v, want ErrRefreshInvalid", err)
	}
}

func TestMemoryStore_Redeem_TwiceIsReuseAndRevokesFamily(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()

	familyID := "family-1"
	firstToken, _, err := s.Issue(ctx, testToken(1, familyID))
	if err != nil {
		t.Fatalf("Issue first: %v", err)
	}
	secondToken, _, err := s.Issue(ctx, testToken(2, familyID))
	if err != nil {
		t.Fatalf("Issue second: %v", err)
	}

	if _, err := s.Redeem(ctx, firstToken); err != nil {
		t.Fatalf("first Redeem: %v", err)
	}
	if _, err := s.Redeem(ctx, firstToken); !errors.Is(err, ports.ErrRefreshReuse) {
		t.Fatalf("second Redeem error = %v, want ErrRefreshReuse", err)
	}

	// The reuse must have revoked the whole family - a still-unredeemed,
	// still-unexpired token from the same family now fails too.
	if _, err := s.Redeem(ctx, secondToken); !errors.Is(err, ports.ErrRefreshInvalid) {
		t.Fatalf("Redeem(secondToken) after reuse = %v, want ErrRefreshInvalid", err)
	}
}

func TestMemoryStore_RevokeFamily_ThenRedeemFails(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()

	familyID := "family-1"
	token, _, err := s.Issue(ctx, testToken(1, familyID))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if err := s.RevokeFamily(ctx, familyID); err != nil {
		t.Fatalf("RevokeFamily: %v", err)
	}

	if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrRefreshInvalid) {
		t.Fatalf("Redeem after RevokeFamily = %v, want ErrRefreshInvalid", err)
	}
}

// TestMemoryStore_Redeem_ConcurrentRedeemsExactlyOneSucceeds proves the
// single-use guarantee under concurrency: exactly one goroutine can flip
// used from false to true under the shared mutex. Run under -race.
func TestMemoryStore_Redeem_ConcurrentRedeemsExactlyOneSucceeds(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testToken(1, "family-1"))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	const n = 50
	var wg sync.WaitGroup
	var successes int32
	var mu sync.Mutex
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if _, err := s.Redeem(ctx, token); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("successes = %d, want exactly 1", successes)
	}
}

func TestMemoryStore_Issue_TokensAreDistinctAndWellFormed(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()

	const n = 10_000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		token, _, err := s.Issue(ctx, testToken(1, "family-1"))
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
	s := NewMemoryStore(Config{TTL: time.Minute})
	wantErr := errors.New("boom")
	s.randRead = func([]byte) (int, error) { return 0, wantErr }

	token, _, err := s.Issue(context.Background(), testToken(1, "family-1"))
	if err == nil {
		t.Fatal("Issue succeeded, want an error")
	}
	if token != "" {
		t.Fatalf("Issue returned token %q on error, want empty", token)
	}
	if len(s.tokens) != 0 {
		t.Fatalf("store has %d tokens after a failed Issue, want 0", len(s.tokens))
	}
}

func TestMemoryStore_DeleteExpired_RemovesOnlyExpired(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }
	ctx := context.Background()

	expiredToken, _, err := s.Issue(ctx, testToken(1, "family-1"))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(2 * time.Minute)
	liveToken, _, err := s.Issue(ctx, testToken(2, "family-2"))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if n := s.DeleteExpired(); n != 1 {
		t.Fatalf("DeleteExpired = %d, want 1", n)
	}

	if _, err := s.Redeem(ctx, expiredToken); !errors.Is(err, ports.ErrRefreshInvalid) {
		t.Fatalf("Redeem(expired) error = %v, want ErrRefreshInvalid", err)
	}
	if _, err := s.Redeem(ctx, liveToken); err != nil {
		t.Fatalf("Redeem(live) after reap: %v", err)
	}
}
