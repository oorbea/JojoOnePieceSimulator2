import { Check } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { InfoHint } from '@/shared/components/presentational/info-hint'
import type { Manga } from '@/shared/lib/zod'

const MANGAS: Manga[] = ['JOJO', 'ONE_PIECE']

type Props = {
  mangas: Manga[]
  isHost: boolean
  onToggle: (manga: Manga) => void
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
// Host-only edits autosave (see lobby-room-container.tsx's
// handleToggleConfigManga) - there's no separate "save" step for this one
// field, unlike the rest of "Lobby settings".
export function MangaRow({ mangas, isHost, onToggle, saving, saved }: Props) {
  const { t } = useTranslation()

  return (
    <YStack width="100%" gap="$1.5">
      <XStack items="center" gap="$1.5">
        <GlowText level="label">{t('game.create.mangasLabel')}</GlowText>
        <InfoHint text={t('game.create.help.mangas')} />
        {saving ? <GlowText level="label" tone="soft">{t('game.config.saving')}</GlowText> : null}
        {!saving && saved ? <GlowText level="label" tone="soft">{t('game.config.saved')}</GlowText> : null}
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
