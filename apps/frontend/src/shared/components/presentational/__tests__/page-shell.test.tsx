import { Text } from 'react-native'
import { renderWithProviders } from '@/test/render'

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

// AquaBackground renders the sky gradient + bubble field. AppShell already
// mounts one for every authenticated route (see app-shell.tsx), so a screen
// living inside it (navPadding=true) must not mount a second — two stacked,
// independently-animated backdrops used to render on top of each other.
describe('PageShell backdrop', () => {
  it('skips its own backdrop by default when navPadding is set (inside AppShell)', async () => {
    const { toJSON } = await renderWithProviders(
      <PageShell navPadding>
        <Text>content</Text>
      </PageShell>
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
      <PageShell navPadding backdrop>
        <Text>content</Text>
      </PageShell>
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
