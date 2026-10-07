package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// maxPushSubscriptionsPerUser caps how many devices one account can keep
// registered; every extra one is a delivery attempt per notification, and a
// user has no legitimate need for more than a handful of phones/browsers.
const maxPushSubscriptionsPerUser = 10

// PushSubscriptionRepository is the Postgres-backed
// ports.IPushSubscriptionRepository adapter over push_subscriptions.
type PushSubscriptionRepository struct {
	queries *db.Queries
}

var _ ports.IPushSubscriptionRepository = (*PushSubscriptionRepository)(nil)

// NewPushSubscriptionRepository builds a PushSubscriptionRepository over pool.
func NewPushSubscriptionRepository(pool *pgxpool.Pool) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{queries: db.New(pool)}
}

func (r *PushSubscriptionRepository) Upsert(ctx context.Context, sub ports.PushSubscription) error {
	owner := pgtype.UUID{Bytes: sub.UserID, Valid: true}
	if err := r.queries.UpsertPushSubscription(ctx, db.UpsertPushSubscriptionParams{
		UserID:   owner,
		Endpoint: sub.Endpoint,
		P256dh:   sub.P256dh,
		Auth:     sub.Auth,
	}); err != nil {
		return fmt.Errorf("upserting push subscription: %w", wrapPgError(err, ports.ErrConstraintViolation))
	}
	// Trimming is housekeeping: a failure here must not fail the registration
	// the user just made, the next upsert trims again.
	if err := r.queries.TrimPushSubscriptionsForUser(ctx, db.TrimPushSubscriptionsForUserParams{
		UserID: owner,
		Limit:  maxPushSubscriptionsPerUser,
	}); err != nil {
		fmt.Printf("trimming push subscriptions of user %s: %v\n", sub.UserID, err)
	}
	return nil
}

func (r *PushSubscriptionRepository) DeleteForUser(ctx context.Context, userID user.UserID, endpoint string) error {
	if err := r.queries.DeletePushSubscriptionForUser(ctx, db.DeletePushSubscriptionForUserParams{
		UserID:   pgtype.UUID{Bytes: userID, Valid: true},
		Endpoint: endpoint,
	}); err != nil {
		return fmt.Errorf("deleting push subscription: %w", err)
	}
	return nil
}

func (r *PushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	if err := r.queries.DeletePushSubscriptionByEndpoint(ctx, endpoint); err != nil {
		return fmt.Errorf("deleting push subscription by endpoint: %w", err)
	}
	return nil
}

func (r *PushSubscriptionRepository) ListByUser(ctx context.Context, userID user.UserID) ([]ports.PushSubscription, error) {
	rows, err := r.queries.ListPushSubscriptionsByUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("listing push subscriptions: %w", err)
	}
	subs := make([]ports.PushSubscription, 0, len(rows))
	for _, row := range rows {
		subs = append(subs, ports.PushSubscription{
			UserID:     user.UserID(row.UserID.Bytes),
			Endpoint:   row.Endpoint,
			P256dh:     row.P256dh,
			Auth:       row.Auth,
			LastSeenAt: row.LastSeenAt.Time,
		})
	}
	return subs, nil
}
