---
title: "Feature: vote buttons + live vote count + tiebreak revote (2026-08-26)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - feature
  - gameplay
---

# Vote buttons + live vote count + revote (2026-08-26)

Closes the first slice of [[game-lobby-todo]]'s §6 second half: nobody could actually vote before
this - `hooks/use-game-commands.ts`'s `vote(option)` sender existed with zero call sites. Two
backend bugs blocked it and were fixed first.

## Backend blockers fixed

1. **`VOTE_CAST.votesCast` was always 0.** `game.Game` gained a private `humanVoteProgress() (cast,
   total int)` (`entities/game/game.go`, right above `VotingComplete`, which now delegates to it) -
   connected humans only, never bots, exactly the population `VotingComplete` waits on. Both emit
   sites (`CastVote`, `castBotVotes`) fill the new `VoteCast.HumanVotesCast`/`HumanVoters` fields
   from it, so a bot's own event still reports human-only numbers. The transport
   (`game_ws_endpoints.go`'s `buildEventFrame`) forwards both into `dto.VoteCastPayload`'s new
   `Voters int` field (`VotesCast` already existed on the wire, just never set).
2. **No voting deadline in the snapshot.** `GameService` gained `votingEnds
   map[GameID]time.Time` and a `VotingEndsAt(id) (time.Time, bool)` accessor - an exact mirror of
   the existing `revealEnds`/`RevealEndsAt` pattern. `dto.GameStateDeadlines{RevealEndsAt,
   VotingEndsAt *time.Time}` replaces `NewGameStateResponse`'s old single `revealEndsAt *time.Time`
   param (a struct, not a second positional pointer - two same-typed pointers next to each other are
   silently swappable at the call site). `GameSnapshotResponse.VotingEndsAt *string` (RFC3339,
   omitempty) is the new field a reconnecting client reads.

## Owner decision made mid-design: the revote now clears the ballot

Found while designing the counter: `Game.CloseVoting` flipped to `TIEBREAK` on a tie WITHOUT
clearing `Round.Ballot`'s votes - so a revote window opened already reading `cast==total`
(`VotingComplete` already true), and the very first changed vote closed it instantly. Owner's call:
`Ballot` gained `Reset()`; `CloseVoting` calls it on the first tie and re-runs `castBotVotes` for
that round (so Versus bot votes aren't silently dropped from the revote tally). Consequence
accepted: bots vote deterministically, so a bot-heavy tie tends to reproduce itself into the coin
flip - existing design, just reached honestly now. `game_service_test.go`'s
`TestCloseVoting_Tie_OpensRevoteThenUsesTiebreaker` had to be rewritten (both players recast in the
revote, not just one) since it depended on the old carry-over behavior.

## Frontend

- `hooks/use-game-commands.ts`'s existing `vote(option)` finally has a caller:
  `lobby-room-container.tsx`'s `handleVote` - first vote goes straight through, changing an
  already-cast vote opens the existing `ConfirmSheet` (new `tone` prop, defaults `'red'` so every
  existing destructive-action caller is unchanged; `'blue'` here since changing a vote isn't
  destructive).
- New `match/vote-bar.tsx`: fixed panel at the bottom of the match view (bottom `position:sticky` on
  web only, last in scroll content on native - the a11y/pointerEvents-leak platform-branch norm, see
  [[a11y-web-leak]]) with a draining `MeterBar` (new shared primitive,
  `shared/components/presentational/meter-bar.tsx`) + seconds, the `cast/total` human count, and the
  options as a keyboard radio group.
- `lib/vote-options.ts` (new): maps a round's raw option ids (`"SURVIVE"`/`"FALL"` for Gauntlet,
  raw team-UUID strings for Versus - see `game.OptionID`'s backend doc) onto
  `{id, labelKey|label, tone, isOwnTeam}`. Versus reuses `lobby-rules.ts`'s `teamTone`, not a second
  tone table. `isOwnTeam` is informational only (a discreet "tu equipo" marker) - voting for the
  rival team stays legal per the domain, never blocked.
- `lib/match-rules.ts` gained `voteProgress(snapshot, live)`: prefers the live `VOTE_CAST` frame's
  absolute counters when they're for the current round, falls back to
  `round.votedParticipantIds` intersected with connected non-bot participants otherwise - the
  fallback is what makes a reconnect mid-vote render correct numbers with zero frames received yet,
  and it mirrors the backend's own `humanVoteProgress` definition exactly.
- `game-socket.store.ts`: `LiveMatchState` gained `votesCast`/`voters` (absolute, never incremented -
  bots emit several `VOTE_CAST` frames in a row the instant a window opens, all `0/N`). `VOTE_CAST`
  is no longer a no-op - it writes those two fields, but only when the frame's `roundIndex` matches
  `live.votingRoundIndex` (guards against a late frame from an already-resolved round).
  `VOTING_OPENED`/`TIEBREAK_OPENED` reset to `0`/`null`; `ROUND_RESOLVED` clears both to `null`.
  `STATE` adopts `game.votingEndsAt` into `live.votingClosesAt` with the exact same "only if nothing
  local is already tracking one" guard the existing `revealEndsAt` adoption uses.
- `VotingStatusBar` lost its own `closesIn` countdown text (now `VoteBar`'s job, no longer shown
  twice) and its hand-rolled `setInterval` in favour of the new shared `shared/hooks/use-now.ts` -
  `ConnectionBanner`'s own hand-rolled interval was migrated to the same hook in the same pass so a
  third copy never appears.

## Keyboard navigation (new standing norm, not just this feature)

See [[norma-teclado.md]] for the full norm this tanda established. Summary as it applies here:
- Vote options are a roving-tabindex radio group (`shared/hooks/use-roving-group.ts`, new): one Tab
  stop for the group, arrows move (wrapping), Home/End jump, Enter/Space votes.
- Roster tiles (`match-roster.tsx`/`participant-tile.tsx`) are now Tab-reachable with the same
  roving group, Enter/Space opens `LoadoutModal`. The hover card
  (`roster-participant.tsx`'s `useHoverTrigger`) turned out to **already** open on keyboard focus on
  web (`useHoverTrigger`'s web `triggerProps` already wire `onFocus`/`onBlur`) - the actual gap was
  only that the tile itself (a plain `YStack`) was never in the tab order at all.
- Single-key shortcuts (`lib/hotkeys.ts`'s pure `resolveHotkey` + `hooks/use-match-hotkeys.ts`'s thin
  listener): `1`-`9` vote the nth option, `S` skips the sorteo - guarded against text-input focus,
  modifier keys, and an open overlay (`blocked` flag).
- **Known gap, not silent**: `blocked` today only reflects `ConfirmSheet` (owned by the container).
  `LoadoutModal`'s open state lives inside `MatchRoster` by design (it's the component that already
  has `mangas` for the modal content) - hotkeys aren't suppressed while it's open yet.
- **Known test-coverage gap**: no automated test for `use-roving-group.ts`'s web-only keyboard branch
  - `hooks/__tests__` always runs under jest's native (react-test-renderer) project per
  `jest.config.js`'s current split, where `Platform.OS !== 'web'` and the hook's web branch never
  engages. Would need either a new jsdom-backed hooks test lane or moving this specific test under
  `shared/lib/__tests__`. Verified manually instead (keyboard-only pass, see below); not yet done in
  this session, flagged for the next one.

## Verification

Backend: `docker compose ... backend-test go test ./...` green (`internal/application/services`,
`internal/domain/entities/game`, `internal/infrastructure/api/dto`, `internal/infrastructure/api/
endpoints` all pass; the redis-backed packages fail without `db-up` running, pre-existing and
unrelated to this diff). Frontend: `pnpm typecheck` clean, `pnpm lint` 0 errors (4426 CRLF warnings
across 108 files are this Windows checkout's `core.autocrlf=true`, pre-existing, not from this
diff), `pnpm test:ci` 43/43 suites, 429/429 tests green (two tests flaked once under full-parallelism
load and passed clean in isolation/with `--maxWorkers=2` - see [[norma-verificacion-docker]]).

**Not yet done this session**: the live `local-up` + `claude-in-chrome` two-browser walkthrough
(vote in one tab, watch the count move in the other; force a tie and confirm the revote opens at
0/2; reload mid-vote and confirm the countdown resumes) and a real keyboard-only pass with the mouse
untouched. Both are next-session verification, not skipped by oversight - flagging explicitly per
[[zettelkasten-workflow]].

**Partially closed 2026-08-28 (`local-up` + `claude-in-chrome`, solo Gauntlet, no second account
available)**: confirmed live, single-participant only - cast a vote on "Sobrevivimos", `0/1 han
votado` incremented and the round resolved/advanced immediately (single connected human = 100%
either way); the hotkey hint (`Teclas: 1-9 para votar, Esc para cerrar`) rendered; casting "Caemos"
on a solo round correctly triggered `GAME_FINISHED` → toast "Esta partida ha terminado." → bounce to
`/play`, matching the documented §6 gap below exactly (no result screen, just a toast). **Still not
verified**: the actual two-browser scenario this note calls out (a second client's live vote count
updating, a real tie forcing a revote, reconnect-mid-vote countdown resume) - no second login was
available this session (Google OAuth only, no guest/dev-bypass auth found) - and the keyboard-only
manual pass. Still open for a session with two real accounts.

## Deliberately still out of scope (§6, unblocked by this tanda but not built)

- Round-resolved feedback (who won, was it a coin flip) - nothing renders `ROUND_RESOLVED` beyond
  clearing the countdown.
- The final result screen - `GAME_FINISHED` still just toasts and bounces to `/play`.
- `ports.IInventory` still has no adapter.
- `GET /games/preview?code=`, `DELETE /games/{id}/participants/me` - still deferred.
- The `Seq` counter on `GameEvent` to detect (not just self-heal) a hub drop.
- `game_endpoints_test.go`/`game_ws_endpoints_test.go` HTTP/WS coverage gap from §2 - still open
  (this tanda added direct `buildEventFrame` tests to the existing `game_ws_endpoints_test.go`, but
  the broader route-level suite from §2 is unrelated and still missing).
- `VOTING_OPENED`/`TIEBREAK_OPENED`'s `closesAt` is still transport-synthesized
  (`time.Now()+window`) separately from the new authoritative `votingEndsAt` - they can drift by
  however long hub delivery took. `forwardEvents` already has `e.svc`/`gameID` in scope to fix this
  cleanly later; not folded into this tanda.
- A backend restart mid-vote still loses `s.timers`/`s.votingEnds` entirely (in-process only, same
  category as the pre-existing `revealEnds` gap) - wedges the lobby in `VOTING` with no client-facing
  deadline and no timer that will ever close the window.

Related: [[game-lobby-todo]], [[gameplay-game-modes]], [[gameplay-application-layer]],
[[game-realtime-transport]], [[game-match-assignment-frontend]], [[norma-teclado]],
[[norma-verificacion-docker]], [[norma-tooltips-y-ayuda-contextual]], [[frontend-stack]].
