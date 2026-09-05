// Split out from refresh.ts so session.store.ts can reference the same
// SecureStore key (to clear it on logout) without importing refresh.ts's
// axios/session-mapping machinery - avoids a circular import between the
// two modules.
export const REFRESH_TOKEN_KEY = 'jops.refresh'
