---
name: feedback_tooltips_norma
description: Norma de proyecto (2026-08-13) - todo botón de la UI lleva tooltip
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 159fec4a-4b3a-4e9a-9ec9-d6c44323beb0
  modified: 2026-08-13T11:48:00.811Z
---

Owner pidió (2026-08-13, pase de UX sobre /play/*) que **todos los botones de la UI lleven
tooltip**, como norma de proyecto para pantallas futuras, no solo para las cuatro revisadas.

**Por qué:** varios controles (privacidad, permitir bots, iconos solo-icono) no dejaban claro qué
hacían sin probarlos.

**Cómo se aplicó:** primitiva reutilizable en vez de anotar cada sitio uno a uno -
`useTooltipTrigger`/`TooltipBubble` (`apps/frontend/src/shared/components/presentational/tooltip.tsx`)
+ prop `tooltip?: string | null` en `GlossButton` que por defecto reutiliza `accessibilityLabel`, así
que todos los `GlossButton` existentes y futuros quedan cubiertos sin tocar sus call sites. Web
muestra por hover/focus, nativo por long-press (no hay hover táctil). `InfoHint` (icono '?') es un
componente aparte para explicar conceptos, no el tooltip de un botón normal. Detalle completo en
`ObsidianVault/norma-tooltips-y-ayuda-contextual.md` del repo — leer esa nota antes de añadir un
botón nuevo que no sea `GlossButton`/`ChannelTile` (los únicos dos que ya llevan tooltip integrado).

**Update 2026-08-13 (mismo día, tercer pase):** el diseño original (bubble como `position:absolute`
anidado) estaba roto de raíz - RN solo compara z-index entre hermanos directos, así que quedaba
recortado por `overflow:hidden` de ancestros o pintado por debajo de hermanos no relacionados.
Reescrito a `Modal transparent` anclado por `.measure()` (mismo patrón que `ConfirmSheet`/
`GlassSelect`), ver `ObsidianVault/norma-tooltips-y-ayuda-contextual.md` para el detalle técnico. Si
se toca `tooltip.tsx` de nuevo: NO volver a un `position:absolute` anidado, por muy tentador que
parezca para un caso "simple".

Ver [[project_i18n_status]] y [[game_lobby_feature_closed_2026-08-13]] (mismo pase de trabajo).
