import { createContext, useContext } from 'react'

import type { ManualSectionId } from './manual-sections'

// Lets any screen open one section of the manual in an overlay without
// navigating away (navigating out of a live game would ask the player to leave
// it). The provider is mounted once, in the signed-in app layout, by the manual
// feature; outside it (login, tests) opening is a no-op.
type OpenManual = (section: ManualSectionId) => void

const noop: OpenManual = () => {}

export const ManualOverlayContext = createContext<OpenManual>(noop)

export function useOpenManual(): OpenManual {
  return useContext(ManualOverlayContext)
}
