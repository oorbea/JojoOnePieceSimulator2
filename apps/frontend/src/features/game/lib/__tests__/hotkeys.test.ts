import { resolveHotkey, type ResolveHotkeyInput } from '@/features/game/lib/hotkeys'

function input(overrides: Partial<ResolveHotkeyInput> = {}): ResolveHotkeyInput {
  return {
    key: '1',
    isTextInputFocused: false,
    blocked: false,
    votingOpen: true,
    optionCount: 2,
    revealing: false,
    summaryOpen: false,
    ...overrides,
  }
}

describe('resolveHotkey', () => {
  it('a digit within range votes that option, zero-indexed', () => {
    expect(resolveHotkey(input({ key: '1' }))).toEqual({ kind: 'vote', index: 0 })
    expect(resolveHotkey(input({ key: '2' }))).toEqual({ kind: 'vote', index: 1 })
  })

  it('an out-of-range digit resolves to null', () => {
    expect(resolveHotkey(input({ key: '9', optionCount: 2 }))).toBeNull()
  })

  it('s skips the sorteo only while it is revealing', () => {
    expect(resolveHotkey(input({ key: 's', revealing: true, votingOpen: false }))).toEqual({ kind: 'skip' })
    expect(resolveHotkey(input({ key: 'S', revealing: true, votingOpen: false }))).toEqual({ kind: 'skip' })
    expect(resolveHotkey(input({ key: 's', revealing: false }))).toBeNull()
  })

  it('s also skips the loadout-summary screen while it is open', () => {
    expect(resolveHotkey(input({ key: 's', summaryOpen: true, votingOpen: false }))).toEqual({ kind: 'skip' })
    expect(resolveHotkey(input({ key: 's', summaryOpen: false, revealing: false }))).toBeNull()
  })

  it('a digit does nothing while voting is not open', () => {
    expect(resolveHotkey(input({ key: '1', votingOpen: false }))).toBeNull()
  })

  it('is suppressed entirely while a text input has focus', () => {
    expect(resolveHotkey(input({ key: '1', isTextInputFocused: true }))).toBeNull()
    expect(resolveHotkey(input({ key: 's', revealing: true, isTextInputFocused: true }))).toBeNull()
  })

  it('is suppressed entirely while an overlay is blocking input', () => {
    expect(resolveHotkey(input({ key: '1', blocked: true }))).toBeNull()
    expect(resolveHotkey(input({ key: 's', revealing: true, blocked: true }))).toBeNull()
  })

  it('a non-digit, non-s key resolves to null', () => {
    expect(resolveHotkey(input({ key: 'a' }))).toBeNull()
    expect(resolveHotkey(input({ key: 'Enter' }))).toBeNull()
  })
})
