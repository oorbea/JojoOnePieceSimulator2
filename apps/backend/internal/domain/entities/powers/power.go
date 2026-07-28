package powers

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type Power struct {
	id            PowerID
	name          string
	description   string
	rarity        enums.PowerRarity
	skills        []string
	picture       string
	pictureThumb  string
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

// SetPictureRenditions replaces the stored main and thumbnail picture keys
// together with the pipeline status that produced them, so the three always
// change as one unit (e.g. a worker moving a Power to READY sets all three
// at once; PENDING/FAILED leave main/thumb untouched).
func (p *Power) SetPictureRenditions(main, thumb string, status enums.PictureStatus) {
	p.picture = main
	p.pictureThumb = thumb
	p.pictureStatus = status
}
