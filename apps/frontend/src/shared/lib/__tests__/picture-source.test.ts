import { fullSource, thumbSource } from '@/shared/lib/picture-source'

describe('thumbSource', () => {
  it('prefers the thumb rendition when present', () => {
    expect(thumbSource({ picture: 'main.webp', pictureThumb: 'thumb.webp' })).toBe('thumb.webp')
  })

  // The worker leaves pictureThumb empty until the transcode finishes, so an
  // empty string (not just a missing key) must fall back to the main
  // rendition - see admin-panel-crud-ux-fixes.md.
  it('falls back to the main rendition when the thumb is an empty string', () => {
    expect(thumbSource({ picture: 'main.webp', pictureThumb: '' })).toBe('main.webp')
  })

  it('returns null when neither rendition exists', () => {
    expect(thumbSource({ picture: '', pictureThumb: '' })).toBeNull()
  })
})

describe('fullSource', () => {
  it('returns the main rendition', () => {
    expect(fullSource({ picture: 'main.webp', pictureThumb: 'thumb.webp' })).toBe('main.webp')
  })

  it('returns null when there is no picture', () => {
    expect(fullSource({ picture: '', pictureThumb: '' })).toBeNull()
  })
})
