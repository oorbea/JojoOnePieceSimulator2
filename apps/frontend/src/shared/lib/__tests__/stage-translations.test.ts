import {
  fromStageTranslationsResponse,
  stageTranslationsFormSchema,
  toStageTranslationsPayload,
} from '@/shared/lib/stage-translations'

describe('toStageTranslationsPayload', () => {
  it('carries every locale, unlike Power which drops blanks', () => {
    const payload = toStageTranslationsPayload({
      'en-GB': { description: 'A stage' },
      'es-ES': { description: 'Un escenario' },
      'ca-ES': { description: 'Un escenari' },
    })

    expect(Object.keys(payload).sort()).toEqual(['ca-ES', 'en-GB', 'es-ES'])
    expect(payload['en-GB']).toEqual({ description: 'A stage' })
  })
})

describe('fromStageTranslationsResponse', () => {
  it('fills every locale absent from the response with an empty form value', () => {
    const values = fromStageTranslationsResponse({ 'en-GB': { description: 'A stage' } })

    expect(values['en-GB']).toEqual({ description: 'A stage' })
    expect(values['es-ES']).toEqual({ description: '' })
    expect(values['ca-ES']).toEqual({ description: '' })
  })
})

describe('stageTranslationsFormSchema', () => {
  const valid = {
    'en-GB': { description: 'A stage' },
    'es-ES': { description: 'Un escenario' },
    'ca-ES': { description: 'Un escenari' },
  }

  it('accepts every locale filled', () => {
    expect(stageTranslationsFormSchema.safeParse(valid).success).toBe(true)
  })

  it('rejects en-GB blank', () => {
    const result = stageTranslationsFormSchema.safeParse({ ...valid, 'en-GB': { description: '' } })
    expect(result.success).toBe(false)
  })

  it('rejects es-ES blank, unlike Power where non-en-GB locales may be entirely blank', () => {
    const result = stageTranslationsFormSchema.safeParse({ ...valid, 'es-ES': { description: '' } })
    expect(result.success).toBe(false)
  })

  it('rejects ca-ES blank', () => {
    const result = stageTranslationsFormSchema.safeParse({ ...valid, 'ca-ES': { description: '' } })
    expect(result.success).toBe(false)
  })
})
