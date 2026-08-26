---
title: "Norma: todo control interactivo navegable por teclado"
tags:
  - norma
  - jojo-onepiece-simulator
  - frontend
---

# Norma: teclado como ciudadano de primera clase (2026-08-26)

A petición del owner ("que absolutamente siempre... sea navegable muy fácilmente e intuitivamente
con las flechas del teclado, el tabulador, el enter, el espacio, etc"), norma general de aquí en
adelante para **toda pantalla nueva**, no solo la de votación que la originó (ver
[[game-vote-buttons-2026-08-26]]): Tab alcanza todo control real, flechas mueven dentro de un grupo
relacionado, Enter/Espacio activan, Esc cierra overlays. Web es la plataforma primaria de esto -
nativo no tiene modelo de foco de teclado que recorrer, así que las piezas web-only se degradan a
no-op en nativo, nunca a un crash.

## Patrón de referencia: roving tabindex

`shared/hooks/use-roving-group.ts` - un grupo de opciones (radio group, roster de tiles, cualquier
lista fija pequeña) es UN solo Tab-stop: exactamente un ítem lleva `tabIndex=0`, el resto `-1`.
Flechas (↑↓←→, con wraparound) mueven cuál; Home/End saltan a los extremos; Enter/Espacio activan
el ítem con foco. `isWeb` (de `shared/lib/web-blur.ts`) gatea toda la mecánica - en nativo
`getItemProps` devuelve solo `{ tabIndex }` y el tap/press existente sigue siendo el único camino.

**Por qué enfoca vía `id` + `document.getElementById(...).focus()` y no vía `ref`**: `GlossButton`
ya usa su propio `ref` interno para el `.measure()` del tooltip (ver
[[norma-tooltips-y-ayuda-contextual]]) y no reenvía uno externo - forzar un segundo ref ahí
significaría reescribir esa primitiva. El id-lookup es la misma escapatoria que cualquier
integración que no controla el árbol de refs de una librería de componentes. Mismo patrón
reutilizado en `ConfirmSheet` para el auto-foco del botón de confirmar al abrirse el sheet (sin
`autoFocus` prop - Tamagui/RN no lo expone).

Dos consumidores hoy: la barra de voto (`vote-bar.tsx`, grupo `role="radiogroup"`/`role="radio"`) y
el roster en partida (`match-roster.tsx`, arrastra el foco entre `ParticipantTile`s en el mismo
orden en que se renderizan, abre el modal de loadout con Enter/Espacio).

## `a11yProps` ganó `checked`/`selected`

`shared/lib/a11y.ts` - necesario para un radio group real (`aria-checked` en web,
`accessibilityState.checked` en nativo). `GlossButton` ganó `a11yRole`/`a11yChecked` porque aplicaba
`a11yProps(label, 'button')` DESPUÉS de `...rest`, así que un `role` pasado desde fuera quedaba
pisado silenciosamente - la norma es centralizar en el helper, no saltárselo desde el call site.

## Atajos de una tecla: guardas obligatorias

`features/game/lib/hotkeys.ts`'s `resolveHotkey` (pura, testeada) es el único sitio donde vive la
lógica de qué tecla hace qué - el listener (`hooks/use-match-hotkeys.ts`) es un wrapper de tres
líneas. Toda tecla de un solo carácter (`1`-`9` para votar, `S` para saltar el sorteo) se suprime
por completo si: hay foco en un `input`/`textarea`/`select`/`contenteditable`, hay una tecla
modificadora pulsada, o el caller marca `blocked` (overlay abierto). Sin esto, un atajo de una tecla
dispara mientras alguien escribe en cualquier campo de texto de la misma pantalla.

**Hueco conocido, no silencioso**: hoy `blocked` solo refleja `ConfirmSheet` (el contenedor lo
controla directamente). El `LoadoutModal` vive dentro de `MatchRoster` (deliberado - es el
componente que ya tiene `mangas` en mano para el contenido del modal), así que los atajos no se
suprimen todavía mientras ese modal está abierto. No es grave (no hay ningún input de texto detrás),
pero está pendiente si se quiere cerrar del todo - requeriría subir ese estado o un mecanismo de
"qué overlay está abierto" compartido.

## Esc cierra overlays

`ConfirmSheet` y `LoadoutModal` ya enrutan `onRequestClose` (el `Modal` de RN lo traduce a Esc en
web y al botón atrás en Android) - verificar que un `Modal` nuevo también lo tenga antes de darlo
por cerrable con teclado.

Related: [[norma-tooltips-y-ayuda-contextual]], [[norma-diseno-ui-ux]], [[a11y-web-leak]],
[[game-vote-buttons-2026-08-26]], [[frontend-stack]].
