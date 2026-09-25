// The concurrency-limiting core of the slow-network catalogue fix (see
// ObsidianVault/entrega-imagenes-red-lenta-2026-09-07.md, root cause a). A
// grid of N cards mounting at once used to start N image fetches at once -
// over HTTP/2 each stream gets bandwidth/N, so on a poor connection *none*
// of them finish. `expo-image`'s `priority` prop is only a hint to the
// native loader, not a limiter (and has no effect at all on web) - the only
// thing that reliably caps concurrency on both platforms is holding back the
// `source` itself until this queue grants a slot.
//
// Pure module state, no React - `use-image-slot.ts` is the React binding.
// Cancellation policy (deliberate, see the plan): an entry still *waiting*
// for a slot is cancelled for free (nothing was requested yet - this is the
// common case, e.g. a debounced search swapping the whole grid). An entry
// already *granted* a slot (in flight) is never evicted by the queue itself,
// on either platform - neither web nor native can resume a partial image
// fetch, so aborting mid-download on a slow link is strictly worse than
// letting it finish. Only real unmount (the caller's own cleanup) drops an
// in-flight grant, and the slot itself is freed once every subscriber of
// that grant has gone away (see "Dedupe" below) - the underlying fetch
// isn't forcibly killed either way.
//
// Dedupe (2026-09-25 fix - see ObsidianVault/sorteo-strip-image-dedupe-bug-
// 2026-09-25.md): several consumers can request the SAME uri at once (the
// sorteo strip shows the same Stand/Devil Fruit picture on many tiles). They
// share one queue slot and one underlying fetch, but EVERY subscriber must
// still get its own onGrant - a caller only mounts its <Image> once notified
// - and the slot must free as soon as the FIRST subscriber settles (loaded
// or errored), not only once every subscriber has settled. Once one
// subscriber's fetch has resolved, the resource is in the platform's own
// image cache (browser HTTP cache / expo-image disk cache), so every other
// subscriber's <Image> resolves from cache near-instantly - it does not
// need to keep holding a concurrency slot, and the watchdog must not fire
// on a subscriber whose image already loaded.

export type ImagePriority = number

// 'grid'/'high': the two concurrency-limited lanes documented above.
// 'hero': for the single foreground image on screen (2026-09-25 sorteo
// reveal-card fix, see ObsidianVault/sorteo-strip-image-dedupe-bug-2026-09-
// 25.md) - it bypasses the queue's concurrency limit entirely instead of
// competing for one of maxConcurrent slots. The bug this fixes: the reel's
// 50 tiles unmount the instant a slot lands, but a tile whose fetch was
// already granted keeps its slot until it settles or its watchdog fires (up
// to highLaneTimeoutMs) - that's the deliberate cancellation policy above,
// working as designed for the grid. The reveal card that mounts right after
// is the only image on screen at that moment and reading it fast matters far
// more than smoothing bandwidth across a strip that no longer exists, so it
// must never wait behind those orphaned slots.
export type QueueLane = 'grid' | 'high' | 'hero'

export type ImageHandle = {
  // Marks this image as done consuming its slot (loaded or errored),
  // freeing it for the next queued entry. Idempotent.
  settle: () => void
  // Re-ranks a still-queued entry (e.g. viewport bucket changed). No-op
  // once granted.
  reprioritize: (priority: ImagePriority) => void
  // Cancels a still-queued entry so it never receives a grant. No-op once
  // granted (see cancellation policy above).
  cancel: () => void
}

type Subscriber = {
  onGrant: () => void
  onTimeout: () => void
  settled: boolean
}

type Entry = {
  key: string
  priority: ImagePriority
  lane: QueueLane
  granted: boolean
  anySettled: boolean
  watchdog: ReturnType<typeof setTimeout> | null
  subscribers: Set<Subscriber>
}

type QueueConfig = {
  maxConcurrent: number
  timeoutMs: number
  highLaneTimeoutMs: number
}

const DEFAULT_CONFIG: QueueConfig = {
  maxConcurrent: 4,
  timeoutMs: 20_000,
  highLaneTimeoutMs: 45_000,
}

let config: QueueConfig = { ...DEFAULT_CONFIG }
const waiting = new Map<string, Entry>()
const inFlight = new Map<string, Entry>()

export function configureImageQueue(partial: Partial<QueueConfig>): void {
  config = { ...config, ...partial }
}

function grant(entry: Entry): void {
  waiting.delete(entry.key)
  entry.granted = true
  inFlight.set(entry.key, entry)
  const timeoutMs = entry.lane === 'high' ? config.highLaneTimeoutMs : config.timeoutMs
  entry.watchdog = setTimeout(() => {
    // Only the subscribers that never got a result of their own are timed
    // out - one that already loaded/errored (anySettled) keeps its result,
    // it must never be flipped back to 'error' just because a *different*
    // subscriber of the same uri is still pending (e.g. a slow/blocked
    // duplicate elsewhere on screen).
    for (const sub of entry.subscribers) {
      if (!sub.settled) sub.onTimeout()
    }
    releaseSlot(entry)
  }, timeoutMs)
  // Every current subscriber gets its own onGrant - each caller mounts its
  // own <Image> and races the platform cache independently.
  for (const sub of entry.subscribers) sub.onGrant()
}

function releaseSlot(entry: Entry): void {
  if (entry.watchdog) {
    clearTimeout(entry.watchdog)
    entry.watchdog = null
  }
  inFlight.delete(entry.key)
  pump()
}

function pump(): void {
  while (inFlight.size < config.maxConcurrent && waiting.size > 0) {
    let best: Entry | null = null
    for (const entry of waiting.values()) {
      if (!best || entry.priority < best.priority) best = entry
    }
    if (!best) break
    grant(best)
  }
}

// Dedupe by key (typically the URI): every caller requesting the same key
// shares one queue slot and one in-flight grant, but each still gets its
// own onGrant/onTimeout callback - see the module doc comment above.
export function enqueueImage(opts: {
  key: string
  priority: ImagePriority
  lane?: QueueLane
  onGrant: () => void
  onTimeout: () => void
}): ImageHandle {
  const { key, priority, lane = 'grid', onGrant, onTimeout } = opts
  const sub: Subscriber = { onGrant, onTimeout, settled: false }

  // 'hero' bypasses the concurrency limit entirely - granted immediately,
  // never enters `waiting`/`inFlight`, so it never counts against
  // maxConcurrent and never waits behind another lane's orphaned slots (see
  // QueueLane's doc). Still gets a watchdog so a genuinely stuck hero fetch
  // eventually flips to 'error' instead of hanging forever.
  if (lane === 'hero') {
    const entry: Entry = {
      key,
      priority,
      lane,
      granted: true,
      anySettled: false,
      watchdog: setTimeout(() => {
        if (!sub.settled) sub.onTimeout()
      }, config.timeoutMs),
      subscribers: new Set([sub]),
    }
    sub.onGrant()
    return handleFor(entry, sub)
  }

  const existing = waiting.get(key) ?? inFlight.get(key)
  if (existing) {
    existing.subscribers.add(sub)
    if (existing.granted) sub.onGrant()
    else if (priority < existing.priority) existing.priority = priority
    return handleFor(existing, sub)
  }

  const entry: Entry = {
    key,
    priority,
    lane,
    granted: false,
    anySettled: false,
    watchdog: null,
    subscribers: new Set([sub]),
  }
  waiting.set(key, entry)
  pump()
  return handleFor(entry, sub)
}

function handleFor(entry: Entry, sub: Subscriber): ImageHandle {
  let released = false
  const leave = () => {
    if (released) return
    released = true
    entry.subscribers.delete(sub)
    if (entry.subscribers.size === 0) {
      if (entry.granted) releaseSlot(entry)
      else waiting.delete(entry.key)
    }
  }
  return {
    settle: () => {
      if (sub.settled) return
      sub.settled = true
      const firstSettle = !entry.anySettled
      entry.anySettled = true
      leave()
      // The underlying resource is now resolved (loaded or errored) for
      // this uri - free the concurrency slot for the rest of the queue as
      // soon as the FIRST subscriber settles, instead of waiting for every
      // duplicate tile to individually finish/unmount.
      if (firstSettle && entry.granted && inFlight.has(entry.key)) releaseSlot(entry)
    },
    reprioritize: (priority) => {
      if (entry.granted) return
      entry.priority = priority
    },
    cancel: () => {
      // The "no-op once granted" policy exists to protect an in-flight
      // download's queue slot from being freed early (see the module doc).
      // 'hero' never held a slot to protect - always release on unmount, or
      // its watchdog timer leaks until it fires.
      if (entry.granted && entry.lane !== 'hero') return
      leave()
    },
  }
}

// visible: order among currently-visible items, lowest wins.
// near: about to scroll into view - still cheap to prefetch.
// far: everything else - queued behind visible/near but never skipped, so
// the grid keeps progressing while the user looks elsewhere (free prefetch).
export function priorityFor(visibility: 'visible' | 'near' | 'far', order: number): ImagePriority {
  if (visibility === 'visible') return order
  if (visibility === 'near') return 10_000 + order
  return 1_000_000 + order
}

// Test-only.
export function __resetImageQueue(): void {
  waiting.clear()
  inFlight.clear()
  config = { ...DEFAULT_CONFIG }
}

export function __imageQueueSnapshot() {
  return {
    waiting: waiting.size,
    inFlight: inFlight.size,
    maxConcurrent: config.maxConcurrent,
  }
}
