---
title: "Frontend: Responsive + Frutiger Aero"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - design
  - decision
---

# Frontend — Responsive + Wii Party / Aero / iOS-gloss

## Decision (superseded 2026-08-04, direction evolved not replaced)

All frontend pages and components must be:

1. **100% responsive** — desktop and mobile, no exceptions.
2. **Wii Party / Wii-era Nintendo as primary aesthetic**, Windows Aero glass and iOS 1–6 glossy icons as secondary layers, at full theme-park intensity: supersaturated Wii palette (`$wiiBlue`, `$sunYellow`, `$meadowGreen`, `$bubblegum`, `$tangerine`, `$grapeSoda`) + polished-plastic neutrals for shape language; Aero translucent glass panels with real `backdrop-filter` blur on web; iOS-6 top-half gloss highlight on every glossy surface. Fredoka (display) + Nunito (body) replaced Inter.
3. The original Frutiger Aero decision (soft gradients, glass, rounded corners) is not reversed — it's the Aero half of this combination. What changed 2026-08-04 is adding Wii Party as the *primary* layer on top, at the owner's request.

## Why

Owner's explicit design requirement (2026-07-31, evolved 2026-08-04). Applies to every screen built from now on.

## How to apply

- Every layout must use responsive patterns (flex/grid, relative units, Tamagui `$md`/`$lg` breakpoints — v4's `mediaQueryDefaultActive` is mobile-first, so write the mobile layout first and override up, not down).
- **Styling entry point is the token/primitive layer, not ad-hoc props.** `apps/frontend/tamagui.config.ts` holds the palette (`tokens.color`) and semantic theme keys (`glassFill`, `channelActive`, `softShadow`, etc. — scheme-aware, defined once on `light`/`dark`, never hardcoded rgba in a screen). `apps/frontend/src/shared/components/presentational/` holds the primitives: `GlassPanel`, `GlossButton`, `WiiCard`, `ChannelTile`, `ChannelBar`, `SpeechBubble`, `GlowText`, `PageShell`, `AquaBackground`. New screens compose these; they don't reinvent glass/gloss/shadow recipes inline.
- Signature element: the glossy channel tile (`ChannelTile`) — solid tone base, gloss cut across the top half, physical `borderBottomWidth` lip that compresses on press. Same press-compression recipe on every `GlossButton`.
- Test on both desktop and mobile viewports, light AND dark theme, before considering any page done.

## RN substitution rules (don't re-derive — no CSS engine here)

Stack is Expo/React Native Web + Tamagui 2.6.2, not Tailwind/CSS. Substitutes used throughout the primitive layer:

- No `::before` → a real absolutely-positioned overlay child (`GlossOverlay`).
- No radial gradient (`@tamagui/linear-gradient` doesn't support it) → 5 concentric circles approximating a Gaussian falloff, `filter: blur()` on web only (`LensFlare`).
- No CSS `inset` box-shadow → two absolute children, a bright border ring + a bottom shading gradient (`InsetRing`/`InsetShade` in `wii-card.tsx`).
- No CSS triangle border → a small rotated square with two borders drawn (`SpeechBubble`'s tail).
- `backdrop-filter` is web-only → a `Platform.OS === 'web'` constant passed via `style=` (never as a Tamagui style prop, or the compiler tries to statically extract it); native compensates with a higher-alpha glass fill instead of blur.
- Gloss/sheen/inset layers are static — only `transform`/`opacity` ever animate, so they cost nothing per frame.
- Animated backdrop (`AquaBackground`/`BubbleField`) is platform-split by filename (`.web.tsx` = CSS `@keyframes`, `.native.tsx` = Reanimated 4 worklets) — zero per-frame JS on either platform. `tsconfig.json` needs `moduleSuffixes: [".web", ".native", ""]` for `tsc` to resolve the platform split the same way Metro does; without it `tsc` can't find the bare `./bubble-field` import.

## Tamagui v4 gotchas hit while building this (don't re-derive)

- **`@tamagui/config/v4`'s `defaultConfig` ships NO `color` token group at all** (only `radius`/`size`/`space`/`zIndex`). This is the root cause of a real bug: `$standPurple`/`$strawHatRed` were used in the original screens but never resolved to anything, because there was no `tokens.color` to register them in. Fixed by adding a `tokens.color` group in `tamagui.config.ts`.
- **Theme lookup wins over token lookup** — a token name that collides with a default theme key (`color*`, `background*`, `border*`, `accent*`, `shadow[1-6]`, bare color-scale names like `blue*`/`red*`) silently becomes unreachable. Screen new token names against this before adding them.
- **Custom keys added to `light`/`dark` themes propagate to all ~294 sub-themes automatically** (Tamagui's `proxyThemeToParents` `Object.assign`s parents first) — safe to add semantic keys once on the two root themes.
- **The `animation` prop from Tamagui v3 docs doesn't exist in this version** — it's `transition` (e.g. `transition="bouncy"`), paired with `enterStyle`/`exitStyle`/`pressStyle`/`hoverStyle`. `animatedBy` is a different prop (selects the animation *driver*, not a named transition).
- **`onlyAllowShorthands: true`** forces shorthand-only for any style prop that HAS a shorthand — `z` not `zIndex`, `self` not `alignSelf`, `minW` not `minWidth`, `pos` is NOT a real shorthand (use `position`, which has none defined and stays available in longhand).
- A custom Tamagui `size`-like variant name can collide with a component's own built-in prop (`scale` collides with the core transform shorthand) — name custom variants defensively (`btnSize`, not `scale`).
- `@tamagui/helpers-icon` (lucide-icons-2's own dependency for `IconProps`) isn't hoisted to a resolvable top-level path under this repo's `node-linker=hoisted` pnpm setup — don't import it directly; declare a local prop-shape type instead.
- `@tamagui/lucide-icons` is deprecated, renamed to `@tamagui/lucide-icons-2` — and it depends on `react-native-svg`, which `expo install` does NOT add automatically; install it explicitly or the production `expo export -p web` bundle fails to resolve it.
- **`app/+html.tsx` has no effect when `app.json`'s `web.output` is `"single"`** (SPA mode) — that customization hook is only consulted by expo-router's per-route static prerendering path. For single-mode, the actual override point is a local `public/index.html` (checked by `@expo/cli` before its built-in template; use `%LANG_ISO_CODE%`/`%WEB_TITLE%` placeholders, omit the auto-injected favicon/CSS/script tags). This is how the PWA manifest link and `<meta name="theme-color">` finally got wired up.
- The extracted `tamagui-web.css` still builds to 0 bytes even after this pass (confirmed via a real `expo export -p web`) — `@tamagui/babel-plugin` CSS extraction isn't producing output in this project, so webfonts on web are carried entirely by `expo-font`'s runtime `@font-face` injection, not by the extracted stylesheet. Not blocking, but don't assume the CSS file has content when debugging web styling.

Related: [[frontend-stack]], [[ADR]], [[norma-diseno-ui-ux]] (obligación de pasar por las skills
frontend-design + ui-ux-pro-max al tocar cualquier pantalla, incluida esta)

## Higiene de layout responsive (2026-08-04) — no re-derivar

Pase de limpieza sobre desktop tras reporte del owner (superposición, texto sin agrupar, botones
inalcanzables, secciones sin centrar). Cuatro causas raíz encontradas en la capa de primitivas, no en
las pantallas — arreglarlas ahí corrigió las tres pantallas de golpe:

- **`GlassPanel` anulaba todo `gap`/`flexDirection`**: envolvía `children` en un único
  `<YStack z="$content" flex={1}>`, dejando al frame un solo hijo en flujo. Cualquier `gap` pasado al
  panel no separaba nada — los bloques de texto de home/profile/login se veían pegados. Fix: los hijos
  se renderizan directos como hijos del frame; `GlossOverlay` sigue absoluto fuera de flujo.
- **Gloss pintaba sobre el texto**: `tamagui.config.ts` tenía `gloss:20 > content:10`. Bajado a
  `gloss:5` (por debajo de `content:10`). Test de regresión en
  `src/shared/lib/__tests__/tokens.test.ts` fija el orden `gloss < content < nav < overlay`.
- **`ChannelBar` no se centraba por encima de 1080px**: las variantes `dock` ponían
  `position:absolute; left:0; right:0` directamente en la píldora, que también tenía
  `maxW:1080; self:'center'` — un absoluto con `left`+`right` fijados ignora `alignSelf`, así que la
  barra (y el botón de logout, empujado por el `flex:1` intermedio) se pegaba al borde izquierdo en
  monitores anchos. Fix: la píldora queda en flujo normal (así `self:'center'` sí funciona) dentro de
  un host absoluto de solo centrado, no interactivo (`pointerEvents:'box-none'`).
- **`AquaBackground` montado dos veces** por ruta autenticada (`AppShell` + `PageShell`). `PageShell`
  ganó una prop `backdrop?: boolean` (default `!navPadding`) para no duplicar el fondo cuando ya vive
  dentro del shell.
- **`ConfirmSheet`** pasó de absolute-en-página a `Modal` real de RN: al ser hermano del contenido y no
  de las `ChannelBar` (que viven un nivel arriba en el árbol), su `z:$overlay` (700) nunca competía de
  verdad con `z:$nav` (500) de las barras — RN sólo compara z-index entre hermanos. `Modal` resuelve
  esto al montar en la raíz.
- Nueva escala de columna por breakpoint en `src/shared/lib/layout.ts` (`columnMaxWidth`): antes solo
  había un salto en `$md` (768) y se quedaba ahí para siempre; ahora crece también en `$lg`/`$xl`,
  tope alineado al `maxW:1080` de la navbar.
- Home: el saludo `"Ready when you are, {firstName}."` se sustituyó por solo `user.username` en
  `GlowText level="hero"` — a petición del owner, sin frase de relleno.

Related: [[a11y-web-leak]] si se vuelve a tocar `a11yProps`.

## Media queries eran max-width, no min-width (2026-08-06) — el bug real detrás del "responsive resuelto"

El pase de 2026-08-04 arregló los síntomas visibles pero no la causa: `@tamagui/config/v4`'s
`defaultConfig.media` define `md`/`lg`/`xl` como **max-width** (`md: {maxWidth:1020}`, verificado en
`node_modules/.pnpm/@tamagui+config@2.6.2*/node_modules/@tamagui/config/dist/esm/media.mjs`), pero
`tamagui.config.ts` hacía `...defaultConfig` sin redefinir `media`, y **todo el código escribía `$md`
como si fuera mobile-first min-width**. El comentario de la línea 27 de esta nota ("v4's
`mediaQueryDefaultActive` is mobile-first") describía solo el *default de activación*, no la dirección
real de la query — engañoso, y probablemente la razón de que nadie lo pillara antes.

Consecuencia real: la fila de nav links con label de `AppShell` (pensada para desktop) se mostraba en
móvil dentro de un `ChannelBar` de `height:64` fijo sin wrap → logo + "JOPS" + 3 items + theme toggle +
logout se recortaban unos sobre otros. El dock inferior (pensado para móvil) solo existía en
`>1020px`. `PageShell`'s `columnMaxWidth` a 390px matcheaba `md`+`lg`+`xl` a la vez y ganaba el último
(`base*1.8`) en vez del primero.

Fix: `tamagui.config.ts` ahora define su propio `media` (min-width real: `sm:640 md:900 lg:1200
xl:1500`, más un alias `maxSm:{maxWidth:639}` para lo que de verdad es "solo móvil") y su propio
`mediaQueryDefaultActive` (todo `false` salvo `maxSm:true` — mobile-first de verdad). Test de guardia en
`src/shared/lib/__tests__/media.test.ts`: cualquier clave que no empiece por `max` debe ser
`minWidth`, y las tiers crecientes deben declararse en orden ascendente.

Segundo cambio del mismo pase — la reserva de espacio para las barras flotantes dejó de ser una
constante (`NAV_BAR_HEIGHT=64`) que cada pantalla pedía a mano vía `PageShell navPadding`: `AppShell`
ahora mide la altura real de ambas barras con `onLayout` y la publica por
`src/shared/lib/nav-insets.tsx` (`NavInsetsProvider`/`useNavInsets`); `PageShell` la lee del contexto en
vez de recibir un prop. Una barra que crece (wrap de contenido) ya no puede quedar tapando la página
porque la reserva crece con ella. `layout.ts`'s `topClearance`/`bottomClearance`/
`desktopBottomClearance` (booleanas) se sustituyeron por `navTopInset`/`navBottomInset` (altura medida o
`null` si la barra no existe en ese tier).

También: la fila de links y el dock pasaron de dos `display`/`$md` independientes a un solo booleano
(`showTopLinks = media.md`) en `AppShell` — estructuralmente imposible que coexistan o que falten los
dos a la vez.
