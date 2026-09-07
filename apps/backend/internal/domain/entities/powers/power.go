package powers

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type Power struct {
	name          string
	description   string
	picture       string
	pictureThumb  string
	pictureCard   string
	pictureLqip   string
	skills        []string
	id            PowerID
	rarity        enums.PowerRarity
	pictureStatus enums.PictureStatus
}

func NewPower(
	id PowerID,
	name string,
	description string,
	rarity enums.PowerRarity,
	skills *[]string,
	picture string,
) (*Power, error) {
	if id.IsNil() {
		return nil, errors.New("id is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if description == "" {
		return nil, errors.New("description is required")
	}
	if !rarity.IsValid() {
		return nil, enums.ErrInvalidRarity
	}
	if skills == nil || len(*skills) < 1 {
		return nil, errors.New("skills are required")
	}
	return &Power{
		id:          id,
		name:        name,
		description: description,
		rarity:      rarity,
		skills:      *skills,
		picture:     picture,
	}, nil
}

func (p Power) ID() PowerID {
	return p.id
}

func (p Power) Name() string {
	return p.name
}

func (p Power) Description() string {
	return p.description
}

func (p Power) Rarity() enums.PowerRarity {
	return p.rarity
}

func (p Power) Skills() []string {
	return p.skills
}

func (p Power) Picture() string {
	return p.picture
}

// PictureThumb returns the stored thumbnail rendition's key, or "" if none
// has been produced yet.
func (p Power) PictureThumb() string {
	return p.pictureThumb
}

// PictureCard returns the stored card rendition's key (128px, the size most
// catalogue grid cells render at), or "" if none has been produced yet.
func (p Power) PictureCard() string {
	return p.pictureCard
}

// PictureLqip returns the stored low-quality placeholder as a complete
// data: URI, or "" if none has been produced yet.
func (p Power) PictureLqip() string {
	return p.pictureLqip
}

// PictureStatus reports where this Power's picture is in the async
// compression pipeline.
func (p Power) PictureStatus() enums.PictureStatus {
	return p.pictureStatus
}

// SetPicture replaces the stored picture key. The empty string means "no
// picture", which NewPower already allows.
func (p *Power) SetPicture(picture string) {
	p.picture = picture
}

// SetPictureRenditions replaces the stored main/thumbnail/card picture keys
// and the LQIP placeholder together with the pipeline status that produced
// them, so they always change as one unit (e.g. a worker moving a Power to
// READY sets all of them at once; PENDING/FAILED leave the renditions
// untouched).
func (p *Power) SetPictureRenditions(main, thumb, card, lqip string, status enums.PictureStatus) {
	p.picture = main
	p.pictureThumb = thumb
	p.pictureCard = card
	p.pictureLqip = lqip
	p.pictureStatus = status
}
