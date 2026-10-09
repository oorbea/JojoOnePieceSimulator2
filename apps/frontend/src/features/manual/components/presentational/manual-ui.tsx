import { createContext, useContext } from 'react'
import type { LayoutChangeEvent } from 'react-native'
import { Paragraph, XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { MeterBar } from '@/shared/components/presentational/meter-bar'

// Small building blocks shared by the manual's sections, composed from the
// app's own glass primitives so the manual reads like the rest of the app.

type SectionCardProps = {
  /** Testing/anchor handle: also what `?section=` deep-links to. */
  id: string
  title: string
  onMeasure?: (y: number) => void
  children: React.ReactNode
}

// Inside the overlay the modal already provides the card and the title, so the
// same section components render bare: just their content.
export const ManualBareContext = createContext(false)

export function SectionCard({ id, title, onMeasure, children }: SectionCardProps) {
  const bare = useContext(ManualBareContext)
  if (bare) {
    return (
      <YStack width="100%" gap="$4" testID={`manual-section-${id}`}>
        {children}
      </YStack>
    )
  }
  return (
    <GlassPanel
      glossy
      elevate={1}
      width="100%"
      p="$5"
      gap="$4"
      testID={`manual-section-${id}`}
      onLayout={(e: LayoutChangeEvent) => onMeasure?.(e.nativeEvent.layout.y)}
    >
      <GlowText level="heading" role="heading">
        {title}
      </GlowText>
      {children}
    </GlassPanel>
  )
}

export function Body({ children }: { children: React.ReactNode }) {
  return (
    <Paragraph color="$panelText" fontSize="$4" maxW={680}>
      {children}
    </Paragraph>
  )
}

export function Muted({ children }: { children: React.ReactNode }) {
  return (
    <Paragraph color="$panelTextSoft" fontSize="$3" maxW={680}>
      {children}
    </Paragraph>
  )
}

type TopicProps = {
  title: string
  children?: React.ReactNode
  /** A real sequence (the round's steps) is numbered; nothing else is. */
  step?: number
}

export function Topic({ title, step, children }: TopicProps) {
  return (
    <XStack gap="$3" items="flex-start">
      {step ? (
        <YStack
          width={28}
          height={28}
          rounded="$circle"
          bg="$wiiBlue"
          items="center"
          justify="center"
          mt="$0.5"
        >
          <GlowText level="label" tone="onColor">
            {step}
          </GlowText>
        </YStack>
      ) : null}
      <YStack flex={1} gap="$1.5">
        <GlowText level="label" tone="ink" fontSize="$5" role="heading">
          {title}
        </GlowText>
        {children}
      </YStack>
    </XStack>
  )
}

export function Chip({ label }: { label: string }) {
  return (
    <GlassPanel tone="plastic" px="$3" py="$1.5" rounded="$pill" elevate={0}>
      <GlowText level="label">{label}</GlowText>
    </GlassPanel>
  )
}

type ChipRowProps = { label?: string; chips: readonly string[] }

export function ChipRow({ label, chips }: ChipRowProps) {
  if (chips.length === 0) return null
  return (
    <YStack gap="$1.5">
      {label ? <GlowText level="label">{label}</GlowText> : null}
      <XStack flexWrap="wrap" gap="$2">
        {chips.map((chip) => (
          <Chip key={chip} label={chip} />
        ))}
      </XStack>
    </YStack>
  )
}

type PctRowProps = {
  label: string
  percent: number
  detail?: string
}

// "28.5%" - no trailing .0 for whole numbers.
export function formatPercent(percent: number): string {
  return `${Number.isInteger(percent) ? percent : percent.toFixed(2).replace(/0$/, '')}%`
}

export function PctRow({ label, percent, detail }: PctRowProps) {
  return (
    <XStack items="center" gap="$3">
      <YStack width={150} $md={{ width: 190 }}>
        <GlowText level="label" tone="ink">
          {label}
        </GlowText>
        {detail ? <GlowText level="label">{detail}</GlowText> : null}
      </YStack>
      <YStack flex={1}>
        <MeterBar value={percent / 100} tone="blue" heightPx={10} a11yLabel={label} />
      </YStack>
      <YStack width={64} items="flex-end">
        <GlowText level="label" tone="ink">
          {formatPercent(percent)}
        </GlowText>
      </YStack>
    </XStack>
  )
}

export function Divider() {
  return <YStack height={1} width="100%" bg="$glassEdge" opacity={0.6} />
}
