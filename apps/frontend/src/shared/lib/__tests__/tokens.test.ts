import tamaguiConfig from '../../../../tamagui.config'

// GlossOverlay (gloss-overlay.tsx) is a decorative absolute sibling of a
// panel's real content — it must always paint *below* that content, or the
// gloss gradient visually covers the text/icons next to it (a real bug that
// shipped with zIndex.gloss=20 > zIndex.content=10). Locks in the ordering
// so a future token edit can't silently reintroduce it.
describe('tamagui zIndex tokens', () => {
  it('keeps gloss below content, and content below the floating nav bars', () => {
    const z = tamaguiConfig.tokensParsed.zIndex as Record<string, { val: number }>

    expect(z.$gloss.val).toBeLessThan(z.$content.val)
    expect(z.$content.val).toBeLessThan(z.$nav.val)
    expect(z.$nav.val).toBeLessThan(z.$overlay.val)
  })
})
