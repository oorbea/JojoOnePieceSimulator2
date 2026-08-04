import { Toaster } from 'burnt/web'

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
const WebToaster = Toaster as unknown as (props: { position?: string }) => React.JSX.Element

export function ToasterMount() {
  return <WebToaster position="bottom-right" />
}
