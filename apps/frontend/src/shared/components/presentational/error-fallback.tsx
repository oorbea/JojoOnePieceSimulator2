import { AlertCircle } from '@tamagui/lucide-icons-2'
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
export function ErrorFallback({ error, onRetry }: { error: AppError; onRetry: () => void }) {
  return (
    <PageShell align="center" plain>
      <WiiCard tone="glass" aspect="square" width={80} items="center" justify="center">
        <AlertCircle size={40} color="$strawHatRedDeep" strokeWidth={2} />
      </WiiCard>
      <SpeechBubble tailSide="bottom" tone="strong" width="100%" maxW={440}>
        <YStack gap="$2">
          <GlowText level="heading" align="center">
            Something broke.
          </GlowText>
          <GlowText level="label" align="center">
            {error.message || 'Try again, or come back in a bit.'}
          </GlowText>
        </YStack>
      </SpeechBubble>
      <GlossButton tone="yellow" btnSize="md" onPress={onRetry} accessibilityLabel="Try again">
        Try Again
      </GlossButton>
    </PageShell>
  )
}
