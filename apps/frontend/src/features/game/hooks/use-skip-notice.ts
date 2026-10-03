import { useEffect, useRef } from 'react'

import i18n from '@/shared/i18n'
import { serverNow } from '@/shared/lib/server-clock'
import { showSuccessToast } from '@/shared/lib/toast'

// A phase that ends this much earlier than its deadline was cut short by a
// majority skip vote, not by its own timer.
const CUT_SHORT_MARGIN_MS = 1500

type Params = {
  state: string | undefined
  revealEndsAt: number | null
  summaryEndsAt: number | null
  connectedHumans: number
}

/** 'reveal' / 'summary' when the previous state ended well before its own
 * deadline, otherwise null. Pure so the rule is testable without React. */
export function skippedPhase(
  prevState: string | undefined,
  nextState: string | undefined,
  prevEndsAt: number | null,
  now: number
): 'reveal' | 'summary' | null {
  if (prevState === nextState || prevEndsAt === null) return null
  if (now >= prevEndsAt - CUT_SHORT_MARGIN_MS) return null
  if (prevState === 'ASSIGNING' && (nextState === 'SUMMARY' || nextState === 'VOTING')) {
    return 'reveal'
  }
  if (prevState === 'SUMMARY' && nextState === 'VOTING') return 'summary'
  return null
}

// Tells everyone when the room skipped a phase by majority, so a sorteo or
// summary that vanishes early doesn't read as a glitch. Solo games stay
// silent - the only human skipping already knows.
export function useSkipNotice({ state, revealEndsAt, summaryEndsAt, connectedHumans }: Params) {
  const prev = useRef<{ state: string | undefined; endsAt: number | null }>({
    state,
    endsAt: null,
  })

  useEffect(() => {
    const phase = connectedHumans > 1 ? skippedPhase(prev.current.state, state, prev.current.endsAt, serverNow()) : null
    if (phase) showSuccessToast(i18n.t(`game.match.skipMajority.${phase}`))
    prev.current = {
      state,
      endsAt: state === 'ASSIGNING' ? revealEndsAt : state === 'SUMMARY' ? summaryEndsAt : null,
    }
  }, [state, revealEndsAt, summaryEndsAt, connectedHumans])
}
