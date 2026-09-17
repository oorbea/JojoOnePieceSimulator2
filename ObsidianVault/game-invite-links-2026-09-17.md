# Lobby invite links (2026-09-17)

## What shipped

The lobby's "share" button now shares a **join link**
(`https://<origin>/join/<token>`) instead of copying the raw join code.
"Copy" still copies the raw 6-character code — both buttons stay on
`JoinCodeCard`, distinct tooltips.

- Backend: `ports.IGameInviteStore` + `infrastructure/gameinvite` (memory) /
  `gameinvite/redis` adapters, copied from `streamticket`'s shape but
  `Lookup` instead of `Redeem` — an invite is **multi-use**, deliberately
  never consumed.
- `services/game_invite.go`: `CreateInvite` (any seated human, not just the
  host), `InviteStatus` (PUBLIC, `VALID`/`EXPIRED` only), `InvitePreview`
  (authenticated, works for PRIVATE lobbies), `JoinByInvite` (idempotent if
  already seated, its own `ErrGameAlreadyStarted` instead of a generic
  state-transition error, deliberately bypasses `JoinByID`'s
  `ErrLobbyPrivate` gate since the token *is* the credential).
- REST: `POST /games/{id}/invite`, `GET /games/invite/{token}/status`
  (public), `GET /games/invite/{token}` (authenticated preview),
  `POST /games/join-invite`. `POST /games/{id}/leave` added alongside it —
  the invite-link "leave your current game and join this one" confirm has
  no open WS socket to send `LEAVE` over.
- Frontend: `/join/[token]` route **outside** `app/(app)/` (same treatment
  as `/login`), `features/game/lib/pending-invite.ts` stashes the token in
  `sessionStorage` across Google's full-page web redirect, `login.tsx`
  resumes it. `features/game/lib/invite-url.ts` is the single place that
  knows the link's shape (`buildInviteUrl`) — a universal/app link later is
  a one-file change. `features/game/lib/share.ts` gained capability-
  detected sharing (`canSystemShare`/`shareInviteLink`) plus a desktop
  popover (`ShareInviteSheet`, reusing `DetailModal`'s chrome) with a
  locally-rendered QR (`react-native-qrcode-svg` over `react-native-svg`,
  no network call — required under the CSP's `default-src 'self'`).
  `features/game/stores/game-invite.store.ts` caches the minted token per
  `(gameId, code)`, reused while >2 min remain.

## Decision worth remembering: revocation is a comparison, not a second write

`GameInvite.Code` freezes the join code at mint time. `JoinByInvite`
(and `InviteStatus`/`InvitePreview`) compare it against the game's
*current* code (`IGameStore.Code`) at read/redeem time — no store write
added inside `RegenerateGameCode`'s critical section. Rejected alternative:
a per-game token set, `DEL`'d on rotation — that needs a second write
inside the same lock `RegenerateGameCode` already holds, and its failure
mode is fail-open (a failed `DEL` leaves tokens alive against a rotated
code, the opposite of what revocation is for). The comparison design fails
closed and needs no reaper for this case (a vanished game already makes
`Code` return `ErrGameNotFound`).

## Deliberate scope cuts (owner-approved)

- **No ban/kick list.** `Kick` only removes the participant; its sole
  barrier to re-entry has always been the host rotating the join code
  (see `RegenerateGameCode`'s own doc comment). An invite link therefore
  can't distinguish "you were kicked" from "the code rotated" — both just
  show "this invite link has expired/no longer works". Adding a real
  ban-list (`game.Game` tracking kicked `UserID`s, checked in
  `joinLocked`) was scoped out as a separate, larger tanda since it would
  also change plain code-join behaviour, not just the invite-link path.
- **Public status endpoint answers nothing but `VALID`/`EXPIRED`.**
  Deliberately does not fold lobby state/fullness/lock into the answer —
  that would leak a lobby fact to an unauthenticated caller (an unfurl bot
  included). Those causes (`GAME_ALREADY_STARTED`, `GAME_FULL`,
  `LOBBY_LOCKED`, etc.) only ever surface at redeem time, behind auth.
- **No rich link preview in chat apps.** The web app is a client-rendered
  SPA with generic `index.html` meta, so WhatsApp/Telegram unfurl a plain
  link, not a card with the lobby's name/host. Server-side meta injection
  would fix it but reopens the same "don't leak lobby facts to an
  unauthenticated bot" question the status endpoint was built to avoid —
  flagged, not built.

## Known constraint if this is ever touched again

`use-google-auth.ts`'s web flow is a **full-page redirect**, not a popup
(COOP from `accounts.google.com` breaks popup completion detection). Its
`redirect_uri` is `origin + pathname` and must be a URL registered with
Google ahead of time — so **login can never be initiated from
`/join/<token>`** (a per-token URL can never be pre-registered). The
invite flow works around this by stashing the token and redirecting to the
already-registered `/login` first, not by trying to fix the redirect URI.

## See also

[[stream-connection-tickets-2026-09-03]] (the short-TTL Redis token
precedent this copies), [[game-lobby-frontend]], [[auth-login-implementation]],
[[session-token-storage-2026-09-05]].
