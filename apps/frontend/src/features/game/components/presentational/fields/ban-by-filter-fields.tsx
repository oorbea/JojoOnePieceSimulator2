import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import type { BannableItem, StandStatKey } from '@/features/game/components/presentational/fields/banlist-field'
import { GlassSelect, type GlassSelectOption } from '@/shared/components/presentational/glass-select'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { fruitTypeSchema, raritySchema, standStatSchema, type FruitType, type Rarity } from '@/shared/contracts/enums'

const RARITIES: Rarity[] = raritySchema.options
const FRUIT_TYPES: FruitType[] = fruitTypeSchema.options
const STAND_STAT_KEYS: StandStatKey[] = [
  'attackPower', 'speed', 'attackRange', 'endurance', 'precision', 'potential',
]

type Props = {
  editable: boolean
  items: BannableItem[]
  onBanMatching: (ids: string[]) => void
}

// Bulk-ban by criteria - "ban every LEGENDARY power", "ban every stand with
// SPD:A", "ban every Zoan fruit" - instead of hunting each one down in
// BanlistField's search box one at a time. Still pure banning underneath
// (adds matched ids to the same flat `banned` list `BanlistField` owns) -
// this is NOT the rarity/fruit-type whitelist that was cut from the create
// form; it's a shortcut for filling the banlist, the semantics stay
// ban-only. A stat filter only ever matches Stands (Devil Fruits have no
// stats) and a fruit-type filter only ever matches Devil Fruits (Stands
// have no fruit type) - setting one silently excludes the other kind from
// the match set, same "dimension implies kind" rule the rest of this
// feature already uses.
export function BanByFilterFields({ editable, items, onBanMatching }: Props) {
  const { t } = useTranslation()
  const [rarities, setRarities] = useState<Rarity[]>([])
  const [fruitTypes, setFruitTypes] = useState<FruitType[]>([])
  const [stats, setStats] = useState<Record<StandStatKey, string | null>>({
    attackPower: null,
    speed: null,
    attackRange: null,
    endurance: null,
    precision: null,
    potential: null,
  })

  const statOptions: GlassSelectOption[] = useMemo(
    () => standStatSchema.options.map((v) => ({ value: v, label: t(`enums.standStat.${v}`) })),
    [t]
  )

  const hasStatFilter = STAND_STAT_KEYS.some((key) => stats[key])
  const hasAnyFilter = rarities.length > 0 || fruitTypes.length > 0 || hasStatFilter

  const matches = useMemo(() => {
    if (!hasAnyFilter) return []
    return items.filter((item) => {
      if (item.kind === 'STAND') {
        if (fruitTypes.length > 0) return false
        if (rarities.length > 0 && !rarities.includes(item.rarity)) return false
        return STAND_STAT_KEYS.every((key) => !stats[key] || item.stats?.[key] === stats[key])
      }
      if (hasStatFilter) return false
      if (rarities.length > 0 && !rarities.includes(item.rarity)) return false
      if (fruitTypes.length > 0 && (!item.fruitType || !fruitTypes.includes(item.fruitType))) return false
      return true
    })
  }, [items, rarities, fruitTypes, stats, hasAnyFilter, hasStatFilter])

  const toggleRarity = (rarity: Rarity) =>
    setRarities((current) => (current.includes(rarity) ? current.filter((r) => r !== rarity) : [...current, rarity]))

  const toggleFruitType = (fruitType: FruitType) =>
    setFruitTypes((current) =>
      current.includes(fruitType) ? current.filter((f) => f !== fruitType) : [...current, fruitType]
    )

  if (!editable) return null

  return (
    <YStack width="100%" gap="$3">
      <GlowText level="label">{t('game.pool.banByFilter.title')}</GlowText>

      <YStack gap="$1.5">
        <GlowText level="label" tone="soft">{t('game.pool.banByFilter.rarities')}</GlowText>
        <XStack gap="$2" flexWrap="wrap">
          {RARITIES.map((rarity) => (
            <GlossButton
              key={rarity}
              tone={rarities.includes(rarity) ? 'blue' : 'glass'}
              btnSize="sm"
              onPress={() => toggleRarity(rarity)}
              accessibilityLabel={t(`enums.rarity.${rarity}`)}
            >
              {t(`enums.rarity.${rarity}`)}
            </GlossButton>
          ))}
        </XStack>
      </YStack>

      <YStack gap="$1.5">
        <GlowText level="label" tone="soft">{t('game.pool.banByFilter.fruitTypes')}</GlowText>
        <XStack gap="$2" flexWrap="wrap">
          {FRUIT_TYPES.map((fruitType) => (
            <GlossButton
              key={fruitType}
              tone={fruitTypes.includes(fruitType) ? 'blue' : 'glass'}
              btnSize="sm"
              onPress={() => toggleFruitType(fruitType)}
              accessibilityLabel={t(`enums.fruitType.${fruitType}`)}
            >
              {t(`enums.fruitType.${fruitType}`)}
            </GlossButton>
          ))}
        </XStack>
      </YStack>

      <YStack gap="$1.5">
        <GlowText level="label" tone="soft">{t('game.pool.banByFilter.standStats')}</GlowText>
        <XStack width="100%" flexWrap="wrap" gap="$2">
          {STAND_STAT_KEYS.map((key) => (
            <YStack key={key} flexBasis={140} grow={1}>
              <GlassSelect
                label={t(`stands.stats.${key}`)}
                options={statOptions}
                value={stats[key]}
                onChange={(value) => setStats((current) => ({ ...current, [key]: value }))}
                clearable
              />
            </YStack>
          ))}
        </XStack>
      </YStack>

      <XStack items="center" gap="$3" flexWrap="wrap">
        <GlossButton
          tone="red"
          btnSize="sm"
          disabled={matches.length === 0}
          onPress={() => onBanMatching(matches.map((m) => m.id))}
          accessibilityLabel={t('game.pool.banByFilter.banMatching')}
        >
          {t('game.pool.banByFilter.banMatching')}
        </GlossButton>
        <GlowText level="label" tone="soft">
          {t('game.pool.banByFilter.matchCount', { count: matches.length })}
        </GlowText>
      </XStack>
    </YStack>
  )
}
