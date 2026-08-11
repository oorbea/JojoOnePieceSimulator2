import { useEffect, useState } from 'react'

// Delays reflecting `value` by `delayMs` — used to turn a search field's
// every-keystroke updates into one query after the user pauses, instead of
// firing a request per key. The input itself should still render the
// immediate `value`; only the value fed into a query key should be
// debounced, e.g. `useStands({ q: useDebouncedValue(search) })`.
export function useDebouncedValue<T>(value: T, delayMs = 300): T {
  const [debounced, setDebounced] = useState(value)

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delayMs)
    return () => clearTimeout(timer)
  }, [value, delayMs])

  return debounced
}
