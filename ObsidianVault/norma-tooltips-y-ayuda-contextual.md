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

## Web: `Modal` de react-native-web se comía el hover/click de TODA la pantalla (2026-08-14, cuarto pase)

El owner reportó que en web los botones "dejaban de funcionar", como si estuvieran constantemente
haciendo hover y dejándolo de hacer. Reproducido: al pasar el ratón por un botón con tooltip, el click
tardaba varios intentos en registrar en cualquier botón de la pantalla, no solo en el que tenía el
tooltip abierto.

**Causa real**: el tercer pase movió `TooltipBubble` a un `Modal` de verdad para librarse de los
problemas de z-index/overflow (bien, ver arriba) — pero en web, `Modal` es una implementación de
`react-native-web`, no un modal nativo del SO. Esa implementación son tres wrappers anidados
(`ModalPortal` → `ModalAnimation` → `ModalFocusTrap` → `ModalContent`), y **dos** de ellos renderizan
sus propios `<div>` `position:fixed; inset:0` cubriendo la pantalla entera:
- `ModalContent` sí reenvía las props sueltas (`rest`) al `View` que renderiza — ahí un
  `pointerEvents="none"` en el `<Modal>` sí llega.
- `ModalAnimation` (la capa que va POR ENCIMA) tiene su propio `<div style={styles.container}>`
  (`position:fixed`, `zIndex:9999`) definido en su propio `StyleSheet`, sin `pointer-events` propio y
  **sin ninguna prop pública para dárselo** — pasar `pointerEvents="none"` al `<Modal>` de RN no llega
  ahí, sólo al hijo de `ModalContent`.

Resultado: mientras el tooltip está `visible`, ese `div` de `ModalAnimation` (pantalla completa,
`pointer-events:auto` por defecto) queda por encima de TODO en el DOM (portal al final de `<body>`).
El ratón, al entrar en el botón, dispara `onHoverIn` → se monta el `Modal` → ese div fixed pasa a estar
encima del cursor → el navegador manda `mouseleave` al botón real (ahora tapado) → `onHoverOut` →
se desmonta el `Modal` → el div desaparece → el ratón vuelve a estar "sobre" el botón → `onHoverIn` de
nuevo… bucle infinito de hover on/off, y cualquier click que cayera a mitad de ese ciclo se lo comía el
div fixed en vez de llegar al botón.

Diagnóstico hecho con `document.elementFromPoint(x, y)` + `getComputedStyle(...).pointerEvents` desde
la consola del navegador sobre el punto de un botón real: devolvía el `<div>` fixed de `ModalAnimation`
(`pointer-events: auto`) en vez del botón, incluso con el `pointerEvents="none"` puesto en `ModalContent`.

**Fix**: en web, `TooltipBubble` ya NO usa `Modal` de RN. Usa `createPortal` de `react-dom` directo a
`document.body` con un `<div style={{position:'fixed', inset:0, pointerEvents:'none'}}>` — el mismo
efecto de "capa de raíz, inmune a overflow/z-index de ancestros" que buscaba el tercer pase, pero sin
pasar por la implementación de `Modal` de `react-native-web` (que no tiene forma de hacerse
transparente al puntero de verdad). Nativo sigue usando `Modal` tal cual - ahí es un modal real del
SO, sin este problema.

**Lección para cualquier `Modal` transparente/decorativo en web** (tooltips, popovers, badges
flotantes - cualquier cosa que NO deba capturar el puntero): no basta con `pointerEvents="none"` en el
`<Modal>` de RN en web. Hay que evitar `Modal` de RN en la rama web y usar un `createPortal` a
`document.body` propio, o verificar explícitamente con `elementFromPoint` que no queda ningún wrapper
intermedio con `pointer-events: auto` tapando la pantalla.

## Quinto pase (2026-08-14, mismo día): la burbuja se salía de pantalla en los bordes

Con el bug del hover ya arreglado, el owner encontró el siguiente: el tooltip del botón "Admin"
(nav superior derecha) se recortaba contra el borde superior de la ventana, y cualquier trigger cerca
del borde derecho lo habría hecho también contra ese lado. Causa: `TooltipBubble` centraba siempre el
eje X sobre el punto medio del trigger (`l: anchor.x + anchor.width/2` + `translateX(-50%)`) y siempre
anclaba por ENCIMA (`t: anchor.y - 8` + `translateY(-100%)`), sin comprobar nunca si quedaba sitio -
ninguno de los dos ejes se clampeaba contra el viewport.

**Fix** (mismo `tooltip.tsx`, sólo rama web): tras el primer render (con la burbuja ya montada en su
posición "natural" sin clampear), un `useLayoutEffect` mide el `offsetWidth`/`offsetHeight` real del
nodo vía un `ref` en el `YStack` y corrige antes del siguiente paint (sin parpadeo visible):
- Horizontal: el centro se clampea a `[mitad+8px, innerWidth-mitad-8px]`.
- Vertical: si no cabe por encima del trigger (`anchor.y - 8 - height < 8px`), se cambia de sitio y se
  ancla DEBAJO (`t: anchor.y + anchor.height + 8`, `translateY(0%)` en vez de `-100%`) en lugar de
  encima.

Nativo no se toca - sigue sin el `translateX(-50%)`/clamp (ancla sin corregir desde la esquina
superior-izquierda, norma ya aceptada en el tercer pase de arriba).

Related: [[a11y-web-leak]], [[frontend-responsive-frutiger-aero]], [[norma-diseno-ui-ux]],
[[game-lobby-frontend]], [[zettelkasten-workflow]].
