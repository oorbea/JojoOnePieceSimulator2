import { Check, Copy, Globe, Lock, Share2 } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { formatCode } from '@/features/game/lib/game-code'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlossOverlay } from '@/shared/components/presentational/gloss-overlay'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  code: string
  isPublic: boolean
  onCopy: () => Promise<'copied' | 'shared' | 'failed'>
  onShare: () => Promise<'copied' | 'shared' | 'failed'>
}

export function JoinCodeCard({ code, isPublic, onCopy, onShare }: Props) {
  const { t } = useTranslation()
  const [justCopied, setJustCopied] = useState(false)

  const handleCopy = async () => {
    const result = await onCopy()
    if (result !== 'failed') {
      setJustCopied(true)
      setTimeout(() => setJustCopied(false), 1500)
    }
  }

  return (
    <GlassPanel glossy elevate={2} width="100%" p="$5" gap="$4" $md={{ flexDirection: 'row', justify: 'space-between' }}>
      <YStack gap="$2" items="center" $md={{ items: 'flex-start' }}>
        <GlowText level="label">{t('game.code.title')}</GlowText>
        <GlowText level="hero" letterSpacing={4}>
          {formatCode(code)}
        </GlowText>
        <XStack items="center" gap="$1.5">
          {isPublic ? <Globe size={14} color="$panelTextSoft" /> : <Lock size={14} color="$panelTextSoft" />}
          <GlowText level="label">{isPublic ? t('game.code.publicHint') : t('game.code.privateHint')}</GlowText>
        </XStack>
      </YStack>

      <XStack gap="$2" items="center">
        <GlossButton
          tone={justCopied ? 'green' : 'glass'}
          btnSize="md"
          shape="circle"
          onPress={handleCopy}
          accessibilityLabel={t('game.code.copy')}
        >
          {justCopied ? <Check size={20} color="white" /> : <Copy size={20} color="$panelText" />}
        </GlossButton>
        <GlossButton tone="glass" btnSize="md" shape="circle" onPress={onShare} accessibilityLabel={t('game.code.share')}>
          <Share2 size={20} color="$panelText" />
        </GlossButton>
      </XStack>
      <GlossOverlay coverage="third" shape="card" />
    </GlassPanel>
  )
}
