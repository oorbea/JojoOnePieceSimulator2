import { act, render } from '@testing-library/react-native'
import { useEffect } from 'react'

import type { RovingItemProps } from '@/shared/hooks/use-roving-group'
import { useRovingGroup } from '@/shared/hooks/use-roving-group'

// The web-only keyboard branch of useRovingGroup, which never engaged under
// the native jest project (Platform.OS !== 'web' there, so getItemProps
// returns only { tabIndex } and none of the key handling below exists). The
// `.web.test.tsx` suffix is what routes this file to the jsdom "logic"
// project - see jest.config.js.
//
// The hook moves focus by looking its target up in the global `document` by
// a `${groupId}-${index}` id rather than through refs (GlossButton owns its
// own ref for tooltip measurement and doesn't forward one). So this test
// injects plain divs carrying those ids: no rendered tree is needed, because
// the hook never touches one.
//
// Rendered through a tiny probe rather than a renderHook helper, matching
// the precedent use-debounced-value.test.tsx already set for this codebase.
// The probe renders `null` deliberately: this project maps `react-native` to
// `react-native-web`, so a react-native <Text> here would resolve to a DOM
// element that @testing-library/react-native's native renderer then rejects
// ("Text strings must be rendered within a <Text> component"). Nothing needs
// to be in the tree anyway - every assertion reads the hook's return value
// or the real `document` the hook focuses through.

const GROUP_ID = 'test-group'
const COUNT = 3

type Api = {
  activeIndex: number
  getItemProps: (index: number) => RovingItemProps
}

// The probe hands its hook value out through a callback from an effect
// rather than assigning to anything in this module's scope directly: the
// react-hooks lint rules reject a component both reassigning an outer
// variable AND mutating an outer object, so the write has to happen in a
// plain function the component merely calls. The effect has no dependency
// array on purpose - it must re-publish after every commit, which is how
// `api()` below stays current across state changes.
let captured: Api | null = null

function publish(next: Api) {
  captured = next
}

function api(): Api {
  if (!captured) throw new Error('Probe has not rendered yet')
  return captured
}

function Probe({ onActivate }: { onActivate: (index: number) => void }) {
  const value = useRovingGroup({ groupId: GROUP_ID, count: COUNT, onActivate })
  useEffect(() => {
    publish(value)
  })
  return null
}

// `await act(async () => ...)` rather than a plain sync `act()`: this version
// of @testing-library/react-native warns on (and does not flush) the sync
// form, so the hook's setState would never commit and every assertion would
// silently read the initial state back. Same shape as use-debounced-value's
// `advance` helper.
async function press(index: number, key: string) {
  const preventDefault = jest.fn()
  await act(async () => {
    api()
      .getItemProps(index)
      .onKeyDown?.({ key, preventDefault })
  })
  return preventDefault
}

describe('useRovingGroup (web keyboard branch)', () => {
  let nodes: HTMLElement[]
  let onActivate: jest.Mock

  beforeEach(async () => {
    document.body.innerHTML = ''
    nodes = []
    for (let i = 0; i < COUNT; i += 1) {
      const el = document.createElement('div')
      el.id = `${GROUP_ID}-${i}`
      // Only focusable elements can become document.activeElement in jsdom.
      el.setAttribute('tabindex', '-1')
      document.body.appendChild(el)
      nodes.push(el)
    }
    onActivate = jest.fn()
    await render(<Probe onActivate={onActivate} />)
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('gives tabIndex 0 only to the active index, -1 to the rest', async () => {
    expect(api().getItemProps(0).tabIndex).toBe(0)
    expect(api().getItemProps(1).tabIndex).toBe(-1)
    expect(api().getItemProps(2).tabIndex).toBe(-1)

    await press(0, 'ArrowRight')

    expect(api().getItemProps(0).tabIndex).toBe(-1)
    expect(api().getItemProps(1).tabIndex).toBe(0)
    expect(api().getItemProps(2).tabIndex).toBe(-1)
  })

  it('exposes the id convention the hook focuses by', () => {
    expect(api().getItemProps(2).id).toBe(`${GROUP_ID}-2`)
  })

  it.each(['ArrowRight', 'ArrowDown'])('moves focus forward on %s', async (key) => {
    await press(0, key)

    expect(api().activeIndex).toBe(1)
    expect(document.activeElement).toBe(nodes[1])
  })

  it.each(['ArrowLeft', 'ArrowUp'])('moves focus backward on %s', async (key) => {
    await press(1, key)

    expect(api().activeIndex).toBe(0)
    expect(document.activeElement).toBe(nodes[0])
  })

  it('wraps forward off the end and backward off the start', async () => {
    await press(COUNT - 1, 'ArrowRight')
    expect(api().activeIndex).toBe(0)
    expect(document.activeElement).toBe(nodes[0])

    await press(0, 'ArrowLeft')
    expect(api().activeIndex).toBe(COUNT - 1)
    expect(document.activeElement).toBe(nodes[COUNT - 1])
  })

  it('jumps to the first item on Home and the last on End', async () => {
    await press(0, 'End')
    expect(api().activeIndex).toBe(COUNT - 1)
    expect(document.activeElement).toBe(nodes[COUNT - 1])

    await press(COUNT - 1, 'Home')
    expect(api().activeIndex).toBe(0)
    expect(document.activeElement).toBe(nodes[0])
  })

  it.each(['Enter', ' '])('activates the focused index on %s', async (key) => {
    await press(2, key)

    expect(onActivate).toHaveBeenCalledTimes(1)
    expect(onActivate).toHaveBeenCalledWith(2)
  })

  it('never activates on plain arrow movement', async () => {
    for (const key of ['ArrowRight', 'ArrowDown', 'ArrowLeft', 'ArrowUp', 'Home', 'End']) {
      await press(0, key)
    }

    expect(onActivate).not.toHaveBeenCalled()
  })

  it('calls preventDefault on every key it handles', async () => {
    const handled = ['ArrowRight', 'ArrowDown', 'ArrowLeft', 'ArrowUp', 'Home', 'End', 'Enter', ' ']
    for (const key of handled) {
      expect(await press(0, key)).toHaveBeenCalled()
    }
  })

  it('leaves unhandled keys alone', async () => {
    const preventDefault = await press(0, 'Tab')

    expect(preventDefault).not.toHaveBeenCalled()
    expect(onActivate).not.toHaveBeenCalled()
    expect(api().activeIndex).toBe(0)
  })

  it('tracks focus moved by the browser itself (onFocus)', async () => {
    await act(async () => {
      api().getItemProps(2).onFocus?.()
    })

    expect(api().activeIndex).toBe(2)
    expect(api().getItemProps(2).tabIndex).toBe(0)
  })
})
