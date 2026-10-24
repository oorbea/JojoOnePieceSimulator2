import type { ComponentRef } from 'react'
import type { ScrollView, View } from 'react-native'

// RN 0.88's `View`/`ScrollView` are plain function components, so `View` as a
// type no longer names the instance a ref receives. These name it instead.
export type ViewRef = ComponentRef<typeof View>
export type ScrollViewRef = ComponentRef<typeof ScrollView>
