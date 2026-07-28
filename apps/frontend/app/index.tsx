import { H1, Paragraph, YStack } from 'tamagui'

// Placeholder landing screen. Routes stay thin shims — real screens will
// render a container component from src/features/<feature>/components/containers.
export default function HomeScreen() {
  return (
    <YStack flex={1} items="center" justify="center" gap="$3" p="$4">
      <H1>JoJo x One Piece Simulator</H1>
      <Paragraph theme="alt2">Frontend scaffold is up.</Paragraph>
    </YStack>
  )
}
