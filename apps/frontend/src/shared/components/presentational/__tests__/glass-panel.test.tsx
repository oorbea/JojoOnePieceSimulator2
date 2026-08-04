import { Text } from 'react-native'
import { renderWithProviders } from '@/test/render'

import { GlassPanel } from '../glass-panel'

type JsonNode = {
  type: string
  props: Record<string, unknown>
  children: (JsonNode | string)[] | null
}

function findByTestId(node: JsonNode, testID: string): JsonNode | null {
  if (node.props?.testID === testID) return node
  for (const child of node.children ?? []) {
    if (typeof child === 'string') continue
    const found = findByTestId(child, testID)
    if (found) return found
  }
  return null
}

// GlassPanel used to wrap every child in a single extra host view
// (`<YStack z="$content" flex={1}>{children}</YStack>`), which left the
// frame with exactly one in-flow child — so any `gap`/`flexDirection` passed
// to GlassPanel silently did nothing between the panel's *real* children,
// which all rendered squeezed together inside that one wrapper. Fixed by
// rendering children as the frame's own direct flex children. This test
// locks that structure in: with no `glossy` overlay in the way, the panel's
// immediate children must be the two Text host nodes themselves — nothing
// wrapping them in between.
describe('GlassPanel', () => {
  it('renders children as the frame’s own direct children, not nested in an extra wrapper', async () => {
    const { toJSON } = await renderWithProviders(
      <GlassPanel testID="panel" gap="$5">
        <Text>First</Text>
        <Text>Second</Text>
      </GlassPanel>
    )

    const panel = findByTestId(toJSON() as JsonNode, 'panel')
    expect(panel).not.toBeNull()

    const children = panel!.children ?? []
    expect(children).toHaveLength(2)

    const texts = children.filter((c): c is JsonNode => typeof c !== 'string')
    expect(texts.map((t) => t.children?.[0])).toEqual(['First', 'Second'])
  })
})
