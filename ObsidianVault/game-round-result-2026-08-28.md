# Round-resolved feedback (vote tally + winner) — 2026-08-28

Closes the "Round-resolved feedback" bullet open since [[game-vote-buttons-2026-08-26]] /
[[game-match-assignment-frontend]] (§6 of [[game-lobby-todo]]): until now nothing rendered
`ROUND_RESOLVED` beyond clearing the countdown, and the game jumped straight from a vote to the next
round's sorteo with no visible outcome at all.

## Owner's product decisions (asked up front, all delegated technically)

1. **Detail level**: per-option vote count **plus a nominal breakdown** (who voted what), rendered as
   **avatars only** (name on hover/tooltip) — consistent with the roster redesign
   ([[sorteo_roster_redesign_2026-08-17]]) already showing avatar+username, not a full stat card.
2. **Placement**: **inline**, replacing `VoteBar` in the exact same slot — no modal, no separate
   result history screen.
3. **Duration**: **automatic + a "skip" button** (client-side only — see below).
4. **Ties**: **yes**, show the tied vote breakdown before the revote opens, not just a bare
   "tie - revote" label.
5. **Explicitly out of scope**: the final-game result screen (`GAME_FINISHED` still just toasts and
   routes to `/play`) — a separate, larger future tanda.

## The two backend changes this actually required

This was **not** a frontend-only feature, contrary to the first read of the ask:

- **`RESOLVING` had no observable window.** `Game.resolveRound` used to set `Round.Result`, emit
  `RoundResolved`, and in the very same call advance the state to `ASSIGNING`/`FINISHED` -
  `GameService.closeVoting` then chained `beginRound` immediately after. A client could never actually
  see the game sitting in `RESOLVING`; the frame arrived and was already stale. Fixed by splitting
  `resolveRound` (parks the Game in `RESOLVING` with `Round.Result` set, nothing else) from a new
  `(*Game) CompleteRound()` (does what the second half used to: `mode.ApplyRoundResult` → `FINISHED`
  + `GameFinished`, or `ASSIGNING`). `GameService.scheduleResultDelay`/`completeRoundAfterResult` hold
  that gap open for `game.ResultDuration` (**6s, fixed, not host-configurable** — same as the sorteo's
  own `RevealDuration`), an exact structural copy of `scheduleRevealDelay`/`openVotingAfterReveal`. A
  new `resultEnds` map + `ResultEndsAt(id)` getter joins `revealEnds`/`votingEnds`, and
  `GameStateResponse.resultEndsAt` mirrors `revealEndsAt`/`votingEndsAt` for a (re)connecting client.
- **A tie's votes were destroyed, not just hidden.** `Game.CloseVoting` calls `Ballot.Reset()` the
  moment a tie opens `TIEBREAK` (deliberately, so the revote starts from an empty ballot — see
  `ballot.go`'s doc) — before this tanda, that vote breakdown was gone forever, all the frontend ever
  had was the domain's `TiebreakUsed` bool. Since the owner wants the tied breakdown *shown*, `Round`
  gained a `TiedVotes map[ParticipantID]OptionID`, populated (via `Ballot.Votes()`'s existing copy
  getter) right before `Reset()`. `dto.GameRoundResponse.TiedVotes` reveals it unconditionally, in
  the round-trip Redis snapshot (`RoundSnapshot.TiedVotes`, same slice-not-map shape as
  `BallotSnapshot.Votes` — see [[game-lobby-persistence]]) so it survives a reconnect mid-`TIEBREAK`.
  This is now the **one deliberate exception** to "votes hidden while a round is live" (see
  [[game-realtime-transport]]'s "votes hidden until a round resolves" section, updated to note it).

What did **not** need to change: once a round resolves, `GameRoundResponse.votes` (participant →
option) was already fully revealed by the existing STATE-after-event rule — the resolved-round tally
itself is pure frontend rendering of data already on the wire.

## Frontend

- `lib/vote-options.ts` gained `voteTally(snapshot, you, votes)` → per-option `{ ...VoteOption, count,
  voterIds }[]` + `maxCount` (floored at 1, avoids a divide-by-zero on a zero-vote round). Placed here
  rather than in `match-rules.ts` as first sketched, to avoid a circular import — `vote-options.ts`
  already imports `currentRound` from `match-rules.ts`, and `voteTally` needs `voteOptions` itself (so
  the result panel's labels/tones/`isOwnTeam` can never drift from the vote bar's). Documented
  in-file; `match-rules.ts` itself is untouched by this tanda beyond its test fixture gaining the two
  new `LiveMatchState` fields.
- New `components/presentational/match/round-result-panel.tsx` (`RoundResultPanel`): a `GlassPanel` in
  the same box/padding as `VoteBar`, one row per option (label + `MeterBar` + count), the winner
  outlined the same way `VoteBar` outlines your own cast vote, a coin-flip badge when
  `result.decidedByCoinFlip`, and a row of `VoterAvatar` per option (new, thin wrapper around the
  existing `ParticipantAvatar` + the same `useTooltipTrigger`/`TooltipBubble` pattern `GlossButton`
  already uses for its own tooltip — so the name-on-hover costs nothing new). Two variants: `'result'`
  (has a winner, shown during `RESOLVING`) and `'tie'` (no winner yet, shown during `TIEBREAK` **above**
  the revote's own `VoteBar`, per the owner's "yes show the tie" call).
- **No client-side countdown bar in the panel** — a deliberate simplification: the server alone
  decides when `RESOLVING` ends (`game.ResultDuration`/`CompleteRound`), so the panel has nothing to
  count down against except by trusting `resultEndsAt`, and the owner's mockup showed no timer bar
  either. "Skip" (`onSkipResult` → the socket store's new `dismissResult()`) only hides the panel
  locally via `live.resultDismissed`; it can never move the game along faster than the server's own
  6s. `resultDismissed` resets on `VOTING_OPENED`/`TIEBREAK_OPENED`/`ROUND_RESOLVED` so the next
  round's panel starts visible again.
- `stores/game-socket.store.ts`: `LiveMatchState` gained `resultEndsAt`/`resultDismissed`, adopted
  from a STATE frame exactly like `revealEndsAt`/`votingClosesAt` (`ROUND_RESOLVED` itself carries no
  `closesAt`, so it always arrives via the STATE frame that immediately follows).
- `match-screen.tsx`: where the vote slot used to be a flat `votingOpen ? <VoteBar/> : null`, it's now
  `showTiedPanel` (above the revote bar) / `showResolvedPanel` (replaces the bar) / `votingOpen`, in
  that order.
- i18n: new `game.match.result.*` keys (`title`, `winner`, `coinFlip`, `tie`, `votesA11y`, `skip`,
  `skipA11y`, `noVotes`) in all three locales (en-GB/es-ES/ca-ES), same nested-dot-path convention as
  the rest of `game.match`.

## Verified

Backend: `go build`/`go vet`/`go test ./...` green in Docker (existing tests updated for the new
`RESOLVING` pause — several had to insert an explicit `CompleteRound()`/`advanceResult(deps)` call
where they used to assume closing a vote landed straight on the next state), plus new tests:
`Game.CompleteRound`'s state guard, `CloseVoting`'s `TiedVotes` capture, `GameService.ResultEndsAt`'s
track/clear/abort-cancels lifecycle, the DTO's `TiedVotes`/`ResultEndsAt` round-trip, and the redis
snapshot round-trip extended to cover `TiedVotes`. Frontend: `typecheck`/`lint`/`test:ci` green in
Docker (445/445 — one `tooltip.test.tsx` failure under full worker parallelism was the documented
Docker flake, reproduced as passing at `--maxWorkers=2`, see [[norma-verificacion-docker]]).

**Not yet done** (flagged, not skipped silently): the live two-browser `claude-in-chrome` walkthrough
for the tie/revote/coin-flip path, and a real solo `local-up` playthrough confirming the 6s auto-
advance and the skip button feel right in practice — same category of follow-up
[[game-vote-buttons-2026-08-26]] already left open and never closed.
