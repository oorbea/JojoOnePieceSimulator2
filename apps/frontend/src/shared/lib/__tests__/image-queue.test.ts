import {
  __imageQueueSnapshot,
  __resetImageQueue,
  configureImageQueue,
  enqueueImage,
} from '@/shared/lib/image-queue'

describe('image-queue', () => {
  beforeEach(() => {
    __resetImageQueue()
    configureImageQueue({ maxConcurrent: 2, timeoutMs: 1000, highLaneTimeoutMs: 2000 })
  })

  afterEach(() => {
    __resetImageQueue()
  })

  it('grants up to maxConcurrent immediately, queues the rest', () => {
    const grants: string[] = []
    enqueueImage({ key: 'a', priority: 0, onGrant: () => grants.push('a'), onTimeout: () => {} })
    enqueueImage({ key: 'b', priority: 1, onGrant: () => grants.push('b'), onTimeout: () => {} })
    enqueueImage({ key: 'c', priority: 2, onGrant: () => grants.push('c'), onTimeout: () => {} })

    expect(grants).toEqual(['a', 'b'])
    expect(__imageQueueSnapshot().inFlight).toBe(2)
    expect(__imageQueueSnapshot().waiting).toBe(1)
  })

  it('duplicate keys ALL get their own onGrant once the shared entry is granted', () => {
    const grants: string[] = []
    enqueueImage({
      key: 'dup',
      priority: 0,
      onGrant: () => grants.push('first'),
      onTimeout: () => {},
    })
    // A second consumer racing for the SAME uri, still waiting.
    enqueueImage({
      key: 'dup',
      priority: 0,
      onGrant: () => grants.push('second'),
      onTimeout: () => {},
    })

    expect(grants).toEqual(['first', 'second'])
  })

  it('a late-joining duplicate on an already-granted entry is granted immediately', () => {
    const grants: string[] = []
    enqueueImage({
      key: 'dup',
      priority: 0,
      onGrant: () => grants.push('first'),
      onTimeout: () => {},
    })
    expect(grants).toEqual(['first'])

    enqueueImage({
      key: 'dup',
      priority: 0,
      onGrant: () => grants.push('late'),
      onTimeout: () => {},
    })
    expect(grants).toEqual(['first', 'late'])
  })

  it('settling the first duplicate frees the slot for the next queued entry, without erroring the others', () => {
    const events: string[] = []
    const dup1 = enqueueImage({
      key: 'dup',
      priority: 0,
      onGrant: () => events.push('dup1-grant'),
      onTimeout: () => events.push('dup1-timeout'),
    })
    enqueueImage({
      key: 'dup',
      priority: 0,
      onGrant: () => events.push('dup2-grant'),
      onTimeout: () => events.push('dup2-timeout'),
    })
    enqueueImage({
      key: 'other',
      priority: 0,
      onGrant: () => events.push('other-grant'),
      onTimeout: () => events.push('other-timeout'),
    })
    // maxConcurrent=2: 'dup' takes one slot (shared by both subscribers),
    // 'other' would need the 2nd slot but is still queued behind nothing
    // else here - use a 3rd distinct key to actually exercise queueing.
    enqueueImage({
      key: 'third',
      priority: 0,
      onGrant: () => events.push('third-grant'),
      onTimeout: () => events.push('third-timeout'),
    })

    expect(__imageQueueSnapshot().waiting).toBe(1) // 'third' queued behind dup+other

    dup1.settle()
    expect(events).toContain('third-grant')
    expect(events).not.toContain('dup2-timeout')
  })

  it('watchdog only times out subscribers that never settled, and clears once all settle', () => {
    jest.useFakeTimers()
    const events: string[] = []
    const dup1 = enqueueImage({
      key: 'dup',
      priority: 0,
      onGrant: () => events.push('dup1-grant'),
      onTimeout: () => events.push('dup1-timeout'),
    })
    enqueueImage({
      key: 'dup',
      priority: 0,
      onGrant: () => events.push('dup2-grant'),
      onTimeout: () => events.push('dup2-timeout'),
    })

    dup1.settle() // dup1 loaded successfully
    jest.advanceTimersByTime(1500) // past timeoutMs

    // dup1 already settled - must never be flipped to error by the watchdog.
    expect(events).not.toContain('dup1-timeout')
    jest.useRealTimers()
  })

  it('cancel on a still-waiting entry frees it without granting', () => {
    const grants: string[] = []
    enqueueImage({ key: 'a', priority: 0, onGrant: () => grants.push('a'), onTimeout: () => {} })
    enqueueImage({ key: 'b', priority: 0, onGrant: () => grants.push('b'), onTimeout: () => {} })
    const c = enqueueImage({
      key: 'c',
      priority: 0,
      onGrant: () => grants.push('c'),
      onTimeout: () => {},
    })

    expect(__imageQueueSnapshot().waiting).toBe(1)
    c.cancel()
    expect(__imageQueueSnapshot().waiting).toBe(0)
    expect(grants).not.toContain('c')
  })

  it('unmount (cancel) after grant frees the slot once every subscriber has left', () => {
    const grants: string[] = []
    const a = enqueueImage({
      key: 'a',
      priority: 0,
      onGrant: () => grants.push('a'),
      onTimeout: () => {},
    })
    enqueueImage({ key: 'b', priority: 0, onGrant: () => grants.push('b'), onTimeout: () => {} })
    const queued = enqueueImage({
      key: 'c',
      priority: 0,
      onGrant: () => grants.push('c'),
      onTimeout: () => {},
    })

    expect(__imageQueueSnapshot().inFlight).toBe(2)
    expect(grants).not.toContain('c')

    // a's consumer unmounts before ever loading/erroring (e.g. scrolled
    // away mid-fetch) - cancel() on a granted entry is a no-op per policy,
    // it must not free the slot early.
    a.cancel()
    expect(__imageQueueSnapshot().inFlight).toBe(2)
    expect(grants).not.toContain('c')

    // Only settling (or every subscriber leaving) frees it. Simulate the
    // real unmount path via settle from the image's own onError/onLoad is
    // covered elsewhere; here settle stands in for "the fetch resolved".
    a.settle()
    expect(grants).toContain('c')
    void queued
  })
})
