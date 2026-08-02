import { Button, H2, Paragraph, YStack } from 'tamagui'

import type { AppError } from '@/shared/api/errors'

// Pure UI — never renders error.stack, only the user-friendly message.
export function ErrorFallback({ error, onRetry }: { error: AppError; onRetry: () => void }) {
  return (
    <YStack flex={1} items="center" justify="center" gap="$4" p="$5" bg="$background">
      <YStack
        maxW={440}
        width="100%"
        gap="$4"
        p="$6"
        rounded="$8"
        items="center"
        bg="rgba(255,255,255,0.6)"
        borderWidth={1}
        borderColor="rgba(255,255,255,0.4)"
        shadowColor="$shadowColor"
        shadowRadius={20}
        shadowOpacity={0.15}
      >
        <Paragraph fontSize="$9">⚠️</Paragraph>
        <H2 text="center">Something went wrong</H2>
        <Paragraph theme="alt2" text="center">
          {error.message || 'An unexpected error occurred. Please try again.'}
        </Paragraph>
        <Button bg="$strawHatRed" color="white" rounded="$10" onPress={onRetry}>
          Try Again
        </Button>
      </YStack>
    </YStack>
  )
}
