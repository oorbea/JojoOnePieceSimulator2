import { Toaster } from 'burnt/web'
import { useMedia } from 'tamagui'

import { DOCK_BAR_BOTTOM_OFFSET, DOCK_BAR_CLEARANCE, DOCK_BAR_HEIGHT } from '@/shared/lib/layout'

// Clears the floating bottom ChannelBar dock — it only renders below the
// `$md` tier (AppShell swaps it for top nav links from `$md` up, see
// app-shell.tsx). ToasterMount is mounted above AppShell in app-providers.tsx
// so it has no NavInsetsProvider to read from; `useMedia()` here is the
// direct equivalent of the same check AppShell makes.
const DOCK_OFFSET = `${DOCK_BAR_BOTTOM_OFFSET + DOCK_BAR_HEIGHT + DOCK_BAR_CLEARANCE}px`
const PLAIN_OFFSET = '24px'

// burnt's web backend (sonner) needs its <Toaster/> mounted once at the
// app root, or every toast() call fails silently with a console warning —
// see burnt's README "Web Support". Native platforms use their own overlay
// and need no mount point, hence the .native.tsx no-op sibling.
//
// The cast below works around a type-only issue: this project's
// `moduleSuffixes: [".web", ".native", ""]` (tsconfig.json, needed for our
// own bubble-field.web/.native-style files) makes tsc resolve burnt/web's
// internal "../build/web" re-export as "../build/web.native" (a real file
// that happens to exist in burnt's package, typed `() => null`) instead of
// "../build/web" (the real sonner-backed Toaster). Metro's bundler-time
// platform resolution isn't affected - only tsc's static type looks at the
// wrong sibling - so the cast restores the correct callable shape without
// changing what actually runs in the browser.
const WebToaster = Toaster as unknown as (props: {
  position?: string
  offset?: string
}) => React.JSX.Element

export function ToasterMount() {
  const media = useMedia()
  return <WebToaster position="bottom-right" offset={media.md ? PLAIN_OFFSET : DOCK_OFFSET} />
}
