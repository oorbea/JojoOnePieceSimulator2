---
title: "Cleanup: socket.io-client removed (2026-09-02)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - cleanup
---

# socket.io-client removed (2026-09-02)

`socket.io-client` sat in `apps/frontend/package.json` from the very first scaffold
(2026-07-28) and was **never imported once**. The realtime transport the game feature
actually uses is the browser's own `WebSocket`. This tanda deleted the dead dep and
fixed the doc drift it left behind.

## What was verified before deleting

`git grep -niE "socket\.io|socketio"` over the whole repo. Outside the lockfile,
`package.json`, the README and the vault there was **zero** matches: no `import`, no
`require`, no `io(`, and no entry in metro / jest / babel / tsconfig / eslint config.
Nothing referenced it, so removal is behaviour-neutral by construction.

## What was removed

- `apps/frontend/package.json`: `"socket.io-client": "^4.8.3"`.
- `apps/frontend/pnpm-lock.yaml`: regenerated via `pnpm install --lockfile-only` (never
  hand-edited — see [[frontend-stack]]'s CLI-only rule). Dropped with it, as transitive
  deps that served nothing else: `socket.io-parser`, `@socket.io/component-emitter`,
  `engine.io-client`, and `engine.io-client`'s `ws` + the `bufferutil` /
  `utf-8-validate` optional-peer warnings. Net diff: −49 lines, and the only added line
  is an unrelated registry `deprecated:` note on `eslint@9.39.5`.

## The real client, for the record

`src/features/game/stores/game-socket.store.ts` — browser-native `WebSocket`, refcounted
connection sharing (many components, one socket), exponential backoff on drop, and a
`RESYNC` on reconnect so a resumed socket re-reads the snapshot instead of trusting a
stale event stream. URL is built by `src/features/game/lib/socket-url.ts` from
`EXPO_PUBLIC_SOCKET_URL` (`ws://localhost:8080/api/v1` locally, also set in
docker-compose and CI), token as a query param because neither the browser's nor RN's
`WebSocket` can set headers — same reasoning as the SSE bridge's `EventSource` URL.
Backend endpoint: `/api/v1/games/{id}/ws`, `coder/websocket`, see
[[game-realtime-transport]].

## Doc drift corrected in the same tanda

Four places had been asserting, long after it stopped being true, that there was no
frontend WS client and/or no backend endpoint:

| Where | What it claimed |
|---|---|
| `apps/frontend/README.md` | "socket.io-client (installed, unwired — the backend has no websocket endpoint yet)"; and `EXPO_PUBLIC_SOCKET_URL` "reserved for when the backend gets a websocket" (it is set and consumed) |
| [[overview]] "Flagged / not done" | dep still installed but unwired, eventual client should target the native protocol |
| [[frontend-stack]] stack paragraph + decisions table | "Install only, no wiring" |
| [[gameplay-application-layer]] "Still not built" | "Frontend: no WebSocket client yet" |
| [[game-realtime-transport]] status + library section | "Done, backend-only. No frontend WebSocket client" |

Lesson worth keeping: a "not done yet" bullet is a claim with an expiry date. When the
tanda that closes it lands, closing the bullet is part of that tanda, not a later
cleanup — otherwise the vault confidently lies for three weeks, as it did here.

## Deliberately NOT touched

The owner explicitly scoped this to `socket.io-client` only, and declined a wider
dependency audit. These also have no direct `import` but are legitimate and stay:

- `expo-crypto` — peer/runtime dep pulled by `expo-auth-session`.
- `react-native-svg` — peer dep of the Tamagui icon set.
- `@expo/metro-runtime` — required by the Expo web bundler entry, not imported by hand.

Related: [[frontend-stack]], [[game-realtime-transport]], [[gameplay-application-layer]],
[[overview]].
