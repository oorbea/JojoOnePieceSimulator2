import { Minus, Plus } from '@tamagui/lucide-icons-2'
import type { ReactNode } from 'react'
import { XStack } from 'tamagui'

import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { SettingRow } from '@/shared/components/presentational/setting-row'

type Props = {
  label: string
  value: number
  min: number
  max: number
  onChange: (value: number) => void
  help?: ReactNode
}

// Shared -/+ stepper used by both the create-lobby form and the config-edit
// panel (team size, voting window) - promoted out of create-lobby-screen.tsx
// so a second screen doesn't duplicate it. Built on `SettingRow` so its
// label+control alignment matches every other field in both forms.
export function NumberStepper({ label, value, min, max, onChange, help }: Props) {
  return (
    <SettingRow label={label} help={help}>
      <XStack items="center" gap="$3">
        <GlossButton
          tone="glass"
          btnSize="sm"
          shape="circle"
          disabled={value <= min}
          onPress={() => onChange(Math.max(min, value - 1))}
          accessibilityLabel={`${label} -`}
        >
          <Minus size={16} color="$panelText" />
        </GlossButton>
        <GlowText level="heading">{value}</GlowText>
        <GlossButton
          tone="glass"
          btnSize="sm"
          shape="circle"
          disabled={value >= max}
          onPress={() => onChange(Math.min(max, value + 1))}
          accessibilityLabel={`${label} +`}
        >
          <Plus size={16} color="$panelText" />
        </GlossButton>
      </XStack>
    </SettingRow>
  )
}
