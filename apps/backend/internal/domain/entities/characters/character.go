package characters

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// Character is the base of a class-table-inheritance hierarchy separate
// from powers.Power: a Character (a JoJo or One Piece character an admin
// authors for the catalogue) is not a Power and never enters the random
// loadout draw, PoolFilter, or a banlist - it is content, not gameplay
// state. JojoCharacter/OnePieceCharacter embed it, mirroring how
// powers.Stand/DevilFruit embed powers.Power.
//
// Unlike Power, Character has no Skills - the owner's decision for this
// content type - and, like game.Stage, Description is a single already
// locale-resolved string rather than a translation map; the picture
// pipeline fields mirror game.Stage's exactly.
type Character struct {
	name           string
	description    string
	picture        string
	pictureThumb   string
	pictureCard    string
	pictureLqip    string
	pictureMediaID string
	id             CharacterID
	manga          enums.Manga
	rarity         enums.PowerRarity
	pictureStatus  enums.PictureStatus
}

// NewCharacter validates and builds a Character. picture is the only
// picture field accepted here - thumb/card/lqip/status start empty/NONE,
// same convention as powers.NewPower/game.NewStage; set them together via
// SetPictureRenditions. rarity is stored and surfaced but, per the owner's
// decision for this phase, is not wired into any draw - there is no
// gachapon yet.
func NewCharacter(id CharacterID, manga enums.Manga, name string, rarity enums.PowerRarity, description string, picture string) (Character, error) {
	if id.IsNil() {
		return Character{}, errors.New("id is required")
	}
	if !manga.IsValid() {
		return Character{}, enums.ErrInvalidManga
	}
	if name == "" {
		return Character{}, errors.New("name is required")
	}
	if !rarity.IsValid() {
		return Character{}, enums.ErrInvalidRarity
	}
	if description == "" {
		return Character{}, errors.New("description is required")
	}
	return Character{id: id, manga: manga, name: name, rarity: rarity, description: description, picture: picture}, nil
}

func (c Character) ID() CharacterID           { return c.id }
func (c Character) Manga() enums.Manga        { return c.manga }
func (c Character) Name() string              { return c.name }
func (c Character) Rarity() enums.PowerRarity { return c.rarity }
func (c Character) Description() string       { return c.description }
func (c Character) Picture() string           { return c.picture }
func (c Character) PictureThumb() string      { return c.pictureThumb }
func (c Character) PictureCard() string       { return c.pictureCard }
func (c Character) PictureLqip() string       { return c.pictureLqip }

// PictureMediaID returns the content-addressed group id media_objects rows
// for this Character's renditions are keyed by, or "" if not backfilled yet.
func (c Character) PictureMediaID() string { return c.pictureMediaID }

// SetMediaID records the content-addressed group id - see
// powers.Power.SetMediaID for why this is separate from
// SetPictureRenditions.
func (c *Character) SetMediaID(mediaID string) { c.pictureMediaID = mediaID }

// PictureStatus reports where this Character's picture is in the async
// compression pipeline.
func (c Character) PictureStatus() enums.PictureStatus { return c.pictureStatus }

// SetPictureRenditions replaces the stored main/thumbnail/card picture keys
// and the LQIP placeholder together with the pipeline status that produced
// them, so they always change as one unit - same pattern as
// powers.Power.SetPictureRenditions/game.Stage.SetPictureRenditions.
func (c *Character) SetPictureRenditions(main, thumb, card, lqip string, status enums.PictureStatus) {
	c.picture = main
	c.pictureThumb = thumb
	c.pictureCard = card
	c.pictureLqip = lqip
	c.pictureStatus = status
}
