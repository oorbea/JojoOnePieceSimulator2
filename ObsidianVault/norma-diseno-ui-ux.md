---
title: Norma — diseño de UI/UX siempre vía skills
tags:
  - norma
  - jojo-onepiece-simulator
  - frontend
---

# Norma: cualquier diseño de UI/UX pasa por `frontend-design` + `ui-ux-pro-max`

Para **cualquier** trabajo que toque el aspecto visual o de interacción del frontend (pantallas
nuevas, componentes nuevos, revisiones de layout, tipografía, color, animación, formularios), cargar
primero las skills `frontend-design` y `ui-ux-pro-max` antes de escribir JSX/estilos.

## Por qué

El proyecto ya tiene un sistema visual propio y deliberado (Wii Party channel UI × Windows Aero
glass × iOS-6 gloss — ver [[frontend-responsive-frutiger-aero]]). Sin esta norma, cada pantalla
nueva corre el riesgo de derivar hacia el "look genérico de IA" (crema+serif, negro+acento único,
broadsheet) que las skills existen precisamente para evitar, y de perder la disciplina de
composición sobre las primitivas ya existentes (`GlassPanel`, `WiiCard`, `GlossButton`, `ChannelTile`,
`PageShell`...) en vez de reinventar recetas de glass/gloss pantalla a pantalla.

## Cómo

1. `frontend-design` primero: fija dirección estética específica al encargo (no aplica si el
   sistema visual del proyecto ya está fijado y sólo hay que componer con las primitivas — en ese
   caso su valor es el checklist de restraint/autocrítica, no un rediseño desde cero).
2. `ui-ux-pro-max`:
   - Si **ya existe** `design-system/<slug>/MASTER.md` en el repo, leerlo y NO regenerar con
     `--persist` salvo `--force` explícito — evita descartar decisiones previas.
   - Si no existe, usarlo puntualmente con `--domain ux`, `--domain color`, `--stack react-native`,
     etc., para reglas concretas (formularios, accesibilidad táctil, iconografía) en vez de
     `--design-system` completo, que generaría un sistema paralelo al ya establecido.
3. La sugerencia de iconos de `ui-ux-pro-max` es Phosphor por defecto — **en este proyecto se
   mantiene `lucide-react-native`**, que es lo que ya usa el shell (`app-shell-container.tsx`).
   Anteponer siempre la consistencia del repo a la sugerencia genérica de la skill.

Ver también [[zettelkasten-workflow]] (norma hermana: registrar decisiones de diseño en el vault).
