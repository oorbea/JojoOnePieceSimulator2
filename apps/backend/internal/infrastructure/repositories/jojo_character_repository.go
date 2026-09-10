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

// JojoCharacterRepository is the pgx/sqlc adapter for
// ports.IJojoCharacterRepository. JoJo characters are stored using class
// table inheritance: a base `characters` row plus a `jojo_characters` row
// sharing the same id - same shape as DevilFruitRepository.
type JojoCharacterRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

var _ ports.IJojoCharacterRepository = (*JojoCharacterRepository)(nil)

func NewJojoCharacterRepository(pool *pgxpool.Pool) *JojoCharacterRepository {
	return &JojoCharacterRepository{pool: pool, queries: db.New(pool)}
}

// Save upserts c's characters/jojo_characters rows, then replaces its
// character_translations rows with translations wholesale - same
// convention as DevilFruitRepository.Save.
func (r *JojoCharacterRepository) Save(ctx context.Context, c *characters.JojoCharacter, translations ports.CharacterTranslations) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			fmt.Printf("rolling back transaction for jojo character %q: %v\n", c.Name(), rollbackErr)
		}
	}()

	q := r.queries.WithTx(tx)

	id, err := q.UpsertCharacter(ctx, db.UpsertCharacterParams{
		ID:            pgtype.UUID{Bytes: c.ID(), Valid: true},
		Manga:         enums.Jojo.String(),
		Name:          c.Name(),
		Rarity:        c.Rarity().String(),
		Picture:       c.Picture(),
		PictureThumb:  c.PictureThumb(),
		PictureCard:   c.PictureCard(),
		PictureStatus: c.PictureStatus().String(),
		PictureLqip:   c.PictureLqip(),
	})
	if err != nil {
		return fmt.Errorf("upserting character %q: %w", c.Name(), wrapPgError(err, ports.ErrJojoCharacterAlreadyExists))
	}

	if err := saveCharacterTranslations(ctx, q, id, translations); err != nil {
		return fmt.Errorf("saving translations for %q: %w", c.Name(), err)
	}

	if err := q.UpsertJojoCharacter(ctx, db.UpsertJojoCharacterParams{
		ID:       id,
		Hamon:    c.Hamon().String(),
		Spin:     c.Spin().String(),
		BattleIq: int16(c.BattleIQ()),
	}); err != nil {
		return fmt.Errorf("upserting jojo character %q: %w", c.Name(), wrapPgError(err, ports.ErrJojoCharacterAlreadyExists))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing jojo character %q: %w", c.Name(), err)
	}
	return nil
}

func (r *JojoCharacterRepository) FindByID(ctx context.Context, id characters.CharacterID, locale enums.Locale) (*characters.JojoCharacter, error) {
	row, err := r.queries.GetJojoCharacterRowByID(ctx, db.GetJojoCharacterRowByIDParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		Locales: fallbackStrings(locale),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ports.ErrJojoCharacterNotFound, id)
		}
		return nil, fmt.Errorf("querying jojo character %s: %w", id, err)
	}
	return buildJojoCharacter(jojoCharacterRowFromGetByID(row))
}

func (r *JojoCharacterRepository) FindByName(ctx context.Context, name string, locale enums.Locale) (*characters.JojoCharacter, error) {
	row, err := r.queries.GetJojoCharacterRowByName(ctx, db.GetJojoCharacterRowByNameParams{
		Name:    name,
		Locales: fallbackStrings(locale),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %q", ports.ErrJojoCharacterNotFound, name)
		}
		return nil, fmt.Errorf("querying jojo character %q: %w", name, err)
	}
	return buildJojoCharacter(jojoCharacterRowFromGetByName(row))
}

func (r *JojoCharacterRepository) GetAll(ctx context.Context, locale enums.Locale) ([]*characters.JojoCharacter, error) {
	rows, err := r.queries.ListJojoCharacterRows(ctx, fallbackStrings(locale))
	if err != nil {
		return nil, fmt.Errorf("listing jojo characters: %w", err)
	}
	return buildJojoCharacters(jojoCharacterRowsFromList(rows))
}

func (r *JojoCharacterRepository) Filter(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale) ([]*characters.JojoCharacter, error) {
	rows, err := r.queries.FilterJojoCharacterRows(ctx, db.FilterJojoCharacterRowsParams{
		Rarity:   enumStrPtr[enums.PowerRarity, db.PowerRarity](filters.Rarity),
		Hamon:    enumStrPtr[enums.HamonLevel, db.HamonLevel](filters.Hamon),
		Spin:     enumStrPtr[enums.SpinLevel, db.SpinLevel](filters.Spin),
		BattleIq: battleIQPtr(filters.BattleIQ),
		Search:   searchPtr(filters.Search),
		Locales:  fallbackStrings(locale),
	})
	if err != nil {
		return nil, fmt.Errorf("filtering jojo characters: %w", err)
	}
	return buildJojoCharacters(jojoCharacterRowsFromFilter(rows))
}

func (r *JojoCharacterRepository) Page(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.JojoCharacter, bool, error) {
	rows, err := r.queries.PageJojoCharacterRows(ctx, db.PageJojoCharacterRowsParams{
		Rarity:    enumStrPtr[enums.PowerRarity, db.PowerRarity](filters.Rarity),
		Hamon:     enumStrPtr[enums.HamonLevel, db.HamonLevel](filters.Hamon),
		Spin:      enumStrPtr[enums.SpinLevel, db.SpinLevel](filters.Spin),
		BattleIq:  battleIQPtr(filters.BattleIQ),
		Search:    searchPtr(filters.Search),
		Locales:   fallbackStrings(locale),
		AfterName: afterName,
		PageLimit: int32(limit + 1),
	})
	if err != nil {
		return nil, false, fmt.Errorf("paging jojo characters: %w", err)
	}

	list, err := buildJojoCharacters(jojoCharacterRowsFromPage(rows))
	if err != nil {
		return nil, false, err
	}
	hasMore := len(list) > limit
	if hasMore {
		list = list[:limit]
	}
	return list, hasMore, nil
}

func (r *JojoCharacterRepository) Count(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale) (int, error) {
	count, err := r.queries.CountJojoCharacterRows(ctx, db.CountJojoCharacterRowsParams{
		Rarity:   enumStrPtr[enums.PowerRarity, db.PowerRarity](filters.Rarity),
		Hamon:    enumStrPtr[enums.HamonLevel, db.HamonLevel](filters.Hamon),
		Spin:     enumStrPtr[enums.SpinLevel, db.SpinLevel](filters.Spin),
		BattleIq: battleIQPtr(filters.BattleIQ),
		Search:   searchPtr(filters.Search),
		Locales:  fallbackStrings(locale),
	})
	if err != nil {
		return 0, fmt.Errorf("counting jojo characters: %w", err)
	}
	return int(count), nil
}

func (r *JojoCharacterRepository) UpdatePicture(ctx context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	err := r.queries.UpdateCharacterPicture(ctx, db.UpdateCharacterPictureParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		Picture:       main,
		PictureThumb:  thumb,
		PictureCard:   card,
		PictureStatus: status.String(),
		PictureLqip:   lqip,
	})
	if err != nil {
		return fmt.Errorf("updating picture for jojo character %s: %w", id, err)
	}
	return nil
}

func (r *JojoCharacterRepository) SetMediaID(ctx context.Context, id characters.CharacterID, mediaID string) error {
	if err := r.queries.UpdateCharacterMediaID(ctx, db.UpdateCharacterMediaIDParams{
		PictureMediaID: mediaID,
		ID:             pgtype.UUID{Bytes: id, Valid: true},
	}); err != nil {
		return fmt.Errorf("setting media id for jojo character %s: %w", id, err)
	}
	return nil
}

func (r *JojoCharacterRepository) Delete(ctx context.Context, id characters.CharacterID) error {
	rowsAffected, err := r.queries.DeleteCharacterByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return fmt.Errorf("deleting jojo character %s: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", ports.ErrJojoCharacterNotFound, id)
	}
	return nil
}

func (r *JojoCharacterRepository) Translations(ctx context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	rows, err := r.queries.GetCharacterTranslations(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("querying translations for jojo character %s: %w", id, err)
	}
	return characterTranslationsFromRows(rows)
}

// battleIQPtr converts an optional byte filter into the *int16 the
// generated smallint query param expects.
func battleIQPtr(v *byte) *int16 {
	if v == nil {
		return nil
	}
	iq := int16(*v)
	return &iq
}
