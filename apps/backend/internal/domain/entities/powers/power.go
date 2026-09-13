package powers

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type Power struct {
	name           string
	description    string
	picture        string
	pictureThumb   string
	pictureCard    string
	pictureLqip    string
	pictureMediaID string
	skills         []string
	id             PowerID
	rarity         enums.PowerRarity
	pictureStatus  enums.PictureStatus
	focalX         float64
	focalY         float64
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
		focalX:      0.5,
		focalY:      0.5,
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

// PictureMediaID returns the content-addressed group id media_objects rows
// for this Power's renditions are keyed by, or "" if it hasn't been
// backfilled/transcoded under the media proxy yet - callers must fall back
// to the presign path in that case.
func (p Power) PictureMediaID() string {
	return p.pictureMediaID
}

// SetMediaID records the content-addressed group id computed for this
// Power's current renditions. Separate from SetPictureRenditions because
// the group id is only known after the worker has both transcoded the
// image AND persisted the resulting media_objects rows - a step that
// happens after (and independently of) the rendition keys themselves.
func (p *Power) SetMediaID(mediaID string) {
	p.pictureMediaID = mediaID
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

// FocalX and FocalY are the normalized (0..1) point of a Power's picture a
// client should keep centered when rendering it with contentFit:'cover' -
// see ObsidianVault/media-proxy-content-addressed.md. No server-side crop:
// this is purely a value clients read through picture-source.ts.
func (p Power) FocalX() float64 { return p.focalX }
func (p Power) FocalY() float64 { return p.focalY }

// SetFocalPoint validates and stores a new focal point. Both coordinates
// must be finite numbers in [0, 1] - out-of-range or NaN/Inf is rejected
// rather than clamped, so a bad client value surfaces as a 400 instead of
// silently landing on an edge.
func (p *Power) SetFocalPoint(x, y float64) error {
	if !isValidFocalCoordinate(x) || !isValidFocalCoordinate(y) {
		return errors.New("focal point must be between 0 and 1")
	}
	p.focalX = x
	p.focalY = y
	return nil
}

func isValidFocalCoordinate(v float64) bool {
	return v >= 0 && v <= 1
}
