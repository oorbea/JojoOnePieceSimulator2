// The backend sends ETag + Cache-Control: private on every read and honors
// If-None-Match with a 304 (apps/backend .../cache_headers.go). This tiny
// in-memory map lets any GET automatically become conditional once it has
// seen a response for that exact URL, without every feature reimplementing
// the header dance.
const etagsByUrl = new Map<string, string>()

export function getKnownEtag(url: string): string | undefined {
  return etagsByUrl.get(url)
}

export function rememberEtag(url: string, etag: string | undefined): void {
  if (etag) etagsByUrl.set(url, etag)
}

export function clearEtags(): void {
  etagsByUrl.clear()
}
