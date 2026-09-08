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
// in-flight grant, and even then the underlying fetch isn't forcibly killed.

export type ImagePriority = number

export type QueueLane = 'grid' | 'high'

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

type Entry = {
  key: string
  priority: ImagePriority
  lane: QueueLane
  onGrant: () => void
  onTimeout: () => void
  granted: boolean
  settled: boolean
  watchdog: ReturnType<typeof setTimeout> | null
  refCount: number
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
    // The fetch is still technically in flight (nothing aborts it - see
    // cancellation policy), but the slot itself is reclaimed so the rest of
    // the grid isn't starved by one stalled/blocked request.
    releaseSlot(entry)
    entry.onTimeout()
  }, timeoutMs)
  entry.onGrant()
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

// Dedupe by key (typically the URI): two cards racing for the same image
// share one queue slot and one in-flight grant via refcount, and only the
// last consumer's `cancel`/`settle` actually tears the entry down.
export function enqueueImage(opts: {
  key: string
  priority: ImagePriority
  lane?: QueueLane
  onGrant: () => void
  onTimeout: () => void
}): ImageHandle {
  const { key, priority, lane = 'grid', onGrant, onTimeout } = opts
  const existing = waiting.get(key) ?? inFlight.get(key)
  if (existing) {
    existing.refCount += 1
    if (existing.granted) onGrant()
    else if (priority < existing.priority) existing.priority = priority
    return handleFor(existing, onGrant)
  }

  const entry: Entry = {
    key,
    priority,
    lane,
    onGrant,
    onTimeout,
    granted: false,
    settled: false,
    watchdog: null,
    refCount: 1,
  }
  waiting.set(key, entry)
  pump()
  return handleFor(entry, onGrant)
}

function handleFor(entry: Entry, ownOnGrant: () => void): ImageHandle {
  let released = false
  const release = () => {
    if (released) return
    released = true
    entry.refCount -= 1
    if (entry.refCount > 0) return
    if (entry.granted) releaseSlot(entry)
    else waiting.delete(entry.key)
  }
  return {
    settle: () => {
      entry.settled = true
      release()
    },
    reprioritize: (priority) => {
      if (entry.granted) return
      entry.priority = priority
    },
    cancel: () => {
      if (entry.granted) return
      release()
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
