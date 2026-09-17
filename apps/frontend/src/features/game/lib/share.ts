import { Platform, Share } from 'react-native'

// No new dependency added: web copies via the browser's built-in Clipboard
// API, native/web sharing both go through RN core's Share (which on web
// falls back to the OS share sheet where available, or throws - caught
// below). Returns which path actually happened so the caller can toast
// accordingly.
export async function shareJoinCode(
  code: string,
  message: string
): Promise<'copied' | 'shared' | 'failed'> {
  if (Platform.OS === 'web') {
    try {
      await navigator.clipboard.writeText(code)
      return 'copied'
    } catch {
      return 'failed'
    }
  }
  try {
    await Share.share({ message })
    return 'shared'
  } catch {
    return 'failed'
  }
}

export type ShareResult = 'shared' | 'copied' | 'unsupported' | 'cancelled' | 'failed'

// canSystemShare reports whether the current platform has an OS-level share
// sheet available - capability detection only, never a device/UA check, so
// the same code path serves phone, tablet and desktop alike. Native always
// has RN's Share module; on web it depends on the browser (mobile Safari/
// Chrome and desktop Chrome/Edge do, Firefox/desktop Safari don't).
export function canSystemShare(): boolean {
  if (Platform.OS !== 'web') return true
  return typeof navigator !== 'undefined' && typeof navigator.share === 'function'
}

// shareInviteLink opens the OS share sheet for url. Must be called directly
// inside a user gesture's onPress handler (never behind an intervening
// await) - both native's Share.share and web's navigator.share require an
// active user-activation and silently/loudly reject otherwise.
export async function shareInviteLink(url: string, message: string): Promise<ShareResult> {
  if (Platform.OS === 'web') {
    if (!canSystemShare()) return 'unsupported'
    try {
      await navigator.share({ title: message, text: message, url })
      return 'shared'
    } catch (err) {
      // The user closing the native share sheet without picking anything
      // throws AbortError - that's a cancel, not a failure, and must not
      // surface an error toast.
      if (err instanceof Error && err.name === 'AbortError') return 'cancelled'
      return 'failed'
    }
  }
  try {
    const result = await Share.share({ message: `${message} ${url}` })
    // RN's Share.share resolves with action 'dismissedAction' (iOS only)
    // when the sheet is dismissed without a pick - also a cancel, not a
    // failure.
    if (result.action === 'dismissedAction') return 'cancelled'
    return 'shared'
  } catch {
    return 'failed'
  }
}

// copyToClipboard backs the desktop share popover's "Copy link" row (and
// anywhere else that needs a plain string on the clipboard, web-only today
// - RN core has no clipboard API, and the popover only ever renders on
// web).
export async function copyToClipboard(text: string): Promise<boolean> {
  if (Platform.OS !== 'web') return false
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}
