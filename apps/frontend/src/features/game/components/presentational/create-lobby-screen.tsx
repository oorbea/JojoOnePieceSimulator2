import { Bot, ChevronLeft, Globe, Lock, Swords, Users } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { BanByFilterFields } from '@/features/game/components/presentational/fields/ban-by-filter-fields'
import { BanlistField, type BannableItem } from '@/features/game/components/presentational/fields/banlist-field'
import { NumberStepper } from '@/features/game/components/presentational/fields/number-stepper'
import { PowerPoolFields } from '@/features/game/components/presentational/fields/power-pool-fields'
import type { PoolFilter } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { InfoHint } from '@/shared/components/presentational/info-hint'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { SettingRow } from '@/shared/components/presentational/setting-row'
import { TooltipBubble, useTooltipTrigger } from '@/shared/components/presentational/tooltip'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import type { GameMode, LobbyVisibility, Manga } from '@/shared/lib/zod'

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
  onAddBan: (id: string) => void
  onRemoveBan: (id: string) => void
  onBanMatching: (ids: string[]) => void
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
  onAddBan,
  onRemoveBan,
  onBanMatching,
  onClearPoolFilter,
  submitting,
  error,
  onSubmit,
}: Props) {
  const { t } = useTranslation()
  const {
    visible: gauntletTooltipVisible,
    anchor: gauntletTooltipAnchor,
    triggerRef: gauntletTooltipRef,
    triggerProps: gauntletTooltipProps,
  } = useTooltipTrigger(t('game.create.help.modeGauntlet'))
  const {
    visible: versusTooltipVisible,
    anchor: versusTooltipAnchor,
    triggerRef: versusTooltipRef,
    triggerProps: versusTooltipProps,
  } = useTooltipTrigger(t('game.create.help.modeVersus'))

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
          ref={gauntletTooltipRef as never}
          interactive
          flexBasis={200}
          grow={1}
          padded
          gap="$2"
          borderColor={mode === 'GAUNTLET' ? ('$wiiBlue' as never) : undefined}
          onPress={() => onChangeMode('GAUNTLET')}
          {...gauntletTooltipProps}
          {...a11yProps(t('enums.gameMode.GAUNTLET'), 'button')}
        >
          <Users size={22} color="$panelText" />
          <GlowText level="heading">{t('enums.gameMode.GAUNTLET')}</GlowText>
          <GlowText level="label">{t('game.create.modeGauntletHint')}</GlowText>
        </WiiCard>
        <WiiCard
          ref={versusTooltipRef as never}
          interactive
          flexBasis={200}
          grow={1}
          padded
          gap="$2"
          borderColor={mode === 'VERSUS' ? ('$strawHatRed' as never) : undefined}
          onPress={() => onChangeMode('VERSUS')}
          {...versusTooltipProps}
          {...a11yProps(t('enums.gameMode.VERSUS'), 'button')}
        >
          <Swords size={22} color="$panelText" />
          <GlowText level="heading">{t('enums.gameMode.VERSUS')}</GlowText>
          <GlowText level="label">{t('game.create.modeVersusHint')}</GlowText>
        </WiiCard>
      </XStack>
      <TooltipBubble visible={gauntletTooltipVisible} label={t('game.create.help.modeGauntlet')} anchor={gauntletTooltipAnchor} />
      <TooltipBubble visible={versusTooltipVisible} label={t('game.create.help.modeVersus')} anchor={versusTooltipAnchor} />

      <GlassPanel glossy p="$5" gap="$4" width="100%" $md={{ flexDirection: 'row' }}>
        <YStack flexBasis={320} grow={1} gap="$4">
          <YStack gap="$1.5">
            <XStack items="center" gap="$1.5">
              <GlowText level="label">{t('game.create.mangasLabel')}</GlowText>
              <InfoHint text={t('game.create.help.mangas')} />
            </XStack>
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
            help={<InfoHint text={t('game.create.help.teamSize')} />}
            value={teamSize}
            min={teamSizeMin}
            max={teamSizeMax}
            onChange={onChangeTeamSize}
          />
        </YStack>

        <YStack flexBasis={320} grow={1} gap="$4">
          <NumberStepper
            label={t('game.create.votingSecondsLabel')}
            help={<InfoHint text={t('game.create.help.votingSeconds')} />}
            value={votingWindowSeconds}
            min={5}
            max={180}
            onChange={onChangeVotingWindow}
          />

          <SettingRow label={t('game.create.privacyLabel')} help={<InfoHint text={t('game.create.help.privacy')} />}>
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
          </SettingRow>

          <YStack gap="$1.5">
            <SettingRow
              label={t('game.create.allowBotsLabel')}
              help={<InfoHint text={t('game.create.help.allowBots')} />}
            >
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
            </SettingRow>
            {mode === 'GAUNTLET' ? (
              <GlowText level="label">{t('game.create.allowBotsGauntletHint')}</GlowText>
            ) : null}
          </YStack>
        </YStack>
      </GlassPanel>

      <PowerPoolFields editable activeCount={poolActiveCount} onClearAll={onClearPoolFilter}>
        <BanByFilterFields editable items={banlistItems} onBanMatching={onBanMatching} />
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
