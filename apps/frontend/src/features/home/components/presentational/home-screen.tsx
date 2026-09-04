import { Apple, Gamepad2, Landmark, Sparkles, User } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { Image } from 'react-native'
import { Paragraph, XStack, YStack } from 'tamagui'

import { ChannelTile } from '@/shared/components/presentational/channel-tile'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { GlossOverlay } from '@/shared/components/presentational/gloss-overlay'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { PageShell } from '@/shared/components/presentational/page-shell'
import type { SessionUser } from '@/shared/stores/session.store'

type Props = {
  user: SessionUser
  onNavigate: (href: string) => void
}

// Every channel now leads somewhere: Play and Profile navigate directly,
// Stands/Devil Fruits/Stages open the read-only public catalogue at
// /catalog/* (see admin's editable equivalent at /admin/*). Profile's icon
// matches the top nav's (`User`) instead of the old `Home` stand-in.
// Labels come from useTranslation() in the component below - this only
// pins the i18n key.
const CHANNELS = [
  { key: 'play', labelKey: 'home.channels.play', tone: 'green' as const, icon: Gamepad2, href: '/play' },
  { key: 'profile', labelKey: 'home.channels.profile', tone: 'blue' as const, icon: User, href: '/profile' },
  { key: 'stands', labelKey: 'home.channels.stands', tone: 'grape' as const, icon: Sparkles, href: '/catalog/stands' },
  {
    key: 'fruits',
    labelKey: 'home.channels.devilFruits',
    tone: 'red' as const,
    icon: Apple,
    href: '/catalog/devil-fruits',
  },
  {
    key: 'stages',
    labelKey: 'home.channels.stages',
    tone: 'yellow' as const,
    icon: Landmark,
    href: '/catalog/stages',
  },
]

// Pure UI — lives inside the authenticated app shell, so it doesn't carry
// its own backdrop or logout button; AppShell already provides both (a
// second logout here would just duplicate the one in the top bar).
export function HomeScreen({ user, onNavigate }: Props) {
  const { t } = useTranslation()
  return (
    <PageShell align="top" scroll maxWidth={720}>
      <GlassPanel
        glossy
        elevate={2}
        width="100%"
        p="$6"
        gap="$5"
        items="center"
        $md={{ flexDirection: 'row' }}
      >
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

        <YStack items="center" gap="$3" $md={{ items: 'flex-start' }}>
          <GlowText level="hero" $md={{ fontSize: '$12' }}>
            {user.username}
          </GlowText>

          <XStack flexWrap="wrap" gap="$2" justify="center" $md={{ justify: 'flex-start' }}>
            <GlassPanel tone="plastic" px="$3" py="$1.5" rounded="$pill" elevate={0}>
              <GlowText level="label">{user.email}</GlowText>
            </GlassPanel>
            <GlassPanel tone="plastic" px="$3" py="$1.5" rounded="$pill" elevate={0}>
              <GlowText level="label">{t(`enums.role.${user.role}`)}</GlowText>
            </GlassPanel>
          </XStack>
        </YStack>
      </GlassPanel>

      <GlowText level="heading" align="center">
        {t('home.pickChannel')}
      </GlowText>

      <XStack flexWrap="wrap" gap="$4" justify="center">
        {CHANNELS.map((channel) => (
          <ChannelTile
            key={channel.key}
            label={t(channel.labelKey)}
            tone={channel.tone}
            icon={channel.icon}
            onPress={() => onNavigate(channel.href)}
          />
        ))}
      </XStack>
    </PageShell>
  )
}
