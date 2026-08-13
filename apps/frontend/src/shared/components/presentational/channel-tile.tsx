import { Lock } from '@tamagui/lucide-icons-2'
import { YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { asToken } from '@/shared/lib/tamagui-token'

import { GlossOverlay } from './gloss-overlay'
import { GlowText } from './glow-text'
import { TooltipBubble, useTooltipTrigger } from './tooltip'
import { InsetRing } from './wii-card'

export type ChannelTileTone = 'blue' | 'grape' | 'red' | 'yellow' | 'green' | 'pink'

// The icon components lucide-icons-2 exports (Home, Sparkles, Apple, Zap,
// Lock, ...) all share this prop shape — declared locally instead of
// depending on @tamagui/helpers-icon directly, since that package isn't
// hoisted to a resolvable path outside lucide-icons-2's own node_modules.
// `color` is typed `any` because the real lucide-icons-2 prop constrains it
// to its own token union, which would make this type incompatible with
// every actual icon component if declared as `string`.
export type IconComponent = React.ComponentType<{
  size?: number
  color?: any
  strokeWidth?: number
}>

const TONE_BG: Record<ChannelTileTone, string> = {
  blue: '$wiiBlue',
  grape: '$grapeSoda',
  red: '$strawHatRed',
  yellow: '$sunYellow',
  green: '$meadowGreen',
  pink: '$bubblegum',
}

type ChannelTileProps = {
  label: string
  tone: ChannelTileTone
  icon: IconComponent
  onPress?: () => void
  /** Renders as a desaturated, disabled "coming soon" slot — honest about
   * what's actually built this pass instead of faking dead navigation. */
  locked?: boolean
}

// The signature element: curved jelly-glass channel tile. Solid tone base,
// gloss cut across the top half, bright inner lip, hover bounce, physical
// press compression — the same recipe GlossButton uses, so the whole UI
// feels consistent under the finger.
export function ChannelTile({
  label,
  tone,
  icon: Icon,
  onPress,
  locked = false,
}: ChannelTileProps) {
  const bg = locked ? '$plasticEdge' : TONE_BG[tone]
  const tooltipLabel = locked ? `${label}, coming soon` : label
  const { visible: tooltipVisible, triggerProps } = useTooltipTrigger(tooltipLabel)

  return (
    <YStack
      {...triggerProps}
      width={96}
      height={96}
      $md={{ width: 112, height: 112 }}
      $lg={{ width: 128, height: 128 }}
      rounded="$card"
      overflow="hidden"
      position="relative"
      items="center"
      justify="center"
      gap="$1.5"
      bg={asToken(bg)}
      borderWidth={1.5}
      borderColor="$glassEdge"
      shadowColor="$softShadow"
      shadowOffset={{ width: 0, height: 8 }}
      shadowRadius={18}
      shadowOpacity={1}
      elevation={5}
      opacity={locked ? 0.65 : 1}
      disabled={locked}
      onPress={locked ? undefined : onPress}
      transition={locked ? undefined : 'bouncy'}
      hoverStyle={locked ? undefined : { scale: 1.05, y: -4 }}
      pressStyle={locked ? undefined : { scale: 0.94, y: 2 }}
      cursor={locked ? 'default' : 'pointer'}
      {...a11yProps(
        locked ? `${label}, coming soon` : label,
        locked ? 'none' : 'button',
        locked ? { disabled: true } : undefined
      )}
    >
      <InsetRing rounded="$card" />
      <TooltipBubble visible={tooltipVisible} label={tooltipLabel} />
      {!locked ? <GlossOverlay coverage="half" shape="card" /> : null}
      {locked ? (
        <Lock size={22} color="$panelTextSoft" />
      ) : (
        <Icon size={26} color="white" strokeWidth={2.5} />
      )}
      <GlowText level="label" tone={locked ? 'soft' : 'onColor'} align="center" fontSize="$2">
        {label}
      </GlowText>
    </YStack>
  )
}
