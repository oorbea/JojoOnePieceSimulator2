package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// StandRepository is the pgx/sqlc adapter for ports.IStandRepository. Stands
// are stored using class table inheritance: a base `powers` row plus a
// `stands` row sharing the same id.
type StandRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

var _ ports.IStandRepository = (*StandRepository)(nil)

func NewStandRepository(pool *pgxpool.Pool) *StandRepository {
	return &StandRepository{pool: pool, queries: db.New(pool)}
}

// Save upserts stand by name: the underlying powers/stands rows, and its
// skills, are fully replaced. It is safe to call repeatedly for the same
// stand (e.g. re-running seed data).
func (r *StandRepository) Save(ctx context.Context, stand *powers.Stand) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			fmt.Printf("rolling back transaction for stand %q: %v\n", stand.Name(), rollbackErr)
		}
	}()

	q := r.queries.WithTx(tx)

	var evolvesFromID pgtype.UUID
	if parent := stand.EvolvesFrom(); parent != nil {
		evolvesFromID = pgtype.UUID{Bytes: parent.ID(), Valid: true}
	}

	id, err := q.UpsertPower(ctx, db.UpsertPowerParams{
		ID:            pgtype.UUID{Bytes: stand.ID(), Valid: true},
		Name:          stand.Name(),
		Description:   stand.Description(),
		Rarity:        stand.Rarity().String(),
		Picture:       stand.Picture(),
		PictureThumb:  stand.PictureThumb(),
		PictureStatus: stand.PictureStatus().String(),
	})
	if err != nil {
		return fmt.Errorf("upserting power %q: %w", stand.Name(), wrapPgError(err, ports.ErrStandAlreadyExists))
	}

	if err := q.DeletePowerSkills(ctx, id); err != nil {
		return fmt.Errorf("clearing skills for %q: %w", stand.Name(), wrapPgError(err, ports.ErrStandAlreadyExists))
	}
	for position, skill := range stand.Skills() {
		if err := q.InsertPowerSkill(ctx, db.InsertPowerSkillParams{
			PowerID:  id,
			Position: int32(position),
			Skill:    skill,
		}); err != nil {
			return fmt.Errorf("inserting skill %q for %q: %w", skill, stand.Name(), wrapPgError(err, ports.ErrStandAlreadyExists))
		}
	}

	if err := q.UpsertStand(ctx, db.UpsertStandParams{
		ID:            id,
		AttackPower:   stand.AttackPower().String(),
		Speed:         stand.Speed().String(),
		AttackRange:   stand.AttackRange().String(),
		Endurance:     stand.Endurance().String(),
		Precision:     stand.Precision().String(),
		Potential:     stand.Potential().String(),
		EvolvesFromID: evolvesFromID,
	}); err != nil {
		return fmt.Errorf("upserting stand %q: %w", stand.Name(), wrapPgError(err, ports.ErrStandAlreadyExists))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing stand %q: %w", stand.Name(), err)
	}
	return nil
}

// FindByID loads the stand with the given id, along with its full
// evolves_from ancestor chain, in a single round trip.
func (r *StandRepository) FindByID(ctx context.Context, id powers.PowerID) (*powers.Stand, error) {
	rows, err := r.queries.GetStandRowsByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("querying stand %s: %w", id, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: %s", ports.ErrStandNotFound, id)
	}

	stands, err := buildStands(standRowsFromGetByID(rows))
	if err != nil {
		return nil, err
	}
	if len(stands) == 0 {
		// Every returned row was an ancestor of `id`, meaning `id` itself was
		// not among the rows - should be unreachable since the query's base
		// case filters on p.id = id.
		return nil, fmt.Errorf("%w: %s", ports.ErrStandNotFound, id)
	}
	return stands[0], nil
}

// UpdatePicture updates only a stand's picture renditions and pipeline
// status, leaving every other column (name, skills, stats, ...) untouched.
// A nil main or thumb leaves that column as-is - used by the PATCH
// .../picture handler to move a stand to PENDING without touching the
// renditions still being served, and by the background compression worker
// to publish new renditions once ready.
func (r *StandRepository) UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error {
	err := r.queries.UpdatePowerPicture(ctx, db.UpdatePowerPictureParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		Picture:       main,
		PictureThumb:  thumb,
		PictureStatus: status.String(),
	})
	if err != nil {
		return fmt.Errorf("updating picture for stand %s: %w", id, err)
	}
	return nil
}

// Delete removes the stand (and its power/skills rows) with the given id.
// Any descendant stand's evolves_from is cleared automatically by the
// schema's ON DELETE SET NULL.
func (r *StandRepository) Delete(ctx context.Context, id powers.PowerID) error {
	rowsAffected, err := r.queries.DeleteStandByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return fmt.Errorf("deleting stand %s: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", ports.ErrStandNotFound, id)
	}
	return nil
}

// FindByName loads the stand with the given name, along with its full
// evolves_from ancestor chain, in a single round trip.
func (r *StandRepository) FindByName(ctx context.Context, name string) (*powers.Stand, error) {
	rows, err := r.queries.GetStandRowsByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("querying stand %q: %w", name, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: %q", ports.ErrStandNotFound, name)
	}

	stands, err := buildStands(standRowsFromGetByName(rows))
	if err != nil {
		return nil, err
	}
	if len(stands) == 0 {
		// Every returned row was an ancestor of `name`, meaning `name`
		// itself was not among the rows - should be unreachable since the
		// query's base case filters on p.name = name.
		return nil, fmt.Errorf("%w: %q", ports.ErrStandNotFound, name)
	}
	return stands[0], nil
}

// GetAll loads every stand in the system.
func (r *StandRepository) GetAll(ctx context.Context) ([]*powers.Stand, error) {
	rows, err := r.queries.ListStandRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing stands: %w", err)
	}
	return buildStands(standRowsFromList(rows))
}

// Filter loads every stand matching the given (all-optional) filters.
func (r *StandRepository) Filter(ctx context.Context, filters ports.StandFilters) ([]*powers.Stand, error) {
	rows, err := r.queries.FilterStandRows(ctx, db.FilterStandRowsParams{
		Rarity:          enumStrPtr[enums.PowerRarity, db.PowerRarity](filters.Rarity),
		AttackPower:     enumStrPtr[enums.StandStat, db.StandStat](filters.AttackPower),
		Speed:           enumStrPtr[enums.StandStat, db.StandStat](filters.Speed),
		AttackRange:     enumStrPtr[enums.StandStat, db.StandStat](filters.AttackRange),
		Endurance:       enumStrPtr[enums.StandStat, db.StandStat](filters.Endurance),
		Precision:       enumStrPtr[enums.StandStat, db.StandStat](filters.Precision),
		Potential:       enumStrPtr[enums.StandStat, db.StandStat](filters.Potential),
		EvolvesFromName: filters.EvolvesFrom,
	})
	if err != nil {
		return nil, fmt.Errorf("filtering stands: %w", err)
	}
	return buildStands(standRowsFromFilter(rows))
}

// enumStrPtr converts an optional domain enum (which knows how to render
// itself via String()) into an optional sqlc-generated string-based enum
// type, for use as a nullable query parameter.
func enumStrPtr[D fmt.Stringer, G ~string](v *D) *G {
	if v == nil {
		return nil
	}
	g := G((*v).String())
	return &g
}
