---
title: "Auditoría de seguridad + Expo SDK 58 (2026-10-24)"
tags:
  - project
  - jojo-onepiece-simulator
  - seguridad
  - dependencias
---

# Auditoría de seguridad + Expo SDK 58 (2026-10-24)

## Qué se hizo
- **Go** `1.26.5` → `1.26.10` (20 vulns de stdlib alcanzables según `govulncheck`), deps directas al día (chi 5.5, pgx 5.11, govips 2.20, goose 3.29...). Imágenes a Alpine 3.23, pnpm pineado a `11.18.0` en los Dockerfiles.
- **chi 5.5**: `ClientIPFromXFF()` sin argumentos ahora hace panic; el equivalente es `ClientIPFromXFFTrustedProxies(1)` (un salto de confianza = NPM). Ver [[csp-y-rate-limit-por-ip-2026-09-05]].
- **r2-worker-proxy**: wrangler/vitest al día + `overrides` (undici, zod) → 0 vulns.
- **Frontend**: Expo SDK 57 → **58** (RN 0.88, React 19.3, Reanimated 4.7, gesture-handler 3), `overrides` en `pnpm-workspace.yaml` para transitivas de tooling.

## Gotchas
- `expo install --fix` falla en Windows (ruta larga): usar `expo install --check` y editar `package.json`. Tras saltar de SDK, `rm -rf node_modules` + reinstalar: quedaban copias viejas de React/RN y Jest fallaba con "Could not locate module react-native/react-private-interface".
- RN 0.88 solo trae tipos estrictos: `View`/`ScrollView` ya no sirven como tipo de ref → `ViewRef`/`ScrollViewRef` en `shared/lib/view-ref.ts`. `onKeyDown` de RN-web recibe un evento DOM, no el tipo de RN (`RovingItemProps.onKeyDown` va con `any`).
- **jest-expo 58 depende de Jest 30** (antes se pinaba 29). Jest 30/jsdom 26 prohíbe redefinir `window.location`: los tests mockean `shared/lib/browser-location.ts`.
- `app.json`: `newArchEnabled` ya no existe en SDK 58. `expo-audio` exige `expo-asset` como dependencia directa. `expo-doctor` 20/20.
- pnpm 11 añadió solo `minimumReleaseAgeExclude` (versiones recién publicadas de Expo) y `'@parcel/watcher': false` en `allowBuilds`.
- `image-size` override a 2.x rompía el bundling web; ya no hace falta (SDK 58 lo resuelve).

## Pendiente
- `pnpm audit` deja 2 high en tooling de build: `node-forge` (sin versión parcheada) y `zod` 3.x (dentro de `@expo/cli`). Repetir al salir un SDK/CLI nuevo.
- Smoke test manual en navegador del SDK 58 (React 19.3 / RN-web 0.21.4): build y tests pasan, no se ha recorrido la app.
