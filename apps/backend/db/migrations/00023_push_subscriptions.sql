-- +goose Up
-- push_subscriptions holds one Web Push subscription per (browser profile,
-- app install): the PushManager endpoint the push service gave that device
-- plus the keys needed to encrypt a payload for it. endpoint is UNIQUE
-- because it IS the device's identity - logging into another account on the
-- same device re-registers the same endpoint, which must move it to the new
-- owner (see db/query/push_subscriptions.sql's upsert) instead of leaving two
-- accounts' notifications landing on one phone. user_id cascades: deleting an
-- account deletes its subscriptions (unlike game_result_participants, there
-- is nothing here worth keeping once nobody owns it).
-- +goose StatementBegin
CREATE TABLE push_subscriptions (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    endpoint     text        NOT NULL UNIQUE,
    p256dh       text        NOT NULL,
    auth         text        NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX push_subscriptions_user_idx ON push_subscriptions (user_id, last_seen_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE push_subscriptions;
-- +goose StatementEnd
