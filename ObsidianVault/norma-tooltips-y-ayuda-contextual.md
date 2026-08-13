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

## Deuda deliberada (no olvidado, decisión de alcance)

- `ChannelBarItem`/`ThemeToggle` (la barra de navegación) y `WiiCard interactive` (las tarjetas de
  modo Gauntlet/Versus, que ya muestran texto completo + hint permanente) **no** recibieron tooltip en
  este pase — fuera del alcance de las cuatro pantallas que el owner señaló, y de valor marginal donde
  ya hay texto visible. Si se retoma: mismo patrón (`useTooltipTrigger` + `TooltipBubble`), sin
  envolver el pressable existente.
- El posicionamiento horizontal de `TooltipBubble` solo se centra pixel-perfect en web (`transform:
  translateX(-50%)`, condicionado por `isWeb` igual que el blur de `WEB_BLUR_STYLE` — RN nativo no
  admite porcentajes en `transform`). En nativo ancla desde el punto medio del trigger sin corregir,
  aceptable para las etiquetas cortas que lleva.

Related: [[a11y-web-leak]], [[frontend-responsive-frutiger-aero]], [[norma-diseno-ui-ux]],
[[game-lobby-frontend]], [[zettelkasten-workflow]].
