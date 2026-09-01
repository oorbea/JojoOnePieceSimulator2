import type { VoteTallyEntry } from '@/features/game/lib/vote-options'
import { voteTally } from '@/features/game/lib/vote-options'
import type { GameSnapshot, GameViewer } from '@/features/game/types/game.types'
import type { GameMode } from '@/shared/lib/zod'

// One round as the final result screen recaps it: the stage that was played
// and how the vote broke down, in the same fixed option order (and with the
// same labels/tones) the vote bar used while it was live.
export type MatchRecapRound = {
  index: number
  stageName: string
  winnerOptionId: string | null
  decidedByCoinFlip: boolean
  entries: VoteTallyEntry[]
  maxCount: number
  /** True when this round never resolved - the game was aborted mid-round.
   * Its votes are shown as they stood, with no winner. */
  unresolved: boolean
}

// One seat's outcome. `won` is deliberately nullable rather than false:
// only VERSUS has per-participant winners (your team either won or didn't).
// GAUNTLET is a co-op run where the whole squad shares one verdict, so
// claiming an individual "lost" there would be simply untrue - the screen
// renders a collective banner instead. Null means "not applicable here".
export type MatchOutcome = {
  participantId: string
  displayName: string
  teamId: string
  bot: boolean
  isSelf: boolean
  won: boolean | null
}

export type MatchRecap = {
  mode: GameMode
  aborted: boolean
  roundsPlayed: number
  /** Raw winning option id: a team id in VERSUS, 'SURVIVE'/'FALL' in
   * GAUNTLET. Null when the game was aborted before anything resolved. */
  winnerOptionId: string | null
  /** VERSUS only: the winning team's name, resolved off the snapshot's own
   * team list (a proper noun, never translated). Null in GAUNTLET, and when
   * there is no winner. */
  winnerTeamName: string | null
  /** GAUNTLET only: did the squad survive its run? Null in VERSUS, and when
   * an aborted game never reached a verdict. */
  squadSurvived: boolean | null
  rounds: MatchRecapRound[]
  outcomes: MatchOutcome[]
}

// matchRecap derives everything the final result screen renders from a
// terminal snapshot. Pure and unit-tested (lib/__tests__/game-result.test.ts)
// - keep any future result-screen decision here rather than inline in the
// component, the same split vote-options.ts already follows for the vote bar.
//
// It reuses voteTally verbatim for every round rather than re-deriving option
// labels, which is only sound because a game's ballot options are stable
// across its rounds in both modes (SURVIVE/FALL in Gauntlet, the same two
// team ids in Versus) - so the last round's option list, which is what
// voteOptions reads, describes every earlier round too. If a mode ever
// varies its options per round, this is the assumption that breaks.
export function matchRecap(snapshot: GameSnapshot, you: GameViewer): MatchRecap {
  const result = snapshot.result
  const aborted = result?.aborted ?? snapshot.state === 'ABORTED'

  const winnerOptionId = result?.winner ? result.winner : null

  const winnerTeamName =
    snapshot.mode === 'VERSUS' && winnerOptionId
      ? (snapshot.teams.find((t) => t.id === winnerOptionId)?.name ?? null)
      : null

  const squadSurvived =
    snapshot.mode === 'GAUNTLET' && winnerOptionId ? winnerOptionId === 'SURVIVE' : null

  const rounds: MatchRecapRound[] = snapshot.rounds.map((round) => {
    // A resolved round reveals its full votes map; a round cut short by an
    // abort may only have the tied breakdown, or nothing at all.
    const votes = round.votes ?? round.tiedVotes ?? {}
    const { entries, maxCount } = voteTally(snapshot, you, votes)
    return {
      index: round.index,
      stageName: round.stage.name,
      winnerOptionId: round.result?.winner ?? null,
      decidedByCoinFlip: round.result?.decidedByCoinFlip ?? false,
      entries,
      maxCount,
      unresolved: !round.result,
    }
  })

  // Prefer the server's own end-of-game seat list (it is the roster as it
  // stood the moment the game ended, which is what an outcome should be
  // about), falling back to the live participant list for a game finished
  // by a backend that predates that field.
  const seats =
    result?.participants && result.participants.length > 0
      ? result.participants.map((p) => ({
          participantId: p.participantId,
          displayName: p.displayName,
          teamId: p.teamId,
          bot: p.bot,
        }))
      : snapshot.participants.map((p) => ({
          participantId: p.id,
          displayName: p.displayName,
          teamId: p.teamId,
          bot: p.kind === 'BOT',
        }))

  const outcomes: MatchOutcome[] = seats.map((seat) => ({
    ...seat,
    isSelf: seat.participantId === you.participantId,
    won: snapshot.mode === 'VERSUS' && winnerOptionId ? seat.teamId === winnerOptionId : null,
  }))

  return {
    mode: snapshot.mode,
    aborted,
    roundsPlayed: result?.roundsPlayed ?? snapshot.rounds.length,
    winnerOptionId,
    winnerTeamName,
    squadSurvived,
    rounds,
    outcomes,
  }
}
