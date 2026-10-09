import { useRouter } from 'expo-router'
import { useCallback, useState } from 'react'

import { MANUAL_RULES } from '@/shared/contracts/rules'
import { ManualOverlayContext } from '@/shared/lib/manual-overlay'
import type { ManualSectionId } from '@/shared/lib/manual-sections'
import { useExitGuardStore } from '@/shared/stores/exit-guard.store'

import { ManualOverlay } from '../presentational/manual-overlay'

// Mounted once in the signed-in app layout. Owns which section is open, hands
// every screen the way to open one (shared/lib/manual-overlay), and renders the
// overlay. "Open the full manual" goes through the exit guard like any other
// navigation, so leaving a live game for it still asks first.
export function ManualOverlayProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const requestExit = useExitGuardStore((state) => state.request)
  const [section, setSection] = useState<ManualSectionId | null>(null)

  const close = useCallback(() => setSection(null), [])

  const openFull = useCallback(
    (id: ManualSectionId) => {
      // Close first: the exit guard's confirmation sheet is its own modal and
      // must not stack on top of this one.
      setSection(null)
      requestExit(() => router.navigate({ pathname: '/manual', params: { section: id } } as never))
    },
    [requestExit, router]
  )

  return (
    <ManualOverlayContext.Provider value={setSection}>
      {children}
      <ManualOverlay section={section} rules={MANUAL_RULES} onClose={close} onOpenFull={openFull} />
    </ManualOverlayContext.Provider>
  )
}
