import { Compass } from '@tamagui/lucide-icons-2'
import { Link } from 'expo-router'

import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { SpeechBubble } from '@/shared/components/presentational/speech-bubble'
import { WiiCard } from '@/shared/components/presentational/wii-card'

// Outside the (app) group, so this has no nav shell — PageShell supplies
// its own animated backdrop.
export default function NotFoundScreen() {
  return (
    <PageShell align="center">
      <WiiCard tone="glass" aspect="square" width={88} items="center" justify="center">
        <Compass size={44} color="$channelActive" strokeWidth={2} />
      </WiiCard>
      <SpeechBubble tailSide="bottom" tone="strong">
        <GlowText level="title" align="center">
          This screen doesn&apos;t exist.
        </GlowText>
      </SpeechBubble>
      <Link href="/" asChild>
        <GlossButton tone="blue" btnSize="lg" flare>
          Go to home
        </GlossButton>
      </Link>
    </PageShell>
  )
}
