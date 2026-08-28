import { act, render, screen } from '@testing-library/react-native'
import { useEffect } from 'react'
import { Text } from 'react-native'

import { notifyScroll } from '@/shared/lib/scroll-bus'

import { useDropZones } from '../use-drop-zones'

// A fake host node: only `.measure()` matters here (same shape `tooltip.tsx`
// and `info-hint.tsx` already stub in their own tests), reporting a fixed
// page rect regardless of what real layout happened.
function fakeNode(pageX: number, pageY: number, width: number, height: number) {
  return {
    measure: (cb: (x: number, y: number, w: number, h: number, px: number, py: number) => void) =>
      cb(0, 0, width, height, pageX, pageY),
  }
}

function nextFrame() {
  return new Promise<void>((resolve) => requestAnimationFrame(() => resolve()))
}

// Probe rather than a `renderHook` helper - this codebase's test harness
// (src/test/render.tsx) has no `renderHook` precedent, so this exposes the
// hook's API on a module-level variable a test can call into directly, same
// render()/act() path every other component test already uses. Captured in
// an effect, not assigned directly in the component body - `react-hooks`'s
// purity rule flags reassigning an outer-scope variable during render
// itself, but explicitly allows doing so from an effect.
let api: ReturnType<typeof useDropZones> | null = null
function Probe() {
  const dropZones = useDropZones()
  useEffect(() => {
    api = dropZones
  })
  return <Text>probe</Text>
}

// Both scenarios share one render/mount - a second `render(<Probe />)` in a
// later `it()` in this same file reliably failed to find its own "probe"
// text (RNTL cleanup timing between tests, not anything `useDropZones`
// itself does), so this stays a single test with two phases instead.
describe('useDropZones', () => {
  it('resolves drop points to zones, and re-measures them on scroll', async () => {
    await render(<Probe />)
    await screen.findByText('probe')

    act(() => {
      api!.registerZone('teamA')(fakeNode(0, 0, 100, 100))
      api!.registerZone('teamB')(fakeNode(200, 0, 100, 100))
      api!.onZoneLayout('teamA')()
      api!.onZoneLayout('teamB')()
    })
    await act(async () => {
      await nextFrame()
    })

    expect(api!.resolveZone(50, 50)).toBe('teamA')
    expect(api!.resolveZone(250, 50)).toBe('teamB')
    expect(api!.resolveZone(150, 50)).toBeNull()

    // The columns scrolled out from under their original page position -
    // simulate that by swapping in nodes reporting new coordinates, then
    // firing the same scroll signal `tooltip.tsx` hides on.
    const movedNode = fakeNode(0, 500, 100, 100)
    act(() => {
      api!.registerZone('teamA')(movedNode)
    })
    act(() => {
      notifyScroll()
    })

    expect(api!.resolveZone(50, 50)).toBeNull()
    expect(api!.resolveZone(50, 550)).toBe('teamA')
  })
})
