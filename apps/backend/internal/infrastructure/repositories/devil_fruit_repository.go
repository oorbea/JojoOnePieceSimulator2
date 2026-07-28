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

// DevilFruitRepository is the pgx/sqlc adapter for ports.IDevilFruitRepository.
// DevilFruits are stored using class table inheritance: a base `powers` row
// plus a `devil_fruits` row sharing the same id.
type DevilFruitRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

var _ ports.IDevilFruitRepository = (*DevilFruitRepository)(nil)

func NewDevilFruitRepository(pool *pgxpool.Pool) *DevilFruitRepository {
	return &DevilFruitRepository{pool: pool, queries: db.New(pool)}
}

// Save upserts fruit by name: the underlying powers/devil_fruits rows, and
// its skills, are fully replaced. It is safe to call repeatedly for the same
// devil fruit (e.g. re-running seed data).
func (r *DevilFruitRepository) Save(ctx context.Context, fruit *powers.DevilFruit) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			fmt.Printf("rolling back transaction for devil fruit %q: %v\n", fruit.Name(), rollbackErr)
		}
	}()

	q := r.queries.WithTx(tx)

	id, err := q.UpsertDevilFruitPower(ctx, db.UpsertDevilFruitPowerParams{
		ID:            pgtype.UUID{Bytes: fruit.ID(), Valid: true},
		Name:          fruit.Name(),
		Description:   fruit.Description(),
		Rarity:        fruit.Rarity().String(),
		Picture:       fruit.Picture(),
		PictureThumb:  fruit.PictureThumb(),
		PictureStatus: fruit.PictureStatus().String(),
	})
	if err != nil {
		return fmt.Errorf("upserting power %q: %w", fruit.Name(), wrapPgError(err, ports.ErrDevilFruitAlreadyExists))
	}

	if err := q.DeletePowerSkills(ctx, id); err != nil {
		return fmt.Errorf("clearing skills for %q: %w", fruit.Name(), wrapPgError(err, ports.ErrDevilFruitAlreadyExists))
	}
	for position, skill := range fruit.Skills() {
		if err := q.InsertPowerSkill(ctx, db.InsertPowerSkillParams{
			PowerID:  id,
			Position: int32(position),
			Skill:    skill,
		}); err != nil {
			return fmt.Errorf("inserting skill %q for %q: %w", skill, fruit.Name(), wrapPgError(err, ports.ErrDevilFruitAlreadyExists))
		}
	}

	if err := q.UpsertDevilFruit(ctx, db.UpsertDevilFruitParams{
		ID:        id,
		FruitType: fruit.FruitType().String(),
	}); err != nil {
		return fmt.Errorf("upserting devil fruit %q: %w", fruit.Name(), wrapPgError(err, ports.ErrDevilFruitAlreadyExists))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing devil fruit %q: %w", fruit.Name(), err)
	}
	return nil
}

// FindByID loads the devil fruit with the given id.
func (r *DevilFruitRepository) FindByID(ctx context.Context, id powers.PowerID) (*powers.DevilFruit, error) {
	row, err := r.queries.GetDevilFruitRowByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ports.ErrDevilFruitNotFound, id)
		}
		return nil, fmt.Errorf("querying devil fruit %s: %w", id, err)
	}
	return buildDevilFruit(devilFruitRowFromGetByID(row))
}

// FindByName loads the devil fruit with the given name.
func (r *DevilFruitRepository) FindByName(ctx context.Context, name string) (*powers.DevilFruit, error) {
	row, err := r.queries.GetDevilFruitRowByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %q", ports.ErrDevilFruitNotFound, name)
		}
		return nil, fmt.Errorf("querying devil fruit %q: %w", name, err)
	}
	return buildDevilFruit(devilFruitRowFromGetByName(row))
}

// GetAll loads every devil fruit in the system.
func (r *DevilFruitRepository) GetAll(ctx context.Context) ([]*powers.DevilFruit, error) {
	rows, err := r.queries.ListDevilFruitRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing devil fruits: %w", err)
	}
	return buildDevilFruits(devilFruitRowsFromList(rows))
}

// Filter loads every devil fruit matching the given (all-optional) filters.
func (r *DevilFruitRepository) Filter(ctx context.Context, filters ports.DevilFruitFilters) ([]*powers.DevilFruit, error) {
	rows, err := r.queries.FilterDevilFruitRows(ctx, db.FilterDevilFruitRowsParams{
		Rarity:    enumStrPtr[enums.PowerRarity, db.PowerRarity](filters.Rarity),
		FruitType: enumStrPtr[enums.FruitType, db.FruitType](filters.FruitType),
	})
	if err != nil {
		return nil, fmt.Errorf("filtering devil fruits: %w", err)
	}
	return buildDevilFruits(devilFruitRowsFromFilter(rows))
}

// UpdatePicture updates only a devil fruit's picture renditions and pipeline
// status, leaving every other column untouched. A nil main or thumb leaves
// that column as-is.
func (r *DevilFruitRepository) UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error {
	err := r.queries.UpdatePowerPicture(ctx, db.UpdatePowerPictureParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		Picture:       main,
		PictureThumb:  thumb,
		PictureStatus: status.String(),
	})
	if err != nil {
		return fmt.Errorf("updating picture for devil fruit %s: %w", id, err)
	}
	return nil
}

// Delete removes the devil fruit (and its power/skills rows) with the given
// id.
func (r *DevilFruitRepository) Delete(ctx context.Context, id powers.PowerID) error {
	rowsAffected, err := r.queries.DeleteDevilFruitByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return fmt.Errorf("deleting devil fruit %s: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", ports.ErrDevilFruitNotFound, id)
	}
	return nil
}
