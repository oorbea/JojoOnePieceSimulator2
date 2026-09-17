import { Platform } from 'react-native'

import { env } from '@/shared/config/env'

// The one place that knows the shape of a lobby invite link. Today it's
// always https://<origin>/join/<token> (Opción A in the design doc) - a
// plain web URL that opens the browser on every platform. Swapping in a
// universal/app link later (once a native build with associatedDomains/
// intentFilters exists) means editing this file only: neither the backend
// nor the /join/[token] route needs to change.
const INVITE_PATH = '/join'

// A bare "http://something" is never acceptable to share (a plain HTTP link
// pasted into WhatsApp is a phishing look-alike waiting to happen) except
// for local dev against localhost, where there's no TLS to have.
function isAcceptableOrigin(origin: string): boolean {
  try {
    const url = new URL(origin)
    if (url.protocol === 'https:') return true
    return (
      url.protocol === 'http:' && (url.hostname === 'localhost' || url.hostname === '127.0.0.1')
    )
  } catch {
    return false
  }
}

// inviteOrigin resolves the origin to build invite links under:
// EXPO_PUBLIC_WEB_ORIGIN if configured, else window.location.origin on web,
// else null (native with nothing configured - see shareInviteLink's
// fallback). Returns null instead of a rejected origin so a
// misconfigured/non-https EXPO_PUBLIC_WEB_ORIGIN fails closed rather than
// building a link nobody should share.
export function inviteOrigin(): string | null {
  const configured = env.EXPO_PUBLIC_WEB_ORIGIN
  if (configured) return isAcceptableOrigin(configured) ? configured : null
  if (Platform.OS === 'web' && typeof window !== 'undefined') {
    return isAcceptableOrigin(window.location.origin) ? window.location.origin : null
  }
  return null
}

// buildInviteUrl builds the shareable join link for token, or null when no
// origin is available (native app with EXPO_PUBLIC_WEB_ORIGIN unset) -
// callers fall back to sharing the raw join code in that case.
export function buildInviteUrl(token: string): string | null {
  const origin = inviteOrigin()
  if (!origin) return null
  return `${origin}${INVITE_PATH}/${encodeURIComponent(token)}`
}
