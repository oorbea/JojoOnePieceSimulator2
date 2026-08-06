import { LogIn } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { Image } from 'react-native'
import { Spinner } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { SpeechBubble } from '@/shared/components/presentational/speech-bubble'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import { logoAsset } from '@/shared/assets'
import type { AppError } from '@/shared/api/errors'

type Props = {
  onSignIn: () => void
  isLoading: boolean
  isReady: boolean
  error: AppError | null
}

// Pure UI — Wii Party channel-menu energy dressed in Aero glass. No
// data/session logic lives here.
export function LoginScreen({ onSignIn, isLoading, isReady, error }: Props) {
  const { t } = useTranslation()
  return (
    <PageShell align="center" maxWidth={440}>
      <GlassPanel
        glossy
        radiusSize="hero"
        elevate={3}
        width="100%"
        p="$6"
        items="center"
        gap="$4"
        transition="bouncy"
        enterStyle={{ scale: 0.9, y: 20, opacity: 0 }}
      >
        <WiiCard aspect="square" width={128} height={128} tone="plastic" interactive={false}>
          <Image
            source={logoAsset}
            style={{ width: '100%', height: '100%' }}
            resizeMode="contain"
          />
        </WiiCard>

        <GlowText level="title" align="center">
          {t('auth.appName')}
        </GlowText>
        <GlowText level="label" align="center">
          {t('auth.signInPrompt')}
        </GlowText>

        <GlossButton
          tone="blue"
          btnSize="lg"
          shape="pill"
          flare
          width="100%"
          disabled={!isReady || isLoading}
          icon={
            isLoading ? () => <Spinner color="white" /> : () => <LogIn size={20} color="white" />
          }
          onPress={onSignIn}
          accessibilityLabel={t('auth.continueWithGoogle')}
        >
          {isLoading ? t('auth.signingIn') : t('auth.continueWithGoogle')}
        </GlossButton>

        {error ? (
          <SpeechBubble tailSide="top" tone="strong" width="100%">
            <GlowText level="label" align="center" color="$strawHatRedDeep">
              {error.message}
            </GlowText>
          </SpeechBubble>
        ) : null}
      </GlassPanel>
    </PageShell>
  )
}
