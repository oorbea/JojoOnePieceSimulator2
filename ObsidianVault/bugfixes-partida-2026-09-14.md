---
title: "Bugfixes de partida (2026-09-14): dock móvil, skip de summary, tooltip en modal, revancha fantasma"
tags:
  - project
  - jojo-onepiece-simulator
  - bugfix
  - game
  - gotcha
---

# Bugfixes de partida — 2026-09-14

Cuatro bugs encontrados jugando en local, cada uno con causa raíz propia. Fix mínimo
en los cuatro, sin rediseño. Verificado con `/verify` en Docker (backend build/vet/
test, frontend typecheck/lint/jest), todo en verde.

## 1. Botones de abajo tapados por el dock móvil

`VoteBar` (`vote-bar.tsx`) era el único elemento anclado abajo de toda la app que no
leía `useNavInsets()`. Usaba `position:'sticky', bottom:0` en web — un `sticky` se
resuelve contra el *scrollport*, no contra el `pb` que aplica `PageShell`, así que se
pegaba al borde del viewport, justo debajo del `ChannelBar` dock (`z:$nav`=500).
Fix: `bottom: navInsets.bottom` en vez de `0`. Ningún otro botón de fase (RoundResult,
LoadoutSummary, Reveal) tenía este problema — están en flujo normal, sí los limpia el
`pb`. Lección: cualquier elemento nuevo `position:sticky` anclado abajo en el juego
tiene que leer `navInsets`, no asumir que el `pb` del `PageShell` lo cubre.

## 2. El skip del resumen de poderes (y del sorteo) no completaba aunque todos pulsaran

**La causa raíz importante, reutilizable.** `Game.revealReady`/`summaryReady`
(el voto de "saltar" sincronizado) vivían sólo en el `*Game` en memoria — nunca se
añadieron a `game.Snapshot` ni al wire de Redis
(`infrastructure/gamestore/redis/wire.go`). `MemoryGameStore` guarda el puntero vivo,
así que ningún test lo detectó; con `REDIS_URL` puesto (el stack real de
docker-compose), cada `withGame` hace Get→Restore y el set de "ready" volvía vacío en
cada comando. Resultado: `ready` nunca llegaba a `total`, y sólo avanzaba el timer del
phase completo.

Es la **misma clase de bug que ya mordió una vez** con `RoundSnapshot.TiedVotes`
(ver nota en `wire.go:142-150`, y el walkthrough en vivo que lo encontró). Patrón:
**todo estado de dominio nuevo que vive fuera de los campos "de siempre" de `Game`
tiene que añadirse explícitamente a `Snapshot()`/`Restore()` Y al wire de Redis** — si
sólo se prueba con `MemoryGameStore` (la mayoría de tests unitarios), el bug queda
invisible porque no hay serialización real de por medio. `MemoryGameStore` NO es un
sustituto válido para probar persistencia; sólo el round-trip de
`gamestore/redis` (encode/decode) lo detecta.

Fix aditivo: `Snapshot.RevealReady`/`SummaryReady []ParticipantID`, poblados en orden
de `g.order` (no iteración de map), con el mismo patrón `omitempty` que ya usa
`PhaseEndsAt`/`TiedVotes` — un payload viejo decodifica a "nadie ha votado", que ya
era el comportamiento por defecto.

**Para la próxima vez que se añada un campo a `Game`:** si el campo debe sobrevivir un
restart o una partida que se sirve desde varias instancias, hace falta tocar 3 sitios:
`Game` (el campo), `Snapshot`/`Restore` (dominio), y `wireGame`/`toWire`/`fromWire`
(Redis). Un test en `redis/wire_test.go` que ejercite el campo con valor no-trivial
(no el caso "todos han votado", que colapsaría igual aunque se perdiera el conteo) es
la única red que lo pilla.

## 3. Tooltip detrás de la modal (battleIQ en el panel de admin)

RNW's `Modal` (usado por los `*-form-modal.tsx` del admin) fija `z-index:9999` sin
prop para cambiarlo. El portal del tooltip (`OverlayPortal` en `tooltip.tsx`) no
ponía ningún z-index — ambos son hijos directos de `<body>`, así que `auto` pierde
siempre contra `9999`. Fix: `zIndex:10000` en el div del portal, manteniendo
`pointerEvents:'none'` (bajarlo sin eso reintroduce el bug del "cuarto pase" descrito
en [[norma-tooltips-y-ayuda-contextual]] — un overlay transparente que se come el
hover/click de toda la pantalla). Arregla todos los tooltips dentro de cualquier
modal, no sólo el de battleIQ.

## 4. Tras la partida, "Volver a las salas" dejaba un fantasma en la revancha

`handleBackToLobbies` (en `lobby-room-container.tsx`) nunca mandaba `LEAVE` — sólo
reseteaba el socket y navegaba, pese al comentario que decía imitar `handleLeave`
(que sí lo manda). Al cerrar el socket sin `LEAVE`, el backend trata la salida como un
`Disconnect`: conserva el asiento y reasigna host si el que se fue era host. El
jugador restante, ya host, podía pedir revancha, y `Rematch` copiaba también al
"fantasma" (con `connected:true` porque `NewHumanParticipant` lo fija así siempre).
Fix: mandar `commands.leave()` antes de resetear el socket, igual que `handleLeave`.
Con el asiento realmente liberado, `reassignHost` promociona a cualquier humano
conectado que quede y la revancha sólo copia a quien sigue en la sala.

Fuera de alcance (anotado, no arreglado): `Game.Disconnect` no emite ningún evento —
una desconexión de un no-host deja rosters obsoletos en los demás clientes hasta el
siguiente STATE. Bug real pero distinto, no reportado por el owner esta vez.

## Relacionado

[[gameplay-domain-design]] · [[game-lobby-persistence]] · [[stage-redis-cache-2026-09-02]] ·
[[norma-tooltips-y-ayuda-contextual]] · [[norma-verificacion-docker]] · [[overview]]
