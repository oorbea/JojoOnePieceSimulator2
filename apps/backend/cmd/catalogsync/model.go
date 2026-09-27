// Command catalogsync is the repeatable, re-runnable tool behind
// db/migrations/00017_seed_catalog_*.sql (see ObsidianVault/
// catalog-seed-from-prod.md). Admins keep filling in Stands/Devil Fruits/
// Characters by hand in prod, in whatever locale they happen to be
// comfortable in; this tool reads prod read-only, works out which
// translations are missing/placeholder/stale against a versioned overlay
// file, and - once that overlay is filled in - generates a goose migration
// that upserts the missing translations everywhere and, in local dev only,
// the whole catalogue (rows + image keys, never image bytes: local and prod
// share one R2 bucket).
//
// Subcommands (see each file's doc for detail):
//   - dump:     prod (read-only) -> db/seed/catalog/snapshot.json
//   - plan:     snapshot.json + translations.json -> pending.json
//   - generate: snapshot.json + translations.json -> a new goose migration
//
// Nothing here touches the local database - applying the generated
// migration is goose's job (make migrate-up), same as any other migration.
package main

import "time"

// seedDir holds every file this tool reads/writes, relative to the backend
// module root (cmd/catalogsync runs with that as its working directory,
// same as every other cmd/*).
const seedDir = "db/seed/catalog"

// locales lists every supported content locale, in the fixed order every
// generated migration and JSON file uses.
var locales = []string{"en-GB", "es-ES", "ca-ES"}

// Translation is one locale's power/character text.
type Translation struct {
	Description string   `json:"description"`
	Skills      []string `json:"skills,omitempty"` // powers only; nil for characters
}

// StandRow is one prod stands+powers row.
type StandRow struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Rarity         string                 `json:"rarity"`
	Picture        string                 `json:"picture"`
	PictureThumb   string                 `json:"picture_thumb"`
	PictureCard    string                 `json:"picture_card"`
	PictureLQIP    string                 `json:"picture_lqip"`
	PictureMediaID string                 `json:"picture_media_id"`
	PictureStatus  string                 `json:"picture_status"`
	FocalX         float64                `json:"focal_x"`
	FocalY         float64                `json:"focal_y"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	AttackPower    string                 `json:"attack_power"`
	Speed          string                 `json:"speed"`
	AttackRange    string                 `json:"attack_range"`
	Endurance      string                 `json:"endurance"`
	Precision      string                 `json:"precision"`
	Potential      string                 `json:"potential"`
	EvolvesFromID  *string                `json:"evolves_from_id,omitempty"`
	Translations   map[string]Translation `json:"translations"`
}

// DevilFruitRow is one prod devil_fruits+powers row.
type DevilFruitRow struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Rarity         string                 `json:"rarity"`
	Picture        string                 `json:"picture"`
	PictureThumb   string                 `json:"picture_thumb"`
	PictureCard    string                 `json:"picture_card"`
	PictureLQIP    string                 `json:"picture_lqip"`
	PictureMediaID string                 `json:"picture_media_id"`
	PictureStatus  string                 `json:"picture_status"`
	FocalX         float64                `json:"focal_x"`
	FocalY         float64                `json:"focal_y"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	FruitType      string                 `json:"fruit_type"`
	Translations   map[string]Translation `json:"translations"`
}

// JojoCharacterRow is one prod characters+jojo_characters row.
type JojoCharacterRow struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Rarity         string                 `json:"rarity"`
	Picture        string                 `json:"picture"`
	PictureThumb   string                 `json:"picture_thumb"`
	PictureCard    string                 `json:"picture_card"`
	PictureLQIP    string                 `json:"picture_lqip"`
	PictureMediaID string                 `json:"picture_media_id"`
	PictureStatus  string                 `json:"picture_status"`
	FocalX         float64                `json:"focal_x"`
	FocalY         float64                `json:"focal_y"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Hamon          string                 `json:"hamon"`
	Spin           string                 `json:"spin"`
	BattleIQ       int                    `json:"battle_iq"`
	Translations   map[string]Translation `json:"translations"`
}

// OnePieceCharacterRow is one prod characters+one_piece_characters row.
type OnePieceCharacterRow struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Rarity          string                 `json:"rarity"`
	Picture         string                 `json:"picture"`
	PictureThumb    string                 `json:"picture_thumb"`
	PictureCard     string                 `json:"picture_card"`
	PictureLQIP     string                 `json:"picture_lqip"`
	PictureMediaID  string                 `json:"picture_media_id"`
	PictureStatus   string                 `json:"picture_status"`
	FocalX          float64                `json:"focal_x"`
	FocalY          float64                `json:"focal_y"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	PhysicalForm    string                 `json:"physical_form"`
	ArmamentHaki    string                 `json:"armament_haki"`
	ObservationHaki string                 `json:"observation_haki"`
	ConquerorHaki   string                 `json:"conqueror_haki"`
	FruitMastery    string                 `json:"fruit_mastery"`
	Translations    map[string]Translation `json:"translations"`
}

// MediaObjectRow mirrors one media_objects row for a picture_media_id the
// dumped rows reference.
type MediaObjectRow struct {
	GroupID     string `json:"group_id"`
	Variant     string `json:"variant"`
	StorageKey  string `json:"storage_key"`
	ContentType string `json:"content_type"`
	Bytes       int64  `json:"bytes"`
	Scope       string `json:"scope"`
}

// StorageObjectRow mirrors one storage_objects row for a picture key the
// dumped rows reference, recording which provider holds it.
type StorageObjectRow struct {
	Key      string `json:"key"`
	Provider string `json:"provider"`
	Bytes    int64  `json:"bytes"`
}

// Snapshot is the full prod dump, written by `dump` and read by `plan` and
// `generate`.
type Snapshot struct {
	DumpedAt           time.Time              `json:"dumped_at"`
	Stands             []StandRow             `json:"stands"`
	DevilFruits        []DevilFruitRow        `json:"devil_fruits"`
	JojoCharacters     []JojoCharacterRow     `json:"jojo_characters"`
	OnePieceCharacters []OnePieceCharacterRow `json:"one_piece_characters"`
	MediaObjects       []MediaObjectRow       `json:"media_objects"`
	StorageObjects     []StorageObjectRow     `json:"storage_objects"`
}

// OverlayTranslation is one hand-authored (or Claude-authored, reviewed in
// the diff) translation entry in translations.json. SourceHash pins it to
// the exact source-locale text it was translated from - see plan's "stale"
// reason.
type OverlayTranslation struct {
	Description string   `json:"description"`
	Skills      []string `json:"skills,omitempty"`
	SourceHash  string   `json:"source_hash"`
}

// Overlay is db/seed/catalog/translations.json: every non-source-locale
// translation this tool has generated so far, keyed by kind, then by the
// row's natural name, then by target locale.
type Overlay struct {
	Stands             map[string]map[string]OverlayTranslation `json:"stands"`
	DevilFruits        map[string]map[string]OverlayTranslation `json:"devil_fruits"`
	JojoCharacters     map[string]map[string]OverlayTranslation `json:"jojo_characters"`
	OnePieceCharacters map[string]map[string]OverlayTranslation `json:"one_piece_characters"`
}

// Sources is db/seed/catalog/sources.json: the confirmed source locale for
// every character row (stands are always es-ES, devil fruits always en-GB -
// see the vault note - so only characters need this).
type Sources struct {
	JojoCharacters     map[string]string `json:"jojo_characters"`
	OnePieceCharacters map[string]string `json:"one_piece_characters"`
}

// PendingEntry is one translation `plan` found missing/placeholder/stale.
type PendingEntry struct {
	Kind         string   `json:"kind"` // stand | devil_fruit | jojo_character | one_piece_character
	Name         string   `json:"name"`
	TargetLocale string   `json:"target_locale"`
	SourceLocale string   `json:"source_locale"`
	SourceText   string   `json:"source_text"`
	SourceSkills []string `json:"source_skills,omitempty"`
	Reason       string   `json:"reason"` // missing | placeholder | stale
	SourceHash   string   `json:"source_hash"`
}
