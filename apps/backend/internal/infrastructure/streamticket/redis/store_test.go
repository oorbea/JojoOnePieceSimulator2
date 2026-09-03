package redis_test

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	streamredis "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket/redis"
)

// newTestStore connects to TEST_REDIS_URL, skipping the test entirely when
// it is unset - same convention as gamestore/redis's tests.
func newTestStore(t *testing.T, ttl time.Duration) *streamredis.Store {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set, skipping redis-backed stream ticket store test")
	}
	s, err := streamredis.New(context.Background(), streamredis.Config{
		URL: url, DialTimeout: 2 * time.Second, OpTimeout: 2 * time.Second, TTL: ttl,
	})
	if err != nil {
		t.Fatalf("connecting to redis: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func testTicket(seed byte) ports.StreamTicket {
	return ports.StreamTicket{
		UserID:   user.UserID{seed, 1},
		Role:     enums.Regular,
		Purpose:  ports.TicketPurposeEvents,
		Resource: "",
	}
}

func TestStore_IssueThenRedeem_ReturnsExactTicket(t *testing.T) {
	s := newTestStore(t, time.Minute)
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

func TestStore_Redeem_SecondTimeFails(t *testing.T) {
	s := newTestStore(t, time.Minute)
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

func TestStore_Redeem_ExpiredFails(t *testing.T) {
	s := newTestStore(t, 100*time.Millisecond)
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testTicket(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrTicketInvalid) {
		t.Fatalf("Redeem error = %v, want ErrTicketInvalid", err)
	}
}

func TestStore_Redeem_UnknownOrEmptyFails(t *testing.T) {
	s := newTestStore(t, time.Minute)
	ctx := context.Background()

	for _, token := range []string{"", "does-not-exist"} {
		if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrTicketInvalid) {
			t.Fatalf("Redeem(%q) error = %v, want ErrTicketInvalid", token, err)
		}
	}
}

// TestStore_Redeem_ConcurrentRedeemsExactlyOneSucceeds proves the Lua
// GET+DEL script is the atomic burn it's meant to be.
func TestStore_Redeem_ConcurrentRedeemsExactlyOneSucceeds(t *testing.T) {
	s := newTestStore(t, time.Minute)
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testTicket(1))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	const n = 25
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
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

func TestStore_Issue_TokensAreWellFormed(t *testing.T) {
	s := newTestStore(t, time.Minute)
	ctx := context.Background()

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
}
