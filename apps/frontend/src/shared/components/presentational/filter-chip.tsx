import { Check } from '@tamagui/lucide-icons-2'
import { XStack } from 'tamagui'

import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'

type Props = {
  label: string
  active: boolean
  onToggle: () => void
  /** Tooltip text - every button gets one (project norm), and for a chip the
   * label alone rarely says what turning it on does. */
  tooltip: string
}

// On/off filter pill that sits in the same row as the GlassSelect filters.
// Same active/inactive vocabulary as the lobby's manga toggles (blue + check
// when on, glass when off) so it reads as one control family; exposed as a
// checkbox since it holds a boolean, not a one-of-many choice.
export function FilterChip({ label, active, onToggle, tooltip }: Props) {
  return (
    <GlossButton
      tone={active ? 'blue' : 'glass'}
      btnSize="sm"
      onPress={onToggle}
      accessibilityLabel={label}
      tooltip={tooltip}
      a11yRole="checkbox"
      a11yChecked={active}
    >
      <XStack items="center" gap="$1.5">
        {active ? <Check size={14} color="white" /> : null}
        <GlowText level="label">{label}</GlowText>
      </XStack>
    </GlossButton>
  )
}
