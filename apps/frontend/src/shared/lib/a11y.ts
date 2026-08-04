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

export function a11yProps(label?: string, role?: AccessibilityRole) {
  if (Platform.OS === 'web') {
    return {
      ...(label ? { 'aria-label': label } : null),
      ...(role ? { role: (ARIA_ROLE[role] ?? role) as never } : null),
    }
  }
  return {
    ...(label ? { accessibilityLabel: label } : null),
    ...(role ? { accessibilityRole: role } : null),
  }
}
