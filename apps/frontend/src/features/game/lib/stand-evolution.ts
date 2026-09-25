import type { StandResponse } from '@/features/stands/types/stands.types'

// standEvolutionChain: a Stand's evolvesFrom is a single self-referencing
// parent pointer (backend: domain/entities/powers/stand.go's evolvesFrom,
// see ObsidianVault docs on the recursive-CTE reads) - STATE already
// delivers the full ancestor chain nested inside loadout.stand.evolvesFrom
// (dto/game_response.go's newGameLoadoutResponse -> NewStandResponse), no
// wire change needed. This walks it and returns it ROOT-FIRST (the most
// base stand in the family, e.g. Silver Chariot, first; the landed stand,
// e.g. Chariot Requiem, last) - the order the sorteo's evolution reveal
// plays in (2026-09-25, "algo esta pasando" feature). A stand with no
// evolvesFrom returns a single-element array (itself), so
// `chain.length - 1` is always the number of evolution steps to reveal.
export function standEvolutionChain(stand: StandResponse): StandResponse[] {
  const chain: StandResponse[] = []
  let current: StandResponse | null = stand
  const seen = new Set<string>()
  while (current) {
    // Defensive only - the backend enforces stands_no_self_evolution and the
    // mapper refuses to build a cycle (stand_mapper.go), so this should
    // never trigger outside a malformed/legacy row. Guards against an
    // infinite loop rather than trusting the wire data blindly.
    if (seen.has(current.id)) break
    seen.add(current.id)
    chain.unshift(current)
    current = current.evolvesFrom
  }
  return chain
}

// standEvolutionSteps: how many evolution phases the reveal must play for
// this stand - 0 for a stand with no evolvesFrom (today's behaviour,
// unchanged), otherwise chain.length - 1. Mirrors the backend's
// powers.Stand.EvolutionDepth() exactly (see reveal.go's RevealPlayer -
// keep both in sync).
export function standEvolutionSteps(stand: StandResponse | null | undefined): number {
  if (!stand) return 0
  return standEvolutionChain(stand).length - 1
}
