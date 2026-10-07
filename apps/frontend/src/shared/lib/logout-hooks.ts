// Work that must happen while the session is still valid, right before it is
// torn down - e.g. telling the server this device no longer wants push
// notifications for the account that is signing out. Features register a hook;
// the session store runs them from clearSession without knowing who they are.
type LogoutHook = () => Promise<void>

const hooks = new Set<LogoutHook>()
let running = false

export function registerLogoutHook(hook: LogoutHook): () => void {
  hooks.add(hook)
  return () => {
    hooks.delete(hook)
  }
}

function withTimeout(work: Promise<void>, ms: number): Promise<void> {
  return new Promise((resolve) => {
    const timer = setTimeout(resolve, ms)
    work.then(
      () => {
        clearTimeout(timer)
        resolve()
      },
      () => {
        clearTimeout(timer)
        resolve()
      }
    )
  })
}

// Runs every hook concurrently, each bounded by `timeoutMs`; a slow or failing
// hook never blocks or breaks the sign-out. Re-entrant calls return at once:
// a hook's own API request can 401 and bounce back into clearSession, which
// would otherwise run the hooks again and loop.
export async function runLogoutHooks(timeoutMs = 3000): Promise<void> {
  if (running) return
  running = true
  try {
    await Promise.all([...hooks].map((hook) => withTimeout(hook(), timeoutMs)))
  } finally {
    running = false
  }
}
