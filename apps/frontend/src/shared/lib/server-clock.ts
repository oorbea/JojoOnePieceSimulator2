// Estimates this device's clock offset from the game server's, so a
// countdown computed against an absolute server deadline (closesAt,
// revealStartedAt, ...) reads the same on every device regardless of how
// skewed its own clock is - the bug a real two-device playtest exposed
// (ObsidianVault: game-frame-deadlines-2026-09-03.md already carried the
// deadline; nothing corrected for the receiving device's own clock).
//
// Every ServerFrame carries `serverTime` - the server's own clock at the
// instant it marshalled that frame (see dto.NewServerFrame on the backend).
// `offset = serverTime - Date.now()` at the moment a frame is received is a
// biased estimate of the true offset: network latency only ever delays
// arrival, so it can only push the observed offset DOWN from the truth,
// never up. The largest offset seen recently is therefore the closest to
// the true offset (the sample that happened to travel fastest) - the same
// "take the min-RTT sample" idea NTP itself uses, simplified to a rolling
// max over one-way estimates since there's no request/response round trip
// here to halve.
//
// Module-level, not React state, deliberately: this is a singleton
// estimate of a physical fact (this device's clock skew), not
// per-component state, and it must be readable from plain functions
// (use-now.ts) as well as components.

const SAMPLE_WINDOW = 20

let offsets: number[] = []

/** Feeds one frame's serverTime into the estimator. receivedAtMs defaults
 * to Date.now() - overridable for tests. Silently ignores an unparseable
 * serverTimeMs (NaN), rather than letting one bad sample poison the whole
 * window. */
export function recordServerTime(serverTimeMs: number, receivedAtMs: number = Date.now()): void {
  if (!Number.isFinite(serverTimeMs)) return
  offsets.push(serverTimeMs - receivedAtMs)
  if (offsets.length > SAMPLE_WINDOW) offsets.shift()
}

/** This device's best current estimate of serverTime - Date.now(), in ms.
 * 0 until the first sample arrives (i.e. behaves exactly like Date.now()
 * for a client that has received no frames yet, or on a platform where
 * WebSocket never connects). */
export function clockOffsetMs(): number {
  if (offsets.length === 0) return 0
  return Math.max(...offsets)
}

/** Date.now(), corrected for this device's estimated clock offset from the
 * server - the drop-in replacement for Date.now() everywhere a countdown is
 * measured against a server-stamped deadline (see shared/hooks/use-now.ts). */
export function serverNow(): number {
  return Date.now() + clockOffsetMs()
}

/** Test-only / session-boundary reset - clears every recorded sample so a
 * fresh estimate builds from scratch. */
export function resetServerClock(): void {
  offsets = []
}
