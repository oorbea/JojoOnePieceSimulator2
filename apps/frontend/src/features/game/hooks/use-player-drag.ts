import { useEffect, useMemo, useRef, useState } from 'react'
import {
  PanResponder,
  Platform,
  type GestureResponderEvent,
  type PanResponderGestureState,
} from 'react-native'

const DRAG_ACTIVATION_DISTANCE = 6

export type DragEndInfo = { pageX: number; pageY: number }

// Props to spread on the drag handle `View`. On web that is a callback ref
// (the hook attaches DOM pointer listeners to the node); on native it is the
// PanResponder's handlers.
type HandleProps = Record<string, unknown>

// Native: `PanResponder` (RN core, no `react-native-gesture-handler` - see
// game-lobby-todo.md §5's note on why that's deliberately avoided here).
//
// Deliberately plain `useState`, not RN's `Animated.ValueXY` - this repo's
// eslint config runs `react-hooks/refs` (flags "accessing a ref value
// during render"), and Animated's imperative API only works by handing out
// a `useRef(...).current` instance and mutating/reading it outside the
// normal render data-flow. `PanResponder.create` itself is a plain memoized
// value (`useMemo`, not `useRef`), so its callbacks can close over
// `enabled`/`onDragEnd` directly with no ref indirection needed either.
function usePlayerDragNative(enabled: boolean, onDragEnd: (info: DragEndInfo) => void) {
  const [translate, setTranslate] = useState({ x: 0, y: 0 })

  const panResponder = useMemo(
    () =>
      PanResponder.create({
        // Never claim on the bare touch-down - only after real movement
        // past the threshold. Claiming on start would steal every tap meant
        // for a nested `GlossButton` (kick/transfer-host).
        onStartShouldSetPanResponder: () => false,
        onMoveShouldSetPanResponder: (
          _evt: GestureResponderEvent,
          gesture: PanResponderGestureState
        ) =>
          enabled &&
          (Math.abs(gesture.dx) > DRAG_ACTIVATION_DISTANCE ||
            Math.abs(gesture.dy) > DRAG_ACTIVATION_DISTANCE),
        onPanResponderMove: (_evt: GestureResponderEvent, gesture: PanResponderGestureState) =>
          setTranslate({ x: gesture.dx, y: gesture.dy }),
        onPanResponderRelease: (evt: GestureResponderEvent) => {
          setTranslate({ x: 0, y: 0 })
          onDragEnd({ pageX: evt.nativeEvent.pageX, pageY: evt.nativeEvent.pageY })
        },
        onPanResponderTerminate: () => setTranslate({ x: 0, y: 0 }),
      }),
    [enabled, onDragEnd]
  )

  const handleProps: HandleProps = enabled ? { ...panResponder.panHandlers } : {}
  return { translate, handleProps }
}

// Web: react-native-web's responder system never negotiated this gesture
// (verified live: `onMoveShouldSetPanResponder` was not called for a mouse
// drag started on the handle, and from the row itself the browser's native
// text-selection `dragstart` cancelled it). Plain DOM pointer events with
// pointer capture are reliable for mouse and touch alike (the handle sets
// `touch-action: none`, so the page scroll doesn't take a touch that starts
// on it) and report `clientX/Y`, the same space `.measure()` returns.
function usePlayerDragWeb(enabled: boolean, onDragEnd: (info: DragEndInfo) => void) {
  const [translate, setTranslate] = useState({ x: 0, y: 0 })
  const [node, setNode] = useState<HTMLElement | null>(null)
  // The latest callback, read only inside the DOM handlers: the caller passes
  // a fresh inline function every render and re-binding the listeners would
  // drop an in-flight gesture.
  const onDragEndRef = useRef(onDragEnd)
  useEffect(() => {
    onDragEndRef.current = onDragEnd
  })

  useEffect(() => {
    if (!enabled || !node) return
    let start: { id: number; x: number; y: number } | null = null
    let active = false

    const onDown = (e: PointerEvent) => {
      if (e.pointerType === 'mouse' && e.button !== 0) return
      start = { id: e.pointerId, x: e.clientX, y: e.clientY }
      active = false
      node.setPointerCapture(e.pointerId)
    }
    const onMove = (e: PointerEvent) => {
      if (!start || e.pointerId !== start.id) return
      const dx = e.clientX - start.x
      const dy = e.clientY - start.y
      if (!active) {
        if (Math.abs(dx) <= DRAG_ACTIVATION_DISTANCE && Math.abs(dy) <= DRAG_ACTIVATION_DISTANCE) {
          return
        }
        active = true
      }
      setTranslate({ x: dx, y: dy })
    }
    const finish = (e: PointerEvent, dropped: boolean) => {
      if (!start || e.pointerId !== start.id) return
      const wasActive = active
      start = null
      active = false
      if (node.hasPointerCapture(e.pointerId)) node.releasePointerCapture(e.pointerId)
      setTranslate({ x: 0, y: 0 })
      if (dropped && wasActive) onDragEndRef.current({ pageX: e.clientX, pageY: e.clientY })
    }
    const onUp = (e: PointerEvent) => finish(e, true)
    const onCancel = (e: PointerEvent) => finish(e, false)

    node.addEventListener('pointerdown', onDown)
    node.addEventListener('pointermove', onMove)
    node.addEventListener('pointerup', onUp)
    node.addEventListener('pointercancel', onCancel)
    return () => {
      node.removeEventListener('pointerdown', onDown)
      node.removeEventListener('pointermove', onMove)
      node.removeEventListener('pointerup', onUp)
      node.removeEventListener('pointercancel', onCancel)
    }
  }, [enabled, node])

  const handleProps: HandleProps = enabled ? { ref: setNode } : {}
  return { translate, handleProps }
}

export const usePlayerDrag = Platform.OS === 'web' ? usePlayerDragWeb : usePlayerDragNative
