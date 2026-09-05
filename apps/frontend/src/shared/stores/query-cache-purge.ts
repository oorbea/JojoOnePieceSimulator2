// Indirection so session.store.ts - imported by nearly everything, including
// plenty of Jest suites that never mount the app's providers - never
// statically pulls in @tanstack/react-query or the native toast lib
// (shared/lib/toast.ts, via providers/query-client.ts's MutationCache).
// query-client.ts registers the real implementation once, at real app
// startup (it's imported by providers/query-provider.tsx, mounted at the
// app root); until then this is a harmless no-op, which is exactly what a
// unit test that never renders <QueryProvider> needs.
let purge: () => Promise<void> = async () => {}

export function registerQueryCachePurge(fn: () => Promise<void>): void {
  purge = fn
}

// Called from session.store.ts's clearSession() on logout.
export function clearPersistedQueryCache(): Promise<void> {
  return purge()
}
