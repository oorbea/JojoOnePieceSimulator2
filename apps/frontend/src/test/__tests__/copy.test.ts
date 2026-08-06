import { readFileSync, readdirSync, statSync } from 'fs'
import { join } from 'path'

// Guards a copy decision: no em dash (—) or en dash (–) in user-visible
// strings — the owner wants text written "more humanly" than that. Strips
// comments first so this only ever flags actual string literals, and
// doesn't punish `// like this — a comment` explaining code elsewhere.
const SRC_ROOT = join(__dirname, '..', '..')
const DASH_RE = /[—–]/

function listSourceFiles(dir: string, out: string[] = []) {
  for (const entry of readdirSync(dir)) {
    if (entry === 'node_modules' || entry === '__tests__') continue
    const full = join(dir, entry)
    const stats = statSync(full)
    if (stats.isDirectory()) {
      listSourceFiles(full, out)
    } else if (
      (/\.(ts|tsx)$/.test(entry) && !/\.test\.[tj]sx?$/.test(entry)) ||
      // i18n locale catalogs (shared/i18n/locales/*.json) carry the same
      // user-visible copy as the rest of src/, just outside .ts/.tsx - they
      // must follow the same em/en dash rule.
      /\.json$/.test(entry)
    ) {
      out.push(full)
    }
  }
  return out
}

// Strips `//...` line comments and `/*...*/` block comments without a full
// parser — good enough for this repo's code, which never puts `//` or `/*`
// inside a string literal that itself needs a dash check.
function stripComments(source: string) {
  return source.replace(/\/\*[\s\S]*?\*\//g, '').replace(/\/\/.*$/gm, '')
}

describe('user-facing copy has no em/en dashes', () => {
  const files = listSourceFiles(SRC_ROOT)

  it('found source files to check', () => {
    expect(files.length).toBeGreaterThan(0)
  })

  it.each(files.map((f) => [f.replace(SRC_ROOT, ''), f]))('%s', (_label, file) => {
    const withoutComments = stripComments(readFileSync(file as string, 'utf8'))
    const offendingLines = withoutComments
      .split('\n')
      .map((line, i) => ({ line: i + 1, text: line }))
      .filter(({ text }) => DASH_RE.test(text))

    expect(offendingLines).toEqual([])
  })
})
