import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { Platform } from 'react-native'

import { devilFruitKeys } from '@/features/devil-fruits/api/devil-fruits.keys'
import { profileKeys } from '@/features/profile/api/profile.keys'
import { stageKeys } from '@/features/stages/api/stages.keys'
import { standKeys } from '@/features/stands/api/stands.keys'
import { clearEtags } from '@/shared/api/etag'
import { env } from '@/shared/config/env'
import { useSessionStore } from '@/shared/stores/session.store'
import { mintEventsTicket } from '@/shared/api/stream-tickets'
import { toAppError } from '@/shared/api/errors'

// Backs off the same shape as the polling this replaces (2s -> 4s -> 8s ...
// capped at 30s), but never gives up permanently - this is now the only
// notification path for picture-pipeline completion on web.
const BASE_RECONNECT_MS = 2_000
const MAX_RECONNECT_MS = 30_000

type PictureEventDTO = {
  kind: 'STAND' | 'DEVIL_FRUIT' | 'USER' | 'STAGE'
  subjectId: string
  status: 'NONE' | 'PENDING' | 'READY' | 'FAILED'
}

// Renderless. Mounted once, app-wide (see app-providers.tsx) - opens an
// EventSource to the backend's SSE stream (events_endpoints.go) whenever a
// session exists, and routes each picture-pipeline event into a query
// invalidation, replacing the polling in use-stands.ts/use-devil-fruits.ts/
// use-profile.ts on web.
//
// React Native has no built-in EventSource, so this is web-only; native
// keeps polling as its fallback (see those hooks' Platform.OS gate).
export function PictureEventsBridge() {
  const accessToken = useSessionStore((state) => state.session?.accessToken ?? null)
  // The stream is admin-only (events_endpoints.go's stream still re-checks
  // this itself). Every non-admin session used to open an EventSource
  // anyway, eat a 403 from the mint, and retry it forever on this same
  // backoff curve - gating the mount here skips even that first attempt.
  const isAdmin = useSessionStore((state) => state.session?.user.role === 'ADMIN')
  const queryClient = useQueryClient()

  const sourceRef = useRef<EventSource | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const hasConnectedBeforeRef = useRef(false)
  // Guards the mint-ticket await connect() makes below: bumped by
  // disconnect() so a mint that resolves after the effect has torn down
  // (session cleared, role changed, unmount) can't open a stale EventSource.
  const cancelledRef = useRef(false)

  useEffect(() => {
    if (Platform.OS !== 'web') return

    const disconnect = () => {
      cancelledRef.current = true
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }
      sourceRef.current?.close()
      sourceRef.current = null
    }

    if (!accessToken || !isAdmin) {
      disconnect()
      hasConnectedBeforeRef.current = false
      reconnectAttemptsRef.current = 0
      return
    }
    cancelledRef.current = false

    const invalidateAll = () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: standKeys.allLocales })
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      void queryClient.invalidateQueries({ queryKey: stageKeys.allLocales })
      void queryClient.invalidateQueries({ queryKey: profileKeys.me })
    }

    // A switch with an explicit branch per kind, not an if/else-as-default -
    // the previous else treated any kind other than STAND/DEVIL_FRUIT as a
    // profile event, so a STAGE event (the backend has emitted kind:"STAGE"
    // since game-stage-content.md landed) was silently invalidating
    // profileKeys.me instead of the stage catalogue.
    const handlePictureEvent = (event: MessageEvent) => {
      const evt = JSON.parse(event.data) as PictureEventDTO
      clearEtags()
      switch (evt.kind) {
        case 'STAND':
          void queryClient.invalidateQueries({ queryKey: standKeys.allLocales })
          break
        case 'DEVIL_FRUIT':
          void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
          break
        case 'STAGE':
          void queryClient.invalidateQueries({ queryKey: stageKeys.allLocales })
          break
        case 'USER':
          void queryClient.invalidateQueries({ queryKey: profileKeys.me })
          break
      }
    }

    const scheduleRetry = () => {
      const attempt = reconnectAttemptsRef.current
      const delay = Math.min(BASE_RECONNECT_MS * 2 ** attempt, MAX_RECONNECT_MS)
      reconnectAttemptsRef.current += 1
      reconnectTimerRef.current = setTimeout(connect, delay)
    }

    // A fresh ticket is minted on every (re)connect - not just once - since
    // EventSource has no way to update credentials on an already-open
    // connection, and the browser's own auto-retry (which we disable below)
    // would otherwise keep hammering a URL whose ticket is already burned or
    // expired.
    async function connect() {
      if (!useSessionStore.getState().session?.accessToken) return

      let ticket: string
      try {
        ticket = await mintEventsTicket()
      } catch (err) {
        if (cancelledRef.current) return
        const status = toAppError(err).status
        // 401: the response interceptor already tried one silent
        // refresh-and-retry (shared/api/interceptors.ts) before this error
        // ever reached here - a 401 surfacing at this layer means that
        // refresh attempt itself failed, so the interceptor has already
        // cleared the session: the effect re-runs with accessToken null and
        // disconnects on its own.
        // 403: this session isn't (or is no longer) an admin - the stream
        // has nothing for it, so stop instead of re-minting forever, unlike
        // the old ?token= loop against a 403 stream.
        if (status === 401 || status === 403) return
        scheduleRetry()
        return
      }
      if (cancelledRef.current) return

      const source = new EventSource(`${env.EXPO_PUBLIC_API_URL}/events?ticket=${encodeURIComponent(ticket)}`)
      sourceRef.current = source

      source.addEventListener('picture', handlePictureEvent)

      source.onopen = () => {
        reconnectAttemptsRef.current = 0
        // A reconnect (not the very first connect) means events may have
        // been missed while disconnected - the stream itself has no replay
        // log, so a full resync is the only way to guarantee nothing was
        // left stuck at PENDING.
        if (hasConnectedBeforeRef.current) invalidateAll()
        hasConnectedBeforeRef.current = true
      }

      source.onerror = () => {
        source.close()
        if (sourceRef.current !== source) return // already superseded
        sourceRef.current = null
        scheduleRetry()
      }
    }

    connect()
    return disconnect
  }, [accessToken, isAdmin, queryClient])

  return null
}
