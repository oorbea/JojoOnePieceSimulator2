import { Image, Modal, Pressable } from 'react-native'
import { useTranslation } from 'react-i18next'
import { YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'

type ImageLightboxProps = {
  visible: boolean
  uri: string | null
  onClose: () => void
}

// Full-screen preview for Stand/Devil Fruit/Stage thumbnails, in both the
// admin panel and the read-only public catalog. Tap anywhere to dismiss -
// same Modal + dimmed backdrop recipe as ConfirmSheet, but no inner
// GlassPanel so the image can use the full viewport instead of being boxed
// in.
export function ImageLightbox({ visible, uri, onClose }: ImageLightboxProps) {
  const { t } = useTranslation()
  if (!uri) return null

  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onClose} statusBarTranslucent>
      <Pressable onPress={onClose} style={{ flex: 1 }} {...a11yProps(t('common.closePreview'), 'button')}>
        <YStack flex={1} items="center" justify="center" bg="rgba(10,12,20,0.9)" p="$4">
          <Image source={{ uri }} style={{ width: '100%', height: '100%' }} resizeMode="contain" />
        </YStack>
      </Pressable>
    </Modal>
  )
}
