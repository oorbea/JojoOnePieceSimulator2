package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Config holds everything needed to reach the Redis instance backing the
// game store.
type Config struct {
	// URL is a redis:// or rediss:// connection string.
	URL string
	// DialTimeout bounds the initial connection + PING done by New.
	DialTimeout time.Duration
	// OpTimeout bounds every individual store operation. Deliberately far
	// larger than infrastructure/cache/redis's REDIS_OP_TIMEOUT (200ms):
	// this store is fail-closed, not a cache, so it should wait rather than
	// fail a vote under transient latency.
	OpTimeout time.Duration
	// TTL is how long a Game survives without being Saved again. Refreshed
	// on every Create/Save - see DeleteExpired's doc for why that alone is
	// enough.
	TTL time.Duration
}

// Store is a Redis-backed ports.IGameStore. Unlike infrastructure/cache's
// decorators, it is fail-CLOSED: every error is returned to the caller,
// never swallowed as a miss, because it is the source of truth for a live
// match, not an optional speedup over one. A corrupt or unreadable payload
// is likewise an error, never treated as "not found" - the key is left in
// place (its TTL will clear it) so it stays available for debugging.
//
// Three keys per game share one TTL, refreshed together on every write:
//
//	jojo:game:id:<gameID>     -> the envelope JSON (the aggregate)
//	jojo:game:code:<CODE>     -> <gameID>, for GetByCode
//	jojo:game:codeof:<gameID> -> <CODE>, for Code and for refreshing the
//	                             code index without a JSON round trip
//
// This shares the "jojo:" root with infrastructure/cache but not its
// <ns>:<gen>: layout - generation-based invalidation is a cache concept,
// and an INCR there would orphan live lobbies here.
type Store struct {
	client    *goredis.Client
	opTimeout time.Duration
	ttl       time.Duration
	now       func() time.Time
}

var _ ports.IGameStore = (*Store)(nil)

// createScript atomically claims a join code and writes both the payload
// and the code index under the same TTL. Returns 0 if the code was already
// taken (by a different, still-live game), 1 on success.
//
// KEYS[1] = id key, KEYS[2] = code key, KEYS[3] = codeof key
// ARGV[1] = id, ARGV[2] = code, ARGV[3] = payload, ARGV[4] = ttl in ms
var createScript = goredis.NewScript(`
if redis.call('SET', KEYS[2], ARGV[1], 'NX', 'PX', ARGV[4]) == false then
	return 0
end
redis.call('SET', KEYS[1], ARGV[3], 'PX', ARGV[4])
redis.call('SET', KEYS[3], ARGV[2], 'PX', ARGV[4])
return 1
`)

// saveScript overwrites the payload and refreshes all three keys' TTL in
// one round trip. Returns 0 if the game was never Create'd (or has already
// expired/been deleted), 1 on success.
//
// KEYS[1] = id key, KEYS[2] = codeof key
// ARGV[1] = payload, ARGV[2] = ttl in ms
var saveScript = goredis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
	return 0
end
redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
local code = redis.call('GET', KEYS[2])
if code then
	redis.call('PEXPIRE', KEYS[2], ARGV[2])
	redis.call('PEXPIRE', 'jojo:game:code:' .. code, ARGV[2])
end
return 1
`)

// deleteScript removes all three keys for a game, resolving the code index
// first so the code key can be found.
//
// KEYS[1] = id key, KEYS[2] = codeof key
var deleteScript = goredis.NewScript(`
local code = redis.call('GET', KEYS[2])
redis.call('DEL', KEYS[1], KEYS[2])
if code then
	redis.call('DEL', 'jojo:game:code:' .. code)
end
return 1
`)

func idKey(id game.GameID) string     { return "jojo:game:id:" + id.String() }
func codeKey(code string) string      { return "jojo:game:code:" + code }
func codeOfKey(id game.GameID) string { return "jojo:game:codeof:" + id.String() }

// New connects to the Redis instance described by cfg and verifies
// reachability with a PING bounded by cfg.DialTimeout, so a misconfigured
// game store fails the deploy instead of silently losing every lobby.
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

	return &Store{client: client, opTimeout: cfg.OpTimeout, ttl: cfg.TTL, now: time.Now}, nil
}

func (s *Store) opContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, s.opTimeout)
}

// Create implements ports.IGameStore.
func (s *Store) Create(ctx context.Context, code string, g *game.Game) error {
	payload, err := encode(g, s.now())
	if err != nil {
		return err
	}
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	res, err := createScript.Run(opCtx, s.client,
		[]string{idKey(g.ID()), codeKey(code), codeOfKey(g.ID())},
		g.ID().String(), code, payload, s.ttl.Milliseconds(),
	).Int()
	if err != nil {
		return fmt.Errorf("creating game %s: %w", g.ID(), err)
	}
	if res == 0 {
		return ports.ErrGameCodeTaken
	}
	return nil
}

// Get implements ports.IGameStore.
func (s *Store) Get(ctx context.Context, id game.GameID) (*game.Game, error) {
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	payload, err := s.client.Get(opCtx, idKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, ports.ErrGameNotFound
		}
		return nil, fmt.Errorf("getting game %s: %w", id, err)
	}
	g, err := decode(payload)
	if err != nil {
		return nil, fmt.Errorf("decoding game %s: %w", id, err)
	}
	return g, nil
}

// GetByCode implements ports.IGameStore.
func (s *Store) GetByCode(ctx context.Context, code string) (*game.Game, error) {
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	idStr, err := s.client.Get(opCtx, codeKey(code)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, ports.ErrGameNotFound
		}
		return nil, fmt.Errorf("resolving code %q: %w", code, err)
	}
	id, err := game.ParseGameID(idStr)
	if err != nil {
		return nil, fmt.Errorf("resolving code %q: %w", code, err)
	}
	return s.Get(ctx, id)
}

// Code implements ports.IGameStore.
func (s *Store) Code(ctx context.Context, id game.GameID) (string, error) {
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	code, err := s.client.Get(opCtx, codeOfKey(id)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", ports.ErrGameNotFound
		}
		return "", fmt.Errorf("getting code for game %s: %w", id, err)
	}
	return code, nil
}

// Save implements ports.IGameStore.
func (s *Store) Save(ctx context.Context, g *game.Game) error {
	payload, err := encode(g, s.now())
	if err != nil {
		return err
	}
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	res, err := saveScript.Run(opCtx, s.client,
		[]string{idKey(g.ID()), codeOfKey(g.ID())},
		payload, s.ttl.Milliseconds(),
	).Int()
	if err != nil {
		return fmt.Errorf("saving game %s: %w", g.ID(), err)
	}
	if res == 0 {
		return ports.ErrGameNotFound
	}
	return nil
}

// Delete implements ports.IGameStore.
func (s *Store) Delete(ctx context.Context, id game.GameID) error {
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	if _, err := deleteScript.Run(opCtx, s.client, []string{idKey(id), codeOfKey(id)}).Result(); err != nil {
		return fmt.Errorf("deleting game %s: %w", id, err)
	}
	return nil
}

// DeleteExpired implements ports.IGameStore. It always returns 0 and does
// nothing: expiry is delegated entirely to Redis's own PX TTL, refreshed on
// every Create/Save, which already implements "remove anything last saved
// more than olderThan ago" - olderThan is therefore ignored, and changing
// GameLobbyTTL only affects newly written keys. There is deliberately no
// SCAN-based sweep (infrastructure/cache/redis rejects KEYS/SCAN for the
// same reason), and since all three keys per game share one TTL there are
// no orphans to sweep. The Reaper stays wired unconditionally against this
// Store - it is a harmless no-op ticker against it.
func (s *Store) DeleteExpired(context.Context, time.Duration) int {
	return 0
}

// Close releases the underlying Redis client.
func (s *Store) Close() error {
	return s.client.Close()
}
