import { LinearGradient } from '@tamagui/linear-gradient'
import { Image } from 'react-native'
import { Button, H1, Paragraph, Spinner, YStack } from 'tamagui'

import type { AppError } from '@/shared/api/errors'

type Props = {
  onSignIn: () => void
  isLoading: boolean
  isReady: boolean
  error: AppError | null
}

// Pure UI — Frutiger Aero: soft gradient backdrop, glossy translucent card,
// rounded everything. No data/session logic lives here.
export function LoginScreen({ onSignIn, isLoading, isReady, error }: Props) {
  return (
    <LinearGradient flex={1} colors={['#bfe9ff', '#e8f6ff', '#f3e8ff']} start={[0, 0]} end={[1, 1]}>
      <YStack flex={1} items="center" justify="center" p="$5">
        <YStack
          maxW={440}
          width="100%"
          gap="$4"
          p="$6"
          rounded="$10"
          items="center"
          bg="rgba(255,255,255,0.55)"
          borderWidth={1}
          borderColor="rgba(255,255,255,0.5)"
          shadowColor="$shadowColor"
          shadowRadius={30}
          shadowOpacity={0.2}
        >
          <Image
            source={require('../../../../../assets/images/logo.png')}
            style={{ width: 120, height: 120, borderRadius: 24 }}
            resizeMode="contain"
          />

          <H1 text="center" color="$standPurple">
            JoJo x One Piece Simulator
          </H1>
          <Paragraph theme="alt2" text="center">
            Sign in with your Google account to continue.
          </Paragraph>

          <Button
            disabled={!isReady || isLoading}
            opacity={!isReady || isLoading ? 0.6 : 1}
            bg="$strawHatRed"
            color="white"
            rounded="$10"
            size="$5"
            width="100%"
            icon={isLoading ? () => <Spinner color="white" /> : undefined}
            onPress={onSignIn}
          >
            {isLoading ? 'Signing in…' : 'Sign in with Google'}
          </Button>

          {error ? (
            <Paragraph color="$red10" text="center">
              {error.message}
            </Paragraph>
          ) : null}
        </YStack>
      </YStack>
    </LinearGradient>
  )
}
