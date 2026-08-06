import { Plus, X } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'

import { GlassField } from './glass-field'
import { GlassPanel } from './glass-panel'
import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'

/**
 * Props for the dynamic skills input used by the admin forms.
 */
type SkillsFieldProps = {
  label: string
  skills: string[]
  onAdd: (skill: string) => void
  onRemove: (index: number) => void
  error?: string
}

/**
 * The one dynamic string[] input the admin forms need (Power.skills).
 * Kept as a small compound component rather than a generic "array field" —
 * it owns its own draft-text state for the pending entry, while the array
 * itself stays fully controlled through onAdd/onRemove so the parent form
 * (react-hook-form via useController) is the single source of truth.
 */
export function SkillsField({ label, skills, onAdd, onRemove, error }: SkillsFieldProps) {
  const [draft, setDraft] = useState('')

  const commit = () => {
    const trimmed = draft.trim()
    if (!trimmed) return
    onAdd(trimmed)
    setDraft('')
  }

  return (
    <YStack width="100%" gap="$2">
      {skills.length > 0 ? (
        <XStack flexWrap="wrap" gap="$2">
          {skills.map((skill, index) => (
            <GlassPanel key={`${skill}-${index}`} tone="plastic" px="$3" py="$1.5" rounded="$pill" elevate={0}>
              <XStack items="center" gap="$2">
                <GlowText level="label">{skill}</GlowText>
                <XStack
                  onPress={() => onRemove(index)}
                  p="$0.5"
                  {...a11yProps(`Remove ${skill}`, 'button')}
                >
                  <X size={14} color="$panelTextSoft" />
                </XStack>
              </XStack>
            </GlassPanel>
          ))}
        </XStack>
      ) : null}

      <XStack gap="$2" items="flex-end">
        <YStack flex={1}>
          <GlassField
            label={label}
            value={draft}
            onChangeText={setDraft}
            placeholder="Add a skill…"
            onSubmitEditing={commit}
            returnKeyType="done"
          />
        </YStack>
        <GlossButton tone="green" btnSize="md" shape="circle" onPress={commit} accessibilityLabel="Add skill">
          <Plus size={20} color="white" />
        </GlossButton>
      </XStack>

      {error ? (
        <GlowText level="label" color="$strawHatRedDeep" {...a11yProps(error, 'alert')}>
          {error}
        </GlowText>
      ) : null}
    </YStack>
  )
}
