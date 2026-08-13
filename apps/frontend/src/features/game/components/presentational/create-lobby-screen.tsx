import { ChevronLeft, Globe, Lock, Swords, Users } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { BanlistField, type BannableItem } from '@/features/game/components/presentational/fields/banlist-field'
import { NumberStepper } from '@/features/game/components/presentational/fields/number-stepper'
import { PowerPoolFields } from '@/features/game/components/presentational/fields/power-pool-fields'
import type { PoolFilter } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import type { FruitType, GameMode, LobbyVisibility, Manga, Rarity } from '@/shared/lib/zod'

type Props = {
  onBack: () => void
  mode: GameMode
  onChangeMode: (mode: GameMode) => void
  mangas: Manga[]
  onToggleManga: (manga: Manga) => void
  teamSize: number
  teamSizeMin: number
  teamSizeMax: number
  onChangeTeamSize: (size: number) => void
  allowBots: boolean
  onToggleAllowBots: () => void
  visibility: LobbyVisibility
  onToggleVisibility: () => void
  votingWindowSeconds: number
  onChangeVotingWindow: (seconds: number) => void
  poolFilter: PoolFilter
  poolActiveCount: number
  banlistItems: BannableItem[]
  onToggleRarity: (rarity: Rarity) => void
  onToggleFruitType: (fruitType: FruitType) => void
  onAddBan: (id: string) => void
  onRemoveBan: (id: string) => void
  onClearPoolFilter: () => void
  submitting: boolean
  error?: string
  onSubmit: () => void
}

export function CreateLobbyScreen({
  onBack,
  mode,
  onChangeMode,
  mangas,
  onToggleManga,
  teamSize,
  teamSizeMin,
  teamSizeMax,
  onChangeTeamSize,
  allowBots,
  onToggleAllowBots,
  visibility,
  onToggleVisibility,
  votingWindowSeconds,
  onChangeVotingWindow,
  poolFilter,
  poolActiveCount,
  banlistItems,
  onToggleRarity,
  onToggleFruitType,
  onAddBan,
  onRemoveBan,
  onClearPoolFilter,
  submitting,
  error,
  onSubmit,
}: Props) {
  const { t } = useTranslation()

  return (
    <PageShell align="top" scroll maxWidth={720}>
      <XStack width="100%" items="center" gap="$3">
        <GlossButton
          tone="glass"
          btnSize="sm"
          shape="circle"
          onPress={onBack}
          accessibilityLabel={t('common.cancel')}
        >
          <ChevronLeft size={18} color="$panelText" />
        </GlossButton>
        <GlowText level="title">{t('game.create.title')}</GlowText>
      </XStack>

      <XStack width="100%" flexWrap="wrap" gap="$3">
        <WiiCard
          interactive
          flexBasis={200}
          grow={1}
          padded
          gap="$2"
          borderColor={mode === 'GAUNTLET' ? ('$wiiBlue' as never) : undefined}
          onPress={() => onChangeMode('GAUNTLET')}
          {...a11yProps(t('enums.gameMode.GAUNTLET'), 'button')}
        >
          <Users size={22} color="$panelText" />
          <GlowText level="heading">{t('enums.gameMode.GAUNTLET')}</GlowText>
          <GlowText level="label">{t('game.create.modeGauntletHint')}</GlowText>
        </WiiCard>
        <WiiCard
          interactive
          flexBasis={200}
          grow={1}
          padded
          gap="$2"
          borderColor={mode === 'VERSUS' ? ('$strawHatRed' as never) : undefined}
          onPress={() => onChangeMode('VERSUS')}
          {...a11yProps(t('enums.gameMode.VERSUS'), 'button')}
        >
          <Swords size={22} color="$panelText" />
          <GlowText level="heading">{t('enums.gameMode.VERSUS')}</GlowText>
          <GlowText level="label">{t('game.create.modeVersusHint')}</GlowText>
        </WiiCard>
      </XStack>

      <GlassPanel glossy p="$5" gap="$4" width="100%" $md={{ flexDirection: 'row' }}>
        <YStack flexBasis={320} grow={1} gap="$4">
          <YStack gap="$1.5">
            <GlowText level="label">{t('game.create.mangasLabel')}</GlowText>
            <XStack gap="$2">
              {(['JOJO', 'ONE_PIECE'] as Manga[]).map((manga) => (
                <GlossButton
                  key={manga}
                  tone={mangas.includes(manga) ? 'blue' : 'glass'}
                  btnSize="sm"
                  onPress={() => onToggleManga(manga)}
                  accessibilityLabel={t(`enums.manga.${manga}`)}
                >
                  {t(`enums.manga.${manga}`)}
                </GlossButton>
              ))}
            </XStack>
          </YStack>

          <NumberStepper
            label={mode === 'GAUNTLET' ? t('game.create.teamSizeGauntletLabel') : t('game.create.teamSizeLabel')}
            value={teamSize}
            min={teamSizeMin}
            max={teamSizeMax}
            onChange={onChangeTeamSize}
          />
        </YStack>

        <YStack flexBasis={320} grow={1} gap="$4">
          <NumberStepper
            label={t('game.create.votingSecondsLabel')}
            value={votingWindowSeconds}
            min={5}
            max={180}
            onChange={onChangeVotingWindow}
          />

          <YStack gap="$1.5">
            <GlowText level="label">{t('game.create.privacyLabel')}</GlowText>
            <GlossButton
              tone="glass"
              btnSize="sm"
              onPress={onToggleVisibility}
              accessibilityLabel={t(`enums.lobbyVisibility.${visibility}`)}
            >
              <XStack items="center" gap="$2">
                {visibility === 'PUBLIC' ? (
                  <Globe size={14} color="$panelText" />
                ) : (
                  <Lock size={14} color="$panelText" />
                )}
                {t(`enums.lobbyVisibility.${visibility}`)}
              </XStack>
            </GlossButton>
          </YStack>

          <YStack gap="$1.5">
            <GlowText level="label">{t('game.create.allowBotsLabel')}</GlowText>
            <GlossButton
              tone={allowBots ? 'blue' : 'glass'}
              btnSize="sm"
              disabled={mode === 'GAUNTLET'}
              onPress={onToggleAllowBots}
              accessibilityLabel={t('game.create.allowBotsLabel')}
            >
              {allowBots ? t('common.save') : t('common.none')}
            </GlossButton>
            {mode === 'GAUNTLET' ? (
              <GlowText level="label">{t('game.create.allowBotsGauntletHint')}</GlowText>
            ) : null}
          </YStack>
        </YStack>
      </GlassPanel>

      <PowerPoolFields
        editable
        standRarities={poolFilter.standRarities as Rarity[]}
        fruitRarities={poolFilter.fruitRarities as Rarity[]}
        fruitTypes={poolFilter.fruitTypes as FruitType[]}
        activeCount={poolActiveCount}
        onToggleRarity={onToggleRarity}
        onToggleFruitType={onToggleFruitType}
        onClearAll={onClearPoolFilter}
      >
        <BanlistField
          editable
          banned={poolFilter.banned}
          items={banlistItems}
          onAddBan={onAddBan}
          onRemoveBan={onRemoveBan}
        />
      </PowerPoolFields>

      {error ? <GlowText level="label" color="$strawHatRedDeep">{error}</GlowText> : null}

      <GlossButton
        tone="green"
        btnSize="lg"
        flare
        width="100%"
        disabled={submitting || mangas.length === 0}
        onPress={onSubmit}
        accessibilityLabel={t('game.create.submit')}
      >
        {submitting ? t('game.create.creating') : t('game.create.submit')}
      </GlossButton>
    </PageShell>
  )
}
