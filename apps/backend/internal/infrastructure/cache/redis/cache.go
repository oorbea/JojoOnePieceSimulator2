// Package redis adapts ports.ICache onto Redis via go-redis. This is the
// only package in the codebase allowed to import the Redis client - callers
// only ever see ports.ICache.
//
// Invalidation is generation-based rather than key-enumeration-based: every
// namespace has a "generation" counter (jojo:<ns>:gen); every real entry is
// stored under a key that embeds the current generation
// (jojo:<ns>:<gen>:<key>). Invalidate(ns) is a single INCR, which atomically
// orphans every previously-written key in that namespace in O(1) - no KEYS
// or SCAN, no unbounded tag set to maintain. Orphaned keys simply expire on
// their own TTL.
//
// Every Get/Set/Delete resolves the current generation and the entry itself
// through a single EVAL, so it costs one round trip and can never observe a
// torn read against a concurrent Invalidate.
package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Config holds everything needed to reach one Redis instance.
type Config struct {
	// URL is a redis:// or rediss:// connection string (see redis.ParseURL).
	URL string
	// DialTimeout bounds the initial connection + PING done by New.
	DialTimeout time.Duration
	// OpTimeout bounds every individual Get/Set/Delete/Invalidate call, so a
	// slow or unreachable Redis can never add more than this to a request's
	// latency - the cache is fail-open, not a new point of failure.
	OpTimeout time.Duration
}

// Cache is the Redis-backed ports.ICache implementation.
type Cache struct {
	client    *redis.Client
	opTimeout time.Duration
	// healthy tracks the last known reachability, purely so a Redis outage
	// logs once on each transition instead of once per failed request.
	healthy atomic.Bool
}

var _ ports.ICache = (*Cache)(nil)

// getScript resolves ns's current generation and returns the value stored
// under it for key, or nil if there is none (including a stale-generation
// entry orphaned by a prior Invalidate).
//
// KEYS[1] = generation key ("jojo:<ns>:gen")
// ARGV[1] = ns, ARGV[2] = key
var getScript = redis.NewScript(`
local gen = redis.call('GET', KEYS[1])
if gen == false then gen = '0' end
return redis.call('GET', 'jojo:' .. ARGV[1] .. ':' .. gen .. ':' .. ARGV[2])
`)

// setScript resolves ns's current generation and stores val under it for
// key, with a TTL in milliseconds.
//
// KEYS[1] = generation key
// ARGV[1] = ns, ARGV[2] = key, ARGV[3] = value, ARGV[4] = ttl in ms
var setScript = redis.NewScript(`
local gen = redis.call('GET', KEYS[1])
if gen == false then gen = '0' end
redis.call('SET', 'jojo:' .. ARGV[1] .. ':' .. gen .. ':' .. ARGV[2], ARGV[3], 'PX', ARGV[4])
return 1
`)

// deleteScript resolves ns's current generation and deletes the entry stored
// under it for key.
//
// KEYS[1] = generation key
// ARGV[1] = ns, ARGV[2] = key
var deleteScript = redis.NewScript(`
local gen = redis.call('GET', KEYS[1])
if gen == false then gen = '0' end
return redis.call('DEL', 'jojo:' .. ARGV[1] .. ':' .. gen .. ':' .. ARGV[2])
`)

// genKey returns the generation counter key for ns.
func genKey(ns string) string {
	return "jojo:" + ns + ":gen"
}

// New connects to the Redis instance described by cfg and verifies
// reachability with a PING bounded by cfg.DialTimeout, so a misconfigured
// deployment fails at startup instead of on the first request.
func New(ctx context.Context, cfg Config) (*Cache, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	c := &Cache{client: client, opTimeout: cfg.OpTimeout}
	c.healthy.Store(true)
	return c, nil
}

// markHealthy logs only on a healthy<->unhealthy transition, so a sustained
// Redis outage produces one log line instead of one per failed request.
func (c *Cache) markHealthy(healthy bool, err error) {
	if c.healthy.Swap(healthy) == healthy {
		return
	}
	if healthy {
		log.Printf("cache: redis reachable again")
	} else {
		log.Printf("cache: redis unreachable, falling back to source of truth: %v", err)
	}
}

func (c *Cache) opContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.opTimeout)
}

// Get implements ports.ICache.
func (c *Cache) Get(ctx context.Context, ns, key string) ([]byte, bool) {
	opCtx, cancel := c.opContext(ctx)
	defer cancel()

	val, err := getScript.Run(opCtx, c.client, []string{genKey(ns)}, ns, key).Text()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.markHealthy(true, nil)
			return nil, false
		}
		c.markHealthy(false, err)
		return nil, false
	}
	c.markHealthy(true, nil)
	return []byte(val), true
}

// Set implements ports.ICache.
func (c *Cache) Set(ctx context.Context, ns, key string, val []byte, ttl time.Duration) {
	opCtx, cancel := c.opContext(ctx)
	defer cancel()

	if _, err := setScript.Run(opCtx, c.client, []string{genKey(ns)}, ns, key, val, ttl.Milliseconds()).Result(); err != nil {
		c.markHealthy(false, err)
		return
	}
	c.markHealthy(true, nil)
}

// Delete implements ports.ICache.
func (c *Cache) Delete(ctx context.Context, ns, key string) {
	opCtx, cancel := c.opContext(ctx)
	defer cancel()

	if _, err := deleteScript.Run(opCtx, c.client, []string{genKey(ns)}, ns, key).Result(); err != nil {
		c.markHealthy(false, err)
		return
	}
	c.markHealthy(true, nil)
}

// Invalidate implements ports.ICache.
func (c *Cache) Invalidate(ctx context.Context, ns string) error {
	opCtx, cancel := c.opContext(ctx)
	defer cancel()

	if err := c.client.Incr(opCtx, genKey(ns)).Err(); err != nil {
		c.markHealthy(false, err)
		return fmt.Errorf("invalidating namespace %q: %w", ns, err)
	}
	c.markHealthy(true, nil)
	return nil
}

// Close implements ports.ICache.
func (c *Cache) Close() error {
	return c.client.Close()
}
