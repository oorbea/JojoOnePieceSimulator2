import { cardSource, fullSource, lqipSource, thumbSource } from '@/shared/lib/picture-source'

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

describe('cardSource', () => {
  it('prefers the card rendition when present', () => {
    expect(
      cardSource({ picture: 'main.webp', pictureThumb: 'thumb.webp', pictureCard: 'card.webp' })
    ).toBe('card.webp')
  })

  // Not backfilled yet (T1's DEFAULT '' migration, or an entity/backend
  // that predates the card rendition) - falls down to thumb, then main.
  it('falls back to the thumb rendition when card is an empty string', () => {
    expect(cardSource({ picture: 'main.webp', pictureThumb: 'thumb.webp', pictureCard: '' })).toBe(
      'thumb.webp'
    )
  })

  it('falls back to the main rendition when card and thumb are both empty', () => {
    expect(cardSource({ picture: 'main.webp', pictureThumb: '', pictureCard: '' })).toBe(
      'main.webp'
    )
  })

  it('falls back to the thumb/main ladder when card is entirely absent', () => {
    expect(cardSource({ picture: 'main.webp', pictureThumb: 'thumb.webp' })).toBe('thumb.webp')
  })

  it('returns null when no rendition exists', () => {
    expect(cardSource({ picture: '', pictureThumb: '', pictureCard: '' })).toBeNull()
  })
})

describe('lqipSource', () => {
  it('returns the lqip data URI when present', () => {
    expect(lqipSource({ pictureLqip: 'data:image/webp;base64,abc' })).toBe(
      'data:image/webp;base64,abc'
    )
  })

  it('normalizes an empty string (not backfilled yet) to null', () => {
    expect(lqipSource({ pictureLqip: '' })).toBeNull()
  })

  it('normalizes a missing field to null', () => {
    expect(lqipSource({})).toBeNull()
  })
})
