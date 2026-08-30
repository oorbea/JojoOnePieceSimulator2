import { useEffect } from 'react'

import { resolveHotkey } from '@/features/game/lib/hotkeys'
import { isWeb } from '@/shared/lib/web-blur'

type UseMatchHotkeysInput = {
  votingOpen: boolean
  optionCount: number
  revealing: boolean
  /** True while the loadout-summary screen (2026-08-30) is showing. */
  summaryOpen: boolean
  /** True while any overlay that should own all keyboard input is showing
   * (ConfirmSheet, LoadoutModal) - hotkeys stay fully suppressed. */
  blocked: boolean
  onVote: (index: number) => void
  onSkipReveal: () => void
  onSkipSummary: () => void
}

const TEXT_INPUT_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT'])

function isTextInputFocused(): boolean {
  if (typeof document === 'undefined') return false
  const el = document.activeElement as HTMLElement | null
  if (!el) return false
  if (TEXT_INPUT_TAGS.has(el.tagName)) return true
  return el.isContentEditable
}

// useMatchHotkeys is a thin addEventListener wrapper around the pure
// lib/hotkeys.ts's resolveHotkey - every judgment call (which keys, which
// guards) lives there so it's unit-testable without a DOM; this hook only
// wires it to the real keydown event and the real focused-element check.
// Web-only: native has no global keydown event to listen for, so the
// effect no-ops there (the isWeb guard mirrors use-roving-group.ts's own).
export function useMatchHotkeys({
  votingOpen,
  optionCount,
  revealing,
  summaryOpen,
  blocked,
  onVote,
  onSkipReveal,
  onSkipSummary,
}: UseMatchHotkeysInput) {
  useEffect(() => {
    if (!isWeb || typeof window === 'undefined') return

    const handler = (e: KeyboardEvent) => {
      if (e.altKey || e.ctrlKey || e.metaKey) return
      const action = resolveHotkey({
        key: e.key,
        isTextInputFocused: isTextInputFocused(),
        blocked,
        votingOpen,
        optionCount,
        revealing,
        summaryOpen,
      })
      if (!action) return
      e.preventDefault()
      if (action.kind === 'vote') onVote(action.index)
      else if (revealing) onSkipReveal()
      else onSkipSummary()
    }

    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [votingOpen, optionCount, revealing, summaryOpen, blocked, onVote, onSkipReveal, onSkipSummary])
}
