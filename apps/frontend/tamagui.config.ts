import { defaultConfig } from '@tamagui/config/v4'
import { createTamagui } from 'tamagui'

// Start from Tamagui's v4 default config (fonts, media queries, animations,
// shorthands) and layer a JoJo/One Piece palette on top so the theme reads as
// "ours" without having to hand-roll the whole token set.
const jojoColors = {
  strawHatRed: '#c0272d',
  strawHatStraw: '#f2c14e',
  standPurple: '#5b2a86',
  standGold: '#d4af37',
  seaBlue: '#1b6ca8',
  inkBlack: '#12131a',
}

const config = createTamagui({
  ...defaultConfig,
  themes: {
    ...defaultConfig.themes,
    light: {
      ...defaultConfig.themes.light,
      color: jojoColors.inkBlack,
      background: '#faf7f0',
      accentColor: jojoColors.strawHatRed,
      accentBackground: jojoColors.strawHatStraw,
    },
    dark: {
      ...defaultConfig.themes.dark,
      color: '#f5f1e8',
      background: jojoColors.inkBlack,
      accentColor: jojoColors.standGold,
      accentBackground: jojoColors.standPurple,
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
