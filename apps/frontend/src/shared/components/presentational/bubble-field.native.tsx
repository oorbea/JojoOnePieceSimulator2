import { useEffect } from 'react'
import { useWindowDimensions } from 'react-native'
import Animated, {
  Easing,
  interpolate,
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withDelay,
  withRepeat,
  withTiming,
} from 'react-native-reanimated'
import { YStack } from 'tamagui'

// Same deterministic geometry as the web variant (see bubble-field.web.tsx)
// so both platforms read as the same scene.
const BUBBLE_GEOMETRY = [
  { size: 22, leftPct: 6, dur: 18000, delay: 0 },
  { size: 14, leftPct: 16, dur: 14000, delay: 2000 },
  { size: 30, leftPct: 26, dur: 22000, delay: 4000 },
  { size: 18, leftPct: 38, dur: 16000, delay: 1000 },
  { size: 26, leftPct: 48, dur: 20000, delay: 6000 },
  { size: 12, leftPct: 58, dur: 13000, delay: 3000 },
  { size: 34, leftPct: 68, dur: 24000, delay: 5000 },
  { size: 16, leftPct: 76, dur: 15000, delay: 7000 },
  { size: 28, leftPct: 84, dur: 19000, delay: 2500 },
  { size: 20, leftPct: 91, dur: 17000, delay: 8000 },
  { size: 24, leftPct: 12, dur: 21000, delay: 500 },
  { size: 15, leftPct: 55, dur: 12000, delay: 9000 },
  { size: 32, leftPct: 96, dur: 23000, delay: 1500 },
  { size: 19, leftPct: 33, dur: 16000, delay: 4500 },
  { size: 27, leftPct: 72, dur: 20000, delay: 6500 },
  { size: 13, leftPct: 44, dur: 14000, delay: 0 },
]

type BubbleProps = { size: number; leftPct: number; dur: number; delay: number; height: number }

function Bubble({ size, leftPct, dur, delay, height }: BubbleProps) {
  const progress = useSharedValue(0)
  const reducedMotion = useReducedMotion()

  useEffect(() => {
    if (reducedMotion) return
    progress.value = withDelay(
      delay,
      withRepeat(withTiming(1, { duration: dur, easing: Easing.linear }), -1, false)
    )
  }, [delay, dur, progress, reducedMotion])

  const style = useAnimatedStyle(() => {
    if (reducedMotion) return { opacity: 0.3 }
    return {
      transform: [
        { translateY: interpolate(progress.value, [0, 1], [height * 0.1, -height * 0.12]) },
        { translateX: Math.sin(progress.value * Math.PI * 2) * 22 },
        { scale: interpolate(progress.value, [0, 1], [0.7, 1.05]) },
      ],
      opacity: interpolate(progress.value, [0, 0.1, 0.9, 1], [0, 0.9, 0.9, 0]),
    }
  })

  return (
    <Animated.View
      style={[
        {
          position: 'absolute',
          bottom: 0,
          left: `${leftPct}%`,
          width: size,
          height: size,
        },
        style,
      ]}
    >
      <YStack
        flex={1}
        rounded="$circle"
        bg="$bubbleFill"
        borderWidth={1.5}
        borderColor="$bubbleEdge"
      >
        <YStack
          position="absolute"
          t="18%"
          l="18%"
          width="30%"
          height="30%"
          rounded="$circle"
          bg="rgba(255,255,255,0.75)"
        />
      </YStack>
    </Animated.View>
  )
}

type BubbleFieldProps = { count: number }

// Reanimated worklets driving translateY/translateX/scale/opacity — the
// whole animation runs on the UI thread, so there is zero per-frame JS cost
// and zero React re-renders once mounted.
export function BubbleField({ count }: BubbleFieldProps) {
  const { height } = useWindowDimensions()
  const bubbles = BUBBLE_GEOMETRY.slice(0, count)

  return (
    <>
      {bubbles.map((bubble, index) => (
        <Bubble key={index} {...bubble} height={height} />
      ))}
    </>
  )
}
