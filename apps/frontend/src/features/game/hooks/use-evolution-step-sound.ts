import { useEffect } from 'react'

import type { EvolvePhase } from '@/features/game/components/presentational/match/power-reveal-card'
import { useSound } from '@/shared/hooks/use-sound'

// Owner-provided clip (2026-09-27, not synthesized like reel-spin/reel-land)
// for the evolution reveal's landing beat. Plays once each time evolvePhase
// transitions into 'step' or 'final' - the same instant PowerRevealCard
// fires its white flash - so every intermediate stage AND the final form
// gets its own hit, not just the last one. Not gated on reducedMotion: the
// flash/shake FX collapse under reduced motion, but the reveal itself (a new
// stage landing) still happens and still deserves an audio cue.
export function useEvolutionStepSound(evolvePhase: EvolvePhase | undefined, enabled: boolean): void {
  const step = useSound(require('../../../../assets/audio/evolution-step.wav'), enabled)

  useEffect(() => {
    if (!enabled) return
    if (evolvePhase === 'step' || evolvePhase === 'final') {
      step.play()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- step is a freshly built stable SharedObject each render; only evolvePhase/enabled should retrigger this
  }, [evolvePhase, enabled])
}
