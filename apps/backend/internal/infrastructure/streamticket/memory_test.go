package streamticket

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

func testTicket(seed byte) ports.StreamTicket {
	return ports.StreamTicket{
		UserID:   user.UserID{seed, 1},
		Role:     enums.Regular,
		Purpose:  ports.TicketPurposeEvents,
		Resource: "",
	}
}

func TestMemoryStore_IssueThenRedeem_ReturnsExactTicket(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()
	want := testTicket(1)

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

func TestMemoryStore_Redeem_SecondTimeFails(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testTicket(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := s.Redeem(ctx, token); err != nil {
		t.Fatalf("first Redeem: %v", err)
	}
	if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrTicketInvalid) {
		t.Fatalf("second Redeem error = %v, want ErrTicketInvalid", err)
	}
}

func TestMemoryStore_Redeem_ExpiredFails(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testTicket(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(2 * time.Minute)
	if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrTicketInvalid) {
		t.Fatalf("Redeem error = %v, want ErrTicketInvalid", err)
	}
}

func TestMemoryStore_Redeem_UnknownOrEmptyFails(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()

	for _, token := range []string{"", "does-not-exist"} {
		if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrTicketInvalid) {
			t.Fatalf("Redeem(%q) error = %v, want ErrTicketInvalid", token, err)
		}
	}
}

// TestMemoryStore_Redeem_ConcurrentRedeemsExactlyOneSucceeds proves the
// single-use guarantee under concurrency: whichever goroutine's delete
// actually removes the map entry is the only one that can succeed. Run
// under -race.
func TestMemoryStore_Redeem_ConcurrentRedeemsExactlyOneSucceeds(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testTicket(1))
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
		token, _, err := s.Issue(ctx, testTicket(1))
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

	token, _, err := s.Issue(context.Background(), testTicket(1))
	if err == nil {
		t.Fatal("Issue succeeded, want an error")
	}
	if token != "" {
		t.Fatalf("Issue returned token %q on error, want empty", token)
	}
	if len(s.tickets) != 0 {
		t.Fatalf("store has %d tickets after a failed Issue, want 0", len(s.tickets))
	}
}

func TestMemoryStore_DeleteExpired_RemovesOnlyExpired(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }
	ctx := context.Background()

	expiredToken, _, err := s.Issue(ctx, testTicket(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(2 * time.Minute)
	liveToken, _, err := s.Issue(ctx, testTicket(2))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if n := s.DeleteExpired(); n != 1 {
		t.Fatalf("DeleteExpired = %d, want 1", n)
	}

	if _, err := s.Redeem(ctx, expiredToken); !errors.Is(err, ports.ErrTicketInvalid) {
		t.Fatalf("Redeem(expired) error = %v, want ErrTicketInvalid", err)
	}
	if _, err := s.Redeem(ctx, liveToken); err != nil {
		t.Fatalf("Redeem(live) after reap: %v", err)
	}
}
