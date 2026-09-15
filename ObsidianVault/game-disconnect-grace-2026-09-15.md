---
title: "Feature: disconnect grace period, in-match auto-vote takeover, resume (2026-09-15)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - feature
  - gameplay
  - gotcha
---

# Disconnect grace period + resume (2026-09-15)

Closed the ghost-player problem the owner flagged: closing a tab kept the seat forever
(`Game.Disconnect` never removed it), the roster went stale on other clients (no event was
emitted), and the player who left had no way back in (no "my active game" endpoint, and the
`games` query segment is excluded from RQ dehydration).

## Design (owner-approved, see the plan-mode transcript for the discussion)

- **45s grace in LOBBY** → a real `Leave` (seat freed) if nobody reconnects.
- **3min grace in a live match** → **relevo en el sitio**: `Game.Abandon` marks the seat
  `abandoned` without removing it - same `ParticipantID`/`userID`/loadout. In Versus it auto-votes
  with its own real loadout (even when `Config.AllowBots()` is false - it's not a bot in the config
  sense, just an orphaned seat); in Gauntlet it casts no vote, but the loadout still counts for the
  squad, which was the whole point of not just deleting the seat.
- Reconnecting at any point clears `abandoned`/`disconnectedAt` and hands full control back.
- `GET /api/v1/games/me` (backed by a new `ports.IGameStore.GamesForUser`) resumes the caller's
  active game - the other half of the fix, since a reload with no seat lost otherwise has nowhere
  to go.

## Load-bearing design decision: `hasConnectedHuman` → `hasActiveHuman`

The single biggest behavior change: **a mere `Disconnect` no longer aborts the Game**. Only
`Abandon` (grace elapsed) does. Before this, a solo player reloading their own app aborted their
own run the instant the socket closed - `checkAbortConditions` fired synchronously inside
`Disconnect`. Moving the abort decision to grace expiry is what makes both the grace period and
`GET /games/me` meaningful at all; without it there'd be nothing left to resume.

## `IGameMode.AutoVotesForAbandoned`

New interface method (Versus `true`, Gauntlet `false`) rather than branching on
`enums.GameModeKind` inside `Game` (which the aggregate's own doc forbids). `castBotVotes`'
guard became "connected bot, or abandoned seat in a mode that auto-votes" - `optionScores` already
summed every team member's loadout regardless of kind, so "votes with its real loadout" and "the
loadout still counts for the Gauntlet squad" both came for free once the guard was right. A new
`Game.CastAutoVoteFor(id)` handles the one case the batch-at-window-open `castBotVotes` can't: a
seat that becomes abandoned *while* a round is already voting (the grace timer firing mid-window).

## The 3-place rule bit a 4th time (and a genuinely new bug was found)

Everything here re-confirms `bugfixes-partida-2026-09-14.md`'s rule: new `Game`/`Participant` state
needs the entity, `Snapshot`/`Restore`, **and** `infrastructure/gamestore/redis/wire.go`, or it
silently vanishes under Redis while `MemoryGameStore`-backed tests stay green. `disconnectedAt`/
`abandoned` got all three from the start this time.

**Bonus find while auditing `wire.go` for that reason**: `wireParticipant` never carried
`AvatarFocalX`/`AvatarFocalY`, even though `game.ParticipantSnapshot` always has. Every
Get→Restore round trip through Redis silently reset a participant's picked focal point to 0,0
(crops jumping to the top-left corner in prod; invisible in any test using `MemoryGameStore`,
which is most of them). Fixed in the same edit, with a regression test
(`TestEncodeDecodeRoundTrip_DisconnectedAbandonedAndFocalPoint`) that would have caught it.

## Grace timers are a separate structure from phase timers

`GameService.timers` is one timer *per GameID* (the current phase). Grace is one timer *per
(GameID, ParticipantID)* - several participants can be mid-grace at once. New `graceTimers
map[graceKey]Timer` under the same `timersMu`. Mirrors the phase-timer machinery closely:
`armGraceTimer`/`cancelGraceTimer`/`cancelGraceTimersForGame`/`rearmGraceTimersLocked`, the last
wired into the same two call sites as `rearmPhaseTimerLocked` (`withGame`, `GetGame`) so a restart
heals exactly the same way. The deadline is **never stored absolute** - only `disconnectedAt` is
persisted, and `deadline = disconnectedAt + graceFor(currentState)` is re-derived every time
(arm, rearm, and inside the fire callback itself). That's what makes a LOBBY disconnect that gets
promoted to an in-match one (host starts the game before the 45s timer fires) land on the 3-minute
window instead of expiring under the wrong rule.

**Also new**: a terminal game must arm no grace timer at all - `Disconnect` checks
state != FINISHED/ABORTED before arming, and `finalizeLocked` cancels every pending grace timer for
the game. Every player closes their tab on the result screen; without this it'd leak one 3-minute
timer per socket on the single highest-volume path in the whole feature.

## `Reconnect` needed a new job: re-electing a host

Nothing before this ever re-elected a host into a lobby that went host-less (every candidate
disconnected/abandoned in turn, `reassignHost` sets `hostID = NilParticipantID`). `Reconnect` now
takes `rng` and calls `reassignHost` if `hostID.IsNil()` - otherwise the first player back into a
host-less lobby would sit with no START button and no way to get one.

## `GET /games/me`: a user can be in two games at once

`joinLocked`'s only duplicate-seat guard is *within* a single Game - nothing stops a user from
being seated in two lobbies simultaneously. `ports.IGameStore.GamesForUser` returns a slice;
`GameService.ActiveGameForUser` picks the first non-terminal, validated candidate (no "last
active" timestamp exists to break a tie more precisely, and it's rare enough not to justify one).

Redis index: `jojo:user:games:<uid>` ZSET, scored `now+ttl` exactly like `jojo:game:public`,
maintained via a plain pipeline (`Store.indexUserGames`) rather than inside the atomic
create/save Lua scripts - deliberately best-effort, since the store's own docs already accept a
stale/over-inclusive membership index (the read path validates every candidate against the Game
itself before trusting it).

## Frontend

- `ParticipantAvatar` (shared by every match roster surface) now dims+desaturates a disconnected
  seat and adds a small bot badge once `abandoned` - no new prop, both ride on `GameParticipant`
  that was already passed in.
- `GET /games/me` → `getMyGame`/`gameKeys.mine`/`useMyGame`, surfaced as a "Jump back in" banner on
  the home screen.
- i18n: `game.lobby.abandoned`, `home.resume.*` across all three locales.

## Verification

Backend: full `go build/vet/test` green in Docker, including the Redis-backed `gamestore/redis`
suite (needs `TEST_REDIS_URL`, already wired in `docker-compose.test.yml`). Frontend: typecheck +
lint (0 errors) + full jest (74 suites / 1407 tests) green. Contracts regenerated with no
unexpected drift (three new WS frames, two new `GameParticipantResponse` fields).

Not done in this tanda (deliberately out of scope, flagged during design): no numeric countdown
shown for the grace window (no wire field carries the grace duration, and hardcoding a duplicate
constant on the frontend felt worse than omitting it); no cross-game "one active game" guard.

Related: [[game-realtime-transport]], [[gameplay-application-layer]], [[game-lobby-persistence]],
[[bugfixes-partida-2026-09-14]], [[gameplay-domain-design]], [[norma-verificacion-docker]].
