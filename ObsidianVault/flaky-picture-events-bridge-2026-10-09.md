---
title: "Flaky test resuelto: picture-events-bridge (2026-10-09)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - testing
  - gotcha
---

# Flaky `picture-events-bridge`: causa raíz y arreglo

**Síntoma (CI, PR #38 y PR #96):** `picture-events-bridge.web.test.tsx` > "a network error minting a ticket backs off and re-mints" fallaba a veces con `mockMintEventsTicket` llamado **2 veces en vez de 1**, antes de avanzar ningún temporizador. Pasaba al relanzar el job y siempre en local.

## Causa raíz

No era el código de producción: era higiene del test.

1. `render()` de `@testing-library/react-native` 14.x es **asíncrono** y solo mete el `unmount` en la cola de limpieza (`addToCleanupQueue`) **después** de que termine su `act` interno.
2. El test llamaba a `render(...)` **sin `await`** y esperaba con un `flush()` de dos microtareas.
3. El `afterEach` automático de RNTL (`flushMicroTasks()` + `cleanup()`) recorre la cola **en ese momento**. Si el `act` del render aún no había terminado (runner de CI cargado), el componente nunca se desmonta y **se filtra al test siguiente**, todavía suscrito al store de sesión.
4. Al empezar el test siguiente, `useSessionStore.setState({ session: adminSession() })` reactiva el efecto del componente huérfano (`mint` nº 1) y el recién montado hace el suyo (`mint` nº 2).

## Cómo se confirmó (no se adivinó)

- Repetir 40 veces con 4 bucles de CPU **no** reproducía el fallo.
- Se instrumentó una copia temporal del test con un `Probe` que cuenta cuántos puentes hay montados al empezar cada test. Con 16 bucles de CPU y 8 procesos de jest en paralelo (~96 ejecuciones): `PROBE-LEAK live=1` en la mayoría de los workers y **3 fallos idénticos al de CI** (`Expected 1, Received 2`).
- Tras el arreglo, 64 ejecuciones con la misma carga: **0 fugas, 0 fallos**.

## Arreglo

- `await render(...)` en los 7 sitios de `picture-events-bridge.web.test.tsx` y en `use-google-auth.web.test.ts` (el otro fichero con el mismo patrón).
- Guarda `src/test/__tests__/render-awaited.test.ts`: falla si algún test tiene una sentencia `render(` / `renderWithProviders(` sin `await` (mismo estilo que la guarda de `copy.test.ts`). Estaba en rojo para esos dos ficheros antes del arreglo.

## Lección

En este repo `render` hay que esperarlo **siempre** (ya lo decía el comentario de `src/test/render.tsx`, pero solo para evitar que `screen` no esté registrado; aquí además hay una fuga entre tests). Un flake que "pasa al relanzar" y no se reproduce repitiendo suele necesitar **instrumentación bajo carga**, no más repeticiones.

Related: [[frontend-stack]], [[norma-verificacion-docker]]
