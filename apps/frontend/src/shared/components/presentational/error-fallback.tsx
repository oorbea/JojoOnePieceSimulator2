import { AlertCircle } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { YStack } from 'tamagui'

import type { AppError } from '@/shared/api/errors'

import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'
import { PageShell } from './page-shell'
import { SpeechBubble } from './speech-bubble'
import { WiiCard } from './wii-card'

// Pure UI — never renders error.stack, only the user-friendly message. This
// is the last resort inside ErrorBoundary, so `plain` skips the animated
// bubble field: a render error in the backdrop must never be able to loop.
// error.message comes from the backend (see shared/api/errors.ts) and is
// still English-only - translating it is a separate, not-yet-done piece
// (it needs the backend to emit error codes, not strings) - only the
// fallbackMessage shown when there's no message at all is translated here.
export function ErrorFallback({ error, onRetry }: { error: AppError; onRetry: () => void }) {
  const { t } = useTranslation()
  return (
    <PageShell align="center" plain>
      <WiiCard tone="glass" aspect="square" width={80} items="center" justify="center">
        <AlertCircle size={40} color="$strawHatRedDeep" strokeWidth={2} />
      </WiiCard>
      <SpeechBubble tailSide="bottom" tone="strong" width="100%" maxW={440}>
        <YStack gap="$2">
          <GlowText level="heading" align="center">
            {t('errorFallback.title')}
          </GlowText>
          <GlowText level="label" align="center">
            {error.message || t('errorFallback.fallbackMessage')}
          </GlowText>
        </YStack>
      </SpeechBubble>
      <GlossButton
        tone="yellow"
        btnSize="md"
        onPress={onRetry}
        accessibilityLabel={t('errorFallback.retry')}
      >
        {t('errorFallback.retry')}
      </GlossButton>
    </PageShell>
  )
}
