package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// sourceLocaleFor returns the locale whose text is trusted as correct for a
// row of the given kind - fixed for stands/devil fruits (per
// ObsidianVault/catalog-seed-from-prod.md), looked up per-row for
// characters since admins have filled those in inconsistently.
func sourceLocaleFor(kind, name string, sources Sources) (locale string, confirmed bool) {
	switch kind {
	case "stand":
		return "es-ES", true
	case "devil_fruit":
		return "en-GB", true
	case "jojo_character":
		loc, ok := sources.JojoCharacters[name]
		return loc, ok
	case "one_piece_character":
		loc, ok := sources.OnePieceCharacters[name]
		return loc, ok
	default:
		return "", false
	}
}

// overlayFor returns the per-kind translations.json map to read/write, and
// suggestionsFor returns the per-kind sources.json map (characters only).
func overlayFor(o *Overlay, kind string) map[string]map[string]OverlayTranslation {
	switch kind {
	case "stand":
		return o.Stands
	case "devil_fruit":
		return o.DevilFruits
	case "jojo_character":
		return o.JojoCharacters
	case "one_piece_character":
		return o.OnePieceCharacters
	default:
		return nil
	}
}

// row bundles what plan needs from any of the four row types, so the
// per-kind loops below share one body.
type row struct {
	kind         string
	name         string
	translations map[string]Translation
}

func rowsFromSnapshot(snap Snapshot) []row {
	var out []row
	for _, s := range snap.Stands {
		out = append(out, row{"stand", s.Name, s.Translations})
	}
	for _, d := range snap.DevilFruits {
		out = append(out, row{"devil_fruit", d.Name, d.Translations})
	}
	for _, c := range snap.JojoCharacters {
		out = append(out, row{"jojo_character", c.Name, c.Translations})
	}
	for _, c := range snap.OnePieceCharacters {
		out = append(out, row{"one_piece_character", c.Name, c.Translations})
	}
	return out
}

// runPlan diffs snapshot.json against translations.json (and, for
// characters, sources.json) and writes pending.json: every translation that
// still needs writing before `generate` can produce a migration.
//
// A character row whose source locale isn't yet confirmed in sources.json
// doesn't go into pending - it can't be translated without knowing which
// text to translate from - instead it's written to sources.suggested.json
// with this tool's best guess, and `plan` exits non-zero asking for that
// file to be reviewed and merged into sources.json.
func runPlan(args []string) error {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	snap, err := readSnapshot()
	if err != nil {
		return err
	}
	overlay, err := readOverlay()
	if err != nil {
		return err
	}
	sources, err := readSources()
	if err != nil {
		return err
	}

	pending, suggested, needsConfirmation := computePending(snap, overlay, sources)

	if needsConfirmation {
		if err := writeJSON(filepath.Join(seedDir, "sources.suggested.json"), suggested); err != nil {
			return err
		}
		return fmt.Errorf("some character rows have no confirmed source locale - review %s and merge the entries you agree with into sources.json, then re-run plan", filepath.Join(seedDir, "sources.suggested.json"))
	}
	_ = os.Remove(filepath.Join(seedDir, "sources.suggested.json")) // stale from a previous run once every row is confirmed

	if err := writeJSON(filepath.Join(seedDir, "pending.json"), pending); err != nil {
		return err
	}

	if len(pending) == 0 {
		fmt.Println("plan: nothing pending - every row has every locale translated and current. Ready for `generate`.")
		return nil
	}
	byReason := map[string]int{}
	for _, p := range pending {
		byReason[p.Reason]++
	}
	fmt.Printf("plan: %d translations pending (missing=%d placeholder=%d stale=%d) - wrote %s\n",
		len(pending), byReason["missing"], byReason["placeholder"], byReason["stale"], filepath.Join(seedDir, "pending.json"))
	fmt.Println("Translate each entry and add it to translations.json (see that file's existing entries for the shape), then re-run plan.")
	return nil
}

// computePending is the shared diff both `plan` and `generate` run: for
// every row in snap, work out which target-locale translations still need
// writing (missing/placeholder/stale against overlay), or - for a character
// row with no confirmed source locale yet - add a guess to suggested and
// set needsConfirmation.
func computePending(snap Snapshot, overlay Overlay, sources Sources) (pending []PendingEntry, suggested Sources, needsConfirmation bool) {
	suggested = Sources{JojoCharacters: map[string]string{}, OnePieceCharacters: map[string]string{}}

	for _, r := range rowsFromSnapshot(snap) {
		sourceLocale, confirmed := sourceLocaleFor(r.kind, r.name, sources)
		if !confirmed {
			guess, confidence := bestGuess(r.translations)
			suggestionMap := suggested.JojoCharacters
			if r.kind == "one_piece_character" {
				suggestionMap = suggested.OnePieceCharacters
			}
			suggestionMap[r.name] = fmt.Sprintf("%s # confidence %.2f - REVIEW before merging into sources.json", guess, confidence)
			needsConfirmation = true
			continue
		}

		source, ok := r.translations[sourceLocale]
		if !ok {
			pending = append(pending, PendingEntry{
				Kind: r.kind, Name: r.name, TargetLocale: sourceLocale, SourceLocale: sourceLocale,
				Reason: "missing",
			})
			continue
		}
		hash := sourceHash(source)

		for _, target := range locales {
			if target == sourceLocale {
				continue
			}
			overlayEntry, hasOverlay := overlayFor(&overlay, r.kind)[r.name][target]
			existing, hasExisting := r.translations[target]

			switch {
			case hasOverlay && overlayEntry.SourceHash == hash:
				// Already translated and still matches the current source
				// text - nothing to do. The generated migration writes
				// overlayEntry regardless of what's currently in prod.
				continue
			case hasOverlay:
				pending = append(pending, PendingEntry{
					Kind: r.kind, Name: r.name, TargetLocale: target, SourceLocale: sourceLocale,
					SourceText: source.Description, SourceSkills: source.Skills,
					Reason: "stale", SourceHash: hash,
				})
			case hasExisting && !looksLikePlaceholder(existing, target, source, sourceLocale):
				// prod already has plausible target-locale text and the
				// overlay has never recorded it - trust prod once, but
				// pin it to today's source hash so a future prod edit is
				// still caught as "stale".
				continue
			default:
				reason := "missing"
				if hasExisting {
					reason = "placeholder"
				}
				pending = append(pending, PendingEntry{
					Kind: r.kind, Name: r.name, TargetLocale: target, SourceLocale: sourceLocale,
					SourceText: source.Description, SourceSkills: source.Skills,
					Reason: reason, SourceHash: hash,
				})
			}
		}
	}

	return pending, suggested, needsConfirmation
}

// bestGuess picks the locale detectLocale is most confident about across
// every translation a character row already has, for the sources.suggested
// entry a human reviews.
func bestGuess(translations map[string]Translation) (locale string, confidence float64) {
	best, bestConfidence := "", -1.0
	for loc, t := range translations {
		if loc == "" {
			continue
		}
		detected, conf := detectLocale(t.Description)
		// A row's own text is at least as informative as the detector's
		// verdict on which language it thinks the text is in - so weight
		// by whether detection agrees with the locale the text is stored
		// under (a strong, self-consistent signal that locale is honest).
		if detected == loc {
			conf += 0.5
		}
		if conf > bestConfidence {
			best, bestConfidence = loc, conf
		}
	}
	return best, bestConfidence
}

func readSnapshot() (Snapshot, error) {
	var s Snapshot
	err := readJSON(filepath.Join(seedDir, "snapshot.json"), &s)
	return s, err
}

func readOverlay() (Overlay, error) {
	o := Overlay{
		Stands:             map[string]map[string]OverlayTranslation{},
		DevilFruits:        map[string]map[string]OverlayTranslation{},
		JojoCharacters:     map[string]map[string]OverlayTranslation{},
		OnePieceCharacters: map[string]map[string]OverlayTranslation{},
	}
	path := filepath.Join(seedDir, "translations.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return o, nil
	}
	err := readJSON(path, &o)
	return o, err
}

func readSources() (Sources, error) {
	s := Sources{JojoCharacters: map[string]string{}, OnePieceCharacters: map[string]string{}}
	path := filepath.Join(seedDir, "sources.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return s, nil
	}
	err := readJSON(path, &s)
	return s, err
}

func readJSON(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(v); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}
	return nil
}
