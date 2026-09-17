import {
  Check,
  Link2,
  Mail,
  MessageCircle,
  QrCode,
  Send,
  X as XIcon,
} from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Linking } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { InviteQr } from '@/features/game/components/presentational/invite-qr'
import { DetailModal } from '@/shared/components/presentational/detail-modal'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  visible: boolean
  onClose: () => void
  url: string
  message: string
  onCopy: () => Promise<boolean>
}

// Desktop (and any web browser without navigator.share) fallback for the
// invite share button - reuses DetailModal's chrome (title/close/scroll)
// instead of inventing another overlay primitive. Every row is a
// full-width GlossButton, which already carries a tooltip from its
// accessibilityLabel per the project's tooltip norm.
export function ShareInviteSheet({ visible, onClose, url, message, onCopy }: Props) {
  const { t } = useTranslation()
  const [justCopied, setJustCopied] = useState(false)
  const [showQr, setShowQr] = useState(false)

  const handleCopy = async () => {
    const ok = await onCopy()
    if (ok) {
      setJustCopied(true)
      setTimeout(() => setJustCopied(false), 1500)
    }
  }

  const encodedUrl = encodeURIComponent(url)
  const encodedMessage = encodeURIComponent(message)

  const rows: { key: string; label: string; icon: React.ReactNode; onPress: () => void }[] = [
    {
      key: 'whatsapp',
      label: t('game.invite.share.whatsapp'),
      icon: <MessageCircle size={20} color="$panelText" />,
      onPress: () => void Linking.openURL(`https://wa.me/?text=${encodedMessage}%20${encodedUrl}`),
    },
    {
      key: 'telegram',
      label: t('game.invite.share.telegram'),
      icon: <Send size={20} color="$panelText" />,
      onPress: () =>
        void Linking.openURL(`https://t.me/share/url?url=${encodedUrl}&text=${encodedMessage}`),
    },
    {
      key: 'x',
      label: t('game.invite.share.x'),
      icon: <XIcon size={20} color="$panelText" />,
      onPress: () =>
        void Linking.openURL(`https://x.com/intent/post?url=${encodedUrl}&text=${encodedMessage}`),
    },
    {
      key: 'email',
      label: t('game.invite.share.email'),
      icon: <Mail size={20} color="$panelText" />,
      onPress: () =>
        void Linking.openURL(
          `mailto:?subject=${encodedMessage}&body=${encodedMessage}%20${encodedUrl}`
        ),
    },
  ]

  return (
    <DetailModal
      visible={visible}
      title={t('game.invite.share.title')}
      onClose={onClose}
      closeA11y={t('common.cancel')}
    >
      <YStack gap="$2">
        <GlossButton
          tone={justCopied ? 'green' : 'glass'}
          btnSize="md"
          width="100%"
          onPress={handleCopy}
          accessibilityLabel={t('game.invite.share.copyLink')}
        >
          <XStack items="center" gap="$2" flex={1}>
            {justCopied ? (
              <Check size={20} color="white" />
            ) : (
              <Link2 size={20} color="$panelText" />
            )}
            <GlowText level="label" color={justCopied ? 'white' : '$panelText'}>
              {justCopied ? t('game.code.linkCopied') : t('game.invite.share.copyLink')}
            </GlowText>
          </XStack>
        </GlossButton>

        <GlossButton
          tone="glass"
          btnSize="md"
          width="100%"
          onPress={() => setShowQr((v) => !v)}
          accessibilityLabel={t('game.invite.share.qr')}
        >
          <XStack items="center" gap="$2" flex={1}>
            <QrCode size={20} color="$panelText" />
            <GlowText level="label">{t('game.invite.share.qr')}</GlowText>
          </XStack>
        </GlossButton>
        {showQr ? (
          <YStack items="center" gap="$2">
            <InviteQr value={url} />
            <GlowText level="label" align="center">
              {t('game.invite.share.qrHint')}
            </GlowText>
          </YStack>
        ) : null}

        {rows.map((row) => (
          <GlossButton
            key={row.key}
            tone="glass"
            btnSize="md"
            width="100%"
            onPress={row.onPress}
            accessibilityLabel={row.label}
          >
            <XStack items="center" gap="$2" flex={1}>
              {row.icon}
              <GlowText level="label">{row.label}</GlowText>
            </XStack>
          </GlossButton>
        ))}
      </YStack>
    </DetailModal>
  )
}
