import { Modal } from 'react-native'
import { YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { a11yProps } from '@/shared/lib/a11y'

type ConfirmSheetProps = {
  visible: boolean
  title: string
  message: string
  confirmLabel: string
  isConfirming?: boolean
  onConfirm: () => void
  onCancel: () => void
}

// Minimal destructive-action confirmation, scoped to this feature only — no
// generic modal system exists yet, and neither "remove avatar" nor "delete
// account" need one. A dimmed full-bleed backdrop plus a centered glass card
// with an explicit Cancel/destructive-confirm pair.
//
// Rendered through RN's own Modal rather than as an absolutely-positioned
// sibling of the page content: an in-page absolute only stacks against its
// own siblings, so it used to sit *below* AppShell's floating nav bars
// (z:$nav, 500) despite asking for z:$overlay (700) — those bars live one
// level up in the tree, where RN's z-index comparison never reaches. Modal
// renders to the native window root (and to a document-level portal on
// react-native-web), so it now actually covers everything, including the
// bars. Tapping the dimmed backdrop cancels; tapping the card doesn't, since
// the card's own press consumes the touch before it reaches the backdrop.
export function ConfirmSheet({
  visible,
  title,
  message,
  confirmLabel,
  isConfirming = false,
  onConfirm,
  onCancel,
}: ConfirmSheetProps) {
  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onCancel}
      statusBarTranslucent
    >
      <YStack
        flex={1}
        items="center"
        justify="center"
        p="$4"
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
