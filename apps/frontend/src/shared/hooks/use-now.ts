import { useEffect, useState } from 'react'

// useNow ticks a `Date.now()` snapshot every intervalMs while active - the
// countdown pattern this codebase had hand-rolled independently in
// voting-status-bar.tsx (1000ms) and connection-banner.tsx (500ms). Both
// were migrated onto this hook in the same pass that added it, so a third
// copy never appears. `active` lets a caller stop the interval entirely
// once there's nothing left to count down (matches the two hand-rolled
// versions' own `if (x === null) return` early-outs).
export function useNow(intervalMs: number, active: boolean): number {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (!active) return
    const id = setInterval(() => setNow(Date.now()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs, active])

  return now
}
