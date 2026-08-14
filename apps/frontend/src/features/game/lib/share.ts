import { Platform, Share } from 'react-native'

// No new dependency added: web copies via the browser's built-in Clipboard
// API, native/web sharing both go through RN core's Share (which on web
// falls back to the OS share sheet where available, or throws - caught
// below). Returns which path actually happened so the caller can toast
// accordingly.
export async function shareJoinCode(code: string, message: string): Promise<'copied' | 'shared' | 'failed'> {
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
