import { LinearGradient } from '@tamagui/linear-gradient'
import { Image } from 'react-native'
import { Button, H2, Paragraph, YStack } from 'tamagui'

import type { SessionUser } from '@/shared/stores/session.store'

type Props = {
  user: SessionUser
  onLogout: () => void
}

// Pure UI post-login placeholder — verifies the auth flow round-trips
// correctly before any real feature screens are built.
export function HomeScreen({ user, onLogout }: Props) {
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
          {user.picture ? (
            <Image
              source={{ uri: user.picture }}
              style={{ width: 96, height: 96, borderRadius: 48 }}
            />
          ) : (
            <YStack
              width={96}
              height={96}
              rounded={48}
              items="center"
              justify="center"
              bg="$standPurple"
            >
              <Paragraph color="white" fontSize="$8">
                {user.completeName.charAt(0).toUpperCase()}
              </Paragraph>
            </YStack>
          )}

          <H2 text="center" color="$standPurple">
            Welcome, {user.completeName}
          </H2>
          <Paragraph theme="alt2" text="center">
            {user.email}
          </Paragraph>
          <Paragraph theme="alt2" text="center">
            Role: {user.role}
          </Paragraph>

          <Button bg="$strawHatRed" color="white" rounded="$10" width="100%" onPress={onLogout}>
            Log out
          </Button>
        </YStack>
      </YStack>
    </LinearGradient>
  )
}
