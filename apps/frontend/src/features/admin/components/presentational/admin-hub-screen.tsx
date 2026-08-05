import { Apple, Sparkles } from '@tamagui/lucide-icons-2'
import { XStack } from 'tamagui'

import { ChannelTile } from '@/shared/components/presentational/channel-tile'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'

type Props = {
  onOpenStands: () => void
  onOpenDevilFruits: () => void
}

// Same Wii channel-selection recipe as HomeScreen — the admin hub is a
// second, smaller "pick a channel" grid, one tile per manageable domain.
export function AdminHubScreen({ onOpenStands, onOpenDevilFruits }: Props) {
  return (
    <PageShell align="top" navPadding scroll maxWidth={720}>
      <GlowText level="title" align="center">
        Admin Panel
      </GlowText>
      <GlowText level="heading" align="center">
        Pick a channel to manage
      </GlowText>

      <XStack flexWrap="wrap" gap="$4" justify="center">
        <ChannelTile label="Stands" tone="grape" icon={Sparkles} onPress={onOpenStands} />
        <ChannelTile label="Devil Fruits" tone="red" icon={Apple} onPress={onOpenDevilFruits} />
      </XStack>
    </PageShell>
  )
}
