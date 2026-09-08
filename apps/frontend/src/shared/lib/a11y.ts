import { AccessibilityInfo, findNodeHandle, Platform, type AccessibilityRole, type View } from 'react-native'

// RN's accessibilityLabel/accessibilityRole pass straight through to the
// DOM on web (Tamagui doesn't translate them), which React logs as unknown
// HTML attributes. Route through this on every pressable so native keeps
// its accessibility props and web gets aria-label/role instead. `role` is
// cast at the call site — Tamagui's web `Role` type isn't exported, but its
// values are the same ARIA role strings RN's AccessibilityRole already uses.
const ARIA_ROLE: Partial<Record<AccessibilityRole, string>> = {
  none: 'presentation',
}

type A11yState = { disabled?: boolean; checked?: boolean; selected?: boolean }

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
      ...(state?.checked !== undefined ? { 'aria-checked': state.checked } : null),
      ...(state?.selected !== undefined ? { 'aria-selected': state.selected } : null),
    }
  }
  return {
    ...(label ? { accessibilityLabel: label } : null),
    ...(role ? { accessibilityRole: role } : null),
    ...(state ? { accessibilityState: state } : null),
  }
}

// Moves focus to a card/element after a UI-driven change (e.g. "Cargar más"
// appending a page) - a single cross-platform entry point instead of every
// call site re-deriving the web/native branch itself.
//
// Web: react-native-web's Pressable forwards its ref to a real DOM node, so
// a plain `.focus()` works.
//
// Native: RN's View has no DOM-style `.focus()` - moving accessibility
// focus needs `AccessibilityInfo.setAccessibilityFocus` with the element's
// native node handle (`findNodeHandle`). A ref that hasn't mounted yet, or
// one belonging to a host component `findNodeHandle` can't resolve, yields
// `null` - silently skip rather than throw, since this only ever fires
// best-effort after a data change, never on a user's direct action.
export function focusElement(el: View | null): void {
  if (!el) return
  if (Platform.OS === 'web') {
    const maybeFocusable = el as unknown as { focus?: () => void }
    if (typeof maybeFocusable.focus === 'function') maybeFocusable.focus()
    return
  }
  const tag = findNodeHandle(el)
  if (tag != null) AccessibilityInfo.setAccessibilityFocus(tag)
}
