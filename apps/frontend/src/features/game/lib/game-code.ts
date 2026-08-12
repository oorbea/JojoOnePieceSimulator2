// Mirrors the backend's codeAlphabet (apps/backend .../application/services/
// game_service.go) - no 0/O/1/I, so a spoken/typed code never gets confused.
export const CODE_ALPHABET = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
export const CODE_LENGTH = 6
const CODE_PATTERN = new RegExp(`^[${CODE_ALPHABET}]+$`)

// normalizeCode uppercases, strips anything outside the alphabet, and caps
// at CODE_LENGTH - safe to call on every keystroke of a join-code input.
export function normalizeCode(raw: string): string {
  return raw
    .toUpperCase()
    .split('')
    .filter((ch) => CODE_ALPHABET.includes(ch))
    .slice(0, CODE_LENGTH)
    .join('')
}

export function isCompleteCode(code: string): boolean {
  return code.length === CODE_LENGTH && CODE_PATTERN.test(code)
}

// formatCode splits a complete code into two groups of three for display,
// e.g. "K7F2QX" -> "K7F 2QX". Purely cosmetic - never used for validation.
export function formatCode(code: string): string {
  if (code.length !== CODE_LENGTH) return code
  return `${code.slice(0, 3)} ${code.slice(3)}`
}
