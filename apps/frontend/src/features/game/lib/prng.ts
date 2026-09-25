// Deterministic seeded PRNG (mulberry32) - the same seed always produces
// the exact same sequence, so every device replaying the same reveal draws
// the identical "random" reel/strip (see reel-geometry.ts's buildReel and
// case-strip.ts's buildCaseStrip, both seeded off the same per-slot hash
// loadout-reveal.ts already computes for revealSpinCycles). Not
// cryptographic - purely reproducible cosmetic randomness.
export function mulberry32(seed: number): () => number {
  let a = seed >>> 0
  return function random() {
    a |= 0
    a = (a + 0x6d2b79f5) | 0
    let t = Math.imul(a ^ (a >>> 15), 1 | a)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

// Fisher-Yates shuffle driven by a seeded PRNG - deterministic per `rand`
// (same generator state in, same permutation out), unlike Array.sort with a
// comparator (whose stability/order isn't guaranteed to be reproducible
// across engines for a "random" comparator).
export function seededShuffle<T>(items: T[], rand: () => number): T[] {
  const arr = items.slice()
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(rand() * (i + 1))
    const tmp = arr[i]
    arr[i] = arr[j]
    arr[j] = tmp
  }
  return arr
}
