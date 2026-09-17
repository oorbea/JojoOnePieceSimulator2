import { Minus, Plus } from '@tamagui/lucide-icons-2'
import { useRef, useState, type ElementRef, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import type { NativeSyntheticEvent, TextInputKeyPressEventData } from 'react-native'
import { Input, XStack } from 'tamagui'

import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { SettingRow } from '@/shared/components/presentational/setting-row'
import { a11yProps } from '@/shared/lib/a11y'

type Props = {
  label: string
  value: number
  min: number
  max: number
  onChange: (value: number) => void
  help?: ReactNode
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n))
}

// Shared -/+ stepper used by both the create-lobby form and the config-edit
// panel (team size, voting window) - promoted out of create-lobby-screen.tsx
// so a second screen doesn't duplicate it. Built on `SettingRow` so its
// label+control alignment matches every other field in both forms.
// `stacked`: the owner reported the label sitting far from the stepper on
// wide screens (`SettingRow`'s `$md` row layout spreads them to a
// `flexBasis:320` column's two ends) - a stepper reads better with its
// label directly above it at every breakpoint, unlike a row-style toggle.
//
// The value itself is also a pressable that swaps to a digits-only text
// input - typing 47 is faster than 37 taps of `+`. Digit-stripping mirrors
// the existing pattern in stage-form-modal.tsx/character-form-modal.tsx
// (`text.replace(/[^0-9]/g, '')`); an out-of-range or empty commit clamps to
// the nearest bound / reverts, never shows an error - the user asked for
// silent clamping, not a validation state.
export function NumberStepper({ label, value, min, max, onChange, help }: Props) {
  const { t } = useTranslation()
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState('')
  const inputRef = useRef<ElementRef<typeof Input>>(null)
  // `onSubmitEditing` (Enter) unmounts the input, which itself fires a
  // trailing `onBlur` for the element losing focus - without this guard
  // that would run `commit` a second time. Reset whenever a fresh edit
  // starts.
  const committedRef = useRef(false)

  function startEditing() {
    committedRef.current = false
    setDraft(String(value))
    setEditing(true)
  }

  function commit() {
    if (committedRef.current) return
    committedRef.current = true
    const digits = draft.replace(/[^0-9]/g, '')
    const parsed = digits === '' ? value : clamp(Number(digits), min, max)
    setEditing(false)
    if (parsed !== value) onChange(parsed)
  }

  function cancel() {
    committedRef.current = true
    setEditing(false)
  }

  function handleKeyPress(e: NativeSyntheticEvent<TextInputKeyPressEventData>) {
    if (e.nativeEvent.key === 'Escape') cancel()
  }

  return (
    <SettingRow label={label} help={help} stacked>
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
        {editing ? (
          <Input
            ref={inputRef as never}
            value={draft}
            onChangeText={(text) => setDraft(text.replace(/[^0-9]/g, ''))}
            onSubmitEditing={commit}
            onBlur={commit}
            onKeyPress={handleKeyPress as never}
            keyboardType="number-pad"
            inputMode="numeric"
            autoFocus
            selectTextOnFocus
            width={64}
            height={40}
            px="$2"
            textAlign="center"
            fontSize="$5"
            fontFamily="$body"
            fontWeight="700"
            bg="$plasticFill"
            borderWidth={1.5}
            borderColor="$glassEdge"
            rounded="$card"
            color="$panelText"
            focusStyle={{ outlineColor: '$channelActive', outlineWidth: 3, outlineStyle: 'solid' }}
            {...a11yProps(`${label}: ${t('game.create.editValueHint')}`, 'spinbutton')}
          />
        ) : (
          <GlossButton
            tone="glass"
            btnSize="sm"
            shape="card"
            minW={64}
            onPress={startEditing}
            accessibilityLabel={`${label}: ${value}`}
            tooltip={t('game.create.editValueHint')}
          >
            <GlowText level="heading">{value}</GlowText>
          </GlossButton>
        )}
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
