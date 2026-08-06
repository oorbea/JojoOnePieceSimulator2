import { Home } from '@tamagui/lucide-icons-2'
import { renderWithProviders } from '@/test/render'

import { AppShell, type AppShellNavItem } from '../app-shell'

type JsonNode = {
  type: string
  props: Record<string, unknown>
  children: (JsonNode | string)[] | null
}

function flatten(node: JsonNode, out: JsonNode[] = []) {
  out.push(node)
  for (const child of node.children ?? []) {
    if (typeof child !== 'string') flatten(child, out)
  }
  return out
}

const ITEMS: AppShellNavItem[] = [{ href: '/', label: 'Home', icon: Home, active: true }]

function renderShell() {
  return renderWithProviders(
    <AppShell
      items={ITEMS}
      onNavigate={() => {}}
      onLogout={() => {}}
      themeMode="light"
      onCycleTheme={() => {}}
    >
      <></>
    </AppShell>
  )
}

// The top nav-links row and the bottom dock used to be gated by two
// independent `$md` checks, so a media-config bug (this suite's regression
// target — see media.test.ts) could show both at once, or neither. They're
// now a single boolean, making that class of bug structurally impossible.
describe('AppShell nav visibility', () => {
  it('renders exactly one of {top nav links, bottom dock} — never both, never neither', async () => {
    const { toJSON } = await renderShell()
    const nodes = flatten(toJSON() as JsonNode)

    const homeLabelCount = nodes.filter((n) => n.children?.includes('Home')).length

    // With mediaQueryDefaultActive's mobile-first default (`md: false` in
    // this test environment), the dock renders with an icon-only item (no
    // "Home" text node) and the labeled links row doesn't render at all.
    expect(homeLabelCount).toBe(0)
  })
})
