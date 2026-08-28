import { useMemo, useState } from 'react'
import { PanResponder, type GestureResponderEvent, type PanResponderGestureState } from 'react-native'

const DRAG_ACTIVATION_DISTANCE = 6

export type DragEndInfo = { pageX: number; pageY: number }

// `PanResponder` is RN-core: it already unifies mouse-drag-as-pointer (web)
// and touch-drag (native) behind one gesture API, so "drag-to-move works on
// both desktop and mobile" doesn't need two implementations or
// `react-native-gesture-handler` (installed but unused elsewhere in this
// repo - see game-lobby-todo.md §5's note on why that's deliberately
// avoided here).
//
// Deliberately plain `useState`, not RN's `Animated.ValueXY` - this repo's
// eslint config runs `react-hooks/refs` (flags "accessing a ref value
// during render"), and Animated's imperative API only works by handing out
// a `useRef(...).current` instance and mutating/reading it outside the
// normal render data-flow - any further use of that value (member access,
// returning it from the hook) trips the rule. `PanResponder.create` itself
// is a plain memoized value (`useMemo`, not `useRef`), so its callbacks can
// close over `enabled`/`onDragEnd` directly with no ref indirection needed
// either - recreated when either changes, which is cheap for a handful of
// roster rows.
export function usePlayerDrag(enabled: boolean, onDragEnd: (info: DragEndInfo) => void) {
  const [translate, setTranslate] = useState({ x: 0, y: 0 })

  const panResponder = useMemo(
    () =>
      PanResponder.create({
        // Never claim on the bare touch/mouse-down - only after real
        // movement past the threshold. Claiming on start would steal every
        // tap meant for a nested `GlossButton` (kick/transfer-host) since
        // RN's responder negotiation favors a parent that wants it
        // immediately over a child that hasn't moved yet.
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

  return { translate, panHandlers: enabled ? panResponder.panHandlers : {} }
}
