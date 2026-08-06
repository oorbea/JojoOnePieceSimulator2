import { createContext, useContext } from 'react'

// Measured (not assumed) reservation for AppShell's floating top bar / bottom
// dock, in px — already inclusive of safe-area inset + offset + clearance
// (see layout.ts's navTopInset/navBottomInset). AppShell publishes this once
// it knows the bars' real rendered height; PageShell consumes it instead of
// re-deriving a fixed NAV_BAR_HEIGHT guess, so a bar that grows (text wraps,
// content overflows one row) never gets silently covered.
//
// The zero default is correct for anything mounted outside AppShell — login,
// +not-found, LoadingScreen — none of which have a floating bar to clear.
export type NavInsets = { top: number; bottom: number }

const ZERO_NAV_INSETS: NavInsets = { top: 0, bottom: 0 }

const NavInsetsContext = createContext<NavInsets>(ZERO_NAV_INSETS)

export const NavInsetsProvider = NavInsetsContext.Provider

export function useNavInsets(): NavInsets {
  return useContext(NavInsetsContext)
}
