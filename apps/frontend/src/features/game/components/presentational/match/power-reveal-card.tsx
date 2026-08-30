import { useTranslation } from 'react-i18next'
import { Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { XStack, YStack } from 'tamagui'

import {
  STAND_STAT_KEYS,
  STAND_STAT_LABELS,
} from '@/features/game/components/presentational/match/loadout-card'
import { PowerBlock } from '@/features/game/components/presentational/match/power-block'
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'
import type { StandResponse } from '@/features/stands/types/stands.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  visible: boolean
  kind: 'stand' | 'devilFruit'
  stand?: StandResponse
  devilFruit?: DevilFruitResponse
  participantName: string
  onSkip: () => void
}

// The sorteo's own big power reveal (owner request, 2026-08-30): when a
// participant's turn lands a Stand or a Devil Fruit, this takes over the
// whole screen for that slot's hold (game.RevealHoldStandMs/
// RevealHoldFruitMs, ~10s/5s at Swift speed) so everyone can actually read
// the art, description and skills before the sorteo moves on - the exact
// gap the pre-2026-08-30 reveal never closed (see reveal.go's superseded
// "this UI never renders a power's description" note). Reuses PowerBlock,
// the same big-card layout LoadoutModal already shows after the match, so
// the two never drift. Carries its own Skip button (bug found 2026-08-30
// manual testing: RevealStage's own Skip sits underneath this Modal, but a
// Modal captures every touch on both native and web - the button was
// visually present but completely unreachable while this card was up).
export function PowerRevealCard({ visible, kind, stand, devilFruit, participantName, onSkip }: Props) {
  const { t } = useTranslation()
  const insets = useSafeAreaInsets()

  const isStand = kind === 'stand'
  const power = isStand ? stand : devilFruit

  return (
    <Modal visible={visible} transparent animationType="fade" statusBarTranslucent>
      <YStack
        flex={1}
        items="center"
        justify="center"
        p="$3"
        pt={insets.top + 12}
        pb={insets.bottom + 12}
        bg="rgba(10,12,20,0.72)"
      >
        <GlassPanel tone="strong" radiusSize="panel" elevate={3} width="100%" maxW={560} p="$4" gap="$3">
          <XStack items="center" justify="center">
            <GlowText level="label" tone="soft">
              {participantName}
            </GlowText>
          </XStack>
          <PowerBlock
            picture={power?.picture}
            name={power?.name}
            rarityLabel={
              isStand
                ? stand
                  ? t(`enums.rarity.${stand.rarity}`)
                  : undefined
                : devilFruit
                  ? t(`enums.fruitType.${devilFruit.fruitType}`)
                  : undefined
            }
            description={power?.description}
            skills={power?.skills}
            fallbackLabel={isStand ? t('game.match.noStand') : t('game.match.noFruit')}
            artHeight={220}
          >
            {isStand && stand ? (
              <XStack flexWrap="wrap" gap="$2" mt="$1">
                {STAND_STAT_KEYS.map((key) => (
                  <YStack key={key} flexBasis={72} grow={1} minW={72} items="center" gap="$0.5">
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
          <XStack justify="center">
            <GlossButton
              tone="glass"
              btnSize="sm"
              onPress={onSkip}
              accessibilityLabel={t('game.match.reveal.skipA11y')}
              tooltip={t('game.match.reveal.skipA11y')}
            >
              {t('game.match.reveal.skip')}
            </GlossButton>
          </XStack>
        </GlassPanel>
      </YStack>
    </Modal>
  )
}
