package cache

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// JojoCharacterRepository decorates a ports.IJojoCharacterRepository with a
// ports.ICache, mirroring devil_fruit_repository.go's
// read-through/write-invalidate shape.
type JojoCharacterRepository struct {
	next         ports.IJojoCharacterRepository
	cache        ports.ICache
	characterTTL time.Duration
	notFoundTTL  time.Duration
}

var _ ports.IJojoCharacterRepository = (*JojoCharacterRepository)(nil)

func NewJojoCharacterRepository(next ports.IJojoCharacterRepository, c ports.ICache, characterTTL, notFoundTTL time.Duration) *JojoCharacterRepository {
	return &JojoCharacterRepository{next: next, cache: c, characterTTL: characterTTL, notFoundTTL: notFoundTTL}
}

func (r *JojoCharacterRepository) Save(ctx context.Context, c *characters.JojoCharacter, translations ports.CharacterTranslations) error {
	if err := r.next.Save(ctx, c, translations); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *JojoCharacterRepository) FindByID(ctx context.Context, id characters.CharacterID, locale enums.Locale) (*characters.JojoCharacter, error) {
	key := idKey(id, locale)
	if data, ok := r.cache.Get(ctx, jojoCharactersNamespace, key); ok {
		if isTombstone(data) {
			return nil, ports.ErrJojoCharacterNotFound
		}
		if c, err := unmarshalJojoCharacter(data); err == nil {
			return c, nil
		}
	}

	c, err := r.next.FindByID(ctx, id, locale)
	if err != nil {
		if errors.Is(err, ports.ErrJojoCharacterNotFound) {
			r.cache.Set(ctx, jojoCharactersNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return nil, err
	}

	if data, err := marshalJojoCharacter(c); err == nil {
		r.cache.Set(ctx, jojoCharactersNamespace, key, data, r.characterTTL)
	}
	return c, nil
}

func (r *JojoCharacterRepository) FindByName(ctx context.Context, name string, locale enums.Locale) (*characters.JojoCharacter, error) {
	key := nameKey(name, locale)
	if data, ok := r.cache.Get(ctx, jojoCharactersNamespace, key); ok {
		if isTombstone(data) {
			return nil, ports.ErrJojoCharacterNotFound
		}
		if c, err := unmarshalJojoCharacter(data); err == nil {
			return c, nil
		}
	}

	c, err := r.next.FindByName(ctx, name, locale)
	if err != nil {
		if errors.Is(err, ports.ErrJojoCharacterNotFound) {
			r.cache.Set(ctx, jojoCharactersNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return nil, err
	}

	if data, err := marshalJojoCharacter(c); err == nil {
		r.cache.Set(ctx, jojoCharactersNamespace, key, data, r.characterTTL)
	}
	return c, nil
}

func (r *JojoCharacterRepository) GetAll(ctx context.Context, locale enums.Locale) ([]*characters.JojoCharacter, error) {
	key := allKey(locale)
	if data, ok := r.cache.Get(ctx, jojoCharactersNamespace, key); ok {
		if list, err := unmarshalJojoCharacters(data); err == nil {
			return list, nil
		}
	}

	list, err := r.next.GetAll(ctx, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalJojoCharacters(list); err == nil {
		r.cache.Set(ctx, jojoCharactersNamespace, key, data, r.characterTTL)
	}
	return list, nil
}

func (r *JojoCharacterRepository) Filter(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale) ([]*characters.JojoCharacter, error) {
	key := jojoCharacterFilterKey(filters, locale)
	if data, ok := r.cache.Get(ctx, jojoCharactersNamespace, key); ok {
		if list, err := unmarshalJojoCharacters(data); err == nil {
			return list, nil
		}
	}

	list, err := r.next.Filter(ctx, filters, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalJojoCharacters(list); err == nil {
		r.cache.Set(ctx, jojoCharactersNamespace, key, data, r.characterTTL)
	}
	return list, nil
}

// Page is a pass-through, deliberately not cached - same reasoning as
// StandRepository.Page (ObsidianVault/catalogue-pagination.md).
func (r *JojoCharacterRepository) Page(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.JojoCharacter, bool, error) {
	return r.next.Page(ctx, filters, locale, afterName, limit)
}

func (r *JojoCharacterRepository) Count(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale) (int, error) {
	return r.next.Count(ctx, filters, locale)
}

// Translations bypasses the cache - same reasoning as
// DevilFruitRepository.Translations.
func (r *JojoCharacterRepository) Translations(ctx context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	return r.next.Translations(ctx, id)
}

func (r *JojoCharacterRepository) Delete(ctx context.Context, id characters.CharacterID) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *JojoCharacterRepository) UpdatePicture(ctx context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	if err := r.next.UpdatePicture(ctx, id, main, thumb, card, lqip, status); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *JojoCharacterRepository) SetMediaID(ctx context.Context, id characters.CharacterID, mediaID string) error {
	if err := r.next.SetMediaID(ctx, id, mediaID); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *JojoCharacterRepository) invalidate(ctx context.Context) {
	if err := r.cache.Invalidate(ctx, jojoCharactersNamespace); err != nil {
		log.Printf("cache: invalidating jojo characters namespace: %v", err)
	}
}
