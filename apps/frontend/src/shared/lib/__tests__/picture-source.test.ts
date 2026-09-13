import {
  cardSource,
  focalFromLocation,
  focalPosition,
  fullSource,
  imageRectForWell,
  lqipSource,
  thumbSource,
} from '@/shared/lib/picture-source'

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

describe('focalPosition', () => {
  it('returns null for the default center (0.5, 0.5)', () => {
    expect(focalPosition({ focalX: 0.5, focalY: 0.5 })).toBeNull()
  })

  it('returns null when the entity carries no focal fields at all', () => {
    expect(focalPosition({})).toBeNull()
    expect(focalPosition(null)).toBeNull()
    expect(focalPosition(undefined)).toBeNull()
  })

  it('converts a non-center focal point to percentage top/left', () => {
    expect(focalPosition({ focalX: 0.25, focalY: 0.75 })).toEqual({ top: '75%', left: '25%' })
  })
})

describe('focalFromLocation', () => {
  it('maps a point inside the rect to a proportional 0..1 pair', () => {
    expect(focalFromLocation(50, 100, 200, 200)).toEqual({ x: 0.25, y: 0.5 })
  })

  it('maps the top-left corner to (0, 0)', () => {
    expect(focalFromLocation(0, 0, 200, 100)).toEqual({ x: 0, y: 0 })
  })

  it('maps the bottom-right corner to (1, 1)', () => {
    expect(focalFromLocation(200, 100, 200, 100)).toEqual({ x: 1, y: 1 })
  })

  it('clamps a location past the rect edges', () => {
    expect(focalFromLocation(-40, 500, 200, 100)).toEqual({ x: 0, y: 1 })
  })

  it('returns the center when the rect has no size yet', () => {
    expect(focalFromLocation(10, 10, 0, 0)).toEqual({ x: 0.5, y: 0.5 })
  })
})

describe('imageRectForWell', () => {
  // This is the "contain" fit math - the same shape RN's resizeMode="contain"
  // uses to letterbox an image inside a well of a different aspect ratio.
  // The focal picker needs its own copy of this so the gesture's coordinate
  // space (locationX/Y relative to the well) matches the image's own
  // displayed rect, not the letterboxed well around it - dividing by the
  // well directly (the original bug) put the crosshair at a different point
  // than the one the user actually touched whenever the aspect ratios
  // differed.
  it('letterboxes a portrait image inside a landscape well (pillarboxed)', () => {
    // 100x200 image (portrait) inside a 300x200 well (landscape): height is
    // the limiting dimension, so the image renders at its full 200px tall
    // and half as wide (100 * (200/200) = 100... scaled by min(300/100,
    // 200/200) = min(3, 1) = 1 -> 100x200, centered with empty bars left/right.
    expect(imageRectForWell({ width: 100, height: 200 }, { width: 300, height: 200 })).toEqual({
      width: 100,
      height: 200,
    })
  })

  it('letterboxes a landscape image inside a portrait well (letterboxed top/bottom)', () => {
    // 200x100 image inside a 200x300 well: width is the limiting dimension
    // (min(200/200, 300/100) = min(1, 3) = 1) -> full 200x100, centered
    // with empty bars above/below.
    expect(imageRectForWell({ width: 200, height: 100 }, { width: 200, height: 300 })).toEqual({
      width: 200,
      height: 100,
    })
  })

  it('fills the well exactly when the aspect ratios already match', () => {
    expect(imageRectForWell({ width: 400, height: 300 }, { width: 200, height: 150 })).toEqual({
      width: 200,
      height: 150,
    })
  })

  it('falls back to the well itself when the natural size is not known yet', () => {
    expect(imageRectForWell(null, { width: 200, height: 150 })).toEqual({
      width: 200,
      height: 150,
    })
  })

  it('falls back to the well itself when the well has no size yet', () => {
    expect(imageRectForWell({ width: 100, height: 100 }, { width: 0, height: 0 })).toEqual({
      width: 0,
      height: 0,
    })
  })
})
