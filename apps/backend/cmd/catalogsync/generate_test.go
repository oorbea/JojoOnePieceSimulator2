package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSqlLit_EscapesQuotes(t *testing.T) {
	got := sqlLit("O'Brien's stand")
	want := "'O''Brien''s stand'"
	if got != want {
		t.Errorf("sqlLit = %q, want %q", got, want)
	}
}

func TestSqlTextArray_EmptyAndPopulated(t *testing.T) {
	if got := sqlTextArray(nil); got != "'{}'::text[]" {
		t.Errorf("sqlTextArray(nil) = %q", got)
	}
	got := sqlTextArray([]string{"a", "it's"})
	want := "ARRAY['a','it''s']::text[]"
	if got != want {
		t.Errorf("sqlTextArray = %q, want %q", got, want)
	}
}

func TestNextMigrationNumber_ScansHighestExisting(t *testing.T) {
	root := t.TempDir()
	migDir := filepath.Join(root, "db", "migrations")
	if err := os.MkdirAll(migDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"00001_init.sql", "00016_picture_focal_point.sql", "notamigration.sql"} {
		if err := os.WriteFile(filepath.Join(migDir, name), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	got, err := nextMigrationNumber()
	if err != nil {
		t.Fatalf("nextMigrationNumber: %v", err)
	}
	if got != "00017" {
		t.Errorf("nextMigrationNumber = %q, want %q", got, "00017")
	}
}

func TestComputePending_StandMissingTranslationIsPending(t *testing.T) {
	snap := Snapshot{
		Stands: []StandRow{{
			ID: "s1", Name: "Star Platinum",
			Translations: map[string]Translation{
				"es-ES": {Description: "Un stand muy fuerte y rápido.", Skills: []string{"Golpes rapidísimos"}},
			},
		}},
	}
	overlay := emptyOverlay()
	sources := emptySources()

	pending, _, needsConfirmation := computePending(snap, overlay, sources)
	if needsConfirmation {
		t.Fatal("stands never need source-locale confirmation")
	}
	if len(pending) != 2 {
		t.Fatalf("pending = %d entries, want 2 (en-GB, ca-ES both missing): %+v", len(pending), pending)
	}
	for _, p := range pending {
		if p.Reason != "missing" || p.SourceLocale != "es-ES" {
			t.Errorf("entry = %+v, want reason=missing source_locale=es-ES", p)
		}
	}
}

func TestComputePending_OverlayMatchingHashIsSatisfied(t *testing.T) {
	source := Translation{Description: "Un stand muy fuerte.", Skills: []string{"Golpe"}}
	snap := Snapshot{
		Stands: []StandRow{{
			ID: "s1", Name: "The World",
			Translations: map[string]Translation{"es-ES": source},
		}},
	}
	overlay := emptyOverlay()
	overlay.Stands["The World"] = map[string]OverlayTranslation{
		"en-GB": {Description: "A very strong stand.", Skills: []string{"Punch"}, SourceHash: sourceHash(source)},
		"ca-ES": {Description: "Un stand molt fort.", Skills: []string{"Cop"}, SourceHash: sourceHash(source)},
	}

	pending, _, _ := computePending(snap, overlay, emptySources())
	if len(pending) != 0 {
		t.Fatalf("pending = %+v, want none - overlay already covers both target locales", pending)
	}
}

func TestComputePending_StaleOverlayIsPending(t *testing.T) {
	oldSource := Translation{Description: "Texto viejo."}
	newSource := Translation{Description: "Texto nuevo, editado por un admin."}
	snap := Snapshot{
		DevilFruits: []DevilFruitRow{{
			ID: "d1", Name: "Gomu Gomu no Mi",
			Translations: map[string]Translation{"en-GB": newSource},
		}},
	}
	overlay := emptyOverlay()
	overlay.DevilFruits["Gomu Gomu no Mi"] = map[string]OverlayTranslation{
		"es-ES": {Description: "Texto viejo traducido.", SourceHash: sourceHash(oldSource)},
	}

	pending, _, _ := computePending(snap, overlay, emptySources())
	found := false
	for _, p := range pending {
		if p.TargetLocale == "es-ES" && p.Reason == "stale" {
			found = true
		}
	}
	if !found {
		t.Fatalf("pending = %+v, want a stale es-ES entry (source text changed since translation)", pending)
	}
}

func TestComputePending_CharacterWithoutConfirmedSourceNeedsConfirmation(t *testing.T) {
	snap := Snapshot{
		JojoCharacters: []JojoCharacterRow{{
			ID: "c1", Name: "Jotaro Kujo",
			Translations: map[string]Translation{"es-ES": {Description: "Un luchador decidido y valiente."}},
		}},
	}
	_, suggested, needsConfirmation := computePending(snap, emptyOverlay(), emptySources())
	if !needsConfirmation {
		t.Fatal("want needsConfirmation = true - no sources.json entry for this character")
	}
	if _, ok := suggested.JojoCharacters["Jotaro Kujo"]; !ok {
		t.Errorf("suggested = %+v, want a guess for Jotaro Kujo", suggested)
	}
}

func TestComputePending_CharacterExistingPlausibleTranslationIsTrusted(t *testing.T) {
	source := Translation{Description: "Un luchador decidido y valiente que protege a su familia."}
	snap := Snapshot{
		JojoCharacters: []JojoCharacterRow{{
			ID: "c1", Name: "Jotaro Kujo",
			Translations: map[string]Translation{
				"es-ES": source,
				"en-GB": {Description: "A determined and brave fighter who protects his family."},
			},
		}},
	}
	sources := emptySources()
	sources.JojoCharacters["Jotaro Kujo"] = "es-ES"

	pending, _, needsConfirmation := computePending(snap, emptyOverlay(), sources)
	if needsConfirmation {
		t.Fatal("source locale is confirmed, should not need confirmation")
	}
	for _, p := range pending {
		if p.TargetLocale == "en-GB" {
			t.Errorf("en-GB already has plausible text, should not be pending: %+v", p)
		}
	}
}

func TestComputePending_PlaceholderCopiedTextIsPending(t *testing.T) {
	source := Translation{Description: "Descripción original en castellano de este personaje."}
	snap := Snapshot{
		JojoCharacters: []JojoCharacterRow{{
			ID: "c1", Name: "Joseph Joestar",
			Translations: map[string]Translation{
				"es-ES": source,
				"en-GB": source, // copy-pasted, not actually translated
			},
		}},
	}
	sources := emptySources()
	sources.JojoCharacters["Joseph Joestar"] = "es-ES"

	pending, _, _ := computePending(snap, emptyOverlay(), sources)
	found := false
	for _, p := range pending {
		if p.TargetLocale == "en-GB" && p.Reason == "placeholder" {
			found = true
		}
	}
	if !found {
		t.Fatalf("pending = %+v, want an en-GB placeholder entry (identical to es-ES source)", pending)
	}
}

func TestResolveWithSourceHint_MergesSnapshotAndOverlay(t *testing.T) {
	overlay := emptyOverlay()
	overlay.Stands["Crazy Diamond"] = map[string]OverlayTranslation{
		"en-GB": {Description: "Restores anything to its former state.", Skills: []string{"Repair"}},
	}
	snapRow := map[string]Translation{
		"es-ES": {Description: "Restaura cualquier cosa a su estado anterior.", Skills: []string{"Reparar"}},
	}

	got := resolveWithSourceHint("stand", "Crazy Diamond", snapRow, overlay)
	if len(got) != 2 {
		t.Fatalf("got %d locales, want 2 (es-ES from snapshot + en-GB from overlay): %+v", len(got), got)
	}
	if got["en-GB"].Description != "Restores anything to its former state." {
		t.Errorf("en-GB = %+v", got["en-GB"])
	}
	if got["es-ES"].Description == "" {
		t.Error("es-ES (source) missing from resolved map")
	}
}

func TestRenderMigration_ProducesGooseShapeAndGuardsWrites(t *testing.T) {
	snap := Snapshot{
		DumpedAt: time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC),
		Stands: []StandRow{{
			ID: "11111111-1111-1111-1111-111111111111", Name: "Star Platinum", Rarity: "LEGENDARY",
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
			AttackPower: "A", Speed: "A", AttackRange: "D", Endurance: "B", Precision: "A", Potential: "INFINITE",
			Translations: map[string]Translation{"es-ES": {Description: "Fuerte.", Skills: []string{"Ora"}}},
		}},
	}
	overlay := emptyOverlay()
	overlay.Stands["Star Platinum"] = map[string]OverlayTranslation{
		"en-GB": {Description: "Strong.", Skills: []string{"Ora"}},
		"ca-ES": {Description: "Fort.", Skills: []string{"Ora"}},
	}

	sql := renderMigration(snap, overlay)

	for _, want := range []string{
		"-- +goose Up", "-- +goose ENVSUB ON", "-- +goose Down",
		"${SEED_CATALOG_FULL:-false}",
		"INSERT INTO powers", "INSERT INTO stands",
		"INSERT INTO power_translations",
		"WHERE EXISTS (SELECT 1 FROM powers WHERE id = '11111111-1111-1111-1111-111111111111' AND updated_at <=",
	} {
		if !strings.Contains(sql, want) {
			t.Errorf("generated migration missing %q\n---\n%s", want, sql)
		}
	}
}

func emptyOverlay() Overlay {
	return Overlay{
		Stands:             map[string]map[string]OverlayTranslation{},
		DevilFruits:        map[string]map[string]OverlayTranslation{},
		JojoCharacters:     map[string]map[string]OverlayTranslation{},
		OnePieceCharacters: map[string]map[string]OverlayTranslation{},
	}
}

func emptySources() Sources {
	return Sources{JojoCharacters: map[string]string{}, OnePieceCharacters: map[string]string{}}
}
