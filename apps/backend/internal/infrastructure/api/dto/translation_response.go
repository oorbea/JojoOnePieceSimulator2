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

// StageTranslationResponse mirrors StageTranslationRequest, for reading back
// every locale's description in an admin edit form.
type StageTranslationResponse struct {
	Description string `json:"description"`
}

// StageTranslationsResponse is the body of the admin-only
// GET /stages/{id}/translations route: every locale's description for one
// Stage, keyed by locale string - same shape as PowerTranslationsResponse,
// without Skills.
type StageTranslationsResponse struct {
	Translations map[string]StageTranslationResponse `json:"translations"`
}

func NewStageTranslationsResponse(t ports.StageTranslations) StageTranslationsResponse {
	out := make(map[string]StageTranslationResponse, len(t))
	for locale, description := range t {
		out[locale.String()] = StageTranslationResponse{Description: description}
	}
	return StageTranslationsResponse{Translations: out}
}
