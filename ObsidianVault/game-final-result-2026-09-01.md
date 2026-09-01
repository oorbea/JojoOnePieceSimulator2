# Pantalla de resultado final + rematch - 2026-09-01

Cierra el último bullet abierto de §6 de [[game-lobby-todo]] y los tres "deliberadamente fuera de
alcance" que [[game-vote-buttons-2026-08-26]] dejó anotados (pantalla de resultado, drift de
`closesAt`, reinicio a media votación). Hasta ahora `GAME_FINISHED` solo lanzaba un toast y echaba a
todo el mundo a `/play`: una partida entera terminaba sin que nadie llegara a ver quién ganó.

## Qué se construyó, por capas

**Backend (4 tandas pequeñas, en este orden):**

1. **`closesAt` autoritativo.** `VOTING_OPENED`/`TIEBREAK_OPENED` sintetizaban su `closesAt` como
   `time.Now()+window` en el transporte, con lo que derivaba tanto como tardara la entrega por el
   hub. Ahora se sella desde el deadline real del servicio al construir el frame, exactamente por la
   vía que [[game-vote-buttons-2026-08-26]] ya había anticipado (`e.svc`/`gameID` estaban en scope
   en `forwardEvents`).
2. **Deadline de fase persistido.** `s.timers`/`s.votingEnds`/`revealEnds` vivían solo en memoria del
   proceso, así que un reinicio a media votación dejaba la sala encallada en `VOTING` sin deadline
   para el cliente y sin ningún timer que fuera a cerrar la ventana nunca. El deadline de la fase
   temporizada vive ahora en el agregado y **se re-arma al cargar**, con lo que un reinicio reanuda
   la ventana en curso en vez de perderla.
3. **Un juego terminado sigue legible.** `finalizeLocked` borraba la partida del store en la misma
   llamada que emitía `GAME_FINISHED`/`GAME_ABORTED`. Ahora la guarda bajo un **TTL corto propio**.
   Esto es lo que hace posible la pantalla de resultado, y de paso permite que esos dos frames
   vuelvan a mandar `STATE` como todos los demás - ver [[game-realtime-transport]], cuya lista de
   excepciones se ha quedado solo con `VOTE_CAST`.
4. **Caso de uso real de rematch.** Comando `REMATCH` + frame `REMATCH_READY`, que lleva a cada
   cliente aún presente en la partida terminada el id de la sala nueva, para que todos sigan al host
   juntos.

**Frontend:**

- `lib/game-result.ts`: `matchRecap()` puro y unit-testeado deriva todo lo que pinta la pantalla a
  partir de un snapshot terminal. Reutiliza `voteTally` tal cual para cada ronda, de modo que
  etiquetas, tonos y orden de opciones coinciden con la barra de voto en vivo.
  - **Supuesto explícito que conviene recordar**: eso solo es correcto porque las opciones de una
    partida son estables entre rondas (`SURVIVE`/`FALL` en Gauntlet, los dos mismos team ids en
    Versus), así que la lista de opciones de la última ronda describe también a las anteriores. Si
    algún modo llega a variar sus opciones por ronda, esta es la suposición que se rompe.
  - `won` es `bool | null`, no `bool`. Solo Versus tiene ganador por participante; Gauntlet es una
    run cooperativa con un veredicto colectivo, así que decirle a alguien "has perdido" ahí sería
    sencillamente falso. `null` significa "no aplica" y la pantalla pinta un banner colectivo.
  - Prefiere la lista de asientos de fin de partida que manda el servidor (el roster tal como estaba
    al terminar, que es de lo que un resultado debería hablar) y cae a la lista viva de
    participantes para partidas terminadas por un backend anterior a ese campo.
- `MatchResultScreen`: veredicto, recap ronda a ronda y outcome por asiento. Un juego terminado o
  abortado **se queda en su ruta**; solo `KICKED` conserva su toast-y-redirect, por ser el único
  caso terminal en el que no queda nada que enseñarte.
- `vote-tally-row.tsx`: la fila de tally se extrae del panel de ronda viva y se comparte, para que
  el recap y el resultado en vivo se rendericen idénticos en vez de divergir.
- Un `REMATCH` rechazado se muestra inline, correlacionado por `requestId` **dentro del mismo bloque
  de ajuste-en-render** que ya usaba el error de config-edit (ver abajo).

## Trampas encontradas (las que costaron tiempo)

**`react-hooks/set-state-in-effect` otra vez.** El primer intento metía el error de rematch en su
propio `useEffect` con un `setState` síncrono dentro - exactamente lo que el commit `acefa7a` ya
había arreglado para el error de config. Se ha plegado al bloque de ajuste-en-render existente,
reutilizando el mismo marcador `configErrorHandledFor`, de modo que un frame de error solo puede
atribuirse una vez. **Norma implícita del repo: en este contenedor, correlacionar un `requestId` con
estado local se hace en ese bloque, no en un efecto.**

**El carril de tests web (`.web.test.tsx`) no funcionaba en absoluto.** El split de
`testMatch`/`testPathIgnorePatterns` en `jest.config.js` estaba puesto, pero la suite **ni siquiera
cargaba**: `Cannot read properties of undefined (reading 'OS')`. Causa: el proyecto `logic` define
`moduleNameMapper`, y eso **reemplaza** el del preset `jest-expo/web`, tirando por la borda el mapeo
`react-native` -> `react-native-web`. Cualquier import con branching de plataforma
(`shared/lib/web-blur.ts`) moría al cargar. Estaba latente porque ningún test previo del proyecto
`logic` había importado uno.
- Y una vez cargaba, todos los asserts con estado fallaban: el helper usaba `act()` síncrono, que
  esta versión de `@testing-library/react-native` ni tolera ni vacía, así que el `setState` del hook
  nunca commiteaba. Los únicos tests "en verde" eran aquellos cuya expectativa coincidía con el
  estado inicial - un falso verde silencioso. Con `await act(async () => ...)` (el idiom que
  `use-debounced-value.test.tsx` ya sentaba), 14/14.
- El probe renderiza `null` a propósito: un `<Text>` de react-native aquí resuelve a un elemento DOM
  que el renderer nativo de RNTL rechaza.

**`fireEvent.press` es asíncrono en RNTL 14.** Un test que hacía dos `press` seguidos sin `await`
dejaba scopes de `act()` solapados que corrompían el act-nesting **para todo el resto del fichero**:
cada `render()` posterior devolvía un árbol vacío y los tests siguientes fallaban con "Unable to
find an element". Los tests pasaban en aislamiento y fallaban juntos, que es la firma de este
problema. Cada press va ahora en su propio `await act(async () => ...)`.
- **Mina latente detectada, no arreglada aquí**: `stand-card`, `stage-card` y `devil-fruit-card`
  tienen el mismo patrón de dos press sin await, y solo sobreviven porque ese test es el último de
  su fichero. Añadir cualquier test detrás los romperá. Anotado para quien toque esos ficheros.
- También sin arreglar: el `hideTimer` de 1800 ms de `useHoverTrigger` no se limpia al desmontar, que
  es el origen real de la flakiness ya documentada de `tooltip.test.tsx`.

**Contadores de fixtures que se daban la vuelta.** `mustStage`/`mustStand`/`mustDevilFruit` en
`game_service_test.go` construían ids escribiendo un contador **`byte`** en `id[15]`. Al pasar de 255
fixtures en el paquete, el contador volvía a 0, el id quedaba todo ceros y los constructores lo
rechazaban con `id is required`. Latente hasta esta tanda: las suites nuevas (finished-TTL,
phase-deadline, rematch) empujaron el paquete por encima de ese techo, y se manifestó como
`TestVotingEndsAt` - **un test preexistente y ajeno, que solo tiene la mala suerte de correr tarde**.
Contadores ahora `uint32` repartidos en los últimos cuatro bytes.
- Lección de método: el fallo apareció **solo después del rebase sobre `develop`**, lo que invitaba a
  culpar a la tanda hermana §4. Lo que lo zanjó fue correr el test contra `origin/develop` puro (que
  pasaba) para demostrar que la regresión era propia.

## Hueco de teclado cerrado (B6)

`blocked` cubre ya los dos overlays. `LoadoutModal` **sigue viviendo dentro de `MatchRoster`** - es
el componente que ya tiene `mangas` en mano para el contenido del modal - y avisa hacia arriba con
un `onModalOpenChange?: (open: boolean) => void` opcional que `MatchScreen` reenvía y el contenedor
espeja en `rosterModalOpen`. Todo abrir/cerrar pasa por un único `openModal()` para que la
notificación no pueda desincronizarse del estado. Patrón a repetir si aparece un tercer overlay: no
subir el estado, añadir otra notificación. Ver [[norma-teclado]].

## Verificación

- Backend (Docker, con Postgres+Redis arriba porque la rama toca `gamestore/redis/wire.go`):
  `go build` y `go vet` limpios, `go test ./...` **14 paquetes ok, 0 fallos**.
- Frontend (carril node en Docker): `typecheck` exit 0, `lint` exit 0 (0 errores),
  `test:ci --maxWorkers=3` **51 suites / 551 tests, 0 fallos**.
- Con el paralelismo por defecto reaparecieron 6 suites en rojo por timeouts de contención (739s y
  776s de runtime), todas en verde al reejecutarlas aisladas en 9-55s, y un conjunto distinto en cada
  corrida - la flakiness ya documentada en [[norma-verificacion-docker]], no una regresión. Bajar a
  `--maxWorkers=3` da una corrida completa estable.

## Pendiente

- El walkthrough en vivo a dos navegadores (empate real, reconexión a media votación, pase
  keyboard-only) sigue abierto desde 2026-08-26; esta tanda no lo desbloquea ni lo cierra.
- Per-power visual FX sigue siendo **plan documentado, no construido** - [[gameplay-power-fx]].
- Las dos minas de tests descritas arriba (`fireEvent.press` sin await en las tres `*-card`, y el
  timer sin limpiar de `useHoverTrigger`).

Related: [[game-lobby-todo]], [[game-vote-buttons-2026-08-26]], [[game-round-result-2026-08-28]],
[[game-realtime-transport]], [[game-match-assignment-frontend]], [[norma-teclado]],
[[norma-verificacion-docker]], [[frontend-stack]], [[gameplay-power-fx]].
