import { Bot, Globe, Lock, Swords, Users } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { BanlistField, type BannableItem } from '@/features/game/components/presentational/fields/banlist-field'
import { NumberStepper } from '@/features/game/components/presentational/fields/number-stepper'
import { PowerPoolFields } from '@/features/game/components/presentational/fields/power-pool-fields'
import type { PoolFilter } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { SettingRow } from '@/shared/components/presentational/setting-row'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import type { GameMode, LobbyVisibility, Manga } from '@/shared/lib/zod'

// Fields mirror create-lobby-screen.tsx's exactly (mode, mangas, team size,
// voting window, privacy, allow bots, pool filter) since UPDATE_CONFIG is a
// full replacement of the same Config shape.
type Props = {
  isHost: boolean
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
  onAddBan: (id: string) => void
  onRemoveBan: (id: string) => void
  onClearPoolFilter: () => void
  saving: boolean
  saved: boolean
  error?: string
  onSubmit: () => void
}

export function LobbyConfigPanel({
  isHost,
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
  onAddBan,
  onRemoveBan,
  onClearPoolFilter,
  saving,
  saved,
  error,
  onSubmit,
}: Props) {
  const { t } = useTranslation()

  return (
    <YStack width="100%" gap="$3">
      {!isHost ? (
        <YStack gap="$1">
          <GlowText level="label">{t('game.config.readOnly')}</GlowText>
          <GlowText level="label">{t('game.config.hostOnly')}</GlowText>
        </YStack>
      ) : null}

      {isHost ? (
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
          </WiiCard>
        </XStack>
      ) : (
        <XStack width="100%" items="center" gap="$2">
          {mode === 'GAUNTLET' ? <Users size={18} color="$panelText" /> : <Swords size={18} color="$panelText" />}
          <GlowText level="heading">{t(`enums.gameMode.${mode}`)}</GlowText>
        </XStack>
      )}

      <GlassPanel glossy p="$5" gap="$4" width="100%" $md={{ flexDirection: 'row' }}>
        <YStack flexBasis={320} grow={1} gap="$4">
          <YStack gap="$1.5">
            <GlowText level="label">{t('game.create.mangasLabel')}</GlowText>
            {isHost ? (
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
            ) : (
              <GlowText level="heading">
                {mangas.map((manga) => t(`enums.manga.${manga}`)).join(', ')}
              </GlowText>
            )}
          </YStack>

          {isHost ? (
            <NumberStepper
              label={mode === 'GAUNTLET' ? t('game.create.teamSizeGauntletLabel') : t('game.create.teamSizeLabel')}
              value={teamSize}
              min={teamSizeMin}
              max={teamSizeMax}
              onChange={onChangeTeamSize}
            />
          ) : (
            <YStack gap="$1.5">
              <GlowText level="label">
                {mode === 'GAUNTLET' ? t('game.create.teamSizeGauntletLabel') : t('game.create.teamSizeLabel')}
              </GlowText>
              <GlowText level="heading">{teamSize}</GlowText>
            </YStack>
          )}
        </YStack>

        <YStack flexBasis={320} grow={1} gap="$4">
          {isHost ? (
            <NumberStepper
              label={t('game.create.votingSecondsLabel')}
              value={votingWindowSeconds}
              min={5}
              max={180}
              onChange={onChangeVotingWindow}
            />
          ) : (
            <YStack gap="$1.5">
              <GlowText level="label">{t('game.create.votingSecondsLabel')}</GlowText>
              <GlowText level="heading">{votingWindowSeconds}</GlowText>
            </YStack>
          )}

          <SettingRow label={t('game.create.privacyLabel')}>
            {isHost ? (
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
            ) : (
              <XStack items="center" gap="$2">
                {visibility === 'PUBLIC' ? (
                  <Globe size={14} color="$panelText" />
                ) : (
                  <Lock size={14} color="$panelText" />
                )}
                <GlowText level="heading">{t(`enums.lobbyVisibility.${visibility}`)}</GlowText>
              </XStack>
            )}
          </SettingRow>

          <YStack gap="$1.5">
            <SettingRow label={t('game.create.allowBotsLabel')}>
              {isHost ? (
                <GlossButton
                  tone={allowBots ? 'blue' : 'glass'}
                  btnSize="sm"
                  disabled={mode === 'GAUNTLET'}
                  onPress={onToggleAllowBots}
                  accessibilityLabel={allowBots ? t('game.create.allowBotsOn') : t('game.create.allowBotsOff')}
                >
                  <XStack items="center" gap="$2">
                    <Bot size={14} color="$panelText" />
                    {allowBots ? t('game.create.allowBotsOn') : t('game.create.allowBotsOff')}
                  </XStack>
                </GlossButton>
              ) : (
                <GlowText level="heading">
                  {allowBots ? t('game.create.allowBotsOn') : t('game.create.allowBotsOff')}
                </GlowText>
              )}
            </SettingRow>
            {isHost && mode === 'GAUNTLET' ? (
              <GlowText level="label">{t('game.create.allowBotsGauntletHint')}</GlowText>
            ) : null}
          </YStack>
        </YStack>
      </GlassPanel>

      <PowerPoolFields editable={isHost} activeCount={poolActiveCount} onClearAll={onClearPoolFilter}>
        <BanlistField
          editable={isHost}
          banned={poolFilter.banned}
          items={banlistItems}
          onAddBan={onAddBan}
          onRemoveBan={onRemoveBan}
        />
      </PowerPoolFields>

      {error ? <GlowText level="label" color="$strawHatRedDeep">{error}</GlowText> : null}

      {isHost ? (
        <GlossButton
          tone="green"
          btnSize="lg"
          flare
          width="100%"
          disabled={saving || mangas.length === 0}
          onPress={onSubmit}
          accessibilityLabel={t('common.save')}
        >
          {saving ? t('game.config.saving') : saved ? t('game.config.saved') : t('common.save')}
        </GlossButton>
      ) : null}
    </YStack>
  )
}
