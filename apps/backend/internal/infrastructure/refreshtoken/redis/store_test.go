package redis_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	refreshredis "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/refreshtoken/redis"
)

// newTestStore connects to TEST_REDIS_URL, skipping the test entirely when
// it is unset - same convention as streamticket/redis's tests.
func newTestStore(t *testing.T, ttl time.Duration) *refreshredis.Store {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set, skipping redis-backed refresh token store test")
	}
	s, err := refreshredis.New(context.Background(), refreshredis.Config{
		URL: url, DialTimeout: 2 * time.Second, OpTimeout: 2 * time.Second, TTL: ttl,
	})
	if err != nil {
		t.Fatalf("connecting to redis: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func testToken(seed byte, familyID string) ports.RefreshToken {
	return ports.RefreshToken{
		UserID:   user.UserID{seed, 1},
		Role:     enums.Regular,
		FamilyID: familyID,
	}
}

func TestStore_IssueThenRedeem_ReturnsExactToken(t *testing.T) {
	s := newTestStore(t, time.Minute)
	ctx := context.Background()
	want := testToken(1, "")

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
	if got.UserID != want.UserID || got.Role != want.Role {
		t.Fatalf("Redeem = %+v, want UserID/Role %+v", got, want)
	}
	if got.FamilyID == "" {
		t.Fatal("Redeem returned an empty FamilyID for a token issued with FamilyID == \"\"")
	}
}

func TestStore_Redeem_UnknownFails(t *testing.T) {
	s := newTestStore(t, time.Minute)
	ctx := context.Background()

	if _, err := s.Redeem(ctx, "does-not-exist"); !errors.Is(err, ports.ErrRefreshInvalid) {
		t.Fatalf("Redeem error = %v, want ErrRefreshInvalid", err)
	}
}

func TestStore_Redeem_ReuseRevokesFamily(t *testing.T) {
	s := newTestStore(t, time.Minute)
	ctx := context.Background()

	first := testToken(1, "")
	token1, _, err := s.Issue(ctx, first)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	redeemed, err := s.Redeem(ctx, token1)
	if err != nil {
		t.Fatalf("first Redeem: %v", err)
	}
	familyID := redeemed.FamilyID

	// Rotate: mint a second token in the same family.
	token2, _, err := s.Issue(ctx, testToken(1, familyID))
	if err != nil {
		t.Fatalf("Issue (rotation): %v", err)
	}

	// Replaying the already-used first token must report reuse and revoke
	// the whole family.
	if _, err := s.Redeem(ctx, token1); !errors.Is(err, ports.ErrRefreshReuse) {
		t.Fatalf("second Redeem of token1 error = %v, want ErrRefreshReuse", err)
	}

	// The second, never-before-redeemed token from the same family must now
	// also fail, since the family was revoked as a side effect of the
	// replay above.
	if _, err := s.Redeem(ctx, token2); !errors.Is(err, ports.ErrRefreshInvalid) {
		t.Fatalf("Redeem of token2 after family revocation error = %v, want ErrRefreshInvalid", err)
	}
}

func TestStore_RevokeFamily_ThenRedeemFails(t *testing.T) {
	s := newTestStore(t, time.Minute)
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testToken(1, ""))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	redeemed, err := s.Redeem(ctx, token)
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}

	token2, _, err := s.Issue(ctx, testToken(1, redeemed.FamilyID))
	if err != nil {
		t.Fatalf("Issue (rotation): %v", err)
	}

	if err := s.RevokeFamily(ctx, redeemed.FamilyID); err != nil {
		t.Fatalf("RevokeFamily: %v", err)
	}

	if _, err := s.Redeem(ctx, token2); !errors.Is(err, ports.ErrRefreshInvalid) {
		t.Fatalf("Redeem after RevokeFamily error = %v, want ErrRefreshInvalid", err)
	}
}

func TestStore_Redeem_ExpiredFails(t *testing.T) {
	s := newTestStore(t, 100*time.Millisecond)
	ctx := context.Background()

	token, _, err := s.Issue(ctx, testToken(1, ""))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	if _, err := s.Redeem(ctx, token); !errors.Is(err, ports.ErrRefreshInvalid) {
		t.Fatalf("Redeem error = %v, want ErrRefreshInvalid", err)
	}
}
