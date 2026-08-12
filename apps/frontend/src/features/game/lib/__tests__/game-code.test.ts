import { formatCode, isCompleteCode, normalizeCode } from '@/features/game/lib/game-code'

describe('normalizeCode', () => {
  it('uppercases and strips ambiguous/invalid characters', () => {
    expect(normalizeCode('ab0oq1x')).toBe('ABQX')
  })

  it('caps at 6 characters', () => {
    expect(normalizeCode('ABCDEFGH')).toBe('ABCDEF')
  })

  it('strips punctuation and whitespace', () => {
    expect(normalizeCode('K7F-2QX!')).toBe('K7F2QX')
  })
})

describe('isCompleteCode', () => {
  it('is true only for a full 6-char code from the alphabet', () => {
    expect(isCompleteCode('K7F2QX')).toBe(true)
    expect(isCompleteCode('K7F2Q')).toBe(false)
    expect(isCompleteCode('')).toBe(false)
  })
})

describe('formatCode', () => {
  it('splits a complete code into two groups of three', () => {
    expect(formatCode('K7F2QX')).toBe('K7F 2QX')
  })

  it('returns an incomplete code unchanged', () => {
    expect(formatCode('K7F')).toBe('K7F')
  })
})
