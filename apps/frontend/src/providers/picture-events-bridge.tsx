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
  const queryClient = useQueryClient()

  const sourceRef = useRef<EventSource | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const hasConnectedBeforeRef = useRef(false)

  useEffect(() => {
    if (Platform.OS !== 'web') return

    const disconnect = () => {
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }
      sourceRef.current?.close()
      sourceRef.current = null
    }

    if (!accessToken) {
      disconnect()
      hasConnectedBeforeRef.current = false
      reconnectAttemptsRef.current = 0
      return
    }

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

    // The token used to authenticate this connection is read fresh here on
    // every (re)connect - not just once - since EventSource has no way to
    // update credentials on an already-open connection, and the browser's
    // own auto-retry (which we disable below) would otherwise keep hammering
    // a URL with a token that may have rotated (re-login) while disconnected.
    const connect = () => {
      const token = useSessionStore.getState().session?.accessToken
      if (!token) return

      const source = new EventSource(`${env.EXPO_PUBLIC_API_URL}/events?token=${encodeURIComponent(token)}`)
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

        const attempt = reconnectAttemptsRef.current
        const delay = Math.min(BASE_RECONNECT_MS * 2 ** attempt, MAX_RECONNECT_MS)
        reconnectAttemptsRef.current += 1
        reconnectTimerRef.current = setTimeout(connect, delay)
      }
    }

    connect()
    return disconnect
  }, [accessToken, queryClient])

  return null
}
