import { existsSync, readFileSync } from 'fs'
import { join } from 'path'

// public/manifest.json is what the browser reads to offer "install" and to
// draw the home-screen icon/name/shortcuts. A typo in a path or a wrong
// declared size fails silently (the icon just doesn't appear), so pin it.
const PUBLIC_DIR = join(__dirname, '..', '..', '..', 'public')
const manifest = JSON.parse(readFileSync(join(PUBLIC_DIR, 'manifest.json'), 'utf8')) as {
  id: string
  name: string
  short_name: string
  start_url: string
  scope: string
  display: string
  orientation: string
  theme_color: string
  background_color: string
  icons: { src: string; sizes: string; type: string; purpose?: string }[]
  screenshots: { src: string; sizes: string; type: string; form_factor?: string }[]
  shortcuts: { name: string; url: string; icons: { src: string; sizes: string }[] }[]
}

// PNG header: 8-byte signature, then the IHDR chunk whose first 8 data bytes
// are width and height (big endian).
function pngSize(file: string): string {
  const header = readFileSync(file)
  return `${header.readUInt32BE(16)}x${header.readUInt32BE(20)}`
}

function publicFile(src: string): string {
  return join(PUBLIC_DIR, src.replace(/^\//, ''))
}

describe('manifest.json', () => {
  it('names the app and starts it from the root, standalone and portrait', () => {
    expect(manifest.name).toBe('Jojo One Piece Simulator')
    expect(manifest.short_name).toBe('JOPS')
    expect(manifest.id).toBe('/')
    expect(manifest.start_url).toBe('/')
    expect(manifest.scope).toBe('/')
    expect(manifest.display).toBe('standalone')
    expect(manifest.orientation).toBe('portrait')
  })

  it('declares real icons at the installable sizes, any and maskable', () => {
    for (const purpose of ['any', 'maskable']) {
      const sizes = manifest.icons.filter((icon) => icon.purpose === purpose).map((i) => i.sizes)
      expect(sizes).toEqual(expect.arrayContaining(['192x192', '512x512']))
    }
  })

  it.each(manifest.icons.map((icon) => [icon.src, icon] as const))(
    'icon %s exists and is the size it declares',
    (src, icon) => {
      expect(existsSync(publicFile(src))).toBe(true)
      expect(icon.type).toBe('image/png')
      expect(pngSize(publicFile(src))).toBe(icon.sizes)
    }
  )

  it('has at least one narrow screenshot, all the same shape, each as declared', () => {
    expect(manifest.screenshots.some((s) => s.form_factor === 'narrow')).toBe(true)
    const shapes = new Set<string>()
    for (const shot of manifest.screenshots) {
      expect(existsSync(publicFile(shot.src))).toBe(true)
      expect(pngSize(publicFile(shot.src))).toBe(shot.sizes)
      const [w, h] = shot.sizes.split('x').map(Number)
      // Chrome's richer install UI rules: both sides >= 320, ratio <= 2.3.
      expect(Math.min(w, h)).toBeGreaterThanOrEqual(320)
      expect(Math.max(w, h) / Math.min(w, h)).toBeLessThanOrEqual(2.3)
      shapes.add((w / h).toFixed(2))
    }
    expect(shapes.size).toBe(1)
  })

  it('has shortcuts that point at app routes with existing icons', () => {
    expect(manifest.shortcuts.map((s) => s.url)).toEqual(['/play/create', '/play/join'])
    for (const shortcut of manifest.shortcuts) {
      for (const icon of shortcut.icons) {
        expect(existsSync(publicFile(icon.src))).toBe(true)
        expect(pngSize(publicFile(icon.src))).toBe(icon.sizes)
      }
    }
  })

  it('uses the brand colours the app shell uses', () => {
    expect(manifest.theme_color).toBe('#00A0E9')
    expect(manifest.background_color).toBe('#EAF7FF')
  })
})
