// Pure timing tables for the end-of-game and end-of-round cinematics - split
// out from the components exactly like reel-geometry.ts/loadout-reveal.ts so
// "do the phases actually cover the full track with no gap or overlap" is
// unit-testable without rendering Reanimated. Audio is the master clock here
// (see ObsidianVault/game-victory-defeat-cinematic-2026-09-14.md): every
// startMs/durationMs below is the beat map of the matching wav
// (defeat-full.wav / victory-full.wav), the visual layer just renders to it.
export type CinematicKind = 'victory' | 'defeat' | 'roundWin' | 'roundLose'

export type CinematicPhase = {
  /** Phase name, unique within a kind's timeline - the renderer switches on
   * this to pick which visual treatment mounts for the phase. */
  name: string
  startMs: number
  durationMs: number
}

export type CinematicTimeline = {
  kind: CinematicKind
  totalMs: number
  /** First ms at which the skip control activates - the owner's pacing
   * decision (2026-09-14): long enough for the first beat to land, short
   * enough nobody feels trapped. Round flashes are too short to skip at all
   * (skipAfterMs === totalMs). */
  skipAfterMs: number
  phases: CinematicPhase[]
}

function timelineFrom(
  kind: CinematicKind,
  skipAfterMs: number,
  beats: [name: string, durationMs: number][]
): CinematicTimeline {
  let cursor = 0
  const phases: CinematicPhase[] = beats.map(([name, durationMs]) => {
    const phase = { name, startMs: cursor, durationMs }
    cursor += durationMs
    return phase
  })
  return { kind, totalMs: cursor, skipAfterMs, phases }
}

const SKIP_AFTER_MS = 2000

const TIMELINES: Record<CinematicKind, CinematicTimeline> = {
  defeat: timelineFrom('defeat', SKIP_AFTER_MS, [
    ['impact', 400],
    ['fracture', 800],
    ['ink', 1000],
    ['verdict', 1800],
    ['silence', 1600],
    ['epilogue', 1600],
  ]),
  victory: timelineFrom('victory', SKIP_AFTER_MS, [
    ['breath', 600],
    ['dawn', 1200],
    ['coronation', 1600],
    ['memory', 2000],
    ['seal', 1200],
    ['epilogue', 1000],
  ]),
  roundWin: timelineFrom('roundWin', 1500, [['flash', 1500]]),
  roundLose: timelineFrom('roundLose', 1500, [['flash', 1500]]),
}

export function cinematicTimeline(kind: CinematicKind): CinematicTimeline {
  return TIMELINES[kind]
}
