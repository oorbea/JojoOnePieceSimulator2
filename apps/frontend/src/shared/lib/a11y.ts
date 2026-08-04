import { Platform, type AccessibilityRole } from 'react-native'

// RN's accessibilityLabel/accessibilityRole pass straight through to the
// DOM on web (Tamagui doesn't translate them), which React logs as unknown
// HTML attributes. Route through this on every pressable so native keeps
// its accessibility props and web gets aria-label/role instead. `role` is
// cast at the call site — Tamagui's web `Role` type isn't exported, but its
// values are the same ARIA role strings RN's AccessibilityRole already uses.
const ARIA_ROLE: Partial<Record<AccessibilityRole, string>> = {
  none: 'presentation',
}

type A11yState = { disabled?: boolean }

// Add every other RN `accessibility*` prop here as it's needed — none of
// them are DOM attributes and Tamagui forwards whatever it doesn't
// recognize straight to the host element on web (see a11y-web-leak in the
// Obsidian vault for the full writeup).
export function a11yProps(label?: string, role?: AccessibilityRole, state?: A11yState) {
  if (Platform.OS === 'web') {
    return {
      ...(label ? { 'aria-label': label } : null),
      ...(role ? { role: (ARIA_ROLE[role] ?? role) as never } : null),
      ...(state?.disabled !== undefined ? { 'aria-disabled': state.disabled } : null),
    }
  }
  return {
    ...(label ? { accessibilityLabel: label } : null),
    ...(role ? { accessibilityRole: role } : null),
    ...(state ? { accessibilityState: state } : null),
  }
}
