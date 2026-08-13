---
title: "Norma: tooltip obligatorio + ayuda contextual ('?')"
tags:
  - norma
  - jojo-onepiece-simulator
  - frontend
---

# Norma: todo botón lleva tooltip; conceptos llevan `InfoHint`

A petición del owner (2026-08-13, pase de UX sobre `/play/*`): **todo botón de la UI debe llevar
tooltip**, de cara adelante también para pantallas nuevas. Se implementó como primitiva reutilizable
en vez de anotar cada call site.

## Qué se construyó

- `shared/components/presentational/tooltip.tsx`: `useTooltipTrigger(label?)` +
  `TooltipBubble({ visible, label })`. Web muestra por `onHoverIn`/`onHoverOut`/`onFocus`/`onBlur`
  (un ratón no tiene long-press); nativo no tiene hover, así que muestra por `onLongPress` con
  auto-ocultado a los ~1.8s. Los handlers se spreadean sobre el MISMO pressable que ya tiene
  `onPress` — envolverlo en un `Pressable` extra le robaría el toque, mismo principio que la regla
  de `pointerEvents`/`a11yProps` de [[a11y-web-leak]].
- `GlossButton` ganó una prop `tooltip?: string | null` que por defecto reutiliza
  `accessibilityLabel` (`tooltip ?? accessibilityLabel`). Con esto, **todos** los botones existentes
  y futuros construidos con `GlossButton` (el 100% de los botones reales del proyecto) quedan
  cubiertos sin tocar sus call sites. `tooltip={null}` opta por que un botón concreto no lo tenga.
  `ChannelTile` (tiles del home) recibió el mismo tratamiento directamente (no pasa por
  `GlossButton`).
- `shared/components/presentational/info-hint.tsx`: icono `?` (`GlossButton tone="glass" btnSize="sm"
  shape="circle"`) para explicar un **concepto** (qué es una opción), no el estado de un control — eso
  ya lo cubre el texto de ayuda en línea existente (`modeGauntletHint`, `allowBotsGauntletHint`, etc.).
  Sirve tooltip por hover Y una `SpeechBubble` persistente al tocar, para que la explicación sea
  alcanzable con un solo tap en cualquier plataforma, no solo en las que tienen hover.
- `shared/components/presentational/setting-row.tsx`: primitiva de alineación
  etiqueta-izquierda/control-derecha (fila en `$md`, apilado en móvil) — arregla el descentrado de
  "Privacidad"/"Permitir bots" reportado por el owner. Causa raíz: el wrapper `YStack items="center"
  justify="center"` de `GlossButton` no tiene ancho propio, así que se estira a la columna y centra el
  pill bajo una etiqueta a la izquierda. `SettingRow` no toca `GlossButton` (20+ call sites dependen
  de su tamaño actual); en su lugar da al control un contenedor `self="flex-start"` que se ajusta al
  contenido.

## Cobertura completa (2026-08-13, segundo pase)

Los items que se dejaron fuera del primer pase también llevan tooltip ahora, a petición explícita del
owner ("ponle tooltips a lo que dudabas"):
- `ChannelBarItem` (nav items de la barra superior/dock inferior + botón de logout en
  `app-shell.tsx`) y `ThemeToggle` — convertido de `styled(YStack, ...)` bare a un wrapper función
  (mismo patrón que `GlossButton`), con prop `tooltip?: string` explícita (no hay un
  `accessibilityLabel` propio del que derivar por defecto, porque los call sites ya construyen su
  propio `a11yProps(...)` fuera del componente).
- `WiiCard interactive` en los selectores de modo Gauntlet/Versus (`create-lobby-screen.tsx`,
  `lobby-config-panel.tsx`) — el tooltip usa la clave `game.create.help.mode` (ya existía en las tres
  locales desde el primer pase, sin usar hasta ahora). **No** se tocó la primitiva `wii-card.tsx`
  en sí: tiene `overflow:'hidden'` propio, así que el bubble se renderiza en un `YStack
  position="relative"` envolvente (mismo `flexBasis`/`grow` que antes llevaba el `WiiCard`), no
  dentro de la tarjeta — meterlo dentro lo habría recortado.

**Bug encontrado y arreglado de paso:** `ChannelTile` (los tiles del home) sí tiene
`overflow:'hidden'` propio, y el tooltip se había metido como hijo directo en el primer pase — el
bubble (que se posiciona por encima de la caja) habría quedado recortado sin que ningún test lo
detectara (RNTL no renderiza layout real). Se corrigió con el mismo patrón de envoltorio
`position:'relative'` sin `overflow:'hidden'` por fuera.

El posicionamiento horizontal de `TooltipBubble` solo se centra pixel-perfect en web (`transform:
  translateX(-50%)`, condicionado por `isWeb` igual que el blur de `WEB_BLUR_STYLE` — RN nativo no
  admite porcentajes en `transform`). En nativo ancla desde el punto medio del trigger sin corregir,
  aceptable para las etiquetas cortas que lleva.

## Reescritura a Modal (2026-08-13, tercer pase) - el diseño con `position:absolute` estaba roto de raíz

El owner reportó tooltips/popovers cortados y "por detrás de otros elementos" en varias pantallas.
Causa real: RN solo compara `zIndex` entre **hermanos directos** (la misma razón por la que
`ConfirmSheet` es un `Modal` de verdad y no un overlay absoluto, ver
[[frontend-responsive-frutiger-aero]]) - una burbuja anidada varios niveles dentro de un formulario
quedaba recortada por el `overflow:hidden` de un ancestro (`GlassPanel`, `WiiCard`, `ChannelTile`) o
pintada por debajo de un hermano no relacionado más abajo en el árbol, según qué contenedor tuviera
su propio contexto de apilamiento. El primer diseño (bubble como hijo absoluto anidado) nunca podía
arreglarse del todo con retoques de z-index.

**Fix**: `useTooltipTrigger`/`TooltipBubble` (`tooltip.tsx`) ahora miden la posición real en pantalla
del trigger vía `.measure()` (RN) y renderizan la burbuja dentro de un `Modal transparent` de verdad,
igual que `ConfirmSheet`/`GlassSelect` - un layer de raíz, inmune a cualquier overflow/scroll/z-index
de ancestros. `GlossButton`, `ChannelTile` y `ChannelBarItem` adjuntan un `ref` interno al nodo host
para la medición; ya no necesitan el wrapper `position:relative` sin `overflow:hidden` que el pase
anterior usó como parche (`ChannelTile` tenía ese wrapper y AÚN ASÍ el tooltip se recortaba - la causa
real nunca fue solo el `overflow:hidden` local). Centrado horizontal sigue siendo solo-web (`transform:
translateX(-50%)`), nativo ancla desde la esquina superior-izquierda del trigger sin el nudge final.

`InfoHint` recibió el mismo tratamiento: el popover persistente (tap-to-toggle) ahora es un `Modal`
anclado por `.measure()` sobre un `View ref={...} collapsable={false}` que envuelve el `GlossButton`
(necesario porque `GlossButton` no es `forwardRef`).

Test (`tooltip.test.tsx`) stubea `triggerRef.current` con un `.measure()` falso, porque
react-test-renderer no implementa el puente nativo real de `measure()`.

## `SpeechBubble`: la cola ya no usa un offset fijo del 18%/40%

Todos los usos reales (`play-hub-screen.tsx`, `login-screen.tsx`, `error-fallback.tsx`,
`info-hint.tsx`) son diálogos centrados de ancho variable - una cola fija a `l:'18%'` solo coincidía
por casualidad con una anchura de caja concreta y se veía descentrada en cualquier otra (el bug
"diálogo mal centrado" del hub). La cola ahora se centra en ambos ejes (mismo truco `transform`
web-only que `TooltipBubble`); en web hay que combinar `rotate="45deg"` y el `translate` en el MISMO
array de `style` (RN aplana estilos por clave completa, no hace merge de arrays `transform` - un
`style` aparte con solo `translateX` se habría comido la rotación).

Related: [[a11y-web-leak]], [[frontend-responsive-frutiger-aero]], [[norma-diseno-ui-ux]],
[[game-lobby-frontend]], [[zettelkasten-workflow]].
