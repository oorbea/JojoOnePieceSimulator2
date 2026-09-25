import type { UseQueryResult } from '@tanstack/react-query'
import { Image as ExpoImage } from 'expo-image'
import { useEffect, useRef } from 'react'

import { applyPoolFilter } from '@/features/game/lib/power-pool'
import { standEvolutionChain } from '@/features/game/lib/stand-evolution'
import type { GameSnapshot } from '@/features/game/types/game.types'
import type { DevilFruitResponse } from '@/features/devil-fruits'
import type { StandResponse } from '@/features/stands'
import { cardSource, thumbSource } from '@/shared/lib/picture-source'

// Warms expo-image's own disk/memory cache for every Stand/Devil Fruit the
// lobby's power pool could roll, well before the sorteo strip needs them
// (see ObsidianVault/sorteo-strip-image-dedupe-bug-2026-09-25.md). The
// strip (case-strip-reel.tsx) shows the SAME ~50 candidates on repeat, so
// once this prefetch lands them in cache, every strip tile resolves near-
// instantly regardless of the app-wide image-queue's concurrency cap - this
// is the only thing in the app that fires image requests outside that
// queue on purpose, and only for lobby-known, non-urgent background work.
//
// Deliberately NOT routed through image-queue.ts: that queue exists to cap
// concurrency for USER-VISIBLE wells competing for bandwidth right now.
// This prefetch has no well to fill and no deadline - it is pure
// background warmup, so it is throttled by its own small batch size
// instead, and yields between batches so it never contends with whatever
// the lobby is actually rendering.
const BATCH_SIZE = 6
const BATCH_DELAY_MS = 150

// Rows without a backfilled picture_media_id fall back to a 15-minute
// presigned R2 URL (see ObsidianVault/media-proxy-content-addressed.md) -
// a lobby that's been open longer than that can be sitting on a stale
// catalogue fetch whose thumb URLs have already expired. Refetching once
// per ASSIGNING entry (not on every render) keeps this cheap.
const CATALOG_STALE_MS = 5 * 60_000

async function prefetchInBatches(urls: string[]): Promise<void> {
  for (let i = 0; i < urls.length; i += BATCH_SIZE) {
    const batch = urls.slice(i, i + BATCH_SIZE)
    await Promise.all(batch.map((uri) => ExpoImage.prefetch(uri, { cachePolicy: 'disk' })))
    if (i + BATCH_SIZE < urls.length) {
      await new Promise((resolve) => setTimeout(resolve, BATCH_DELAY_MS))
    }
  }
}

// usePowerPoolPrefetch: call once the lobby/match screen has both the
// snapshot's poolFilter and the full catalogue loaded. Re-fires whenever
// the resolved candidate set's ids OR urls change (a config edit, a
// catalogue refetch picking up fresh presigned URLs, or - for Versus's
// ReassignsEachRound - a new round) so a stale/expired presigned URL from
// before never lingers as the only warmed copy.
export function usePowerPoolPrefetch(
  snapshot: GameSnapshot | null,
  standsQuery: Pick<UseQueryResult<StandResponse[]>, 'data' | 'dataUpdatedAt' | 'refetch'>,
  devilFruitsQuery: Pick<UseQueryResult<DevilFruitResponse[]>, 'data' | 'dataUpdatedAt' | 'refetch'>
): void {
  const lastKeyRef = useRef<string | null>(null)
  const lastAssigningRefetchRef = useRef<string | null>(null)

  // Refresh the catalogue once per ASSIGNING window if it's old enough that
  // its presigned thumb URLs may already have expired - see
  // CATALOG_STALE_MS above. Keyed on the round/id, not just `state`, so a
  // reconnect that re-delivers the same ASSIGNING snapshot doesn't refetch
  // twice.
  useEffect(() => {
    if (!snapshot || snapshot.state !== 'ASSIGNING') return
    const assigningKey = `${snapshot.id}:${snapshot.rounds.length}`
    if (lastAssigningRefetchRef.current === assigningKey) return
    const oldestUpdate = Math.min(
      standsQuery.dataUpdatedAt || 0,
      devilFruitsQuery.dataUpdatedAt || 0
    )
    if (oldestUpdate === 0 || Date.now() - oldestUpdate < CATALOG_STALE_MS) return
    lastAssigningRefetchRef.current = assigningKey
    void standsQuery.refetch()
    void devilFruitsQuery.refetch()
  }, [snapshot, standsQuery, devilFruitsQuery])

  const stands = standsQuery.data
  const fruits = devilFruitsQuery.data

  useEffect(() => {
    if (!snapshot || !stands?.length || !fruits?.length) return

    const filtered = applyPoolFilter(stands, fruits, snapshot.config.poolFilter)
    const urls = [
      ...filtered.stands.map((s) => thumbSource(s)),
      ...filtered.fruits.map((f) => thumbSource(f)),
    ].filter((u): u is string => !!u)
    if (urls.length === 0) return

    // Re-run when the candidate set's ids OR any url actually changes
    // (e.g. a refetch above picked up fresh presigned URLs for the same
    // ids) - not on every unrelated snapshot update (READY_COUNT ticks,
    // chat, etc. all reuse the same GameSnapshot type).
    const key = urls.slice().sort().join('|')
    if (key === lastKeyRef.current) return
    lastKeyRef.current = key

    let cancelled = false
    void (async () => {
      if (cancelled) return
      await prefetchInBatches(urls)
    })()

    return () => {
      cancelled = true
    }
  }, [snapshot, stands, fruits])

  // Second, separate warmup for the CARD (512px) rendition - the reveal
  // card (power-reveal-card.tsx) shows cardSource(), not the thumb the loop
  // above warms, and only for the participants' own already-landed
  // loadouts plus every intermediate stage of a Stand's evolution chain
  // (2026-09-25 "algo esta pasando" feature) - a landed stand is an
  // ordinary pool entry, so its base/intermediate stages aren't necessarily
  // in the pool filter at all. Small, fixed-size set (players * ~1-4
  // stages), so no batching needed. Keyed the same way as above so a
  // reconnect/refetch with fresh presigned URLs re-warms instead of
  // silently reusing an expired one.
  const cardKeyRef = useRef<string | null>(null)
  useEffect(() => {
    if (!snapshot) return
    const urls = new Set<string>()
    for (const participant of snapshot.participants) {
      const loadout = participant.loadout
      if (!loadout) continue
      if (loadout.stand) {
        for (const stage of standEvolutionChain(loadout.stand)) {
          const url = cardSource(stage)
          if (url) urls.add(url)
        }
      }
      if (loadout.devilFruit) {
        const url = cardSource(loadout.devilFruit)
        if (url) urls.add(url)
      }
    }
    if (urls.size === 0) return
    const key = Array.from(urls).sort().join('|')
    if (key === cardKeyRef.current) return
    cardKeyRef.current = key

    let cancelled = false
    void (async () => {
      if (cancelled) return
      await prefetchInBatches(Array.from(urls))
    })()

    return () => {
      cancelled = true
    }
  }, [snapshot])
}
