package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
)

// runDump connects to prod (or whatever --source-dsn points at) with a
// READ ONLY transaction - this tool must never be able to write to the
// source database - and writes db/seed/catalog/snapshot.json. It talks SQL
// directly rather than going through the repositories package: every
// repository method resolves one locale with a fallback chain, and this
// needs every locale's row as stored, unresolved.
func runDump(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("dump", flag.ExitOnError)
	dsn := fs.String("source-dsn", os.Getenv("SOURCE_DATABASE_URL"), "prod (or other source) Postgres DSN; defaults to $SOURCE_DATABASE_URL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dsn == "" {
		return fmt.Errorf("--source-dsn or SOURCE_DATABASE_URL is required")
	}

	pool, err := postgres.NewPool(ctx, *dsn)
	if err != nil {
		return fmt.Errorf("connecting to source database: %w", err)
	}
	defer pool.Close()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Errorf("beginning read-only transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // read-only: a rollback is always safe and always correct here

	snap := Snapshot{DumpedAt: time.Now().UTC()}

	if snap.Stands, err = dumpStands(ctx, tx); err != nil {
		return err
	}
	if snap.DevilFruits, err = dumpDevilFruits(ctx, tx); err != nil {
		return err
	}
	if snap.JojoCharacters, err = dumpJojoCharacters(ctx, tx); err != nil {
		return err
	}
	if snap.OnePieceCharacters, err = dumpOnePieceCharacters(ctx, tx); err != nil {
		return err
	}

	keys, groupIDs := referencedStorage(snap)
	if snap.StorageObjects, err = dumpStorageObjects(ctx, tx, keys); err != nil {
		return err
	}
	if snap.MediaObjects, err = dumpMediaObjects(ctx, tx, groupIDs); err != nil {
		return err
	}

	if err := writeJSON(filepath.Join(seedDir, "snapshot.json"), snap); err != nil {
		return err
	}
	fmt.Printf("dumped %d stands, %d devil fruits, %d jojo characters, %d one piece characters (%d storage objects, %d media objects)\n",
		len(snap.Stands), len(snap.DevilFruits), len(snap.JojoCharacters), len(snap.OnePieceCharacters),
		len(snap.StorageObjects), len(snap.MediaObjects))
	return nil
}

// dumpStands reads every stand's core+picture+stat columns and its
// power_translations rows.
func dumpStands(ctx context.Context, tx pgx.Tx) ([]StandRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT p.id, p.name, p.rarity::text, p.picture, p.picture_thumb, p.picture_card,
		       p.picture_lqip, p.picture_media_id, p.picture_status::text, p.focal_x, p.focal_y,
		       p.created_at, p.updated_at,
		       s.attack_power::text, s.speed::text, s.attack_range::text, s.endurance::text,
		       s."precision"::text, s.potential::text, s.evolves_from_id
		FROM stands s JOIN powers p ON p.id = s.id
		ORDER BY p.name`)
	if err != nil {
		return nil, fmt.Errorf("querying stands: %w", err)
	}
	defer rows.Close()

	var out []StandRow
	for rows.Next() {
		var r StandRow
		var evolvesFrom *string
		if err := rows.Scan(&r.ID, &r.Name, &r.Rarity, &r.Picture, &r.PictureThumb, &r.PictureCard,
			&r.PictureLQIP, &r.PictureMediaID, &r.PictureStatus, &r.FocalX, &r.FocalY,
			&r.CreatedAt, &r.UpdatedAt,
			&r.AttackPower, &r.Speed, &r.AttackRange, &r.Endurance, &r.Precision, &r.Potential, &evolvesFrom); err != nil {
			return nil, fmt.Errorf("scanning stand row: %w", err)
		}
		r.EvolvesFromID = evolvesFrom
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	translations, err := dumpPowerTranslations(ctx, tx, "STAND")
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Translations = translations[out[i].ID]
	}
	return out, nil
}

func dumpDevilFruits(ctx context.Context, tx pgx.Tx) ([]DevilFruitRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT p.id, p.name, p.rarity::text, p.picture, p.picture_thumb, p.picture_card,
		       p.picture_lqip, p.picture_media_id, p.picture_status::text, p.focal_x, p.focal_y,
		       p.created_at, p.updated_at, d.fruit_type::text
		FROM devil_fruits d JOIN powers p ON p.id = d.id
		ORDER BY p.name`)
	if err != nil {
		return nil, fmt.Errorf("querying devil fruits: %w", err)
	}
	defer rows.Close()

	var out []DevilFruitRow
	for rows.Next() {
		var r DevilFruitRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Rarity, &r.Picture, &r.PictureThumb, &r.PictureCard,
			&r.PictureLQIP, &r.PictureMediaID, &r.PictureStatus, &r.FocalX, &r.FocalY,
			&r.CreatedAt, &r.UpdatedAt, &r.FruitType); err != nil {
			return nil, fmt.Errorf("scanning devil fruit row: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	translations, err := dumpPowerTranslations(ctx, tx, "DEVIL_FRUIT")
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Translations = translations[out[i].ID]
	}
	return out, nil
}

// dumpPowerTranslations returns every power_translations row for powers of
// kind, keyed by power id then locale.
func dumpPowerTranslations(ctx context.Context, tx pgx.Tx, kind string) (map[string]map[string]Translation, error) {
	rows, err := tx.Query(ctx, `
		SELECT t.power_id, t.locale::text, t.description, t.skills
		FROM power_translations t JOIN powers p ON p.id = t.power_id
		WHERE p.kind = $1`, kind)
	if err != nil {
		return nil, fmt.Errorf("querying power translations: %w", err)
	}
	defer rows.Close()

	out := make(map[string]map[string]Translation)
	for rows.Next() {
		var powerID, locale string
		var tr Translation
		if err := rows.Scan(&powerID, &locale, &tr.Description, &tr.Skills); err != nil {
			return nil, fmt.Errorf("scanning power translation: %w", err)
		}
		if out[powerID] == nil {
			out[powerID] = make(map[string]Translation)
		}
		out[powerID][locale] = tr
	}
	return out, rows.Err()
}

func dumpJojoCharacters(ctx context.Context, tx pgx.Tx) ([]JojoCharacterRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT c.id, c.name, c.rarity::text, c.picture, c.picture_thumb, c.picture_card,
		       c.picture_lqip, c.picture_media_id, c.picture_status::text, c.focal_x, c.focal_y,
		       c.created_at, c.updated_at, j.hamon::text, j.spin::text, j.battle_iq
		FROM jojo_characters j JOIN characters c ON c.id = j.id
		ORDER BY c.name`)
	if err != nil {
		return nil, fmt.Errorf("querying jojo characters: %w", err)
	}
	defer rows.Close()

	var out []JojoCharacterRow
	for rows.Next() {
		var r JojoCharacterRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Rarity, &r.Picture, &r.PictureThumb, &r.PictureCard,
			&r.PictureLQIP, &r.PictureMediaID, &r.PictureStatus, &r.FocalX, &r.FocalY,
			&r.CreatedAt, &r.UpdatedAt, &r.Hamon, &r.Spin, &r.BattleIQ); err != nil {
			return nil, fmt.Errorf("scanning jojo character row: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	translations, err := dumpCharacterTranslations(ctx, tx, "JOJO")
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Translations = translations[out[i].ID]
	}
	return out, nil
}

func dumpOnePieceCharacters(ctx context.Context, tx pgx.Tx) ([]OnePieceCharacterRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT c.id, c.name, c.rarity::text, c.picture, c.picture_thumb, c.picture_card,
		       c.picture_lqip, c.picture_media_id, c.picture_status::text, c.focal_x, c.focal_y,
		       c.created_at, c.updated_at, o.physical_form::text, o.armament_haki::text,
		       o.observation_haki::text, o.conqueror_haki::text, o.fruit_mastery::text
		FROM one_piece_characters o JOIN characters c ON c.id = o.id
		ORDER BY c.name`)
	if err != nil {
		return nil, fmt.Errorf("querying one piece characters: %w", err)
	}
	defer rows.Close()

	var out []OnePieceCharacterRow
	for rows.Next() {
		var r OnePieceCharacterRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Rarity, &r.Picture, &r.PictureThumb, &r.PictureCard,
			&r.PictureLQIP, &r.PictureMediaID, &r.PictureStatus, &r.FocalX, &r.FocalY,
			&r.CreatedAt, &r.UpdatedAt, &r.PhysicalForm, &r.ArmamentHaki, &r.ObservationHaki,
			&r.ConquerorHaki, &r.FruitMastery); err != nil {
			return nil, fmt.Errorf("scanning one piece character row: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	translations, err := dumpCharacterTranslations(ctx, tx, "ONE_PIECE")
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Translations = translations[out[i].ID]
	}
	return out, nil
}

// dumpCharacterTranslations mirrors dumpPowerTranslations for
// character_translations (description only, no skills).
func dumpCharacterTranslations(ctx context.Context, tx pgx.Tx, manga string) (map[string]map[string]Translation, error) {
	rows, err := tx.Query(ctx, `
		SELECT t.character_id, t.locale::text, t.description
		FROM character_translations t JOIN characters c ON c.id = t.character_id
		WHERE c.manga = $1`, manga)
	if err != nil {
		return nil, fmt.Errorf("querying character translations: %w", err)
	}
	defer rows.Close()

	out := make(map[string]map[string]Translation)
	for rows.Next() {
		var characterID, locale string
		var tr Translation
		if err := rows.Scan(&characterID, &locale, &tr.Description); err != nil {
			return nil, fmt.Errorf("scanning character translation: %w", err)
		}
		if out[characterID] == nil {
			out[characterID] = make(map[string]Translation)
		}
		out[characterID][locale] = tr
	}
	return out, rows.Err()
}

// referencedStorage collects every picture key (main/thumb/card) and every
// picture_media_id group id the snapshot's rows reference, so dumping
// storage_objects/media_objects can be scoped to exactly those - not to
// every row in either table, which would also pull in unrelated user
// avatars.
func referencedStorage(snap Snapshot) (keys []string, groupIDs []string) {
	keySet := map[string]struct{}{}
	groupSet := map[string]struct{}{}
	add := func(picture, thumb, card, mediaID string) {
		for _, k := range []string{picture, thumb, card} {
			if k != "" {
				keySet[k] = struct{}{}
			}
		}
		if mediaID != "" {
			groupSet[mediaID] = struct{}{}
		}
	}
	for _, s := range snap.Stands {
		add(s.Picture, s.PictureThumb, s.PictureCard, s.PictureMediaID)
	}
	for _, d := range snap.DevilFruits {
		add(d.Picture, d.PictureThumb, d.PictureCard, d.PictureMediaID)
	}
	for _, c := range snap.JojoCharacters {
		add(c.Picture, c.PictureThumb, c.PictureCard, c.PictureMediaID)
	}
	for _, c := range snap.OnePieceCharacters {
		add(c.Picture, c.PictureThumb, c.PictureCard, c.PictureMediaID)
	}
	for k := range keySet {
		keys = append(keys, k)
	}
	for g := range groupSet {
		groupIDs = append(groupIDs, g)
	}
	return keys, groupIDs
}

func dumpStorageObjects(ctx context.Context, tx pgx.Tx, keys []string) ([]StorageObjectRow, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	rows, err := tx.Query(ctx, `SELECT key, provider, bytes FROM storage_objects WHERE key = ANY($1) ORDER BY key`, keys)
	if err != nil {
		return nil, fmt.Errorf("querying storage objects: %w", err)
	}
	defer rows.Close()

	var out []StorageObjectRow
	for rows.Next() {
		var r StorageObjectRow
		if err := rows.Scan(&r.Key, &r.Provider, &r.Bytes); err != nil {
			return nil, fmt.Errorf("scanning storage object: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func dumpMediaObjects(ctx context.Context, tx pgx.Tx, groupIDs []string) ([]MediaObjectRow, error) {
	if len(groupIDs) == 0 {
		return nil, nil
	}
	rows, err := tx.Query(ctx, `
		SELECT group_id, variant, storage_key, content_type, bytes, scope
		FROM media_objects WHERE group_id = ANY($1) ORDER BY group_id, variant`, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("querying media objects: %w", err)
	}
	defer rows.Close()

	var out []MediaObjectRow
	for rows.Next() {
		var r MediaObjectRow
		if err := rows.Scan(&r.GroupID, &r.Variant, &r.StorageKey, &r.ContentType, &r.Bytes, &r.Scope); err != nil {
			return nil, fmt.Errorf("scanning media object: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
