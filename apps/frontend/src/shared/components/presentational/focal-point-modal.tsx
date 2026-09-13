import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { ScrollView, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { notifyScroll } from '@/shared/lib/scroll-bus'

import { FocalPointPicker } from './focal-point-picker'
import { GlassPanel } from './glass-panel'
import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'

type Props = {
  visible: boolean
  uri: string | null
  x: number
  y: number
  onConfirm: (x: number, y: number) => void
  /** Present only when this modal was reopened from the edit form
   * ("Ajustar encuadre") - the admin can back out to the value it had
   * before reopening. Omitted for the mandatory post-upload framing step,
   * where there's nothing to cancel back to and the backdrop/Esc commit
   * the current draft instead of dismissing (see onRequestClose below). */
  onCancel?: () => void
}

// Same Modal + dimmed-backdrop + centered GlassPanel recipe as
// ConfirmSheet/*FormModal (a real RN Modal, not an absolute overlay - RN
// only compares zIndex between direct siblings, which is exactly the bug
// ConfirmSheet's own history worked around). Holds its own local draft so
// dragging the crosshair never writes through to the form/network on every
// move - the value only reaches the caller once, via onConfirm.
export function FocalPointModal({ visible, uri, x, y, onConfirm, onCancel }: Props) {
  const { t } = useTranslation()
  const insets = useSafeAreaInsets()
  const [draft, setDraft] = useState({ x, y })
  const [wasVisible, setWasVisible] = useState(visible)

  // Re-seed the draft whenever the modal is (re)opened against a possibly
  // different starting value - not on every `x`/`y` change, since those are
  // this modal's own edits echoed back through the caller once confirmed.
  // Adjusted directly during render off a "was it already visible" key
  // (same pattern lobby-room-container.tsx uses) rather than a useEffect,
  // which would call setState synchronously inside the effect body
  // (react-hooks/set-state-in-effect).
  if (visible !== wasVisible) {
    setWasVisible(visible)
    if (visible) setDraft({ x, y })
  }

  const commit = () => onConfirm(draft.x, draft.y)

  // RN translates onRequestClose to Esc (web) and the back button
  // (Android) - per norma-teclado.md every Modal must route it. Confirming
  // is never destructive (worst case the value is the default center), so
  // the mandatory post-upload framing step commits the current draft
  // instead of trapping the admin with no way out; the edit-form reopening
  // restores the previous value instead, matching its Cancel button.
  const onRequestClose = onCancel ?? commit

  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onRequestClose} statusBarTranslucent>
      <YStack
        flex={1}
        items="center"
        justify="center"
        p="$4"
        pt={insets.top + 16}
        pb={insets.bottom + 16}
        bg="rgba(10,12,20,0.45)"
        {...a11yProps(t('focalPoint.modalTitle'), 'alert')}
      >
        <GlassPanel
          tone="strong"
          radiusSize="panel"
          elevate={3}
          width="100%"
          maxW={480}
          maxH="90%"
          p="$5"
          gap="$4"
        >
          <GlowText level="heading" align="center">
            {t('focalPoint.modalTitle')}
          </GlowText>

          <ScrollView
            flex={1}
            minH={0}
            keyboardShouldPersistTaps="handled"
            onScroll={notifyScroll}
            scrollEventThrottle={16}
          >
            <FocalPointPicker
              uri={uri}
              x={draft.x}
              y={draft.y}
              onChange={(nextX, nextY) => setDraft({ x: nextX, y: nextY })}
            />
          </ScrollView>

          <YStack gap="$2">
            <GlossButton tone="blue" btnSize="md" onPress={commit} accessibilityLabel={t('focalPoint.confirm')}>
              {t('focalPoint.confirm')}
            </GlossButton>
            {onCancel ? (
              <GlossButton tone="glass" btnSize="md" onPress={onCancel} accessibilityLabel={t('common.cancel')}>
                {t('common.cancel')}
              </GlossButton>
            ) : null}
          </YStack>
        </GlassPanel>
      </YStack>
    </Modal>
  )
}
