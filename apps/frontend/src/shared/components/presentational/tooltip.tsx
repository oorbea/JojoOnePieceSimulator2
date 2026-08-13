import { useRef, useState } from 'react'

import { isWeb } from '@/shared/lib/web-blur'

import { GlassPanel } from './glass-panel'
import { GlowText } from './glow-text'

const NATIVE_AUTO_HIDE_MS = 1800

// Cross-platform tooltip trigger: web shows on hover or keyboard focus (a
// mouse/keyboard user has no long-press affordance); native has no hover at
// all, so it shows on long-press with an auto-hide timer instead. The
// returned `triggerProps` are meant to be spread onto the SAME pressable
// that already owns `onPress` - wrapping it in an extra Pressable would
// steal the tap, the same reasoning as this project's `pointerEvents`/
// a11yProps ownership rule (see ObsidianVault/a11y-web-leak.md).
export function useTooltipTrigger(label?: string) {
  const [visible, setVisible] = useState(false)
  const hideTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const clearHideTimer = () => {
    if (hideTimer.current) {
      clearTimeout(hideTimer.current)
      hideTimer.current = null
    }
  }

  if (!label) {
    return { visible: false, triggerProps: {} as Record<string, unknown> }
  }

  const triggerProps = isWeb
    ? {
        onHoverIn: () => setVisible(true),
        onHoverOut: () => setVisible(false),
        onFocus: () => setVisible(true),
        onBlur: () => setVisible(false),
      }
    : {
        onLongPress: () => {
          setVisible(true)
          clearHideTimer()
          hideTimer.current = setTimeout(() => setVisible(false), NATIVE_AUTO_HIDE_MS)
        },
      }

  return { visible, triggerProps }
}

type TooltipBubbleProps = { visible: boolean; label?: string }

// The bubble - absolutely positioned above its trigger, inert to the
// pointer (`style.pointerEvents`, never the top-level prop) so it can never
// eat the hover/press meant for the button underneath it. Horizontal
// centering uses a web-only `transform` (RN's native transform doesn't take
// percentage values) - on native it anchors from the trigger's horizontal
// midpoint instead, close enough for the short labels these carry.
export function TooltipBubble({ visible, label }: TooltipBubbleProps) {
  if (!visible || !label) return null

  return (
    <GlassPanel
      tone="strong"
      elevate={0}
      px="$2.5"
      py="$1.5"
      rounded="$pill"
      position="absolute"
      b="100%"
      l="50%"
      mb="$2"
      z="$overlay"
      maxW={220}
      style={{
        pointerEvents: 'none',
        ...(isWeb ? ({ transform: [{ translateX: '-50%' }] } as object) : null),
      }}
    >
      <GlowText level="label" fontSize="$1" numberOfLines={2}>
        {label}
      </GlowText>
    </GlassPanel>
  )
}
