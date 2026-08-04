import { defaultConfig } from '@tamagui/config/v4'
import { createFont, createTamagui } from 'tamagui'

// Wii Party × Windows Aero × iOS-gloss design system.
//
// @tamagui/config/v4's defaultConfig ships radius/size/space/zIndex tokens
// but NO `color` token group at all — that's the whole reason `$standPurple`
// and `$strawHatRed` used to resolve to nothing below. Registering a real
// `tokens.color` group fixes those dead references as a side effect of this
// rewrite.
//
// Brand hexes here are duplicated in public/manifest.json (theme_color /
// background_color) — keep them in sync if either changes.
//
// Naming rule: Tamagui's theme lookup wins over token lookup, so no token
// name below may collide with a default theme key. Forbidden prefixes:
// color*, background*, border*, accent*, shadow[1-6], placeholderColor,
// outlineColor, and the bare color-scale names (blue*, red*, green*,
// yellow*, black*, white*). Every key below has been screened against that.
const jojoColors = {
  // --- Wii supersaturated primaries, each with a `*Deep` shade used as the
  // physical button lip (borderBottomColor) that sells "3D" with no shadow
  // hacks.
  wiiBlue: '#00A0E9',
  wiiBlueDeep: '#0072B5',
  sunYellow: '#FFD400',
  sunYellowDeep: '#E09A00',
  meadowGreen: '#5DC21E',
  meadowGreenDeep: '#3B8F0E',
  bubblegum: '#FF5FA2',
  bubblegumDeep: '#DB2E77',
  tangerine: '#FF8A1F',
  tangerineDeep: '#DE6300',
  grapeSoda: '#9B5CF6',
  grapeSodaDeep: '#6D2FD1',

  // --- JoJo / One Piece brand carry-over. strawHatRed/standPurple are the
  // two tokens that were referenced but never registered before this file.
  strawHatRed: '#E63946',
  strawHatRedDeep: '#A81D2A',
  strawHatStraw: '#FFD86B',
  standPurple: '#7A3FB5',
  standGold: '#F2C744',
  seaBlue: '#1B6CA8',
  inkBlack: '#12131A',

  // --- polished-plastic neutrals (scheme-invariant, so plain tokens)
  plasticWhite: '#FFFFFF',
  plasticOffWhite: '#F4FAFF',
  plasticEdge: '#C9DDE9',
}

const config = createTamagui({
  ...defaultConfig,
  fonts: {
    ...defaultConfig.fonts,
    // Display face: Fredoka, used with restraint (titles, channel labels).
    heading: createFont({
      family: 'Fredoka, Fredoka_500Medium, Fredoka_600SemiBold, sans-serif',
      size: {
        1: 12,
        2: 14,
        3: 15,
        4: 16,
        5: 18,
        6: 20,
        7: 23,
        8: 26,
        9: 30,
        10: 36,
        11: 44,
        12: 54,
      },
      lineHeight: {
        1: 16,
        2: 18,
        3: 20,
        4: 22,
        5: 24,
        6: 26,
        7: 29,
        8: 32,
        9: 36,
        10: 42,
        11: 50,
        12: 60,
      },
      weight: {
        1: '500',
        2: '500',
        3: '600',
        4: '600',
        5: '600',
        6: '600',
        7: '600',
        8: '600',
        9: '600',
        10: '600',
        11: '600',
        12: '600',
      },
      letterSpacing: {
        1: 0,
        2: 0,
        3: -0.1,
        4: -0.1,
        5: -0.2,
        6: -0.2,
        7: -0.3,
        8: -0.3,
        9: -0.4,
        10: -0.4,
        11: -0.5,
        12: -0.5,
      },
      face: {
        500: { normal: 'Fredoka_500Medium' },
        600: { normal: 'Fredoka_600SemiBold' },
      },
    }),
    // Body / UI face: Nunito — rounded terminals to agree with Fredoka, but
    // a real text face at 16px/1.5 line-height. Replaces Inter.
    body: createFont({
      family: 'Nunito, Nunito_500Medium, Nunito_700Bold, sans-serif',
      size: {
        1: 11,
        2: 12,
        3: 13,
        4: 14,
        5: 15,
        6: 16,
        7: 18,
        8: 20,
        9: 23,
        10: 27,
        11: 32,
        12: 38,
      },
      lineHeight: {
        1: 16,
        2: 17,
        3: 19,
        4: 21,
        5: 22,
        6: 24,
        7: 26,
        8: 28,
        9: 31,
        10: 35,
        11: 40,
        12: 46,
      },
      weight: {
        1: '500',
        2: '500',
        3: '500',
        4: '500',
        5: '500',
        6: '500',
        7: '700',
        8: '700',
        9: '700',
        10: '700',
        11: '700',
        12: '700',
      },
      letterSpacing: {
        1: 0.1,
        2: 0.1,
        3: 0.05,
        4: 0,
        5: 0,
        6: 0,
        7: -0.1,
        8: -0.1,
        9: -0.2,
        10: -0.2,
        11: -0.3,
        12: -0.3,
      },
      face: {
        500: { normal: 'Nunito_500Medium' },
        700: { normal: 'Nunito_700Bold' },
      },
    }),
  },
  tokens: {
    ...defaultConfig.tokens,
    color: {
      ...jojoColors,
    },
    // Extend, never replace — Tamagui internals (Button/Input sizing, etc)
    // rely on the numeric keys already present in defaultConfig.tokens.radius
    // and .zIndex.
    radius: {
      ...defaultConfig.tokens.radius,
      chip: 14,
      card: 22,
      panel: 28,
      hero: 32,
      bubble: 26,
      pill: 9999,
      circle: 9999,
    },
    zIndex: {
      ...defaultConfig.tokens.zIndex,
      backdrop: 0,
      content: 10,
      gloss: 20,
      nav: 500,
      overlay: 700,
    },
    // tokens.size / tokens.space are intentionally left untouched — `space`
    // is derived from `size` upstream via sizeToSpace, so adding a key to
    // one and not the other yields a half-resolving token. One-off
    // dimensions use raw numbers (the existing codebase convention, e.g.
    // `maxW={440}`).
  },
  themes: {
    ...defaultConfig.themes,
    light: {
      ...defaultConfig.themes.light,
      color: jojoColors.inkBlack,
      background: '#EAF7FF',
      accentColor: '#FFFFFF',
      accentBackground: jojoColors.wiiBlue,

      // Sky gradient the animated backdrop reads.
      pageFrom: '#B8ECFF',
      pageMid: '#E9F8FF',
      pageTo: '#F1E9FF',

      // Soap-bubble field.
      bubbleFill: 'rgba(255,255,255,0.36)',
      bubbleEdge: 'rgba(255,255,255,0.70)',

      // Aero glass — web gets real backdrop-filter blur at these alphas;
      // native has no blur, so glassFillNative raises alpha instead.
      glassFill: 'rgba(255,255,255,0.55)',
      glassFillStrong: 'rgba(255,255,255,0.75)',
      glassFillNative: 'rgba(255,255,255,0.86)',
      glassEdge: 'rgba(255,255,255,0.85)',
      glassEdgeInner: 'rgba(255,255,255,0.60)',

      // iOS-6 gloss overlay stops.
      glossPeak: 'rgba(255,255,255,0.90)',
      glossFade: 'rgba(255,255,255,0.28)',
      glossNil: 'rgba(255,255,255,0)',

      panelText: '#16283C',
      panelTextSoft: '#4A6480',

      plasticFill: '#FFFFFF',
      plasticEdgeColor: '#C9DDE9',

      channelFill: 'rgba(255,255,255,0.62)',
      channelActive: '#00A0E9',

      softShadow: 'rgba(18,60,92,0.18)',
      hardShadow: 'rgba(12,44,74,0.30)',

      glowColor: 'rgba(0,160,233,0.55)',
      textGlow: 'rgba(255,255,255,0.85)',
    },
    dark: {
      ...defaultConfig.themes.dark,
      color: '#EAF4FF',
      background: '#0A1626',
      accentColor: '#FFFFFF',
      accentBackground: jojoColors.wiiBlue,

      // "Night Wii" — deep indigo/navy sky, same saturated tones elsewhere
      // so the playfulness survives the theme switch.
      pageFrom: '#071A30',
      pageMid: '#0D2A47',
      pageTo: '#1A1230',

      bubbleFill: 'rgba(160,215,255,0.14)',
      bubbleEdge: 'rgba(170,220,255,0.30)',

      glassFill: 'rgba(20,30,48,0.58)',
      glassFillStrong: 'rgba(16,24,40,0.78)',
      glassFillNative: 'rgba(16,24,40,0.90)',
      glassEdge: 'rgba(150,205,255,0.38)',
      glassEdgeInner: 'rgba(120,180,240,0.18)',

      glossPeak: 'rgba(255,255,255,0.34)',
      glossFade: 'rgba(255,255,255,0.10)',
      glossNil: 'rgba(255,255,255,0)',

      panelText: '#EAF4FF',
      panelTextSoft: '#9FB8D4',

      plasticFill: '#152437',
      plasticEdgeColor: '#28405C',

      channelFill: 'rgba(14,24,40,0.66)',
      channelActive: '#38B6F2',

      softShadow: 'rgba(0,0,0,0.45)',
      hardShadow: 'rgba(0,0,0,0.65)',

      glowColor: 'rgba(120,215,255,0.45)',
      textGlow: 'rgba(0,0,0,0.55)',
    },
  },
})

// Augment Tamagui's module so `styled()`/`GetProps` and JSX style props pick
// up this app's tokens/themes instead of the library defaults.
type AppConfig = typeof config

declare module '@tamagui/web' {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type -- required declaration-merging shape
  interface TamaguiCustomConfig extends AppConfig {}
}

export default config
