import { Text } from 'react-native'
import { renderWithProviders } from '@/test/render'

import { NavInsetsProvider } from '@/shared/lib/nav-insets'

import { PageShell } from '../page-shell'

type JsonNode = {
  type: string
  props: Record<string, unknown>
  children: (JsonNode | string)[] | null
}

function flatten(node: JsonNode, out: JsonNode[] = []) {
  out.push(node)
  for (const child of node.children ?? []) {
    if (typeof child !== 'string') flatten(child, out)
  }
  return out
}

// jest.setup.ts mocks @tamagui/linear-gradient's LinearGradient down to a
// plain RN View, so its host `type` is just "View" here — the `colors` prop
// (unique to LinearGradient among this tree's Views) is what actually
// identifies AquaBackground's sky gradient.
function countGradients(root: JsonNode) {
  return flatten(root).filter((n) => Array.isArray(n.props?.colors)).length
}

function styleOf(node: JsonNode): Record<string, unknown> {
  const style = node.props?.style
  return Object.assign({}, ...(Array.isArray(style) ? style : [style]).filter(Boolean))
}

// Finds PageShell's own padded content YStack (not the outer flex:1 wrapper,
// nor anything AquaBackground/ScrollView render) by its unique paddingTop.
function paddingOf(root: JsonNode) {
  const withPadding = flatten(root).find((n) => styleOf(n).paddingTop !== undefined)
  const style = withPadding ? styleOf(withPadding) : {}
  return { top: style.paddingTop, bottom: style.paddingBottom }
}

// AquaBackground renders the sky gradient + bubble field. AppShell already
// mounts one for every authenticated route (see app-shell.tsx) and publishes
// its measured bar reservation through NavInsetsProvider, so a screen living
// inside it must not mount a second backdrop — two stacked, independently-
// animated backdrops used to render on top of each other.
describe('PageShell backdrop', () => {
  it('skips its own backdrop by default when a NavInsetsProvider reservation is present (inside AppShell)', async () => {
    const { toJSON } = await renderWithProviders(
      <NavInsetsProvider value={{ top: 120, bottom: 0 }}>
        <PageShell>
          <Text>content</Text>
        </PageShell>
      </NavInsetsProvider>
    )

    expect(countGradients(toJSON() as JsonNode)).toBe(0)
  })

  it('renders its own backdrop by default when not inside AppShell (login, not-found, ...)', async () => {
    const { toJSON } = await renderWithProviders(
      <PageShell>
        <Text>content</Text>
      </PageShell>
    )

    expect(countGradients(toJSON() as JsonNode)).toBeGreaterThan(0)
  })

  it('honors an explicit backdrop override either way', async () => {
    const shown = await renderWithProviders(
      <NavInsetsProvider value={{ top: 120, bottom: 0 }}>
        <PageShell backdrop>
          <Text>content</Text>
        </PageShell>
      </NavInsetsProvider>
    )
    expect(countGradients(shown.toJSON() as JsonNode)).toBeGreaterThan(0)

    const hidden = await renderWithProviders(
      <PageShell backdrop={false}>
        <Text>content</Text>
      </PageShell>
    )
    expect(countGradients(hidden.toJSON() as JsonNode)).toBe(0)
  })
})

describe('PageShell nav clearance', () => {
  it('takes its padding from NavInsetsProvider when one reserves space', async () => {
    const { toJSON } = await renderWithProviders(
      <NavInsetsProvider value={{ top: 120, bottom: 90 }}>
        <PageShell>
          <Text>content</Text>
        </PageShell>
      </NavInsetsProvider>
    )

    expect(paddingOf(toJSON() as JsonNode)).toEqual({ top: 120, bottom: 90 })
  })

  it('falls back to plain breathing room without a provider', async () => {
    const { toJSON } = await renderWithProviders(
      <PageShell>
        <Text>content</Text>
      </PageShell>
    )

    // Zero safe-area insets in the test environment, so this is exactly the
    // +16 breathing-room fallback.
    expect(paddingOf(toJSON() as JsonNode)).toEqual({ top: 16, bottom: 16 })
  })
})
