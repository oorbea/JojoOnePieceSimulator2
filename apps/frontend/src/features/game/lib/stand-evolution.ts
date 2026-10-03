import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'
import type { StandResponse } from '@/features/stands/types/stands.types'

// Anything with a single self-referencing evolvesFrom parent pointer: a Stand
// (domain/entities/powers/stand.go's evolvesFrom) or, since power effects, a
// DevilFruit whose game-attached parent is its pre-evolution (Hito Hito no
// mi: Model Nika <- Gomu Gomu no mi, see game.LinkFruitEvolutions).
type Evolvable<T> = { id: string; evolvesFrom: T | null }

// evolutionChain: STATE already delivers the full ancestor chain nested
// inside loadout.stand.evolvesFrom / loadout.devilFruit.evolvesFrom
// (dto/game_response.go's newGameLoadoutResponse), no wire change needed.
// This walks it and returns it ROOT-FIRST (the most base power in the
// family, e.g. Silver Chariot, first; the final one, e.g. Chariot Requiem,
// last) - the order the sorteo's evolution reveal plays in (2026-09-25,
// "algo esta pasando" feature). A power with no evolvesFrom returns a
// single-element array (itself), so `chain.length - 1` is always the number
// of evolution steps to reveal, and an element's index is its
// EvolutionDepth().
export function evolutionChain<T extends Evolvable<T>>(power: T): T[] {
  const chain: T[] = []
  let current: T | null = power
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

// evolutionSteps: how many evolution phases the reveal must play for this
// power - 0 for one with no evolvesFrom (today's behaviour, unchanged),
// otherwise chain.length - 1. Mirrors the backend's EvolutionDepth() exactly
// (see reveal.go's RevealPlayer - keep both in sync).
export function evolutionSteps<T extends Evolvable<T>>(power: T | null | undefined): number {
  if (!power) return 0
  return evolutionChain(power).length - 1
}

export function standEvolutionChain(stand: StandResponse): StandResponse[] {
  return evolutionChain(stand)
}

export function standEvolutionSteps(stand: StandResponse | null | undefined): number {
  return evolutionSteps(stand)
}

export function devilFruitEvolutionChain(fruit: DevilFruitResponse): DevilFruitResponse[] {
  return evolutionChain(fruit)
}

export function devilFruitEvolutionSteps(fruit: DevilFruitResponse | null | undefined): number {
  return evolutionSteps(fruit)
}
