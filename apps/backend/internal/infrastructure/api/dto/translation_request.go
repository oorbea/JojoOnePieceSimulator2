package dto

import (
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// TranslationRequest is one locale's content in a StandRequest/
// DevilFruitRequest's "translations" map, keyed by locale string
// ("en-GB", "es-ES", "ca-ES").
type TranslationRequest struct {
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
}

// validateTranslations parses and validates a "translations" map shared by
// StandRequest/DevilFruitRequest: every key must be a supported locale,
// en-GB must be present with a non-empty description and at least one
// skill, and every other present locale must also have a non-empty
// description and at least one skill (a locale is either fully translated
// or absent - no partial/blank overrides).
func validateTranslations(m map[string]TranslationRequest) (ports.PowerTranslations, []string) {
	var errs []string
	out := make(ports.PowerTranslations, len(m))

	for key, t := range m {
		locale, err := enums.ParseLocale(key)
		if err != nil {
			errs = append(errs, fmt.Sprintf("translations: unsupported locale %q", key))
			continue
		}
		if t.Description == "" {
			errs = append(errs, fmt.Sprintf("translations.%s.description is required", key))
			continue
		}
		if len(t.Skills) == 0 {
			errs = append(errs, fmt.Sprintf("translations.%s.skills are required", key))
			continue
		}
		skills := append([]string(nil), t.Skills...)
		out[locale] = ports.PowerContent{Description: t.Description, Skills: skills}
	}

	if _, ok := out[enums.EnGB]; !ok {
		errs = append(errs, "translations.en-GB is required")
	}

	return out, errs
}

// StageTranslationRequest is one locale's content in a StageRequest's
// "translations" map - just a description, unlike TranslationRequest, since
// a Stage has no skills.
type StageTranslationRequest struct {
	Description string `json:"description"`
}

// validateStageTranslations parses and validates a StageRequest's
// "translations" map: every key must be a supported locale, and - per the
// owner's decision, unlike Power translations where only en-GB is mandatory -
// every one of enums.Locales() must be present with a non-empty description.
func validateStageTranslations(m map[string]StageTranslationRequest) (ports.StageTranslations, []string) {
	var errs []string
	out := make(ports.StageTranslations, len(m))

	for key, t := range m {
		locale, err := enums.ParseLocale(key)
		if err != nil {
			errs = append(errs, fmt.Sprintf("translations: unsupported locale %q", key))
			continue
		}
		if t.Description == "" {
			errs = append(errs, fmt.Sprintf("translations.%s.description is required", key))
			continue
		}
		out[locale] = t.Description
	}

	for _, l := range enums.Locales() {
		if _, ok := out[l]; !ok {
			errs = append(errs, fmt.Sprintf("translations.%s is required", l))
		}
	}

	return out, errs
}
