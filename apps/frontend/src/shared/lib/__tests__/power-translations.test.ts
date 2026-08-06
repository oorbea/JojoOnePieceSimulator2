import {
  createEmptyTranslationsForm,
  fromTranslationsResponse,
  powerTranslationsFormSchema,
  toTranslationsPayload,
} from '@/shared/lib/power-translations'

describe('toTranslationsPayload', () => {
  it('drops locales that are entirely blank and keeps the ones that are filled', () => {
    const payload = toTranslationsPayload({
      ...createEmptyTranslationsForm(),
      'en-GB': { description: 'A stand', skills: ['Ora Ora Ora'] },
    })

    expect(payload).toEqual({ 'en-GB': { description: 'A stand', skills: ['Ora Ora Ora'] } })
  })

  it('keeps every locale that has content', () => {
    const payload = toTranslationsPayload({
      'en-GB': { description: 'A stand', skills: ['Punch'] },
      'es-ES': { description: 'Un stand', skills: ['Puñetazo'] },
      'ca-ES': { description: '', skills: [] },
    })

    expect(Object.keys(payload).sort()).toEqual(['en-GB', 'es-ES'])
  })
})

describe('fromTranslationsResponse', () => {
  it('fills every locale absent from the response with an empty form value', () => {
    const values = fromTranslationsResponse({ 'en-GB': { description: 'A stand', skills: ['Punch'] } })

    expect(values['en-GB']).toEqual({ description: 'A stand', skills: ['Punch'] })
    expect(values['es-ES']).toEqual({ description: '', skills: [] })
    expect(values['ca-ES']).toEqual({ description: '', skills: [] })
  })
})

describe('powerTranslationsFormSchema', () => {
  const valid = {
    'en-GB': { description: 'A stand', skills: ['Punch'] },
    'es-ES': { description: '', skills: [] },
    'ca-ES': { description: '', skills: [] },
  }

  it('accepts en-GB filled and the other locales entirely blank', () => {
    expect(powerTranslationsFormSchema.safeParse(valid).success).toBe(true)
  })

  it('rejects en-GB entirely blank', () => {
    const result = powerTranslationsFormSchema.safeParse({
      ...valid,
      'en-GB': { description: '', skills: [] },
    })
    expect(result.success).toBe(false)
  })

  it('rejects a locale with a description but no skills (half-filled)', () => {
    const result = powerTranslationsFormSchema.safeParse({
      ...valid,
      'es-ES': { description: 'Un stand', skills: [] },
    })
    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0].message).toBe('validation.translationIncomplete')
    }
  })

  it('rejects a locale with skills but no description (half-filled)', () => {
    const result = powerTranslationsFormSchema.safeParse({
      ...valid,
      'ca-ES': { description: '', skills: ['Puny'] },
    })
    expect(result.success).toBe(false)
  })
})
