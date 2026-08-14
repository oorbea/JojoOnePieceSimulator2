import { useRef, useState } from 'react'
import { Modal } from 'react-native'
import { YStack } from 'tamagui'

import { isWeb } from '@/shared/lib/web-blur'

import { GlassPanel } from './glass-panel'
import { GlowText } from './glow-text'

const NATIVE_AUTO_HIDE_MS = 1800

type Anchor = { x: number; y: number; width: number; height: number }

type Measurable = {
  measure?: (
    callback: (x: number, y: number, width: number, height: number, pageX: number, pageY: number) => void
  ) => void
}

// Cross-platform tooltip trigger: web shows on hover or keyboard focus (a
// mouse/keyboard user has no long-press affordance); native has no hover at
// all, so it shows on long-press with an auto-hide timer instead.
//
// The bubble itself renders through a root-level `Modal` (see
// `TooltipBubble` below), anchored to the trigger's on-screen position via
// `.measure()` - NOT a nested `position:absolute` inside the trigger's own
// tree. RN only compares z-index between direct siblings (the exact reason
// `ConfirmSheet` is a real `Modal` instead of an absolute overlay, per
// ObsidianVault/frontend-responsive-frutiger-aero.md) - a tooltip a few
// levels deep inside a form would otherwise render clipped by an ancestor's
// `overflow:hidden` or painted over by a later, unrelated sibling.
export function useTooltipTrigger(label?: string) {
  const ref = useRef<Measurable | null>(null)
  const [visible, setVisible] = useState(false)
  const [anchor, setAnchor] = useState<Anchor | null>(null)
  const hideTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  if (!label) {
    return { visible: false, anchor: null, triggerRef: ref, triggerProps: {} as Record<string, unknown> }
  }

  const clearHideTimer = () => {
    if (hideTimer.current) {
      clearTimeout(hideTimer.current)
      hideTimer.current = null
    }
  }

  const show = () => {
    ref.current?.measure?.((_x, _y, width, height, pageX, pageY) => {
      setAnchor({ x: pageX, y: pageY, width, height })
      setVisible(true)
    })
  }

  const triggerProps = isWeb
    ? {
        onHoverIn: show,
        onHoverOut: () => setVisible(false),
        onFocus: show,
        onBlur: () => setVisible(false),
      }
    : {
        onLongPress: () => {
          show()
          clearHideTimer()
          hideTimer.current = setTimeout(() => setVisible(false), NATIVE_AUTO_HIDE_MS)
        },
      }

  return { visible, anchor, triggerRef: ref, triggerProps }
}

type TooltipBubbleProps = { visible: boolean; label?: string; anchor: Anchor | null }

// Floats above EVERYTHING (its own `Modal` layer), anchored to the
// trigger's measured screen position. Centering above the anchor uses a
// web-only `transform` (RN's native transform doesn't take percentage
// values, so native anchors from the trigger's top-left corner instead -
// fine for the short labels these carry). Inert to the pointer throughout,
// so it never eats the hover/press meant for the button underneath.
export function TooltipBubble({ visible, label, anchor }: TooltipBubbleProps) {
  if (!visible || !label || !anchor) return null

  // pointerEvents="none" on Modal itself (not just its children): RNW's Modal
  // renders a fixed, full-viewport wrapper div with no pointer-events style of
  // its own, so it swallows hover/click for the whole screen even though the
  // bubble content underneath is already inert - that stole hover off the
  // trigger button the instant the tooltip opened, flipping visible on/off
  // forever and blocking every click behind it.
  return (
    <Modal visible transparent animationType="none" statusBarTranslucent pointerEvents="none">
      <YStack flex={1} style={{ pointerEvents: 'none' }}>
        <YStack
          position="absolute"
          t={Math.max(anchor.y - 8, 4)}
          l={anchor.x + anchor.width / 2}
          maxW={220}
          style={{
            pointerEvents: 'none',
            ...(isWeb ? { transform: [{ translateX: '-50%' }, { translateY: '-100%' }] } : null),
          }}
        >
          <GlassPanel tone="strong" elevate={2} px="$2.5" py="$1.5" rounded="$pill">
            <GlowText level="label" fontSize="$1" numberOfLines={2}>
              {label}
            </GlowText>
          </GlassPanel>
        </YStack>
      </YStack>
    </Modal>
  )
}
