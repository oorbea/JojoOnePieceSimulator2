const { getDefaultConfig } = require('expo/metro-config')
const { withTamagui } = require('@tamagui/metro-plugin')

// isCSSEnabled lets Tamagui extract atomic CSS on web instead of shipping
// runtime style objects — this is what makes the web build fast.
const config = getDefaultConfig(__dirname, { isCSSEnabled: true })

// burnt's web toaster (sonner) ships as .mjs - required per burnt's README
// "Expo Web" section, or Metro fails to resolve it on web builds.
config.resolver.sourceExts.push('mjs')

module.exports = withTamagui(config, {
  components: ['tamagui'],
  config: './tamagui.config.ts',
  outputCSS: './tamagui-web.css',
})
