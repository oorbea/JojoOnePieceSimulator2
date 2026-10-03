---
title: "Tanda de playtest (2026-10-03): ajustes de sala, bots, skip por mayoría, +10s, stage en el sorteo"
tags:
  - project
  - jojo-onepiece-simulator
  - game
  - bugfix
  - gotcha
---

# Tanda de playtest — 2026-10-03

Nueve cosas encontradas en una partida real; todo commiteado directo en `develop`, un commit por ítem. Enfoque deliberado: lo más simple que funcione, reutilizando lo que ya existía.

## Qué cambió

- **Editar la sala**: el panel ya existía (`LobbyConfigPanel`, host-only, `UPDATE_CONFIG`) pero estaba colapsado debajo de los rosters con una etiqueta genérica. Ahora el desplegable va **encima de los rosters** y el host lo ve como "Editar ajustes de la sala". Bug colateral: `Game.Reconfigure` dejaba bots descartados en `Team.members` (los misma-modo reutilizan los `Team`); ahora hace `RemoveMember`.
- **Bots**: el backend ya estaba (`ADD_BOT`/`REMOVE_BOT`, solo Versus, solo host). Faltaba UI: botón "Añadir bot" por equipo (deshabilitado con pista si `allowBots` está off) y "Quitar bot" en la fila del bot. Gauntlet sigue sin bots por diseño.
- **Skip por mayoría estricta** (`ready*2 > total`, `strictMajority` en `game.go`) para sorteo y resumen. Antes exigía el 100% de humanos conectados. Aviso: `useSkipNotice` muestra un toast cuando la fase termina bastante antes de su deadline (solo con >1 humano conectado). La etiqueta del botón pasa a `listos/necesarios`. **No** se tocó la votación de ronda (decisión del dueño: solo sorteo + resumen).
- **+10 s del host en votación**: comando `EXTEND_VOTING` → `GameService.ExtendVoting` re-arma el timer con `armPhaseTimer(deadline+10s)` (reutiliza el re-armado existente, **sin estado nuevo que persistir**) → evento `VotingExtended` → frame `VOTING_EXTENDED{closesAt}`. Sin límite de usos. Gotcha: un STATE posterior **no** mueve un `votingClosesAt` ya fijado en el cliente, por eso el frame es la única vía.
- **Volver al sorteo en directo**: `useLoadoutReveal.rewatch()` incrementa un contador metido en el `runKey`; reutiliza `seekRevealTimeline` + `serverNow()` (misma matemática que una reconexión). No deshace el `REVEAL_READY` ya enviado. El botón sale en `VotingStatusBar` mientras el servidor sigue en ASSIGNING.
- **Ficha de un poder**: en `LoadoutModal`, botón "Ver ficha completa" bajo cada poder → `DetailModal` con `StandDetail`/`DevilFruitDetail` del catálogo. El hover card (`TooltipCard`) es `pointerEvents:none` por diseño, así que **no** puede alojar clicks; solo el modal.
- **`enums.hakiLevel.undefined`**: el slot sintético `hakiSet` (solo para el reveal, sin label ni valor) se colaba en la lista del modal y caía en el `default` de `enumNamespace`. Ahora se filtra.
- **Stats solapadas en móvil**: filas escalares con `flex={1}` en label y valor; `minW:280` de los `PowerBlock` solo desde `$md`.
- **Stage en el sorteo (Versus)**: el stage antes solo se elegía en `OpenVoting`. Ahora `Game.PrepareUpcomingStage` lo elige al asignar loadouts (`upcomingStage` en `Game`) y `OpenVoting` lo consume. Va en `Snapshot`/`Restore`/`redis/wire.go` y en el DTO (`GameSnapshotResponse.upcomingStage`). UI: `StageAnnouncement` ("Escenario del combate: solo el lugar donde se desarrollará, no da ventaja") grande en la intro del sorteo, compacto durante los jugadores y en el resumen; el `StageBanner` de votación lleva la misma etiqueta en Versus. Gauntlet intacto.

## El bug "en Versus ronda 2 sale un stage pero no hay sorteo"

**No se pudo reproducir** el sorteo ausente con un humano + bot (la ronda 2 sí arrancaba, tras saltar la ronda 1). Lo que sí era real: durante ASSIGNING de la ronda ≥2, `match-screen` pintaba `currentRound().stage`, que es el de la ronda **anterior** (la `Round` nueva solo se crea en `OpenVoting`). Eso encaja con "aparece un stage" antes/sin sorteo. Arreglado: en ASSIGNING nunca se pinta el `StageBanner`; Versus muestra el stage anunciado. Si reaparece con 2+ humanos, sospechosos pendientes: `shouldReveal`/`assignmentSeq` en la segunda `LOADOUTS_ASSIGNED` y campos `live.*` no reseteados en `ROUND_RESOLVED`.

## Gotchas aprendidos

- **typegen necesita registrar cada payload** en `cmd/typegen/registry.go`: `FramePayloads` solo no basta. Si falta, `ws.ts` referencia `xxxPayloadSchema` sin definirlo y la app entera rompe en runtime con `... is not defined` (`pnpm typecheck` lo habría pillado; lo pilló el walkthrough en vivo antes). `frame_table_test.go` mantiene su propia lista de constantes: añadir ahí también.
- **El barrel `@/features/stands` arrastra los containers** (cliente API, validación de env): importarlo desde un componente presentacional rompe sus tests con "Invalid environment configuration". Para `StandDetail`/`DevilFruitDetail` se importa por ruta directa (excepción documentada en el propio import).
- **`burnt` es módulo nativo sin fallback JS**: cualquier test cuyo grafo llegue a `shared/lib/toast.ts` fallaba con `Cannot find native module 'Burnt'`; ahora hay mock global en `jest.setup.ts`.
- `GlowText` no acepta `flexShrink`/`textAlign="right"` (solo `flex` y `align` left/center): alinear a la derecha con un contenedor `YStack items="flex-end"`.
- Docker en Windows: `docker run -v "$(pwd):/repo"` desde Git Bash monta vacío; usar PowerShell con ruta explícita. Frontend local sirve en `:8081` (nginx), no en `:3000`.
