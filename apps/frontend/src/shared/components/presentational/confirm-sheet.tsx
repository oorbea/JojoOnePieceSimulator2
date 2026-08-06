import { Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'

import { GlassPanel } from './glass-panel'
import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'

type ConfirmSheetProps = {
  visible: boolean
  title: string
  message: string
  confirmLabel: string
  isConfirming?: boolean
  onConfirm: () => void
  onCancel: () => void
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
}: ConfirmSheetProps) {
  const insets = useSafeAreaInsets()

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
              tone="red"
              btnSize="md"
              disabled={isConfirming}
              onPress={onConfirm}
              accessibilityLabel={confirmLabel}
            >
              {isConfirming ? 'Working…' : confirmLabel}
            </GlossButton>
            <GlossButton
              tone="glass"
              btnSize="md"
              disabled={isConfirming}
              onPress={onCancel}
              accessibilityLabel="Cancel"
            >
              Cancel
            </GlossButton>
          </YStack>
        </GlassPanel>
      </YStack>
    </Modal>
  )
}
