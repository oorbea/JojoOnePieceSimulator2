import { forwardRef } from 'react'
import { TextInput as RNTextInput } from 'react-native'
import { Input, useTheme, YStack } from 'tamagui'

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
  function GlassField({ label, error, id, height, multiline, value, onChangeText, ...rest }, ref) {
    const theme = useTheme()

    return (
      <YStack width="100%" gap="$1.5">
        <GlowText level="label">{label}</GlowText>
        <YStack position="relative" rounded="$card" overflow="hidden">
          <InsetRing rounded="$card" />
          {multiline ? (
            // Tamagui's web Input silently drops `multiline` and always renders a
            // single-line <input>, so a real RN TextInput is used here instead -
            // on web react-native-web maps it to a scrollable <textarea>.
            <RNTextInput
              ref={ref as never}
              value={value as string | undefined}
              onChangeText={onChangeText}
              multiline
              textAlignVertical="top"
              scrollEnabled
              placeholderTextColor={theme.panelTextSoft?.val}
              style={{
                height: (height as number) ?? 90,
                paddingHorizontal: 16,
                paddingTop: 10,
                fontSize: 16,
                fontFamily: 'Nunito, Nunito_500Medium, Nunito_700Bold, sans-serif',
                color: theme.panelText?.val,
                backgroundColor: theme.plasticFill?.val,
                borderWidth: 1.5,
                borderColor: error ? theme.strawHatRedDeep?.val : theme.glassEdge?.val,
                borderRadius: 22,
              }}
              {...a11yProps(label)}
              {...(rest as object)}
            />
          ) : (
            <Input
              ref={ref}
              value={value}
              onChangeText={onChangeText}
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
              {...a11yProps(label)}
            />
          )}
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
