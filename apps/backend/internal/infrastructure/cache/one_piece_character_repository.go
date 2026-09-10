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

// OnePieceCharacterRepository decorates a ports.IOnePieceCharacterRepository
// with a ports.ICache - same shape as JojoCharacterRepository.
type OnePieceCharacterRepository struct {
	next         ports.IOnePieceCharacterRepository
	cache        ports.ICache
	characterTTL time.Duration
	notFoundTTL  time.Duration
}

var _ ports.IOnePieceCharacterRepository = (*OnePieceCharacterRepository)(nil)

func NewOnePieceCharacterRepository(next ports.IOnePieceCharacterRepository, c ports.ICache, characterTTL, notFoundTTL time.Duration) *OnePieceCharacterRepository {
	return &OnePieceCharacterRepository{next: next, cache: c, characterTTL: characterTTL, notFoundTTL: notFoundTTL}
}

func (r *OnePieceCharacterRepository) Save(ctx context.Context, c *characters.OnePieceCharacter, translations ports.CharacterTranslations) error {
	if err := r.next.Save(ctx, c, translations); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *OnePieceCharacterRepository) FindByID(ctx context.Context, id characters.CharacterID, locale enums.Locale) (*characters.OnePieceCharacter, error) {
	key := idKey(id, locale)
	if data, ok := r.cache.Get(ctx, onePieceCharactersNamespace, key); ok {
		if isTombstone(data) {
			return nil, ports.ErrOnePieceCharacterNotFound
		}
		if c, err := unmarshalOnePieceCharacter(data); err == nil {
			return c, nil
		}
	}

	c, err := r.next.FindByID(ctx, id, locale)
	if err != nil {
		if errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
			r.cache.Set(ctx, onePieceCharactersNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return nil, err
	}

	if data, err := marshalOnePieceCharacter(c); err == nil {
		r.cache.Set(ctx, onePieceCharactersNamespace, key, data, r.characterTTL)
	}
	return c, nil
}

func (r *OnePieceCharacterRepository) FindByName(ctx context.Context, name string, locale enums.Locale) (*characters.OnePieceCharacter, error) {
	key := nameKey(name, locale)
	if data, ok := r.cache.Get(ctx, onePieceCharactersNamespace, key); ok {
		if isTombstone(data) {
			return nil, ports.ErrOnePieceCharacterNotFound
		}
		if c, err := unmarshalOnePieceCharacter(data); err == nil {
			return c, nil
		}
	}

	c, err := r.next.FindByName(ctx, name, locale)
	if err != nil {
		if errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
			r.cache.Set(ctx, onePieceCharactersNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return nil, err
	}

	if data, err := marshalOnePieceCharacter(c); err == nil {
		r.cache.Set(ctx, onePieceCharactersNamespace, key, data, r.characterTTL)
	}
	return c, nil
}

func (r *OnePieceCharacterRepository) GetAll(ctx context.Context, locale enums.Locale) ([]*characters.OnePieceCharacter, error) {
	key := allKey(locale)
	if data, ok := r.cache.Get(ctx, onePieceCharactersNamespace, key); ok {
		if list, err := unmarshalOnePieceCharacters(data); err == nil {
			return list, nil
		}
	}

	list, err := r.next.GetAll(ctx, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalOnePieceCharacters(list); err == nil {
		r.cache.Set(ctx, onePieceCharactersNamespace, key, data, r.characterTTL)
	}
	return list, nil
}

func (r *OnePieceCharacterRepository) Filter(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) ([]*characters.OnePieceCharacter, error) {
	key := onePieceCharacterFilterKey(filters, locale)
	if data, ok := r.cache.Get(ctx, onePieceCharactersNamespace, key); ok {
		if list, err := unmarshalOnePieceCharacters(data); err == nil {
			return list, nil
		}
	}

	list, err := r.next.Filter(ctx, filters, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalOnePieceCharacters(list); err == nil {
		r.cache.Set(ctx, onePieceCharactersNamespace, key, data, r.characterTTL)
	}
	return list, nil
}

func (r *OnePieceCharacterRepository) Page(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.OnePieceCharacter, bool, error) {
	return r.next.Page(ctx, filters, locale, afterName, limit)
}

func (r *OnePieceCharacterRepository) Count(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) (int, error) {
	return r.next.Count(ctx, filters, locale)
}

func (r *OnePieceCharacterRepository) Translations(ctx context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	return r.next.Translations(ctx, id)
}

func (r *OnePieceCharacterRepository) Delete(ctx context.Context, id characters.CharacterID) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *OnePieceCharacterRepository) UpdatePicture(ctx context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	if err := r.next.UpdatePicture(ctx, id, main, thumb, card, lqip, status); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *OnePieceCharacterRepository) SetMediaID(ctx context.Context, id characters.CharacterID, mediaID string) error {
	if err := r.next.SetMediaID(ctx, id, mediaID); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *OnePieceCharacterRepository) invalidate(ctx context.Context) {
	if err := r.cache.Invalidate(ctx, onePieceCharactersNamespace); err != nil {
		log.Printf("cache: invalidating one piece characters namespace: %v", err)
	}
}
