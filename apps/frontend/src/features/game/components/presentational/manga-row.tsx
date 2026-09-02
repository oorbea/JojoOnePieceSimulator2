import { Check } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { InfoHint } from '@/shared/components/presentational/info-hint'
import type { Manga } from '@/shared/contracts/enums'

const MANGAS: Manga[] = ['JOJO', 'ONE_PIECE']

type GroupProps = {
  labelKey: string
  helpKey: string
  mangas: Manga[]
  isHost: boolean
  onToggle: (manga: Manga) => void
}

function MangaToggleGroup({ labelKey, helpKey, mangas, isHost, onToggle }: GroupProps) {
  const { t } = useTranslation()
  return (
    <YStack width="100%" gap="$1.5">
      <XStack items="center" gap="$1.5">
        <GlowText level="label">{t(labelKey)}</GlowText>
        <InfoHint text={t(helpKey)} />
      </XStack>
      <XStack flexWrap="wrap" gap="$2">
        {MANGAS.map((manga) => {
          const active = mangas.includes(manga)
          return (
            <GlossButton
              key={manga}
              tone={active ? 'blue' : 'glass'}
              btnSize="sm"
              disabled={!isHost}
              onPress={isHost ? () => onToggle(manga) : undefined}
              accessibilityLabel={t(`enums.manga.${manga}`)}
              tooltip={isHost ? t(`enums.manga.${manga}`) : null}
            >
              <XStack items="center" gap="$1.5">
                {active ? <Check size={14} color="$panelText" /> : null}
                <GlowText level="label">{t(`enums.manga.${manga}`)}</GlowText>
              </XStack>
            </GlossButton>
          )
        })}
      </XStack>
    </YStack>
  )
}

type Props = {
  stageMangas: Manga[]
  powerMangas: Manga[]
  isHost: boolean
  onToggleStageManga: (manga: Manga) => void
  onTogglePowerManga: (manga: Manga) => void
  saving?: boolean
  saved?: boolean
}

// Always-visible manga selector for the main lobby screen (not the
// collapsed "Lobby settings" panel) - the owner reported that which
// manga(s) a lobby draws powers from wasn't legible without opening
// settings. A visible check mark on the active chip is the affordance fix
// (color alone was the only signal before): flexWrap lets the two chips
// wrap under the label on narrow screens instead of overflowing a
// non-wrapping row, which is what happened to this same field inside the
// config panel (long manga names, e.g. "JoJo's Bizarre Adventure").
//
// Two independent toggle groups (Stages / Powers), not one - the owner
// wants to be able to run e.g. Stages from both mangas while powers stay
// JoJo-only. Host-only edits autosave (see lobby-room-container.tsx's
// handleToggleConfigManga) - there's no separate "save" step for this
// field, unlike the rest of "Lobby settings".
export function MangaRow({
  stageMangas,
  powerMangas,
  isHost,
  onToggleStageManga,
  onTogglePowerManga,
  saving,
  saved,
}: Props) {
  const { t } = useTranslation()

  return (
    <YStack width="100%" gap="$2.5">
      <XStack items="center" gap="$1.5">
        {saving ? <GlowText level="label" tone="soft">{t('game.config.saving')}</GlowText> : null}
        {!saving && saved ? <GlowText level="label" tone="soft">{t('game.config.saved')}</GlowText> : null}
      </XStack>
      <MangaToggleGroup
        labelKey="game.create.stageMangasLabel"
        helpKey="game.create.help.stageMangas"
        mangas={stageMangas}
        isHost={isHost}
        onToggle={onToggleStageManga}
      />
      <MangaToggleGroup
        labelKey="game.create.powerMangasLabel"
        helpKey="game.create.help.powerMangas"
        mangas={powerMangas}
        isHost={isHost}
        onToggle={onTogglePowerManga}
      />
    </YStack>
  )
}
