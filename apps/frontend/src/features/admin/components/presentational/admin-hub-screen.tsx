import { Apple, Map, Sparkles } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import { ChannelTile } from '@/shared/components/presentational/channel-tile'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'

type Props = {
  onOpenStands: () => void
  onOpenDevilFruits: () => void
  onOpenStages: () => void
}

// Same Wii channel-selection recipe as HomeScreen — the admin hub is a
// second, smaller "pick a channel" grid, one tile per manageable domain.
export function AdminHubScreen({ onOpenStands, onOpenDevilFruits, onOpenStages }: Props) {
  const { t } = useTranslation()
  return (
    <PageShell align="top" scroll maxWidth={720}>
      <GlowText level="title" align="center">
        {t('admin.title')}
      </GlowText>
      <GlowText level="heading" align="center">
        {t('admin.pickChannel')}
      </GlowText>

      <XStack flexWrap="wrap" gap="$4" justify="center">
        <ChannelTile label={t('admin.stands')} tone="grape" icon={Sparkles} onPress={onOpenStands} />
        <ChannelTile label={t('admin.devilFruits')} tone="red" icon={Apple} onPress={onOpenDevilFruits} />
        <ChannelTile label={t('admin.stages')} tone="blue" icon={Map} onPress={onOpenStages} />
      </XStack>
    </PageShell>
  )
}
