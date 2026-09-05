---
title: CSP header + real per-IP rate limiting (2026-09-05)
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - deploy
  - security
  - decision
  - resolved
aliases:
  - csp-y-rate-limit-por-ip
---

# CSP header + real per-IP rate limiting (2026-09-05)

> [!success] Status: **DONE, shipped, committed on `develop`**
> Closes two items carried as future tandas: [[session-token-storage-2026-09-05]]'s *No CSP
> header* gap, and the rightmost-XFF switch [[stream-connection-tickets-2026-09-03]] flagged as
> *its own future tanda, not folded into this one*. Two commits, no unrelated changes mixed in:
> `fix(backend): key rate limits on the real client IP behind the proxy` →
> `feat(deploy): send security headers incl. CSP from the frontend nginx`.

Owner wanted the pragmatic fix, not the maximal one: enforce CSP from day one (no Report-Only
phase, no nonces), trust `X-Forwarded-For` without a trusted-proxy CIDR list.

## The rate-limit bug: site-wide limits, not per-client

`router.go` installed only `middleware.ClientIPFromRemoteAddr` (stock chi v5.3.2). Prod's backend
container publishes no port — reachable only through Nginx Proxy Manager on `public-net` — so
`RemoteAddr` was **always NPM's own container IP**, for every caller on the internet. `ratelimit.go`'s
`keyByClientIP` already had a doc comment anticipating this exact problem, and
`defaultRateLimitGlobalPerIP` had already been bumped 120→240 specifically to buy headroom against
it (see [[stream-connection-tickets-2026-09-03]]). The real severity: `RATE_LIMIT_LOGIN_PER_IP=10`
and `RATE_LIMIT_REFRESH_PER_IP=30` were consequently **global** limits for the whole site — with
`JWT_TTL=15m` and a silent refresh-on-401 (see [[session-token-storage-2026-09-05]]), 30
refreshes/min shared across every logged-in user is a real throttling hazard, not a theoretical one.

**Fix**: chain `middleware.ClientIPFromXFF()` (stock chi, zero custom code) after
`ClientIPFromRemoteAddr` in `router.go`. Called with **no trusted-prefix arguments**, it takes the
**rightmost** XFF entry — the entry NPM itself appends via
`proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for` — and a client-forged XFF value lands
to its *left*, ignored. No XFF header at all (dev, direct hits) → falls back to `RemoteAddr`
unchanged. `keyByClientIP` needed zero changes — it already just reads `middleware.GetClientIP`.

Accepted residual risk: anything that reaches `backend:8080` directly on `public-net` can forge XFF
for a fresh bucket per request. Prod publishes no ports there; only NPM and sibling containers are
on that network. If a second untrusted hop is ever added in front of NPM, switch to
`ClientIPFromXFF(trustedCIDRs...)` with NPM's actual address, not the no-argument form — noted
directly in `ratelimit.go`'s doc comment now, replacing the *switch to XFF* TODO it used to carry.

New tests in `ratelimit_test.go` (rightmost-entry independence, spoofed-left-entry rejection,
no-XFF fallback, same-rightmost-different-RemoteAddr sharing a bucket) — none of the existing tests
had ever set XFF or a custom `RemoteAddr`, so this bug had zero test coverage before.

`defaultRateLimitGlobalPerIP` stays at 240 — it's per real client now, i.e. more generous than
originally intended rather than less; lowering it is a separate call, left alone.

## The CSP: what actually shipped

`deployments/docker/nginx.frontend.conf` set only COOP. Verified against a **real**
`docker compose build frontend` + inspecting the exported `dist/index.html` inside the image
(not something you can know from source alone with Expo's `output: single`):

- **No inline `<script>`** in the export — one `<style id=expo-reset>` inline block plus one
  external `<script src=/_expo/static/js/web/...>`. So `script-src 'self' 'wasm-unsafe-eval'`
  needed **no** `'unsafe-inline'` or hash — the sha256-hash fallback the owner pre-approved for
  this case (in case of an inline bootstrap script) turned out unnecessary.
- `style-src 'self' 'unsafe-inline'` **is** required and not avoidable without nonces (ruled out):
  the inline reset style plus `@tamagui/core`'s `inject-styles.cjs` (`styleEl.innerHTML = css` at
  runtime) plus react-native-web's `createCSSStyleSheet.js` (creates `<style>` nodes at runtime,
  no nonce hook).
- `connect-src 'self' ${CSP_CONNECT_EXTRA}` — templated via nginx's stock envsubst entrypoint
  (`/etc/nginx/templates/*.template` → `/etc/nginx/conf.d`, `nginx:alpine`'s
  `20-envsubst-on-templates.sh`), new compose var `CSP_CONNECT_EXTRA`: dev =
  `http://localhost:8080 ws://localhost:8080` (base `docker-compose.yml`), prod =
  `wss://jojo-one-piece-simulator.duckdns.org` (`docker-compose.prod.yml` override) — prod's REST
  is same-origin already, only the `wss://` socket needs naming explicitly rather than betting on
  browsers treating `'self'` as covering a scheme-different same-origin upgrade.
- `img-src https: data: blob:` stays broad rather than enumerating the storage fallback chain's
  three provider hosts ([[storage-fallback-chain]]) plus raw `lh3.googleusercontent.com` avatar
  URLs — images aren't the XSS vector this policy defends against.
- Also added while the header block was open (owner opted in): `X-Content-Type-Options: nosniff`,
  `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy` denying
  geolocation/microphone/camera/payment/usb/interest-cohort.

> [!warning] nginx `add_header` doesn't merge across levels — and this file already had that bug
> `add_header` in a `location` block **replaces** every `add_header` inherited from `server`, it
> doesn't add to it. The pre-existing COOP header was **already silently dropped** on
> `/_expo/static/*`, `/sw.js` and `/manifest.json` before this change, unnoticed until now. Fix:
> the full header block is now repeated verbatim in every `location` that sets its own
> `add_header`, with a comment on the `server`-level block explaining why. Do **not** factor this
> into an `include`d snippet file under `conf.d/` — `nginx:alpine` auto-includes every
> `/etc/nginx/conf.d/*.conf` at `http` level, so an envsubst-rendered snippet there would apply
> twice globally.

## Verification

Real Docker build (`docker compose build frontend`) confirmed the no-inline-script fact above.
Full dev stack up (`docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build`):
`curl -sI` on `/` and `/sw.js` showed every header present, `connect-src` correctly interpolated to
the dev value. Browser load of `/login` (Claude-in-Chrome): zero console messages of any kind, no
CSP violations — confirms `style-src 'unsafe-inline'` covers Tamagui/react-native-web's runtime
injection as expected. Full Google-login / catalogue-image / game-socket walkthrough was **not**
done live (needs the owner's manual two-account Google login per
[[auth_google_only_no_dev_bypass]]) — flagged here as the one thing to still eyeball in a real
session, watching DevTools for any CSP violation report on those specific flows.

Backend: `go vet ./...` + full `go test ./...` clean via Docker
([[feedback_backend_tests_via_docker]]), including the four new XFF rate-limit tests.

Related: [[session-token-storage-2026-09-05]], [[stream-connection-tickets-2026-09-03]],
[[auth-hardening-2026-09-02]], [[storage-fallback-chain]], [[norma-verificacion-docker]].
