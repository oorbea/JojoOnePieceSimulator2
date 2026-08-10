import { forwardRef } from 'react'
import { Input, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'

import { GlowText } from './glow-text'
import { InsetRing } from './wii-card'

type GlassFieldProps = Omit<React.ComponentProps<typeof Input>, 'onChangeText'> & {
  label: string
  error?: string
  onChangeText?: (text: string) => void
}

// The one text-input primitive the design system was still missing —
// visible label always (never placeholder-only), the same recessed-glass
// inset ring every other surface uses, a visible focus ring, and an error
// rendered directly under the field (never batched at the top of a form).
export const GlassField = forwardRef<React.ElementRef<typeof Input>, GlassFieldProps>(
  function GlassField({ label, error, id, height, multiline, ...rest }, ref) {
    return (
      <YStack width="100%" gap="$1.5">
        <GlowText level="label">{label}</GlowText>
        <YStack position="relative" rounded="$card" overflow="hidden">
          <InsetRing rounded="$card" />
          <Input
            ref={ref}
            multiline={multiline}
            {...rest}
            id={id}
            bg="$plasticFill"
            borderWidth={1.5}
            borderColor={error ? '$strawHatRedDeep' : '$glassEdge'}
            rounded="$card"
            height={height ?? 52}
            px="$4"
            fontSize="$5"
            fontFamily="$body"
            color="$panelText"
            placeholderTextColor="$panelTextSoft"
            focusStyle={{ outlineColor: '$channelActive', outlineWidth: 3, outlineStyle: 'solid' }}
            {...(multiline ? { textAlignVertical: 'top' as const, pt: '$2' } : {})}
            {...a11yProps(label)}
          />
        </YStack>
        {error ? (
          <GlowText level="label" color="$strawHatRedDeep" {...a11yProps(error, 'alert')}>
            {error}
          </GlowText>
        ) : null}
      </YStack>
    )
  }
)
