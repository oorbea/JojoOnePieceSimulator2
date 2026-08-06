package dto

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"

// TranslationResponse mirrors TranslationRequest, for reading back every
// locale's content in an admin edit form.
type TranslationResponse struct {
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
}

// PowerTranslationsResponse is the body of the admin-only
// GET /stands/{id}/translations and GET /devil-fruits/{id}/translations
// routes: every locale's content for one Power, keyed by locale string.
// Unlike the public StandResponse/DevilFruitResponse (one resolved
// description/skills), this always carries every locale at once, so an
// admin can edit them side by side.
type PowerTranslationsResponse struct {
	Translations map[string]TranslationResponse `json:"translations"`
}

func NewPowerTranslationsResponse(t ports.PowerTranslations) PowerTranslationsResponse {
	out := make(map[string]TranslationResponse, len(t))
	for locale, content := range t {
		out[locale.String()] = TranslationResponse{Description: content.Description, Skills: content.Skills}
	}
	return PowerTranslationsResponse{Translations: out}
}
