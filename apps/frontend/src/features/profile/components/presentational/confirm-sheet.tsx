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
export function ConfirmSheet({
  visible,
  title,
  message,
  confirmLabel,
  isConfirming = false,
  onConfirm,
  onCancel,
}: ConfirmSheetProps) {
  if (!visible) return null

  return (
    <YStack
      position="absolute"
      t={0}
      l={0}
      r={0}
      b={0}
      z="$overlay"
      items="center"
      justify="center"
      p="$4"
      style={{ pointerEvents: 'auto' }}
      bg="rgba(10,12,20,0.45)"
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
  )
}
