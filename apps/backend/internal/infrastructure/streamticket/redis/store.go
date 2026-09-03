// Package redis provides a Redis-backed ports.IStreamTicketStore, for
// multi-instance deployments where the process minting a ticket (behind a
// normal authenticated POST) and the process redeeming it (the stream
// handler) may not be the same instance. Mirrors gamestore/redis's Config/
// New/opContext shape.
package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// ticketTokenBytes mirrors streamticket.MemoryStore's entropy budget - 32
// bytes (256 bits), base64url-encoded to 43 characters.
const ticketTokenBytes = 32

// Config holds everything needed to reach the Redis instance backing the
// stream ticket store.
type Config struct {
	// URL is a redis:// or rediss:// connection string.
	URL string
	// DialTimeout bounds the initial connection + PING done by New.
	DialTimeout time.Duration
	// OpTimeout bounds every individual store operation. Fail-closed, like
	// gamestore/redis - a ticket store is a source of truth for whether a
	// connection may proceed, not an optional speedup.
	OpTimeout time.Duration
	// TTL is how long a minted ticket stays redeemable.
	TTL time.Duration
}

// Store is a Redis-backed ports.IStreamTicketStore. Fail-closed: every
// error is returned to the caller, never swallowed as a miss.
//
//	jojo:stream-ticket:<token> -> the envelope JSON (StreamTicket)
//
// Shares the "jojo:" root with gamestore/redis but not infrastructure/
// cache's <ns>:<gen>: layout, for the same reason gamestore doesn't: a
// cache generation bump must not be able to orphan a live ticket.
type Store struct {
	client    *goredis.Client
	opTimeout time.Duration
	ttl       time.Duration
}

var _ ports.IStreamTicketStore = (*Store)(nil)

func ticketKey(token string) string { return "jojo:stream-ticket:" + token }

// redeemScript atomically fetches and deletes the ticket in one round trip,
// so two concurrent Redeems of the same token can never both see a value -
// the burn. Not GETDEL: that needs Redis >= 6.2, and every other adapter in
// this codebase is already Lua-script-shaped, so EVAL keeps the pattern
// consistent and version-proof.
//
// KEYS[1] = ticket key
var redeemScript = goredis.NewScript(`
local v = redis.call('GET', KEYS[1])
if v then
	redis.call('DEL', KEYS[1])
end
return v
`)

// New connects to the Redis instance described by cfg and verifies
// reachability with a PING bounded by cfg.DialTimeout, so a misconfigured
// deploy fails at boot instead of silently minting tickets nobody can ever
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

// Issue implements ports.IStreamTicketStore.
func (s *Store) Issue(ctx context.Context, t ports.StreamTicket) (string, time.Time, error) {
	buf := make([]byte, ticketTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("generating stream ticket: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)

	payload, err := encode(t)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(s.ttl)

	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	// NX: a collision across 256 bits of entropy would mean something is
	// badly wrong (a broken RNG) - surfacing it as an error is correct,
	// never silently overwriting another caller's live ticket.
	ok, err := s.client.SetNX(opCtx, ticketKey(token), payload, s.ttl).Result()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("storing stream ticket: %w", err)
	}
	if !ok {
		return "", time.Time{}, errors.New("stream ticket token collision")
	}
	return token, expiresAt, nil
}

// Redeem implements ports.IStreamTicketStore.
func (s *Store) Redeem(ctx context.Context, token string) (ports.StreamTicket, error) {
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	res, err := redeemScript.Run(opCtx, s.client, []string{ticketKey(token)}).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ports.StreamTicket{}, ports.ErrTicketInvalid
		}
		return ports.StreamTicket{}, fmt.Errorf("redeeming stream ticket: %w", err)
	}
	payload, ok := res.(string)
	if !ok || payload == "" {
		return ports.StreamTicket{}, ports.ErrTicketInvalid
	}
	t, err := decode([]byte(payload))
	if err != nil {
		return ports.StreamTicket{}, fmt.Errorf("decoding stream ticket: %w", err)
	}
	return t, nil
}

// Close releases the underlying Redis client.
func (s *Store) Close() error {
	return s.client.Close()
}
