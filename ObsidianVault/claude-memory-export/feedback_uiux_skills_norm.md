---
name: feedback_uiux_skills_norm
description: User requires frontend-design + ui-ux-pro-max skills for any UI/UX design work on this project
metadata: 
  node_type: memory
  type: feedback
  originSessionId: edbd1834-0c9f-48a8-b841-a0eced841c8b
  modified: 2026-08-04T13:11:46.894Z
---

For any UI/UX design work on this project (new screens, components, layout/typography/color/animation
changes), load the `frontend-design` and `ui-ux-pro-max` skills before writing JSX/styles.

**Why:** the project already has a deliberate visual system (Wii Party channel UI × Windows Aero
glass × iOS-6 gloss, see [[a11y_web_leak]] and the in-repo ObsidianVault note
`frontend-responsive-frutiger-aero.md`). Skipping the skills risks drifting toward generic
AI-design defaults and reinventing glass/gloss recipes instead of composing the existing primitives.

**How to apply:** `frontend-design` for aesthetic direction (mostly a restraint/critique checklist
here since the visual system is already fixed); `ui-ux-pro-max` for targeted `--domain`/`--stack`
lookups (ux, color, react-native) rather than a full `--design-system` regen if a MASTER.md already
exists. Icon library stays `lucide-react-native` (matches existing shell) even though the skill
defaults to suggesting Phosphor.

Recorded permanently in the project's own ObsidianVault as `norma-diseno-ui-ux.md`, linked from
`overview.md` and `frontend-responsive-frutiger-aero.md`. See [[project_context]].
