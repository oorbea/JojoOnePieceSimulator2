// Thin seam over `window.location`. jsdom (jest 30) makes `window.location`
// non-configurable, so tests can no longer swap it for a fake; they mock this
// module instead. Web-only: callers must already be on `Platform.OS === 'web'`.
export const browserLocation = {
  get origin(): string {
    return window.location.origin
  },
  get pathname(): string {
    return window.location.pathname
  },
  get hash(): string {
    return window.location.hash
  },
  assign(url: string): void {
    window.location.assign(url)
  },
}
