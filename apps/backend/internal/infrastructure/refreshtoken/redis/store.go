// Package redis provides a Redis-backed ports.IRefreshTokenStore, for
// multi-instance deployments where the process minting/rotating a refresh
// token (login, or a previous refresh) and the process redeeming it (the
// next refresh request) may not be the same instance. Mirrors
// streamticket/redis's Config/New/opContext shape.
package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// refreshTokenBytes mirrors streamticket/redis's entropy budget - 32 bytes
// (256 bits), base64url-encoded to 43 characters. Used both for the token
// itself and, when a new chain starts, for the family id.
const refreshTokenBytes = 32

// Config holds everything needed to reach the Redis instance backing the
// refresh token store.
type Config struct {
	// URL is a redis:// or rediss:// connection string.
	URL string
	// DialTimeout bounds the initial connection + PING done by New.
	DialTimeout time.Duration
	// OpTimeout bounds every individual store operation. Fail-closed, like
	// streamticket/redis - a refresh token store is a source of truth for
	// whether a session may continue, not an optional speedup.
	OpTimeout time.Duration
	// TTL is how long a minted token (and the family marker refreshed
	// alongside it) stays redeemable.
	TTL time.Duration
}

// Store is a Redis-backed ports.IRefreshTokenStore. Fail-closed: every
// error is returned to the caller, never swallowed as a miss.
//
//	jojo:refresh:<token>      -> HASH{user_id, role, family_id, used}
//	jojo:refresh-fam:<famID>  -> marker key, present = family alive
//
// Shares the "jojo:" root with gamestore/redis and streamticket/redis but
// not infrastructure/cache's <ns>:<gen>: layout, for the same reason
// gamestore doesn't: a cache generation bump must not be able to orphan a
// live refresh token.
type Store struct {
	client    *goredis.Client
	opTimeout time.Duration
	ttl       time.Duration
}

var _ ports.IRefreshTokenStore = (*Store)(nil)

func tokenKey(token string) string     { return "jojo:refresh:" + token }
func familyKey(familyID string) string { return "jojo:refresh-fam:" + familyID }

// redeemScript atomically checks-and-burns a refresh token, so two
// concurrent Redeems of the same token can never both succeed, and a
// replayed (already-used) token revokes its whole family in the same round
// trip.
//
// KEYS[1] = token hash key
// KEYS[2] = family marker key
var redeemScript = goredis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
	return {err = 'invalid'}
end
local used = redis.call('HGET', KEYS[1], 'used')
if used == '1' then
	redis.call('DEL', KEYS[2])
	return {err = 'reuse'}
end
if redis.call('EXISTS', KEYS[2]) == 0 then
	return {err = 'invalid'}
end
redis.call('HSET', KEYS[1], 'used', '1')
local user_id = redis.call('HGET', KEYS[1], 'user_id')
local role = redis.call('HGET', KEYS[1], 'role')
local family_id = redis.call('HGET', KEYS[1], 'family_id')
return {user_id, role, family_id}
`)

// New connects to the Redis instance described by cfg and verifies
// reachability with a PING bounded by cfg.DialTimeout, so a misconfigured
// deploy fails at boot instead of silently minting tokens nobody can ever
// redeem.
func New(ctx context.Context, cfg Config) (*Store, error) {
	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}
	client := goredis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return &Store{client: client, opTimeout: cfg.OpTimeout, ttl: cfg.TTL}, nil
}

func (s *Store) opContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, s.opTimeout)
}

func randomToken() (string, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// Issue implements ports.IRefreshTokenStore.
func (s *Store) Issue(ctx context.Context, t ports.RefreshToken) (string, time.Time, error) {
	familyID := t.FamilyID
	if familyID == "" {
		fid, err := randomToken()
		if err != nil {
			return "", time.Time{}, fmt.Errorf("generating refresh token family id: %w", err)
		}
		familyID = fid
	}

	token, err := randomToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generating refresh token: %w", err)
	}

	expiresAt := time.Now().Add(s.ttl)

	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	pipe := s.client.TxPipeline()
	pipe.HSet(opCtx, tokenKey(token), map[string]interface{}{
		"user_id":   t.UserID.String(),
		"role":      t.Role.String(),
		"family_id": familyID,
		"used":      "0",
	})
	pipe.Expire(opCtx, tokenKey(token), s.ttl)
	// Refresh the family marker's TTL on every issue/rotation, so a family
	// that keeps refreshing never expires out from under an active session
	// while an idle one still dies on schedule.
	pipe.Set(opCtx, familyKey(familyID), "1", s.ttl)
	if _, err := pipe.Exec(opCtx); err != nil {
		return "", time.Time{}, fmt.Errorf("storing refresh token: %w", err)
	}

	return token, expiresAt, nil
}

// Redeem implements ports.IRefreshTokenStore.
//
// The family key the atomic script needs to check/revoke isn't known until
// the token's family_id is read out of its hash, so Redeem does a plain,
// non-mutating HGET first to resolve it. That read carries no race risk:
// it never changes state, and the script re-validates everything (EXISTS,
// used, family EXISTS) atomically before committing to a decision - a
// token that gets redeemed by a concurrent caller between the HGET and the
// EVAL is simply caught by the script's own checks.
func (s *Store) Redeem(ctx context.Context, token string) (ports.RefreshToken, error) {
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	tKey := tokenKey(token)
	fid, err := s.client.HGet(opCtx, tKey, "family_id").Result()
	if err != nil && !errors.Is(err, goredis.Nil) {
		return ports.RefreshToken{}, fmt.Errorf("looking up refresh token family: %w", err)
	}

	res, err := redeemScript.Run(opCtx, s.client, []string{tKey, familyKey(fid)}).Result()
	if err != nil {
		switch err.Error() {
		case "invalid":
			return ports.RefreshToken{}, ports.ErrRefreshInvalid
		case "reuse":
			return ports.RefreshToken{}, ports.ErrRefreshReuse
		}
		return ports.RefreshToken{}, fmt.Errorf("redeeming refresh token: %w", err)
	}

	vals, ok := res.([]interface{})
	if !ok || len(vals) != 3 {
		return ports.RefreshToken{}, fmt.Errorf("redeeming refresh token: unexpected script result %v", res)
	}
	userIDStr, ok1 := vals[0].(string)
	roleStr, ok2 := vals[1].(string)
	familyIDStr, ok3 := vals[2].(string)
	if !ok1 || !ok2 || !ok3 {
		return ports.RefreshToken{}, fmt.Errorf("redeeming refresh token: malformed script result %v", res)
	}

	userID, err := user.ParseUserID(userIDStr)
	if err != nil {
		return ports.RefreshToken{}, fmt.Errorf("parsing refresh token user id: %w", err)
	}
	role, err := enums.ParseUserRole(roleStr)
	if err != nil {
		return ports.RefreshToken{}, fmt.Errorf("parsing refresh token role: %w", err)
	}

	return ports.RefreshToken{UserID: userID, Role: role, FamilyID: familyIDStr}, nil
}

// RevokeFamily implements ports.IRefreshTokenStore.
func (s *Store) RevokeFamily(ctx context.Context, familyID string) error {
	opCtx, cancel := s.opContext(ctx)
	defer cancel()
	if err := s.client.Del(opCtx, familyKey(familyID)).Err(); err != nil {
		return fmt.Errorf("revoking refresh token family: %w", err)
	}
	return nil
}

// Close releases the underlying Redis client.
func (s *Store) Close() error {
	return s.client.Close()
}
