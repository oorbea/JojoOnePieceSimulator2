import { Text } from 'react-native'
import { renderWithProviders } from '@/test/render'

import { ChannelBar } from '../channel-bar'

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

function styleOf(node: JsonNode): Record<string, unknown> {
  const style = node.props?.style
  return Object.assign({}, ...(Array.isArray(style) ? style : [style]).filter(Boolean))
}

// `dock="top"|"bottom"` used to set `position:absolute; left:0; right:0`
// directly on the pill itself, over-constraining it against its own
// `maxW:1080; self:center` — past 1080px wide, the bar hugged the left edge
// instead of centering. Docking now wraps the (still self-centering, still
// in normal flow) pill in a non-interactive absolute host. This locks that
// structure in.
describe('ChannelBar docking', () => {
  it('positions only the host absolutely, keeping the pill itself in normal flow', async () => {
    const { toJSON } = await renderWithProviders(
      <ChannelBar dock="top" t={16} testID="pill">
        <Text>Item</Text>
      </ChannelBar>
    )

    const nodes = flatten(toJSON() as JsonNode)
    const pill = nodes.find((n) => n.props?.testID === 'pill')
    expect(pill).toBeTruthy()

    const pillStyle = styleOf(pill!)
    // The pill must not carry its own absolute positioning any more — that
    // was the exact bug (left:0 + right:0 + maxWidth fighting each other).
    expect(pillStyle.position).not.toBe('absolute')

    // Some ancestor must be the absolute, non-interactive centering host.
    const host = nodes.find((n) => {
      const s = styleOf(n)
      return s.position === 'absolute' && s.left === 0 && s.right === 0
    })
    expect(host).toBeTruthy()
    expect(styleOf(host!).pointerEvents).toBe('box-none')
  })

  it('renders in normal flow (no absolute host) when not docked', async () => {
    const { toJSON } = await renderWithProviders(
      <ChannelBar testID="pill">
        <Text>Item</Text>
      </ChannelBar>
    )

    const nodes = flatten(toJSON() as JsonNode)
    // GlossOverlay itself is a legitimate absolute decorative child — the
    // thing that must be absent when not docked is the box-none centering
    // host, which only ChannelBar's own docking path renders.
    const host = nodes.find((n) => styleOf(n).pointerEvents === 'box-none')
    expect(host).toBeUndefined()
  })
})
