---
title: "Live two-browser walkthrough: tie/revote/coin-flip + reconnect + keyboard-only (2026-09-02)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - gameplay
  - bug
---

# Live two-browser walkthrough (2026-09-02)

Closes the walkthrough flagged open since 2026-08-26 ([[game-vote-buttons-2026-08-26]],
[[game-round-result-2026-08-28]], [[game-lobby-todo]] §6) - the live `local-up` + `claude-in-chrome`
pass for the tie/revote/coin-flip path and a keyboard-only manual check, plus a reconnect-mid-vote
test. Real two-browser test with **two distinct real Google accounts** (auth is Google-only, no dev
bypass exists - see "Auth blocker" below), not the bot/single-account workaround prior sessions used.

## Setup

`local-up` stack (postgres/redis/backend/frontend in Docker), two Chrome tabs each with an isolated
live session (host `trollotron`, guest `meck_porky`), Versus 1v1 lobby, no bots, default voting
window. Two tabs in the same Chrome profile share `localStorage` (auth token lives there, see
`secure-storage.ts`) - the trick that keeps both sessions genuinely live at once is: log in tab A,
let it finish loading (token now in memory, WS open), *then* log in tab B (overwrites the shared
`localStorage` token, but tab A never re-reads storage unless it reloads/navigates again). Confirmed
working: both tabs stayed independently authenticated and connected for the whole session.

## Auth blocker (read before attempting this again)

`POST /auth/google` is the only login route (`apps/backend/internal/infrastructure/api/endpoints/auth_endpoints.go`);
`AuthService.LoginWithGoogle` always uses the real `GoogleVerifier` in `main.go` wiring - the
`fakeGoogleVerifier` only exists in `auth_service_test.go`, never wired into a runnable server. There
is **no dev-mode bypass**. A live multi-account walkthrough therefore requires the owner to actually
log in two real Google accounts in two browser tabs - Claude cannot do this alone (entering
third-party credentials is out of bounds). If this needs automating without owner involvement in the
future, the concrete option is an `AUTH_DEV_BYPASS`-style env var wiring a fake verifier for
`local-up` only, never for prod - not built, owner explicitly deferred this in favor of doing the
login manually this time.

## What passed

- **Sync on lobby join**: second account joining by code updated the host's screen live (Team A/B
  counts) - already covered by [[game-lobby-persistence]] but re-confirmed end-to-end through real
  auth this time.
- **Reconnect mid-vote**: reloaded the guest tab (full page navigate, not just a WS drop) with ~22s
  left on an open VOTING window. Reconnected straight back into the same voting screen, correct
  remaining time, own vote state intact, "0/2 han votado" preserved. No re-vote needed, no stuck
  loading state.
- **Tie -> revote -> coin-flip, end-to-end**: two real votes cast for different options (Team A vs
  Team B, 1-1) correctly opened `TIEBREAK`. A second 0-0 tie (nobody voted before the revote's timer
  ran out) correctly escalated to the coin-flip path (`ports.ITiebreaker`). The **RESOLVING** result
  screen correctly showed "Team B wins", "Decided by a coin flip", and the per-option tally
  ("Team A 1 / Team B 1") - `RoundResultPanel`'s `'result'` variant and `voteTally()` are both fine.
- **Keyboard-only, partial**: the sorteo reveal's `S` hotkey (skip) worked from a live keypress with
  no mouse involved, confirmed via `get_page_text` showing the state actually advanced. Full
  1-9 vote-hotkey + Tab roving-tabindex coverage (already unit-tested per [[norma-teclado]]) was
  **not** re-verified live this session - the browser extension disconnected mid-round-3 wait (a
  tooling hiccup, not an app bug - see below) and the match ended (host cancelled) before it could be
  retried. Flag as still-open if a fully clean live keyboard pass matters before shipping further
  in-match UI.

## Bug found: tied-vote breakdown never renders during TIEBREAK

**This is the one real regression this session found**, and it directly contradicts what
[[game-round-result-2026-08-28]] documented as shipped and verified.

Repro: round 1, two real human votes cast for different options -> genuine 1-1 tie -> game correctly
enters `TIEBREAK` (`Voto de desempate abierto` header, VoteBar shows "Empate - revotación"). Per the
owner's explicit 2026-08-28 call, a `RoundResultPanel` in its `'tie'` variant should render **above**
that VoteBar, showing the tied round's per-option counts + voter avatars
(`match-screen.tsx`'s `showTiedPanel = snapshot.state === 'TIEBREAK' && !!round?.tiedVotes`). It did
not. Confirmed via `get_page_text` (real rendered DOM text, not a screenshot misread) on **both**
clients during the live TIEBREAK window - no vote counts, no avatars, nothing between the header and
the VoteBar.

Notably the *other* variant of the same component (`'result'`, shown during `RESOLVING` after the
tiebreak coin-flip resolves) rendered correctly with the right tally in the same match - so
`RoundResultPanel`/`voteTally()` are not fundamentally broken; something specific to the
TIEBREAK-variant data path (or its gating condition) is failing.

**Lead, not yet confirmed**: `apps/frontend/src/features/game/stores/game-socket.store.ts`'s
`SERVER_FRAME.STATE` handler casts `frame.payload as GameStateResponse` (`{game, you}` - confirmed by
hitting `GET /api/v1/games/:id` live) but then stores the **whole** `payload` into `state.snapshot`
(`{ snapshot: payload }`), while `snapshot.state`/`snapshot.rounds` consumers elsewhere
(`match-rules.ts`'s `currentRound`, `match-screen.tsx`) expect `snapshot` to already be a flat
`GameSnapshot`. Whether the WS wire payload is actually already flattened at the protocol level
(unlike the REST DTO) - which would make this a non-issue - was not verified live before the session
ended.

**Delegated to a subagent** (per owner instruction, 2026-09-02: send bugs found during this
walkthrough to a subagent to investigate/fix/commit atomically, no co-author, so live testing could
keep iterating) rather than fixed inline. See git log around this date for the actual root cause and
fix - update this note once known if it turns out to be something other than the lead above.

### Subagent investigation result (2026-09-02)

The `{game, you}`-flattening lead above is **not** the bug: `use-game-socket.ts` already does
`snapshot: snapshot?.game ?? null` before handing `snapshot` to `MatchScreen`/`match-rules.ts` - the
flattening was already correct.

Traced the whole pipeline instead: domain (`Game.CloseVoting` sets `round.TiedVotes` before the state
flips to TIEBREAK, same mutation), `dto.NewGameStateResponse`/`NewGameRoundResponse` (already covered
by `TestNewGameStateResponse_TiedVotes`, and now also by a new
`TestCastVote_Tie_DTOStateCarriesTiedVotes` in `game_service_test.go` that drives the tie through the
*real* `GameService.CastVote` path instead of hand-building a TIEBREAK `Game` via direct domain
calls - passes), and `GameService.withGame`'s publish/persist ordering (initially suspected a
resend-races-ahead-of-Save bug against the Redis store; disproven - `GetGame` and `withGame` share the
same per-`GameID` mutex, so a resend can never observe the store before the mutation that triggered it
has been fully saved). Every layer checked out correct on the current `develop` HEAD.

Could not re-run the live two-browser repro (Google OAuth has no dev bypass - see "Auth blocker"
above - and the subagent has no real Google credentials), so the exact live failure couldn't be
re-triggered. What *was* found: `match-screen.tsx`'s `showTiedPanel`/`showResolvedPanel` gating
(`snapshot.state === 'TIEBREAK' && !!round?.tiedVotes && !live.resultDismissed`) had **zero** test
coverage - no component test for `MatchScreen`/`RoundResultPanel` at all, unlike every other piece of
view logic in this feature (`voteProgress`, `currentRound`, `hasAllLoadouts` all live in
`match-rules.ts` and are unit-tested there). Extracted it into `match-rules.ts` as
`roundResultPanelVariant(snapshot, live): 'tie' | 'result' | null`, wired `match-screen.tsx` to use it,
and added 6 unit tests covering the 'tie' case, the 0-0-tie no-panel case, the 'result' case, the
dismissed case, no-round, and stale-tiedVotes-outside-TIEBREAK. All pass on current code.

Bottom line: the current code, as far as every layer could be independently verified, is correct. If
the live repro is still reproducible on a fresh `local-up` build, the next step is a real live
walkthrough with `get_page_text` + `read_network_requests` capturing the actual WS `STATE` frame bytes
during the TIEBREAK window (not just the rendered DOM) to see whether the wire payload itself is
missing `tiedVotes` in that specific run - which would point back at something environment/timing
specific this session couldn't reproduce statically (e.g. a stale frontend bundle in that particular
`local-up` container predating this feature, or a genuine race not covered by the mutex reasoning
above that only a live capture would show).

## Environment note (not an app bug)

The Chrome extension MCP connection dropped mid-session during a long `wait` sequence (round-3
summary countdown) - "Chrome extension disconnected mid-operation", tab group lost. Re-running
`tabs_context_mcp` recovered the same tab IDs and both sessions were still alive/authenticated after
reconnecting, but the in-progress round-3 keyboard test was lost when the host later cancelled the
match. Transient tooling issue, not reproduced deliberately, not investigated further.
