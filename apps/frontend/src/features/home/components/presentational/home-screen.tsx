import { Apple, Home as HomeIcon, Sparkles, Zap } from '@tamagui/lucide-icons-2'
import { Image } from 'react-native'
import { Paragraph, XStack, YStack } from 'tamagui'

import { ChannelTile } from '@/shared/components/presentational/channel-tile'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { GlossOverlay } from '@/shared/components/presentational/gloss-overlay'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { SpeechBubble } from '@/shared/components/presentational/speech-bubble'
import type { SessionUser } from '@/shared/stores/session.store'

type Props = {
  user: SessionUser
  onLogout: () => void
}

// The domains behind the empty channel slots aren't built yet — showing
// them locked and honest beats faking navigation to routes that don't
// exist (typedRoutes would refuse to compile a fake href anyway).
const CHANNELS = [
  { key: 'profile', label: 'Profile', tone: 'blue' as const, icon: HomeIcon, locked: false },
  { key: 'stands', label: 'Stands', tone: 'grape' as const, icon: Sparkles, locked: true },
  { key: 'fruits', label: 'Devil Fruits', tone: 'red' as const, icon: Apple, locked: true },
  { key: 'powers', label: 'Powers', tone: 'yellow' as const, icon: Zap, locked: true },
]

// Pure UI — now lives inside the authenticated app shell, so it doesn't
// carry its own gradient/nav; PageShell + AppShell already provide those.
export function HomeScreen({ user, onLogout }: Props) {
  const firstName = user.completeName.split(' ')[0] ?? user.completeName

  return (
    <PageShell align="top" navPadding scroll maxWidth={720}>
      <GlassPanel
        glossy
        elevate={2}
        width="100%"
        p="$6"
        gap="$5"
        $md={{ flexDirection: 'row', items: 'center' }}
      >
        <YStack items="center" gap="$3">
          <YStack width={96} height={96} rounded="$circle" overflow="hidden" position="relative">
            <InsetRing rounded="$circle" />
            <GlossOverlay coverage="third" shape="circle" />
            {user.picture ? (
              <Image source={{ uri: user.picture }} style={{ width: '100%', height: '100%' }} />
            ) : (
              <YStack flex={1} items="center" justify="center" bg="$grapeSoda">
                <Paragraph color="white" fontSize="$8" fontWeight="800">
                  {user.completeName.charAt(0).toUpperCase()}
                </Paragraph>
              </YStack>
            )}
          </YStack>
        </YStack>

        <YStack flex={1} gap="$3">
          <SpeechBubble tailSide="left">
            <GlowText level="heading">Ready when you are, {firstName}.</GlowText>
          </SpeechBubble>

          <XStack flexWrap="wrap" gap="$2" justify="center" $md={{ justify: 'flex-start' }}>
            <GlassPanel tone="plastic" px="$3" py="$1.5" rounded="$pill" elevate={0}>
              <GlowText level="label">{user.email}</GlowText>
            </GlassPanel>
            <GlassPanel tone="plastic" px="$3" py="$1.5" rounded="$pill" elevate={0}>
              <GlowText level="label">{user.role}</GlowText>
            </GlassPanel>
          </XStack>
        </YStack>
      </GlassPanel>

      <GlowText level="heading" align="center">
        Pick a channel
      </GlowText>

      <XStack flexWrap="wrap" gap="$4" justify="center">
        {CHANNELS.map((channel) => (
          <ChannelTile
            key={channel.key}
            label={channel.label}
            tone={channel.tone}
            icon={channel.icon}
            locked={channel.locked}
          />
        ))}
      </XStack>

      <GlossButton tone="red" btnSize="md" onPress={onLogout} accessibilityLabel="Log out">
        Log out
      </GlossButton>
    </PageShell>
  )
}
