---
title: "Feature: manual de normas y convenios (2026-10-09)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - backend
  - feature
  - gameplay
---

# Manual de normas y convenios

Página `/manual` (item "Manual" en la barra, entre Jugar y Perfil), en en-GB/es-ES/ca-ES. Explica cómo se juega, los modos, las habilidades, las probabilidades del sorteo, los efectos de los poderes, el Battle IQ, los **convenios de combate** que usan los votantes y las normas de etiqueta. Requisito del dueño: **todo lo que dice tiene que ser verdad en el código**, por eso no se escribe a mano.

## Cómo se garantiza que es verdad

- `game.BuildManualRules()` (`game/manual_rules.go`) calcula los datos desde las mismas tablas que usa el juego y `cmd/typegen` los emite a `apps/frontend/src/shared/contracts/rules.ts` (`MANUAL_RULES`). Si una regla cambia y no se regenera, falla el job `contracts` de CI ([[contratos-tipos-generados]]).
- **Matrices de evolución**: se calculan ejecutando el resolver real (`resolvePowerEffects`) sobre cada combinación etapa × nivel con poderes sintéticos. El manual no reimplementa la lógica.
- **Mínimos de estadística**: `statFloorRules` describe cada disparador con un `Subject` (mismo tipo que los convenios), no con closures, así el resolver y el manual leen un único valor.
- **Probabilidades**: salen de `DefaultAssignmentWeights()` normalizadas a %. Son *antes de efectos*; el manual lo dice.
- **Límites de partida** (tamaños de equipo, 3 rondas de Versus, ventana de voto 30 s [5-180], +10 s del host): `ManualLimits`, desde las constantes de `game/config.go`. `votingExtension` se movió de la capa de aplicación a `game.VotingExtensionSeconds` para que el dominio pueda exportarlo.
- **Textos**: a mano, pero `features/manual/lib/__tests__/rules-copy.test.ts` exige copy en los 3 idiomas para **cada** convenio, categoría, grupo, resultado y nivel que aparezca en `rules.ts`, y que los `{{parámetros}}` coincidan entre idiomas. Una regla nueva en Go sin traducir rompe CI.
- La tabla de bandas de Battle IQ del frontend (`shared/lib/battle-iq.ts`) ya no está copiada a mano: sale de `rules.ts`.

## Convenios de combate (tabla en código, `game/combat_conventions.go`)

Dato puro, hoy solo lo consume el manual; pensado para que lo usen el desempate por LLM y el bot (→ ver [[TODO]]). Acordados con el dueño:

- **Ver Stands**: usuarios de Stand, de cualquier Spin o de cualquier Haki de Observación.
- **Tocar a un Logia**: Haki de Armadura, Stand, Spin >= GOLDEN. Hamon solo si tiene sentido que la energía solar llegue (criterio de los votantes).
- **Dañar a un Stand**: otro Stand, Armadura, Conquistador, Logia, cualquier Spin.
- **Paro de tiempo**: The World, Star Platinum: The World, The World (Steel Ball Run) y King Crimson (tratado como paro). Se mueven en él otro parador, Spin INFINITE y Conquistador YONKO_PLUS.
- **Made in Heaven** (acelera, no para): un parador lo frena unos segundos (más cortos de lo normal), Spin INFINITE le sigue el ritmo, Conquistador YONKO_PLUS frena su aceleración, Observación YONKO_PLUS predice dónde está.
- **Conquistador**: diferencia >= 2 niveles noquea (NONE cuenta como nivel); 1 nivel solo dificulta moverse; Spin INFINITE resiste; Hamon PERFECT lo suaviza un escalón.
- **Weather Report** debilita a los usuarios de fruta.
- **Yami Yami no mi**: tocando al usuario o a su Stand anula frutas y Stands (no Spin/Hamon/Haki; puede tocar Logias). No anula los absolutos, Go Beyond, los paradores de tiempo ni Made in Heaven.
- **Spin INFINITE** vence cualquier defensa salvo GER, Wonder of U y otro Spin INFINITE. **Spin INFINITE vs Spin INFINITE no tiene ganador evidente**: lo deciden los votantes.
- **Stands absolutos** (solo los supera una lista): GER <- Soft & Wet: Go Beyond; Wonder of U <- Go Beyond, GER; D4C: Love Train <- Spin INFINITE, GER, Go Beyond, Wonder of U; Tusk: Act 4 <- GER, Wonder of U.

Test de coherencia (`combat_conventions_test.go`): un absoluto está en la lista de excepciones de Spin INFINITE exactamente cuando su propia lista de "quién lo supera" no incluye Spin INFINITE. Ese test destapó una contradicción entre dos reglas del dueño (Tusk A4 "vencido por otro Spin INFINITE" vs "dos Spin INFINITE se anulan"); se resolvió con el duelo a criterio de los votantes y quitando a Spin INFINITE de la lista de Tusk A4.

Los nombres de los convenios se suman a `PowerEffectRuleNames()`, así que `db/migrations/power_effect_names_test.go` los cruza con los seeds (Wonder of U, Yami Yami, Weather Report... están seedeados; solo Ball Breaker y Soft & Wet: Go Beyond son nombres solo-prod).

## Frontend (`src/features/manual/`)

- Página larga con índice (columna lateral sticky en web ancho, chips envueltos en estrecho) y scroll a cada sección; `?section=<id>` deep-link y la URL se actualiza al elegir. `PageShell` ganó `scrollRef`.
- La posición de cada sección se suma de tres `onLayout` (fila, columna, sección); el scroll del deep link espera 120 ms a que reporten.
- Solo las secciones con datos lo reciben por props: pantalla presentacional pura, el contenedor importa `MANUAL_RULES`.
- Se numera solo la secuencia real (dibujar, votar, resolver); lo demás no.
- `shared/lib/loadout-slots.ts`: slot (wire) -> clave de rasgo/nivel i18n. Los tres mapas duplicados de `reveal-stage.tsx`, `loadout-modal.tsx` y `trait-chips.tsx` NO se refactorizaron (usan claves camelCase propias).

## Gotchas

- El sticky del índice tiene que sumar `useNavInsets().top`: con `top: 16` la barra flotante tapaba el primer botón. Solo se vio mirándolo en vivo.
- En el entorno local el frontend publica en **:8081** (por el `.env`), no en 3000 como dice el skill `local-up`; la imagen es de producción (nginx), no hay hot reload: cada cambio de UI exige `up -d --build frontend`, y el service worker tarda un par de segundos en refrescar.
- `resize_window` de la herramienta de Chrome no cambia el viewport; para ver 390 px se embebe `/manual` en un iframe del mismo origen (el compose dev lo permite).
- La guarda `copy.test.ts` prohíbe guiones largos/medios en textos de usuario: el rango de bandas usa `0-69`, no `0–69`.
- Contenido generado con Tamagui: no pasar `aria-level` en los headings (solo `role="heading"`), por la fuga de props a DOM de [[a11y-web-leak]].

## Limitaciones y decisiones

- **Las cadenas de evolución del manual están declaradas en código** (`manualStandChains`/`manualFruitChains`: Tusk A1-A4, Soft & Wet -> Go Beyond, Ball Breaker, Gomu -> Nika) porque el `evolves_from` de los Stands vive en la base de datos, no en código. Los niveles de cada etapa sí vienen de las tablas reales. Un test (`TestManualRules_CoverEveryTier`) falla si una etapa con nivel falta en las cadenas. → ver [[TODO]].
- **La regla de "mayoría estricta" para saltar sorteo/resumen** y el cara o cruz del segundo empate están solo en el texto, no generados.
- **Battle IQ**: no se inventaron ejemplos canónicos por banda; solo se explican las bandas. El comentario de `loadout_builder.go` dice que un 255 es ~25x más raro que un 130 en la banda alta, pero con vida media de 30 puntos son ~18x (`0.5^(125/30)`); el manual evita citar la cifra.
- **No hay enlaces "?" contextuales** desde el lobby/votación/sorteo: navegar a `/manual` desde una partida activa dispara el aviso de abandonar la sala (exit-guard). → ver [[TODO]].
- Los nombres de Stands/frutas se muestran tal cual están en el catálogo (sin traducir).

## Verificación (2026-10-09)

Backend y frontend completos en Docker (`/verify`); en vivo con `local-up` + Chrome en es-ES: índice con scroll y URL, deep link en carga en frío (`?section=conventions`), sticky, sin claves crudas, 8 secciones y 20 convenios en el DOM, y en 390 px (iframe) el índice pasa a chips y el dock inferior marca Manual. No se probó el cambio a en-GB/ca-ES en vivo (lo cubren los tests de copy).

Related: [[gameplay-power-effects]], [[gameplay-game-modes]], [[gameplay-versus-inventory-characters]], [[norma-diseno-ui-ux]], [[norma-tooltips-y-ayuda-contextual]], [[frontend-responsive-frutiger-aero]], [[contratos-tipos-generados]]
