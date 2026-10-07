import { THEME_COLORS, applyThemeColor } from '../theme-color'

describe('applyThemeColor', () => {
  beforeEach(() => {
    document.head.innerHTML = ''
  })

  it('pins every theme-color meta to the resolved colour and drops media', () => {
    document.head.innerHTML =
      '<meta name="theme-color" content="#111111" media="(prefers-color-scheme: light)">' +
      '<meta name="theme-color" content="#222222" media="(prefers-color-scheme: dark)">'

    applyThemeColor('dark')

    const metas = document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')
    expect(metas).toHaveLength(2)
    metas.forEach((meta) => {
      expect(meta.content).toBe(THEME_COLORS.dark)
      expect(meta.hasAttribute('media')).toBe(false)
    })
  })

  it('creates the meta when the page has none', () => {
    applyThemeColor('light')

    const metas = document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')
    expect(metas).toHaveLength(1)
    expect(metas[0].content).toBe(THEME_COLORS.light)
  })
})
