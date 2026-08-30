import { Sparkles } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Image, Pressable } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { ImageLightbox } from '@/shared/components/presentational/image-lightbox'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'

export type PowerBlockProps = {
  picture?: string
  name?: string
  rarityLabel?: string
  description?: string
  skills?: string[]
  fallbackLabel: string
  /** Taller art area for the sorteo's own full-screen reveal card (owner
   * request, 2026-08-30) - LoadoutModal keeps the original 160. */
  artHeight?: number
  children?: React.ReactNode
}

// One power's full breakdown: large art, name + rarity/type pill,
// description, skills - and, for a Stand, the caller's own stat grid slotted
// in as `children`. Shows `fallbackLabel` when the loadout didn't draw one
// (e.g. NoStandWeight won the roll) rather than an empty block. Extracted
// out of loadout-modal.tsx (2026-08-30) so the sorteo's own big power-reveal
// card (power-reveal-card.tsx) can reuse the exact same layout instead of
// duplicating it. Tapping the art opens ImageLightbox full-screen, same
// recipe as the admin Stand/DevilFruit cards (owner request, 2026-08-30).
export function PowerBlock({
  picture,
  name,
  rarityLabel,
  description,
  skills,
  fallbackLabel,
  artHeight = 160,
  children,
}: PowerBlockProps) {
  const { t } = useTranslation()
  const [previewOpen, setPreviewOpen] = useState(false)
  return (
    <GlassPanel tone="plastic" p="$3" gap="$2" rounded="$panel">
      <Pressable
        onPress={() => setPreviewOpen(true)}
        disabled={!picture}
        {...a11yProps(name ? t('game.match.loadout.viewImageA11y', { name }) : fallbackLabel, 'imagebutton')}
      >
        <YStack
          width="100%"
          height={artHeight}
          rounded="$card"
          overflow="hidden"
          position="relative"
          bg="$plasticEdge"
        >
          <InsetRing rounded="$card" />
          {name ? (
            picture ? (
              <Image source={{ uri: picture }} style={{ width: '100%', height: '100%' }} />
            ) : (
              <YStack flex={1} items="center" justify="center">
                <Sparkles size={36} color="$standPurple" />
              </YStack>
            )
          ) : (
            <YStack flex={1} items="center" justify="center">
              <GlowText level="label" tone="soft">
                {fallbackLabel}
              </GlowText>
            </YStack>
          )}
        </YStack>
      </Pressable>
      <ImageLightbox visible={previewOpen} uri={picture ?? null} onClose={() => setPreviewOpen(false)} />

      {name ? (
        <>
          <XStack items="center" justify="space-between" gap="$2">
            <GlowText level="heading" numberOfLines={1} flex={1}>
              {name}
            </GlowText>
            {rarityLabel ? (
              <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
                <GlowText level="label">{rarityLabel}</GlowText>
              </GlassPanel>
            ) : null}
          </XStack>
          {description ? <GlowText level="label">{description}</GlowText> : null}
          {skills && skills.length > 0 ? (
            <YStack gap="$1">
              <GlowText level="label" tone="soft" fontSize="$1">
                {t('stands.skills')}
              </GlowText>
              <XStack flexWrap="wrap" gap="$1.5">
                {skills.map((skill) => (
                  <GlassPanel
                    key={skill}
                    tone="plastic"
                    px="$2"
                    py="$0.5"
                    rounded="$pill"
                    elevate={0}
                  >
                    <GlowText level="label" fontSize="$1">
                      {skill}
                    </GlowText>
                  </GlassPanel>
                ))}
              </XStack>
            </YStack>
          ) : null}
          {children}
        </>
      ) : null}
    </GlassPanel>
  )
}
