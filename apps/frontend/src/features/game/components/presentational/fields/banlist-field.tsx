import { Ban, X } from '@tamagui/lucide-icons-2'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { searchPoolItems } from '@/features/game/lib/pool-stats'
import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { a11yProps } from '@/shared/lib/a11y'
import type { FruitType, Rarity, StandStat } from '@/shared/contracts/enums'

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
  /** Renders a clear-banlist GlossButton next to the "Banned powers" label
   * when set AND editable AND banned is non-empty. The UI-visible pool
   * filter is banlist-only, so "clear banlist" and "clear the whole pool
   * filter" are the same action from the container's point of view. */
  onClearBanlist?: () => void
}

// GlassField search box over Stands + Devil Fruits - typing filters
// candidates by name (case/diacritic-insensitive, multi-token AND, via
// searchPoolItems), tapping one bans it; already-banned items render as
// removable chips below. Search text is local UI state (like
// LobbyRoomScreen's configExpanded), the banned-id list itself is owned by
// the container via onAddBan/onRemoveBan.
export function BanlistField({
  editable,
  banned,
  items,
  onAddBan,
  onRemoveBan,
  onClearBanlist,
}: Props & { onAddBan: (id: string) => void; onRemoveBan: (id: string) => void }) {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')

  const byId = useMemo(() => new Map(items.map((item) => [item.id, item])), [items])

  const { results, total } = useMemo(
    () => searchPoolItems(items, banned, search, MAX_RESULTS),
    [items, banned, search]
  )
  const hasQuery = search.trim().length > 0

  return (
    <YStack width="100%" gap="$2">
      <XStack width="100%" items="center" justify="space-between" gap="$2">
        <GlowText level="label">{t('game.pool.banlist')}</GlowText>
        {editable && onClearBanlist && banned.length > 0 ? (
          <GlossButton
            tone="glass"
            btnSize="sm"
            onPress={onClearBanlist}
            accessibilityLabel={t('game.pool.banlistClear')}
            tooltip={t('game.pool.banlistClearHint')}
          >
            <XStack items="center" gap="$1.5">
              <Ban size={14} color="$panelText" />
              <GlowText level="label">{t('game.pool.banlistClear')}</GlowText>
            </XStack>
          </GlossButton>
        ) : null}
      </XStack>

      {editable ? (
        <>
          <XStack width="100%" items="flex-end" gap="$2">
            <YStack flex={1}>
              <GlassField
                label={t('game.pool.banlistSearch')}
                value={search}
                onChangeText={setSearch}
                placeholder={t('game.pool.banlistSearch')}
              />
            </YStack>
            {hasQuery ? (
              <GlossButton
                tone="glass"
                btnSize="sm"
                shape="circle"
                onPress={() => setSearch('')}
                accessibilityLabel={t('game.pool.banlistClearSearch')}
                tooltip={t('game.pool.banlistClearSearchHint')}
              >
                <X size={14} color="$panelText" />
              </GlossButton>
            ) : null}
          </XStack>

          {hasQuery ? (
            results.length > 0 ? (
              <YStack width="100%" gap="$1">
                <GlowText level="label">
                  {t('game.pool.banlistResultCount', { shown: results.length, total })}
                </GlowText>
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
                    <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
                      <GlowText level="label">
                        {item.kind === 'STAND' ? t('game.pool.kindStand') : t('game.pool.kindDevilFruit')}
                      </GlowText>
                    </GlassPanel>
                  </XStack>
                ))}
              </YStack>
            ) : (
              <GlowText level="label">{t('game.pool.banlistNoMatches')}</GlowText>
            )
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
