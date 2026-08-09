import { Text } from 'react-native'
import { renderWithProviders } from '@/test/render'

import { GlossButton } from '../gloss-button'

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

function flattenStyle(style: unknown): Record<string, unknown> {
  const list = Array.isArray(style) ? style : [style]
  return Object.assign({}, ...list.filter(Boolean))
}

// A `shape="circle"` button used to inherit the size table's horizontal
// padding (meant for pill/card text labels) on top of `aspectRatio: 1`,
// which squeezed an icon-only circle down to a sliver a few pixels wide —
// the `Plus` icon in the Skills field rendered as a tiny dot instead of a
// "+". This locks in that a circle button is square with zero horizontal
// padding, regardless of `btnSize`.
describe('GlossButton', () => {
  it('renders a circle shape as square with no horizontal padding', async () => {
    const { toJSON } = await renderWithProviders(
      <GlossButton testID="add-skill" tone="green" btnSize="md" shape="circle" onPress={() => {}}>
        <Text>+</Text>
      </GlossButton>
    )

    const button = findByTestId(toJSON() as JsonNode, 'add-skill')
    expect(button).not.toBeNull()

    const style = flattenStyle(button!.props.style)
    expect(style.width).toBe(style.height)
    expect(style.paddingLeft ?? 0).toBe(0)
    expect(style.paddingRight ?? 0).toBe(0)
  })

  it('keeps the size table’s horizontal padding for a pill shape', async () => {
    const { toJSON } = await renderWithProviders(
      <GlossButton testID="pill-btn" tone="green" btnSize="md" shape="pill" onPress={() => {}}>
        <Text>Label</Text>
      </GlossButton>
    )

    const button = findByTestId(toJSON() as JsonNode, 'pill-btn')
    expect(button).not.toBeNull()

    const style = flattenStyle(button!.props.style)
    expect(Number(style.paddingLeft ?? 0)).toBeGreaterThan(0)
  })
})
