import { Sparkles, X } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { Image, Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { ScrollView, XStack, YStack } from 'tamagui'

import { ParticipantAvatar } from '@/features/game/components/presentational/match/participant-avatar'
import {
  STAND_STAT_KEYS,
  STAND_STAT_LABELS,
} from '@/features/game/components/presentational/match/loadout-card'
import { loadoutSlots } from '@/features/game/lib/match-rules'
import type { GameParticipant } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import type { Manga } from '@/shared/lib/zod'

type Props = {
  visible: boolean
  participant: GameParticipant | null
  isSelf: boolean
  mangas: Manga[]
  onClose: () => void
}

// The near-fullscreen breakdown a tap on a ParticipantTile opens: same data
// as the old inline LoadoutCard, just rendered large enough to actually
// read the descriptions/skills that card never had room for. Single
// column on phones, Stand | Devil Fruit side by side from $md up.
export function LoadoutModal({ visible, participant, isSelf, mangas, onClose }: Props) {
  const { t } = useTranslation()
  const insets = useSafeAreaInsets()

  if (!participant) return null
  const loadout = participant.loadout
  const stand = loadout?.stand
  const fruit = loadout?.devilFruit
  const slots = loadout ? loadoutSlots(loadout, mangas) : []
  const scalarSlots = slots.filter((s) => s.key !== 'stand' && s.key !== 'devilFruit')
  const hasStandSlot = slots.some((s) => s.key === 'stand')
  const hasFruitSlot = slots.some((s) => s.key === 'devilFruit')

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
      statusBarTranslucent
    >
      <YStack
        flex={1}
        items="center"
        justify="center"
        p="$3"
        pt={insets.top + 12}
        pb={insets.bottom + 12}
        bg="rgba(10,12,20,0.55)"
        onPress={onClose}
      >
        <GlassPanel
          tone="strong"
          radiusSize="panel"
          elevate={3}
          width="100%"
          height="100%"
          maxW={880}
          p="$4"
          gap="$3"
          onPress={(e: { stopPropagation?: () => void }) => e.stopPropagation?.()}
        >
          <XStack items="center" gap="$2.5">
            <ParticipantAvatar participant={participant} size={48} isSelf={isSelf} />
            <GlowText level="heading" flex={1} numberOfLines={1}>
              {participant.displayName}
            </GlowText>
            <GlossButton
              tone="glass"
              shape="circle"
              btnSize="sm"
              onPress={onClose}
              accessibilityLabel={t('game.match.loadout.closeA11y')}
            >
              <X size={18} />
            </GlossButton>
          </XStack>

          <ScrollView flex={1} minH={0}>
            <YStack gap="$3" pb="$2" $md={{ flexDirection: 'row', flexWrap: 'wrap' }}>
              {hasStandSlot ? (
                <YStack flex={1} minW={280} gap="$2">
                  <PowerBlock
                    picture={stand?.picture}
                    name={stand?.name}
                    rarityLabel={stand ? t(`enums.rarity.${stand.rarity}`) : undefined}
                    description={stand?.description}
                    skills={stand?.skills}
                    fallbackLabel={t('game.match.noStand')}
                  >
                    {stand ? (
                      <XStack flexWrap="wrap" gap="$2" mt="$1">
                        {STAND_STAT_KEYS.map((key) => (
                          <YStack
                            key={key}
                            flexBasis={72}
                            grow={1}
                            minW={72}
                            items="center"
                            gap="$0.5"
                          >
                            <GlowText level="label" tone="soft" fontSize="$1">
                              {STAND_STAT_LABELS[key]}
                            </GlowText>
                            <GlowText level="heading" fontSize="$5">
                              {t(`enums.standStat.${stand[key]}`)}
                            </GlowText>
                          </YStack>
                        ))}
                      </XStack>
                    ) : null}
                  </PowerBlock>
                </YStack>
              ) : null}

              {hasFruitSlot ? (
                <YStack flex={1} minW={280} gap="$2">
                  <PowerBlock
                    picture={fruit?.picture}
                    name={fruit?.name}
                    rarityLabel={fruit ? t(`enums.fruitType.${fruit.fruitType}`) : undefined}
                    description={fruit?.description}
                    skills={fruit?.skills}
                    fallbackLabel={t('game.match.noFruit')}
                  />
                </YStack>
              ) : null}

              {scalarSlots.length > 0 ? (
                <YStack width="100%" gap="$1.5">
                  {scalarSlots.map((slot) => (
                    <XStack
                      key={slot.key}
                      items="center"
                      justify="space-between"
                      px="$3"
                      py="$2"
                      rounded="$card"
                      bg="$plasticFill"
                    >
                      <GlowText level="label">{slot.i18nKey ? t(slot.i18nKey) : slot.key}</GlowText>
                      <GlowText level="label" tone={slot.value === 'NONE' ? 'soft' : undefined}>
                        {t(`enums.${enumNamespace(slot.key)}.${slot.value}`)}
                      </GlowText>
                    </XStack>
                  ))}
                </YStack>
              ) : null}
            </YStack>
          </ScrollView>
        </GlassPanel>
      </YStack>
    </Modal>
  )
}

// Mirrors trait-chips.tsx's own enumNamespace - kept local since the modal
// needs it for the same scalar slots, and the two are unlikely to grow
// enough shared surface to justify a shared module yet.
function enumNamespace(key: string): string {
  switch (key) {
    case 'spin':
      return 'spinLevel'
    case 'hamon':
      return 'hamonLevel'
    case 'fruitMastery':
      return 'fruitMastery'
    case 'physicalForm':
      return 'physicalForm'
    default:
      return 'hakiLevel'
  }
}

type PowerBlockProps = {
  picture?: string
  name?: string
  rarityLabel?: string
  description?: string
  skills?: string[]
  fallbackLabel: string
  children?: React.ReactNode
}

// One power's full breakdown: large art, name + rarity/type pill,
// description, skills - and, for a Stand, the caller's own stat grid slotted
// in as `children`. Shows `fallbackLabel` when the loadout didn't draw one
// (e.g. NoStandWeight won the roll) rather than an empty block.
function PowerBlock({
  picture,
  name,
  rarityLabel,
  description,
  skills,
  fallbackLabel,
  children,
}: PowerBlockProps) {
  const { t } = useTranslation()
  return (
    <GlassPanel tone="plastic" p="$3" gap="$2" rounded="$panel">
      <YStack
        width="100%"
        height={160}
        rounded="$card"
        overflow="hidden"
        position="relative"
        bg="$plasticEdge"
      >
        <InsetRing rounded="$card" />
        {name ? (
          picture ? (
            <Image source={{ uri: picture }} style={{ width: '100%', height: '100%' }} />
          ) : (
            <YStack flex={1} items="center" justify="center">
              <Sparkles size={36} color="$standPurple" />
            </YStack>
          )
        ) : (
          <YStack flex={1} items="center" justify="center">
            <GlowText level="label" tone="soft">
              {fallbackLabel}
            </GlowText>
          </YStack>
        )}
      </YStack>

      {name ? (
        <>
          <XStack items="center" justify="space-between" gap="$2">
            <GlowText level="heading" numberOfLines={1} flex={1}>
              {name}
            </GlowText>
            {rarityLabel ? (
              <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
                <GlowText level="label">{rarityLabel}</GlowText>
              </GlassPanel>
            ) : null}
          </XStack>
          {description ? <GlowText level="label">{description}</GlowText> : null}
          {skills && skills.length > 0 ? (
            <YStack gap="$1">
              <GlowText level="label" tone="soft" fontSize="$1">
                {t('stands.skills')}
              </GlowText>
              <XStack flexWrap="wrap" gap="$1.5">
                {skills.map((skill) => (
                  <GlassPanel
                    key={skill}
                    tone="plastic"
                    px="$2"
                    py="$0.5"
                    rounded="$pill"
                    elevate={0}
                  >
                    <GlowText level="label" fontSize="$1">
                      {skill}
                    </GlowText>
                  </GlassPanel>
                ))}
              </XStack>
            </YStack>
          ) : null}
          {children}
        </>
      ) : null}
    </GlassPanel>
  )
}
