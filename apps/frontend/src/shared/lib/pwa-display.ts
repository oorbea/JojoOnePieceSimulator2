// True when the app runs as an installed PWA (home-screen icon, no browser
// chrome) rather than in a browser tab. Always false off web.
export function isStandaloneDisplay(): boolean {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') return false
  // `navigator.standalone` is the iOS Safari spelling of the same thing.
  const iosStandalone = (navigator as Navigator & { standalone?: boolean }).standalone === true
  return iosStandalone || window.matchMedia?.('(display-mode: standalone)').matches === true
}
