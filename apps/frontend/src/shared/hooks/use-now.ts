import { useEffect, useState } from 'react'

import { serverNow } from '@/shared/lib/server-clock'

// useNow ticks a serverNow() snapshot every intervalMs while active - the
// countdown pattern this codebase had hand-rolled independently in
// voting-status-bar.tsx (1000ms) and connection-banner.tsx (500ms). Both
// were migrated onto this hook in the same pass that added it, so a third
// copy never appears. `active` lets a caller stop the interval entirely
// once there's nothing left to count down (matches the two hand-rolled
// versions' own `if (x === null) return` early-outs).
//
// serverNow() (Date.now() corrected for this device's estimated clock
// offset from the server, see shared/lib/server-clock.ts) rather than
// Date.now() directly - a countdown here is always measured against a
// server-stamped deadline (closesAt/*EndsAt), so a client whose own clock
// is skewed used to read a countdown that was wrong by exactly that skew.
// This is what fixed the >5s discrepancy a real two-device playtest showed
// on the loadout-summary countdown between a PC and a phone with slightly
// different clocks.
export function useNow(intervalMs: number, active: boolean): number {
  const [now, setNow] = useState(() => serverNow())

  useEffect(() => {
    if (!active) return
    const id = setInterval(() => setNow(serverNow()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs, active])

  return now
}
