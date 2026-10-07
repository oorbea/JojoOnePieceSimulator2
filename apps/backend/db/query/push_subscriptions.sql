-- name: UpsertPushSubscription :exec
INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
VALUES ($1, $2, $3, $4)
ON CONFLICT (endpoint) DO UPDATE
SET user_id      = EXCLUDED.user_id,
    p256dh       = EXCLUDED.p256dh,
    auth         = EXCLUDED.auth,
    last_seen_at = now();

-- name: ListPushSubscriptionsByUser :many
SELECT id, user_id, endpoint, p256dh, auth, created_at, last_seen_at
FROM push_subscriptions
WHERE user_id = $1
ORDER BY last_seen_at DESC;

-- name: DeletePushSubscriptionForUser :exec
DELETE FROM push_subscriptions
WHERE user_id = $1 AND endpoint = $2;

-- name: DeletePushSubscriptionByEndpoint :exec
DELETE FROM push_subscriptions
WHERE endpoint = $1;

-- name: TrimPushSubscriptionsForUser :exec
-- Keeps only the $2 most recently seen subscriptions of a user, so a user
-- cannot accumulate unbounded endpoints (each one costs a delivery attempt
-- per notification).
DELETE FROM push_subscriptions AS target
WHERE target.user_id = $1
  AND target.id NOT IN (
      SELECT keep.id
      FROM push_subscriptions AS keep
      WHERE keep.user_id = $1
      ORDER BY keep.last_seen_at DESC
      LIMIT $2
  );
