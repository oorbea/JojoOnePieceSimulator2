import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlowText } from '@/shared/components/presentational/glow-text'

// The "To Be Continued ⟵" card (owner decision, 2026-09-25 playtest
// feedback - see ObsidianVault/game-victory-defeat-cinematic-2026-09-14.md):
// the direct JoJo meme reference, deliberately left untranslated in all
// three locales (game.result.cinematic.epilogueLabel) - translating it
// loses the reference. Mounted during the defeat cinematic's epilogue
// beat; the CALLER wraps this in a full-bleed Animated.View and drives
// ITS opacity/transform (transform/opacity only, per project norm) - this
// component itself is static content and takes no position prop of its
// own, so it's safe to nest inside that wrapper without a stray
// zero-size positioning context.
export function TbcArrow() {
  const { t } = useTranslation()
  return (
    <YStack flex={1} justify="flex-end" items="flex-start" p="$4">
      <XStack
        items="center"
        gap="$2"
        borderWidth={2}
        borderColor="rgba(231,217,184,0.55)"
        rounded="$card"
        px="$3"
        py="$2"
        bg="rgba(20,10,6,0.35)"
      >
        <GlowText level="heading" color="#e7d9b8">
          ⟵
        </GlowText>
        <GlowText level="label" color="#e7d9b8" fontFamily="$display">
          {t('game.result.cinematic.epilogueLabel')}
        </GlowText>
      </XStack>
    </YStack>
  )
}
