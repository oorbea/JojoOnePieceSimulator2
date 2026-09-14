import { cinematicTimeline } from '@/features/game/lib/outcome-cinematic'
import type { CinematicKind } from '@/features/game/lib/outcome-cinematic'

const KINDS: CinematicKind[] = ['victory', 'defeat', 'roundWin', 'roundLose']

describe('cinematicTimeline', () => {
  it.each(KINDS)('%s: phases are contiguous and sum to the total duration', (kind) => {
    const timeline = cinematicTimeline(kind)

    expect(timeline.phases.length).toBeGreaterThan(0)
    expect(timeline.phases[0].startMs).toBe(0)

    let cursor = 0
    for (const phase of timeline.phases) {
      expect(phase.startMs).toBe(cursor)
      expect(phase.durationMs).toBeGreaterThan(0)
      cursor += phase.durationMs
    }
    expect(cursor).toBe(timeline.totalMs)
  })

  it('defeat runs 7.2s and only allows skip from 2s in', () => {
    const timeline = cinematicTimeline('defeat')
    expect(timeline.totalMs).toBe(7200)
    expect(timeline.skipAfterMs).toBe(2000)
  })

  it('victory runs 7.6s and only allows skip from 2s in', () => {
    const timeline = cinematicTimeline('victory')
    expect(timeline.totalMs).toBe(7600)
    expect(timeline.skipAfterMs).toBe(2000)
  })

  it('round flashes run 1.5s and cannot be skipped early', () => {
    for (const kind of ['roundWin', 'roundLose'] as const) {
      const timeline = cinematicTimeline(kind)
      expect(timeline.totalMs).toBe(1500)
      expect(timeline.skipAfterMs).toBe(timeline.totalMs)
    }
  })

  it('names every defeat phase in the pacted beat order', () => {
    expect(cinematicTimeline('defeat').phases.map((p) => p.name)).toEqual([
      'impact',
      'fracture',
      'ink',
      'verdict',
      'silence',
      'epilogue',
    ])
  })

  it('names every victory phase in the pacted beat order', () => {
    expect(cinematicTimeline('victory').phases.map((p) => p.name)).toEqual([
      'breath',
      'dawn',
      'coronation',
      'memory',
      'seal',
      'epilogue',
    ])
  })
})
