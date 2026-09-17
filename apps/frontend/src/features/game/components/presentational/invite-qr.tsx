import QRCode from 'react-native-qrcode-svg'
import { YStack } from 'tamagui'

type Props = {
  value: string
  size?: number
}

// Renders entirely client-side via react-native-svg (an inline <svg> on
// web) - no network call, no external QR-image service. Required under the
// app's CSP (default-src 'self': an <img src> pointed at a third-party QR
// API would be blocked outright), and the right call anyway: the invite URL
// never leaves the device to get turned into a picture.
export function InviteQr({ value, size = 200 }: Props) {
  return (
    <YStack items="center" p="$3" bg="white" rounded="$card">
      <QRCode value={value} size={size} />
    </YStack>
  )
}
