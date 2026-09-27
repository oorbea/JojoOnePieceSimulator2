package main

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// stopwords is a small, high-frequency word list per locale, used only to
// pick which of the three languages a chunk of prose is most likely written
// in - not a real language-ID model, but enough to flag "this es-ES row
// looks like English" or "this ca-ES row is byte-identical to the es-ES
// one" for a human (or Claude) to double check in `plan`'s output.
var stopwords = map[string]map[string]struct{}{
	"en-GB": set("the", "and", "of", "to", "a", "in", "is", "his", "her", "with", "that", "he", "she", "it", "can", "an", "as", "by", "from", "was", "were"),
	"es-ES": set("el", "la", "los", "las", "de", "que", "y", "un", "una", "con", "su", "sus", "es", "por", "para", "del", "al", "se", "lo", "en"),
	"ca-ES": set("el", "la", "els", "les", "de", "que", "i", "un", "una", "amb", "seu", "seva", "es", "per", "del", "al", "se", "no", "en", "aquest"),
}

func set(words ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}

var wordRe = regexp.MustCompile(`[a-zàèéíòóúüïç]+`)

// detectLocale scores text against each locale's stopword list and returns
// the best match plus its confidence (matched stopword tokens / total
// tokens). Short text (a name, a one-word skill) yields a low-confidence
// guess - callers should treat anything under ~0.15 as "unknown", not act
// on it silently.
func detectLocale(text string) (locale string, confidence float64) {
	tokens := wordRe.FindAllString(strings.ToLower(text), -1)
	if len(tokens) == 0 {
		return "", 0
	}

	best, bestScore := "", -1
	// Iterate `locales` (not the map) so ties break deterministically.
	for _, loc := range locales {
		score := 0
		for _, tok := range tokens {
			if _, ok := stopwords[loc][tok]; ok {
				score++
			}
		}
		if score > bestScore {
			best, bestScore = loc, score
		}
	}
	return best, float64(bestScore) / float64(len(tokens))
}

// minPlausibleDescriptionLength is well under the shortest real content
// description in this catalogue (~80 chars for the shortest genuine
// devil-fruit blurb) but well over stub text admins leave behind
// ("placeholder", "A", "-", "TBD") - see catalogsync's own snapshot for the
// length distribution this was picked against.
const minPlausibleDescriptionLength = 30

// looksLikePlaceholder reports whether target's text is suspiciously not a
// real, own-locale translation: empty, too short to be real content (a
// stub like "placeholder" or "A" left behind while the real text was only
// written for the source locale), byte-identical to source's text (a lazy
// copy across the locale tabs), or confidently detected as source's
// language instead of target's own.
func looksLikePlaceholder(target Translation, targetLocale string, source Translation, sourceLocale string) bool {
	if targetLocale == sourceLocale {
		return false
	}
	text := strings.TrimSpace(target.Description)
	if text == "" || len(text) < minPlausibleDescriptionLength {
		return true
	}
	if strings.EqualFold(text, strings.TrimSpace(source.Description)) {
		return true
	}
	detected, confidence := detectLocale(target.Description)
	return confidence >= 0.2 && detected == sourceLocale && detected != targetLocale
}

// sourceHash fingerprints a source translation so a generated migration can
// later be regenerated deterministically, and so `plan` can tell a
// still-current overlay entry from one whose source text changed in prod
// since it was translated ("stale").
func sourceHash(t Translation) string {
	sum := sha256.Sum256([]byte(t.Description + "\x00" + strings.Join(t.Skills, "\x00")))
	return hex.EncodeToString(sum[:])
}
