import { readFileSync, readdirSync, statSync } from 'fs'
import { join } from 'path'

// Guards a test-hygiene decision: `render()` from @testing-library/react-native
// (14.x) is async, and it only queues the component's `unmount` for the
// automatic afterEach cleanup AFTER its internal `act` has finished. A test
// that calls it without `await` can reach that cleanup first, so the component
// is never unmounted and leaks into the next test, still subscribed to
// whatever stores it reads. Under a loaded CI runner this showed up as
// picture-events-bridge's "network error ... re-mints" test seeing two mints:
// the previous test's bridge was still mounted and reconnected when the next
// test set an admin session (see ObsidianVault/flaky-picture-events-bridge-
// 2026-10-09.md). Always `await render(...)` / `await renderWithProviders(...)`.
const SRC_ROOT = join(__dirname, '..', '..')

function listTestFiles(dir: string, out: string[] = []) {
  for (const entry of readdirSync(dir)) {
    if (entry === 'node_modules') continue
    const full = join(dir, entry)
    if (statSync(full).isDirectory()) listTestFiles(full, out)
    else if (/\.test\.[tj]sx?$/.test(entry)) out.push(full)
  }
  return out
}

// Keeps the newlines of a block comment so reported line numbers match the file.
function stripComments(source: string) {
  return source
    .replace(/\/\*[\s\S]*?\*\//g, (block) => block.replace(/[^\n]/g, ''))
    .replace(/\/\/.*$/gm, '')
}

// A statement that starts with a bare render call: nothing awaits it, returns
// it or assigns it.
const BARE_RENDER_RE = /^\s*(render|renderWithProviders)\(/

describe('tests await render()', () => {
  const files = listTestFiles(SRC_ROOT)

  it('found test files to check', () => {
    expect(files.length).toBeGreaterThan(0)
  })

  it.each(files.map((f) => [f.replace(SRC_ROOT, ''), f]))('%s', (_label, file) => {
    const offendingLines = stripComments(readFileSync(file as string, 'utf8'))
      .split('\n')
      .map((text, i) => ({ line: i + 1, text: text.trim() }))
      .filter(({ text }) => BARE_RENDER_RE.test(text))

    expect(offendingLines).toEqual([])
  })
})
