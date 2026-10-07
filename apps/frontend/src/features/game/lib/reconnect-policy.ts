import type { SocketStatus } from '@/features/game/stores/game-socket.store'

// What the room does when the device signals it may have been cut off from
// the server. Pure so the policy is testable apart from the hook/store.

// App back in the foreground (tab visible, PWA reopened). A closed socket is
// retried immediately instead of waiting out the backoff. A socket that still
// reads "open" may have missed frames while the page was hidden - mobile
// browsers throttle or freeze background tabs - so ask the server for a fresh
// snapshot (the same RESYNC a reconnect sends).
export function foregroundAction(status: SocketStatus): 'retry' | 'resync' {
  return status === 'open' ? 'resync' : 'retry'
}

// Network came back. Only a socket that is not open needs a nudge; an open one
// either never noticed the outage or recovers through its own close event.
export function onlineAction(status: SocketStatus): 'retry' | 'none' {
  return status === 'open' ? 'none' : 'retry'
}
