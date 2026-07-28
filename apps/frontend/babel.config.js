module.exports = function (api) {
  api.cache(true)
  return {
    presets: ['babel-preset-expo'],
    plugins: [
      [
        '@tamagui/babel-plugin',
        {
          components: ['tamagui'],
          config: './tamagui.config.ts',
          logTimings: true,
        },
      ],
      // reanimated v4 moved its worklet transform into react-native-worklets;
      // it must be the last plugin in the list.
      'react-native-worklets/plugin',
    ],
  }
}
