import { Sparkles } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { TraitChip } from '@/features/game/components/presentational/match/trait-chips'
import { loadoutSlots, type LoadoutSlot } from '@/features/game/lib/match-rules'
import type { GameParticipant } from '@/features/game/types/game.types'
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'
import type { StandResponse } from '@/features/stands/types/stands.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { LazyImage } from '@/shared/components/presentational/lazy-image'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import type { Manga } from '@/shared/contracts/enums'
import { cardSource, lqipSource } from '@/shared/lib/picture-source'

type Props = {
  participant: GameParticipant
  isSelf: boolean
  mangas: Manga[]
}

// Exported for LoadoutModal, which renders the same stat grid at a larger
// size rather than duplicating the key/label table.
export const STAND_STAT_KEYS = [
  'attackPower',
  'speed',
  'attackRange',
  'endurance',
  'precision',
  'potential',
] as const
export const STAND_STAT_LABELS: Record<(typeof STAND_STAT_KEYS)[number], string> = {
  attackPower: 'PWR',
  speed: 'SPD',
  attackRange: 'RNG',
  endurance: 'END',
  precision: 'PRE',
  potential: 'DEV',
}

// A render-order row: the two big art blocks (Stand/DevilFruit) render on
// their own full-width row; every scalar chip slot between/around them
// collapses into a single flex-wrap row with its neighbours, same as the
// old flat TraitChips row - only now each chip still reveals individually
// via its own `index` into loadoutSlots, so a slot "between" two blocks
// isn't held back by them.
type Row =
  | { kind: 'block'; key: 'stand' | 'devilFruit'; index: number }
  | { kind: 'chips'; entries: { slot: LoadoutSlot; index: number }[] }

function buildRows(slots: LoadoutSlot[]): Row[] {
  const rows: Row[] = []
  slots.forEach((slot, index) => {
    if (slot.key === 'stand' || slot.key === 'devilFruit') {
      rows.push({ kind: 'block', key: slot.key, index })
      return
    }
    const last = rows[rows.length - 1]
    if (last && last.kind === 'chips') last.entries.push({ slot, index })
    else rows.push({ kind: 'chips', entries: [{ slot, index }] })
  })
  return rows
}

// Renders one participant's finished loadout - the sorteo ceremony (spin,
// reveal, suspense) happens entirely in RevealStage before this card ever
// mounts, so there's nothing left for this card to animate.
export function LoadoutCard({ participant, isSelf, mangas }: Props) {
  const { t } = useTranslation()
  const loadout = participant.loadout
  const slots = loadout ? loadoutSlots(loadout, mangas) : []
  const rows = buildRows(slots)

  return (
    <WiiCard
      padded
      width={220}
      gap="$2.5"
      borderColor={isSelf ? '$wiiBlue' : undefined}
      borderWidth={isSelf ? 2.5 : undefined}
    >
      <XStack items="center" gap="$1.5" flexWrap="wrap">
        <GlowText level="label" numberOfLines={1} flex={1}>
          {participant.displayName}
        </GlowText>
        {isSelf ? (
          <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
            <GlowText level="label">{t('game.lobby.you')}</GlowText>
          </GlassPanel>
        ) : null}
      </XStack>

      <YStack width="100%" gap="$2.5">
        {rows.map((row, i) => {
          if (row.kind === 'block') {
            return row.key === 'stand' ? (
              <StandBlock key="stand" stand={loadout?.stand} visible t={t} />
            ) : (
              <DevilFruitBlock key="devilFruit" devilFruit={loadout?.devilFruit} visible t={t} />
            )
          }
          return (
            <XStack key={`chips-${i}`} gap="$1.5" flexWrap="wrap">
              {row.entries.map((e) => (
                <TraitChip key={e.slot.key} slot={e.slot} />
              ))}
            </XStack>
          )
        })}
      </YStack>
    </WiiCard>
  )
}

type TFunc = (key: string) => string

// Reserves the block's usual height even before its slot is revealed, so
// later slots (a haki chip row, the DevilFruit block, ...) don't jump the
// layout down as each one pops in - only the CONTENT swaps in at reveal
// time, the box itself is always there once this manga's slot exists at all.
function StandBlock({ stand, visible, t }: { stand?: StandResponse; visible: boolean; t: TFunc }) {
  return (
    <YStack gap="$1.5">
      {visible ? (
        stand ? (
          <LazyImage
            uri={cardSource(stand)}
            lqip={lqipSource(stand)}
            height={110}
            pictureStatus={stand.pictureStatus}
            fallback={<Sparkles size={26} color="$standPurple" />}
          />
        ) : (
          <YStack
            width="100%"
            height={110}
            rounded="$card"
            overflow="hidden"
            items="center"
            justify="center"
            bg="$plasticEdge"
          >
            <GlowText level="label" tone="soft">
              {t('game.match.noStand')}
            </GlowText>
          </YStack>
        )
      ) : (
        <YStack width="100%" height={110} rounded="$card" overflow="hidden" bg="$plasticEdge" />
      )}

      {visible && stand ? (
        <YStack gap="$1.5">
          <XStack items="center" justify="space-between">
            <GlowText level="label" numberOfLines={1} flex={1}>
              {stand.name}
            </GlowText>
            <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
              <GlowText level="label" fontSize="$1">
                {t(`enums.rarity.${stand.rarity}`)}
              </GlowText>
            </GlassPanel>
          </XStack>
          <XStack flexWrap="wrap" gap="$1.5">
            {STAND_STAT_KEYS.map((key) => (
              <YStack key={key} flexBasis={56} grow={1} minW={56} items="center" gap="$0.5">
                <GlowText level="label" tone="soft" fontSize="$1">
                  {STAND_STAT_LABELS[key]}
                </GlowText>
                <GlowText level="label" fontSize="$3">
                  {t(`enums.standStat.${stand[key]}`)}
                </GlowText>
              </YStack>
            ))}
          </XStack>
        </YStack>
      ) : null}
    </YStack>
  )
}

function DevilFruitBlock({
  devilFruit,
  visible,
  t,
}: {
  devilFruit?: DevilFruitResponse
  visible: boolean
  t: TFunc
}) {
  return (
    <YStack gap="$1.5">
      {visible ? (
        devilFruit ? (
          <LazyImage
            uri={cardSource(devilFruit)}
            lqip={lqipSource(devilFruit)}
            height={90}
            pictureStatus={devilFruit.pictureStatus}
            fallback={<Sparkles size={22} color="$tangerine" />}
          />
        ) : (
          <YStack
            width="100%"
            height={90}
            rounded="$card"
            overflow="hidden"
            items="center"
            justify="center"
            bg="$plasticEdge"
          >
            <GlowText level="label" tone="soft">
              {t('game.match.noFruit')}
            </GlowText>
          </YStack>
        )
      ) : (
        <YStack width="100%" height={90} rounded="$card" overflow="hidden" bg="$plasticEdge" />
      )}

      {visible && devilFruit ? (
        <XStack items="center" justify="space-between" gap="$1.5">
          <GlowText level="label" numberOfLines={1} flex={1}>
            {devilFruit.name}
          </GlowText>
          <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
            <GlowText level="label" fontSize="$1">
              {t(`enums.fruitType.${devilFruit.fruitType}`)}
            </GlowText>
          </GlassPanel>
        </XStack>
      ) : null}
    </YStack>
  )
}
