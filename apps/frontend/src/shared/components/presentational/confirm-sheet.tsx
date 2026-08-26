import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'

import { GlassPanel, isWebPlatform } from './glass-panel'
import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'

const CONFIRM_BUTTON_ID = 'confirm-sheet-confirm-button'

type ConfirmSheetProps = {
  visible: boolean
  title: string
  message: string
  confirmLabel: string
  isConfirming?: boolean
  onConfirm: () => void
  onCancel: () => void
  /** Confirm button tone - defaults to 'red' (every existing caller is a
   * destructive action). Pass 'blue' or similar for a non-destructive
   * confirmation, e.g. changing an already-cast vote. */
  tone?: 'red' | 'blue' | 'glass'
}

// Shared destructive-action confirmation — promoted out of the profile
// feature (features/profile/components/presentational/confirm-sheet.tsx,
// which keeps its own copy) so admin features (Stands, Devil Fruits) get
// the same dimmed-backdrop + centered glass card pattern without a
// cross-feature import. See that file for the full Modal-vs-absolute
// z-index rationale — identical here.
export function ConfirmSheet({
  visible,
  title,
  message,
  confirmLabel,
  isConfirming = false,
  onConfirm,
  onCancel,
  tone = 'red',
}: ConfirmSheetProps) {
  const { t } = useTranslation()
  const insets = useSafeAreaInsets()

  // Web-only: a keyboard user opening this sheet via Enter/Space on a
  // GlossButton would otherwise land with focus nowhere in particular
  // (Modal doesn't move it) and have to Tab in blind. GlossButton owns its
  // own internal ref (for the tooltip's .measure() anchoring) and doesn't
  // forward an external one, so this focuses via the DOM id instead - the
  // same escape hatch use-roving-group.ts uses for the same reason. A
  // short timeout lets the Modal's own mount/animation settle first (an
  // immediate focus() call can be swallowed mid-transition on some
  // browsers).
  useEffect(() => {
    if (!isWebPlatform || !visible || typeof document === 'undefined') return
    const id = setTimeout(() => document.getElementById(CONFIRM_BUTTON_ID)?.focus(), 50)
    return () => clearTimeout(id)
  }, [visible])

  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onCancel} statusBarTranslucent>
      <YStack
        flex={1}
        items="center"
        justify="center"
        p="$4"
        pt={insets.top + 16}
        pb={insets.bottom + 16}
        bg="rgba(10,12,20,0.45)"
        onPress={isConfirming ? undefined : onCancel}
        {...a11yProps(title, 'alert')}
      >
        <GlassPanel
          tone="strong"
          radiusSize="panel"
          elevate={3}
          width="100%"
          maxW={380}
          p="$5"
          gap="$4"
          onPress={(e: { stopPropagation?: () => void }) => e.stopPropagation?.()}
        >
          <GlowText level="heading" align="center">
            {title}
          </GlowText>
          <GlowText level="label" align="center">
            {message}
          </GlowText>
          <YStack gap="$2">
            <GlossButton
              id={isWebPlatform ? CONFIRM_BUTTON_ID : undefined}
              tone={tone}
              btnSize="md"
              disabled={isConfirming}
              onPress={onConfirm}
              accessibilityLabel={confirmLabel}
            >
              {isConfirming ? t('common.working') : confirmLabel}
            </GlossButton>
            <GlossButton
              tone="glass"
              btnSize="md"
              disabled={isConfirming}
              onPress={onCancel}
              accessibilityLabel={t('common.cancel')}
            >
              {t('common.cancel')}
            </GlossButton>
          </YStack>
        </GlassPanel>
      </YStack>
    </Modal>
  )
}
