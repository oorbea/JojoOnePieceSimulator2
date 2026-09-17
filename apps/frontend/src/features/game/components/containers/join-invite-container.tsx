import { Redirect, useLocalSearchParams, useRouter } from 'expo-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import {
  JoinInviteScreen,
  type JoinInviteStatus,
} from '@/features/game/components/presentational/join-invite-screen'
import { useInvitePreview } from '@/features/game/hooks/use-invite-preview'
import { useInviteStatus } from '@/features/game/hooks/use-invite-status'
import { useJoinGameByInvite } from '@/features/game/hooks/use-join-by-invite'
import { useLeaveGameById } from '@/features/game/hooks/use-leave-game-by-id'
import { useMyGame } from '@/features/game/hooks/use-my-game'
import { inviteErrorKey } from '@/features/game/lib/invite-error'
import { clearPendingInvite, setPendingInvite } from '@/features/game/lib/pending-invite'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import { ConfirmSheet } from '@/shared/components/presentational/confirm-sheet'
import { toAppError } from '@/shared/api/errors'
import { useSessionStore } from '@/shared/stores/session.store'

export function JoinInviteContainer() {
  const router = useRouter()
  const { t } = useTranslation()
  const { token } = useLocalSearchParams<{ token: string }>()
  const session = useSessionStore((state) => state.session)
  const isHydrated = useSessionStore((state) => state.isHydrated)
  const [switchConfirmed, setSwitchConfirmed] = useState(false)

  const status = useInviteStatus(token ?? '')
  // Only worth previewing once the public check already says VALID, and
  // only once signed in - a stranger's token gets nothing beyond
  // status.data.status (see InvitePreview's own auth requirement).
  const preview = useInvitePreview(token ?? '', !!session && status.data?.status === 'VALID')
  const mine = useMyGame(!!session)
  const join = useJoinGameByInvite()
  const leaveCurrent = useLeaveGameById()

  if (!isHydrated || status.isPending) return <LoadingScreen />

  if (status.data?.status === 'EXPIRED') {
    return (
      <JoinInviteScreen
        status="error"
        errorKey="game.invite.error.expired"
        onJoin={() => undefined}
        onBack={() => router.replace('/play' as never)}
      />
    )
  }

  if (!session) {
    if (token) setPendingInvite(token)
    return <Redirect href="/login" />
  }
  clearPendingInvite()

  if (preview.isError) {
    return (
      <JoinInviteScreen
        status="error"
        errorKey={inviteErrorKey(toAppError(preview.error))}
        onJoin={() => undefined}
        onBack={() => router.replace('/play' as never)}
      />
    )
  }

  // Already in the target lobby (a stale tab, a double-tap on the link) -
  // just resume it, no confirm needed.
  if (mine.data && preview.data && mine.data.game.id === preview.data.gameId) {
    router.replace(`/play/${preview.data.gameId}` as never)
    return <LoadingScreen />
  }

  // Seated somewhere else: confirm before pulling them out of it. Skipped
  // once the user has already confirmed once this mount (avoids asking
  // again while the leave+join mutation is in flight).
  if (mine.data && preview.data && !switchConfirmed) {
    return (
      <>
        <JoinInviteScreen
          status="preview"
          preview={preview.data}
          onJoin={() => undefined}
          onBack={() => router.replace('/play' as never)}
        />
        <ConfirmSheet
          visible
          title={t('game.invite.switch.title')}
          message={t('game.invite.switch.message')}
          confirmLabel={t('game.invite.switch.confirm')}
          tone="blue"
          onConfirm={() => {
            setSwitchConfirmed(true)
            leaveCurrent.mutate(mine.data!.game.id, {
              onSuccess: () => {
                join.mutate(token!, {
                  onSuccess: (data) => router.replace(`/play/${data.game.id}` as never),
                })
              },
            })
          }}
          onCancel={() => router.replace('/play' as never)}
        />
      </>
    )
  }

  const screenStatus: JoinInviteStatus =
    join.isPending || leaveCurrent.isPending ? 'joining' : preview.data ? 'preview' : 'checking'

  if (join.isError) {
    return (
      <JoinInviteScreen
        status="error"
        errorKey={inviteErrorKey(toAppError(join.error))}
        onJoin={() => undefined}
        onBack={() => router.replace('/play' as never)}
      />
    )
  }

  return (
    <JoinInviteScreen
      status={screenStatus}
      preview={preview.data}
      onJoin={() => {
        if (!token) return
        join.mutate(token, {
          onSuccess: (data) => router.replace(`/play/${data.game.id}` as never),
        })
      }}
      onBack={() => router.replace('/play' as never)}
    />
  )
}
