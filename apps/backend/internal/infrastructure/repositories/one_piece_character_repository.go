package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// OnePieceCharacterRepository is the pgx/sqlc adapter for
// ports.IOnePieceCharacterRepository - same shape as JojoCharacterRepository.
type OnePieceCharacterRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

var _ ports.IOnePieceCharacterRepository = (*OnePieceCharacterRepository)(nil)

func NewOnePieceCharacterRepository(pool *pgxpool.Pool) *OnePieceCharacterRepository {
	return &OnePieceCharacterRepository{pool: pool, queries: db.New(pool)}
}

func (r *OnePieceCharacterRepository) Save(ctx context.Context, c *characters.OnePieceCharacter, translations ports.CharacterTranslations) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			fmt.Printf("rolling back transaction for one piece character %q: %v\n", c.Name(), rollbackErr)
		}
	}()

	q := r.queries.WithTx(tx)

	id, err := q.UpsertCharacter(ctx, db.UpsertCharacterParams{
		ID:            pgtype.UUID{Bytes: c.ID(), Valid: true},
		Manga:         enums.OnePiece.String(),
		Name:          c.Name(),
		Rarity:        c.Rarity().String(),
		Picture:       c.Picture(),
		PictureThumb:  c.PictureThumb(),
		PictureCard:   c.PictureCard(),
		PictureStatus: c.PictureStatus().String(),
		PictureLqip:   c.PictureLqip(),
	})
	if err != nil {
		return fmt.Errorf("upserting character %q: %w", c.Name(), wrapPgError(err, ports.ErrOnePieceCharacterAlreadyExists))
	}

	if err := saveCharacterTranslations(ctx, q, id, translations); err != nil {
		return fmt.Errorf("saving translations for %q: %w", c.Name(), err)
	}

	if err := q.UpsertOnePieceCharacter(ctx, db.UpsertOnePieceCharacterParams{
		ID:              id,
		PhysicalForm:    c.PhysicalForm().String(),
		ArmamentHaki:    c.ArmamentHaki().String(),
		ObservationHaki: c.ObservationHaki().String(),
		ConquerorHaki:   c.ConquerorHaki().String(),
		FruitMastery:    c.FruitMastery().String(),
	}); err != nil {
		return fmt.Errorf("upserting one piece character %q: %w", c.Name(), wrapPgError(err, ports.ErrOnePieceCharacterAlreadyExists))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing one piece character %q: %w", c.Name(), err)
	}
	return nil
}

func (r *OnePieceCharacterRepository) FindByID(ctx context.Context, id characters.CharacterID, locale enums.Locale) (*characters.OnePieceCharacter, error) {
	row, err := r.queries.GetOnePieceCharacterRowByID(ctx, db.GetOnePieceCharacterRowByIDParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		Locales: fallbackStrings(locale),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ports.ErrOnePieceCharacterNotFound, id)
		}
		return nil, fmt.Errorf("querying one piece character %s: %w", id, err)
	}
	return buildOnePieceCharacter(onePieceCharacterRowFromGetByID(row))
}

func (r *OnePieceCharacterRepository) FindByName(ctx context.Context, name string, locale enums.Locale) (*characters.OnePieceCharacter, error) {
	row, err := r.queries.GetOnePieceCharacterRowByName(ctx, db.GetOnePieceCharacterRowByNameParams{
		Name:    name,
		Locales: fallbackStrings(locale),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %q", ports.ErrOnePieceCharacterNotFound, name)
		}
		return nil, fmt.Errorf("querying one piece character %q: %w", name, err)
	}
	return buildOnePieceCharacter(onePieceCharacterRowFromGetByName(row))
}

func (r *OnePieceCharacterRepository) GetAll(ctx context.Context, locale enums.Locale) ([]*characters.OnePieceCharacter, error) {
	rows, err := r.queries.ListOnePieceCharacterRows(ctx, fallbackStrings(locale))
	if err != nil {
		return nil, fmt.Errorf("listing one piece characters: %w", err)
	}
	return buildOnePieceCharacters(onePieceCharacterRowsFromList(rows))
}

func (r *OnePieceCharacterRepository) Filter(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) ([]*characters.OnePieceCharacter, error) {
	rows, err := r.queries.FilterOnePieceCharacterRows(ctx, db.FilterOnePieceCharacterRowsParams{
		Rarity:          enumStrPtr[enums.PowerRarity, db.PowerRarity](filters.Rarity),
		PhysicalForm:    enumStrPtr[enums.PhysicalForm, db.PhysicalForm](filters.PhysicalForm),
		ArmamentHaki:    enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ArmamentHaki),
		ObservationHaki: enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ObservationHaki),
		ConquerorHaki:   enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ConquerorHaki),
		FruitMastery:    enumStrPtr[enums.FruitMastery, db.FruitMastery](filters.FruitMastery),
		Search:          searchPtr(filters.Search),
		Locales:         fallbackStrings(locale),
	})
	if err != nil {
		return nil, fmt.Errorf("filtering one piece characters: %w", err)
	}
	return buildOnePieceCharacters(onePieceCharacterRowsFromFilter(rows))
}

func (r *OnePieceCharacterRepository) Page(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.OnePieceCharacter, bool, error) {
	rows, err := r.queries.PageOnePieceCharacterRows(ctx, db.PageOnePieceCharacterRowsParams{
		Rarity:          enumStrPtr[enums.PowerRarity, db.PowerRarity](filters.Rarity),
		PhysicalForm:    enumStrPtr[enums.PhysicalForm, db.PhysicalForm](filters.PhysicalForm),
		ArmamentHaki:    enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ArmamentHaki),
		ObservationHaki: enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ObservationHaki),
		ConquerorHaki:   enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ConquerorHaki),
		FruitMastery:    enumStrPtr[enums.FruitMastery, db.FruitMastery](filters.FruitMastery),
		Search:          searchPtr(filters.Search),
		Locales:         fallbackStrings(locale),
		AfterName:       afterName,
		PageLimit:       int32(limit + 1),
	})
	if err != nil {
		return nil, false, fmt.Errorf("paging one piece characters: %w", err)
	}

	list, err := buildOnePieceCharacters(onePieceCharacterRowsFromPage(rows))
	if err != nil {
		return nil, false, err
	}
	hasMore := len(list) > limit
	if hasMore {
		list = list[:limit]
	}
	return list, hasMore, nil
}

func (r *OnePieceCharacterRepository) Count(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) (int, error) {
	count, err := r.queries.CountOnePieceCharacterRows(ctx, db.CountOnePieceCharacterRowsParams{
		Rarity:          enumStrPtr[enums.PowerRarity, db.PowerRarity](filters.Rarity),
		PhysicalForm:    enumStrPtr[enums.PhysicalForm, db.PhysicalForm](filters.PhysicalForm),
		ArmamentHaki:    enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ArmamentHaki),
		ObservationHaki: enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ObservationHaki),
		ConquerorHaki:   enumStrPtr[enums.HakiLevel, db.HakiLevel](filters.ConquerorHaki),
		FruitMastery:    enumStrPtr[enums.FruitMastery, db.FruitMastery](filters.FruitMastery),
		Search:          searchPtr(filters.Search),
		Locales:         fallbackStrings(locale),
	})
	if err != nil {
		return 0, fmt.Errorf("counting one piece characters: %w", err)
	}
	return int(count), nil
}

func (r *OnePieceCharacterRepository) UpdatePicture(ctx context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	err := r.queries.UpdateCharacterPicture(ctx, db.UpdateCharacterPictureParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		Picture:       main,
		PictureThumb:  thumb,
		PictureCard:   card,
		PictureStatus: status.String(),
		PictureLqip:   lqip,
	})
	if err != nil {
		return fmt.Errorf("updating picture for one piece character %s: %w", id, err)
	}
	return nil
}

func (r *OnePieceCharacterRepository) SetMediaID(ctx context.Context, id characters.CharacterID, mediaID string) error {
	if err := r.queries.UpdateCharacterMediaID(ctx, db.UpdateCharacterMediaIDParams{
		PictureMediaID: mediaID,
		ID:             pgtype.UUID{Bytes: id, Valid: true},
	}); err != nil {
		return fmt.Errorf("setting media id for one piece character %s: %w", id, err)
	}
	return nil
}

func (r *OnePieceCharacterRepository) Delete(ctx context.Context, id characters.CharacterID) error {
	rowsAffected, err := r.queries.DeleteCharacterByID(ctx, db.DeleteCharacterByIDParams{
		ID:    pgtype.UUID{Bytes: id, Valid: true},
		Manga: enums.OnePiece.String(),
	})
	if err != nil {
		return fmt.Errorf("deleting one piece character %s: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", ports.ErrOnePieceCharacterNotFound, id)
	}
	return nil
}

func (r *OnePieceCharacterRepository) Translations(ctx context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	rows, err := r.queries.GetCharacterTranslations(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("querying translations for one piece character %s: %w", id, err)
	}
	return characterTranslationsFromRows(rows)
}
