// Browser/OS chrome colour (Android status bar, standalone title bar) for each
// resolved app theme. Hexes mirror tamagui.config.ts (light: wiiBlue, dark:
// the dark theme `background`) and the `theme-color` metas in public/index.html
// - keep the three in sync.
export const THEME_COLORS = {
  light: '#00A0E9',
  dark: '#0A1626',
} as const

export type ResolvedTheme = keyof typeof THEME_COLORS

// public/index.html ships two `theme-color` metas split by
// `prefers-color-scheme`, which only follows the OS. The in-app toggle can
// force the opposite theme, so once the app knows the resolved theme it pins
// every theme-color meta to one value (dropping `media`) instead.
export function applyThemeColor(theme: ResolvedTheme, doc: Document | undefined = globalThis.document) {
  if (!doc) return
  const color = THEME_COLORS[theme]
  const metas = doc.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')
  if (metas.length === 0) {
    const meta = doc.createElement('meta')
    meta.name = 'theme-color'
    meta.content = color
    doc.head.appendChild(meta)
    return
  }
  metas.forEach((meta) => {
    meta.removeAttribute('media')
    meta.content = color
  })
}
