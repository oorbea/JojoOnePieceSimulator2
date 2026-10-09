import i18next from 'i18next'

import enGB from '@/shared/i18n/locales/en-GB.json'

import { subjectLabel } from '../subject-label'

// eslint-disable-next-line import/no-named-as-default-member -- default import is correct; same pattern as src/test/render.tsx
const i18n = i18next.createInstance()
void i18n.init({
  resources: { 'en-GB': { translation: enGB } },
  lng: 'en-GB',
  interpolation: { escapeValue: false },
})
const t = i18n.t.bind(i18n) as (key: string, options?: Record<string, unknown>) => string

describe('subjectLabel', () => {
  it('names the broad kinds of power user', () => {
    expect(subjectLabel(t, { kind: 'ANY_STAND' })).toBe('Stand users')
    expect(subjectLabel(t, { kind: 'ANY_FRUIT' })).toBe('Devil Fruit users')
    expect(subjectLabel(t, { kind: 'FRUIT_TYPE', fruitTypes: ['LOGIA'] })).toBe('Logia fruit users')
    expect(
      subjectLabel(t, { kind: 'FRUIT_TYPE', fruitTypes: ['MYTHICAL_ZOAN', 'ANCIENT_ZOAN'] })
    ).toBe('Mythical Zoan / Ancient Zoan fruit users')
  })

  it('says "any level" when the minimum is the lowest level that counts', () => {
    expect(subjectLabel(t, { kind: 'SLOT_MIN', slot: 'SPIN', min: 'BASIC' })).toBe(
      'Spin (any level)'
    )
    expect(subjectLabel(t, { kind: 'SLOT_MIN', slot: 'OBSERVATION_HAKI', min: 'PRIVATE' })).toBe(
      'Observation Haki (any level)'
    )
    expect(subjectLabel(t, { kind: 'SLOT_MIN', slot: 'HAMON', min: 'BASIC' })).toBe(
      'Hamon (any level)'
    )
  })

  it('spells out a higher minimum with the UI wording of the level', () => {
    expect(subjectLabel(t, { kind: 'SLOT_MIN', slot: 'SPIN', min: 'GOLDEN' })).toBe(
      'Spin Golden or higher'
    )
    expect(subjectLabel(t, { kind: 'SLOT_MIN', slot: 'CONQUEROR_HAKI', min: 'YONKO_PLUS' })).toBe(
      "Conqueror's Haki Yonko+ or higher"
    )
  })

  it('lists named powers, labelling a named group', () => {
    expect(subjectLabel(t, { kind: 'POWERS', names: ['Weather Report'] })).toBe('Weather Report')
    expect(
      subjectLabel(t, {
        kind: 'POWERS',
        group: 'TIME_STOPPERS',
        names: ['The World', 'King Crimson'],
      })
    ).toBe('Time stoppers (The World, King Crimson)')
  })
})
