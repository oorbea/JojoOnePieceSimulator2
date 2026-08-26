import { teamTone } from '@/features/game/lib/lobby-rules'
import { currentRound } from '@/features/game/lib/match-rules'
import type { GameSnapshot, GameViewer } from '@/features/game/types/game.types'
import type { GlossButtonTone } from '@/shared/components/presentational/gloss-button'

export type VoteOption = {
  id: string
  /** Gauntlet: an i18n key (`game.vote.gauntlet.SURVIVE`/`FALL`) - the copy
   * is the owner's own wording, never derived from the wire value directly.
   * Versus: undefined - `label` carries the team's name instead, since team
   * names are never translated (see i18n-multi-language's "names are never
   * translated" rule). Exactly one of labelKey/label is set. */
  labelKey?: string
  label?: string
  tone: GlossButtonTone
  /** Versus only: this option is the caller's own team. Informational only
   * - voting for the rival team stays legal, per the domain (no self-vote
   * restriction, see gameplay-game-modes.md). Always false in Gauntlet. */
  isOwnTeam: boolean
}

// GAUNTLET_TONES intentionally mirrors the domain's own SURVIVE/FALL
// semantics (green = good outcome, red = bad), not the generic team-tone
// palette below - a squad's Gauntlet vote isn't a team choice.
const GAUNTLET_TONES: Record<string, GlossButtonTone> = {
  SURVIVE: 'green',
  FALL: 'red',
}

// voteOptions maps a round's raw option ids (round.Ballot.Options - literal
// "SURVIVE"/"FALL" strings for Gauntlet, raw team-UUID strings for Versus,
// see game.OptionID's doc) onto what the vote bar actually renders. Pure and
// unit-tested (lib/__tests__/vote-options.test.ts) - keep any future
// mode-specific vote copy/tone decision here, not inline in a component.
export function voteOptions(snapshot: GameSnapshot, you: GameViewer): VoteOption[] {
  const round = currentRound(snapshot)
  if (!round) return []

  if (snapshot.mode === 'GAUNTLET') {
    return round.options.map((id) => ({
      id,
      labelKey: `game.vote.gauntlet.${id}`,
      tone: GAUNTLET_TONES[id] ?? 'glass',
      isOwnTeam: false,
    }))
  }

  return round.options.map((id) => {
    const teamIndex = snapshot.teams.findIndex((t) => t.id === id)
    const team = teamIndex >= 0 ? snapshot.teams[teamIndex] : undefined
    return {
      id,
      // Team name is a proper noun - never translated, straight off the
      // snapshot. Falls back to the raw id only if a round somehow carries
      // an option with no matching team (shouldn't happen; keeps the UI
      // from rendering blank rather than crashing).
      label: team?.name ?? id,
      // Reuses lobby-rules.ts's teamTone (the same tone TeamColumn/
      // MatchRoster already assign that team) rather than a second tone
      // table - GlossButtonTone and TeamTone overlap on exactly the four
      // values teamTone ever returns.
      tone: teamIndex >= 0 ? teamTone(teamIndex) : 'glass',
      isOwnTeam: id === you.teamId,
    }
  })
}
