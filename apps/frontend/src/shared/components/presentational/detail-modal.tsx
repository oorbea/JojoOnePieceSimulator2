import { X } from '@tamagui/lucide-icons-2'
import { useEffect } from 'react'
import { Modal, Platform } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { ScrollView, XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { notifyScroll } from '@/shared/lib/scroll-bus'

import { GlassPanel } from './glass-panel'
import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'

type DetailModalProps = {
  visible: boolean
  title: string
  onClose: () => void
  closeA11y: string
  /** Extra actions rendered under the content, e.g. an admin's "Edit"
   * button. Omitted entirely for read-only viewers. */
  footer?: React.ReactNode
  children: React.ReactNode
}

// The "card brought to the foreground" view for a single Stand/Devil
// Fruit/Stage — same full-window Modal + centered glass card recipe as
// ConfirmSheet/StandFormModal, so it inherits the same backdrop dismiss,
// Esc-to-close (web) and safe-area handling without reinventing any of it.
// Body content is per-feature (StandDetail/DevilFruitDetail/StageDetail);
// this only owns the chrome: title, close, optional footer, scroll.
export function DetailModal({ visible, title, onClose, closeA11y, footer, children }: DetailModalProps) {
  const insets = useSafeAreaInsets()

  // Esc closes overlays (norma-teclado.md) - RN's Modal has no built-in key
  // handling on web, so this wires it directly to the DOM while visible.
  useEffect(() => {
    if (Platform.OS !== 'web' || !visible || typeof document === 'undefined') return
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [visible, onClose])

  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onClose} statusBarTranslucent>
      <YStack
        flex={1}
        items="center"
        justify="center"
        p="$4"
        pt={insets.top + 16}
        pb={insets.bottom + 16}
        bg="rgba(10,12,20,0.45)"
        onPress={onClose}
        {...a11yProps(title, 'none')}
      >
        <GlassPanel
          tone="strong"
          radiusSize="panel"
          elevate={3}
          width="100%"
          maxW={560}
          maxH="90%"
          p="$5"
          gap="$4"
          onPress={(e: { stopPropagation?: () => void }) => e.stopPropagation?.()}
        >
          <XStack items="center" justify="space-between" gap="$3">
            <GlowText level="heading" flex={1} numberOfLines={2}>
              {title}
            </GlowText>
            <GlossButton
              tone="glass"
              btnSize="sm"
              shape="circle"
              onPress={onClose}
              accessibilityLabel={closeA11y}
            >
              <X size={16} color="$panelText" />
            </GlossButton>
          </XStack>

          <ScrollView
            flex={1}
            minH={0}
            keyboardShouldPersistTaps="handled"
            onScroll={notifyScroll}
            scrollEventThrottle={16}
          >
            <YStack gap="$4" pb="$2">
              {children}
            </YStack>
          </ScrollView>

          {footer ? <YStack gap="$2">{footer}</YStack> : null}
        </GlassPanel>
      </YStack>
    </Modal>
  )
}
