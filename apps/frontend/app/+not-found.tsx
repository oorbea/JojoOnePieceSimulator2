import { Link } from 'expo-router'
import { Button, H2, YStack } from 'tamagui'

export default function NotFoundScreen() {
  return (
    <YStack flex={1} items="center" justify="center" gap="$3" p="$4">
      <H2>This screen doesn&apos;t exist.</H2>
      <Link href="/" asChild>
        <Button>Go to home</Button>
      </Link>
    </YStack>
  )
}
