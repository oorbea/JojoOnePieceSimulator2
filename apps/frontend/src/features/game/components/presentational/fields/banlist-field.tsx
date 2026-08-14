import { X } from '@tamagui/lucide-icons-2'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { a11yProps } from '@/shared/lib/a11y'
import type { FruitType, Rarity, StandStat } from '@/shared/lib/zod'

/** Every `StandResponse` stat field `BanByFilterFields` can filter on. */
export type StandStatKey =
  'attackPower' | 'speed' | 'attackRange' | 'endurance' | 'precision' | 'potential'

// A power that can be banned - built by the container from
// @/features/stands's useStands() and @/features/devil-fruits's
// useDevilFruits() results (cross-feature imports go through those
// barrels, never their internal paths). `kind` is display-only, both power
// kinds share the same PowerID space server-side so `banned` is a flat
// list of ids regardless of kind. `rarity`/`fruitType`/`stats` are carried
// alongside the id/name so `BanByFilterFields` can compute "ban every power
// matching this rarity/fruit-type/stand-stat" without a second query -
// `BanlistField` itself only ever reads `id`/`name`/`kind`.
export type BannableItem = {
  id: string
  name: string
  kind: 'STAND' | 'DEVIL_FRUIT'
  rarity: Rarity
  fruitType?: FruitType
  stats?: Record<StandStatKey, StandStat>
}

const MAX_RESULTS = 6

type Props = {
  editable: boolean
  banned: string[]
  items: BannableItem[]
}

// GlassField search box over Stands + Devil Fruits - typing filters
// candidates by name, tapping one bans it; already-banned items render as
// removable chips below. Search text is local UI state (like
// LobbyRoomScreen's configExpanded), the banned-id list itself is owned by
// the container via onAddBan/onRemoveBan.
export function BanlistField({
  editable,
  banned,
  items,
  onAddBan,
  onRemoveBan,
}: Props & { onAddBan: (id: string) => void; onRemoveBan: (id: string) => void }) {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')

  const byId = useMemo(() => new Map(items.map((item) => [item.id, item])), [items])

  const results = useMemo(() => {
    const query = search.trim().toLowerCase()
    if (!query) return []
    return items
      .filter((item) => !banned.includes(item.id) && item.name.toLowerCase().includes(query))
      .slice(0, MAX_RESULTS)
  }, [items, banned, search])

  return (
    <YStack width="100%" gap="$2">
      <GlowText level="label">{t('game.pool.banlist')}</GlowText>

      {editable ? (
        <>
          <GlassField
            label={t('game.pool.banlistSearch')}
            value={search}
            onChangeText={setSearch}
            placeholder={t('game.pool.banlistSearch')}
          />
          {results.length > 0 ? (
            <YStack width="100%" gap="$1">
              {results.map((item) => (
                <XStack
                  key={item.id}
                  width="100%"
                  items="center"
                  justify="space-between"
                  py="$2"
                  px="$3"
                  rounded="$card"
                  bg="$plasticFill"
                  cursor="pointer"
                  onPress={() => {
                    onAddBan(item.id)
                    setSearch('')
                  }}
                  {...a11yProps(item.name, 'button')}
                >
                  <GlowText level="label">{item.name}</GlowText>
                </XStack>
              ))}
            </YStack>
          ) : null}
        </>
      ) : null}

      {banned.length === 0 ? (
        <GlowText level="label">{t('game.pool.banlistEmpty')}</GlowText>
      ) : (
        <XStack width="100%" flexWrap="wrap" gap="$2">
          {banned.map((id) => {
            const item = byId.get(id)
            const label = item?.name ?? id
            return (
              <GlassPanel key={id} tone="plastic" px="$3" py="$1.5" rounded="$pill" elevate={0}>
                <XStack items="center" gap="$2">
                  <GlowText level="label">{label}</GlowText>
                  {editable ? (
                    <XStack
                      onPress={() => onRemoveBan(id)}
                      cursor="pointer"
                      {...a11yProps(t('game.pool.removeBan', { name: label }), 'button')}
                    >
                      <X size={14} color="$panelText" />
                    </XStack>
                  ) : null}
                </XStack>
              </GlassPanel>
            )
          })}
        </XStack>
      )}
    </YStack>
  )
}
