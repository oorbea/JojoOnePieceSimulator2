-- +goose Up
-- game_mode mirrors enums.GameModeKind.String() (GAUNTLET/VERSUS), same
-- convention as every other domain enum type in this schema.
-- +goose StatementBegin
CREATE TYPE game_mode AS ENUM ('GAUNTLET', 'VERSUS');
-- +goose StatementEnd

-- game_results is the persistent counterpart to ports.IGameHistory.Record,
-- called once by GameService.finalizeLocked right before a finished/aborted
-- Game is deleted from its (in-memory or Redis) live store. game_id is the
-- PK on purpose: it makes Record idempotent (ON CONFLICT DO NOTHING, see
-- db/query/game_history.sql), since finalizeLocked's call is best-effort
-- and a retry must not fail. winner stays plain text, not an enum: it's
-- genuinely polymorphic - SURVIVE/FALL for Gauntlet, a TeamID for Versus,
-- empty if aborted before any round resolved - see game.GameResult.Winner.
-- +goose StatementBegin
CREATE TABLE game_results (
    game_id       uuid        PRIMARY KEY,
    mode          game_mode   NOT NULL,
    winner        text        NOT NULL DEFAULT '',
    rounds_played integer     NOT NULL CHECK (rounds_played >= 0),
    aborted       boolean     NOT NULL,
    recorded_at   timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX game_results_recorded_at_idx ON game_results (recorded_at DESC);
-- +goose StatementEnd

-- game_result_participants lets history answer "what did I play", not just
-- "what happened" - see game.GameResult.Participants. user_id is nullable
-- with ON DELETE SET NULL, never CASCADE: a user deleting their account
-- must not erase the other players' match history, and display_name is
-- snapshotted so the row stays readable once user_id goes null.
-- participant_id/team_id reference nothing - both are per-game, in-memory-
-- only identities with no table of their own.
-- +goose StatementBegin
CREATE TABLE game_result_participants (
    game_id        uuid    NOT NULL REFERENCES game_results (game_id) ON DELETE CASCADE,
    participant_id uuid    NOT NULL,
    user_id        uuid    REFERENCES users (id) ON DELETE SET NULL,
    display_name   text    NOT NULL,
    team_id        uuid    NOT NULL,
    is_bot         boolean NOT NULL,
    PRIMARY KEY (game_id, participant_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX game_result_participants_user_idx ON game_result_participants (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE game_result_participants;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE game_results;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE game_mode;
-- +goose StatementEnd
