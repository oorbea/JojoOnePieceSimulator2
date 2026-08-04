import { Button, YStack, type ButtonProps } from 'tamagui'

import { asToken } from '@/shared/lib/tamagui-token'

import { GlossOverlay } from './gloss-overlay'
import { LensFlare } from './lens-flare'

export type GlossButtonTone =
  'blue' | 'red' | 'yellow' | 'green' | 'pink' | 'orange' | 'grape' | 'glass'

export type GlossButtonSize = 'sm' | 'md' | 'lg'
export type GlossButtonShape = 'pill' | 'card' | 'circle'

const TONE_STYLES: Record<
  GlossButtonTone,
  { bg: string; borderBottomColor: string; color: string }
> = {
  blue: { bg: '$wiiBlue', borderBottomColor: '$wiiBlueDeep', color: 'white' },
  red: { bg: '$strawHatRed', borderBottomColor: '$strawHatRedDeep', color: 'white' },
  yellow: { bg: '$sunYellow', borderBottomColor: '$sunYellowDeep', color: '$inkBlack' },
  green: { bg: '$meadowGreen', borderBottomColor: '$meadowGreenDeep', color: 'white' },
  pink: { bg: '$bubblegum', borderBottomColor: '$bubblegumDeep', color: 'white' },
  orange: { bg: '$tangerine', borderBottomColor: '$tangerineDeep', color: 'white' },
  grape: { bg: '$grapeSoda', borderBottomColor: '$grapeSodaDeep', color: 'white' },
  glass: { bg: '$glassFillStrong', borderBottomColor: '$glassEdgeInner', color: '$panelText' },
}

const SIZE_STYLES: Record<GlossButtonSize, { height: number; px: string; fontSize: string }> = {
  sm: { height: 40, px: '$3.5', fontSize: '$3' },
  md: { height: 52, px: '$5', fontSize: '$5' },
  lg: { height: 64, px: '$6', fontSize: '$6' },
}

const SHAPE_STYLES: Record<GlossButtonShape, { rounded: string; aspectRatio?: number }> = {
  pill: { rounded: '$pill' },
  card: { rounded: '$card' },
  circle: { rounded: '$circle', aspectRatio: 1 },
}

type GlossButtonProps = Omit<ButtonProps, 'size'> & {
  tone?: GlossButtonTone
  /** Named `btnSize`, not `size` — `size` is a Tamagui Button prop, and
   * `scale` collides with the built-in transform shorthand. */
  btnSize?: GlossButtonSize
  shape?: GlossButtonShape
  /** Renders a single lens-flare glow behind this button. Use on ONE
   * primary CTA per screen — never on every button. */
  flare?: boolean
}

// Physical 3D Wii/iOS button: a hard offset "lip" in the tone's deep shade
// that compresses on press while the button drops and scales down, plus a
// bounce on hover. Built on Tamagui's Button (not styled(Button)) so its
// icon slot, disabled semantics, and focus ring keep working exactly as
// Tamagui intends — only the visual recipe is layered on top.
export function GlossButton({
  tone = 'blue',
  btnSize = 'md',
  shape = 'pill',
  flare = false,
  children,
  disabled,
  ...rest
}: GlossButtonProps) {
  const toneStyle = TONE_STYLES[tone]
  const sizeStyle = SIZE_STYLES[btnSize]
  const shapeStyle = SHAPE_STYLES[shape]

  return (
    <YStack position="relative" items="center" justify="center">
      {flare && !disabled ? <LensFlare size={btnSize === 'lg' ? 'md' : 'sm'} /> : null}
      <Button
        {...rest}
        disabled={disabled}
        overflow="hidden"
        position="relative"
        height={sizeStyle.height}
        px={asToken(sizeStyle.px)}
        fontSize={asToken(sizeStyle.fontSize)}
        rounded={asToken(shapeStyle.rounded)}
        aspectRatio={shapeStyle.aspectRatio}
        borderWidth={1.5}
        borderColor="$glassEdge"
        borderBottomWidth={5}
        bg={asToken(toneStyle.bg)}
        borderBottomColor={asToken(toneStyle.borderBottomColor)}
        color={asToken(toneStyle.color)}
        fontWeight="800"
        shadowColor="$hardShadow"
        shadowOffset={{ width: 0, height: 8 }}
        shadowRadius={18}
        shadowOpacity={1}
        elevation={5}
        transition="bouncy"
        hoverStyle={{ scale: 1.05, y: -2, shadowRadius: 26 }}
        pressStyle={{ scale: 0.94, y: 3, borderBottomWidth: 2, shadowRadius: 6 }}
        focusStyle={{ outlineColor: '$channelActive', outlineWidth: 3, outlineStyle: 'solid' }}
        disabledStyle={{ opacity: 0.55, borderBottomWidth: 3 }}
        accessibilityRole="button"
      >
        <GlossOverlay coverage="half" shape={shape === 'circle' ? 'circle' : 'pill'} />
        {children}
      </Button>
    </YStack>
  )
}
