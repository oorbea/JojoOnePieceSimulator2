export type HotkeyAction = { kind: 'vote'; index: number } | { kind: 'skip' }

export type ResolveHotkeyInput = {
  key: string
  /** True while a text input/textarea/select/contenteditable has focus -
   * digits and 's' must never hijack normal typing. */
  isTextInputFocused: boolean
  /** True while a modal/overlay that should own all keyboard input is open
   * (ConfirmSheet, LoadoutModal, ...) - the caller decides what counts as
   * blocking, this function only respects the flag. */
  blocked: boolean
  /** True when a vote is actually open right now (VOTING or TIEBREAK) -
   * without this, '1'/'2' would silently do nothing anyway, but resolving
   * them to an action the caller can't act on is worse than resolving to
   * null explicitly. */
  votingOpen: boolean
  /** Also required for a digit to resolve - out-of-range digits (more keys
   * than options, e.g. '9' with 2 options) resolve to null rather than an
   * out-of-bounds index. */
  optionCount: number
  /** True while the sorteo overlay is actually showing - 's' only means
   * anything during the reveal, there's nothing to skip once voting has
   * opened. */
  revealing: boolean
  /** True while the loadout-summary screen (2026-08-30) is showing - 's'
   * also skips that, same key as the sorteo's own skip since the two never
   * overlap. */
  summaryOpen: boolean
}

// resolveHotkey is the single source of truth for every match-screen single-
// key shortcut, kept as a pure function so every guard is unit-testable
// without a DOM. The caller (use-match-hotkeys.ts) is a thin
// addEventListener wrapper around this - all judgment calls live here.
export function resolveHotkey(input: ResolveHotkeyInput): HotkeyAction | null {
  if (input.isTextInputFocused || input.blocked) return null

  const key = input.key.toLowerCase()

  if (key === 's') return input.revealing || input.summaryOpen ? { kind: 'skip' } : null

  if (!input.votingOpen) return null
  if (!/^[1-9]$/.test(key)) return null
  const index = Number(key) - 1
  if (index >= input.optionCount) return null
  return { kind: 'vote', index }
}
