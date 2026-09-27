package enums

import "errors"

type Locale byte

// EnGB is the zero value (Locale's iota order is a wire/storage detail,
// left alone by the 2026-09-27 default-locale switch to avoid reordering
// every generated contract and DB enum) - it is no longer the mandatory
// locale. See DefaultLocale.
const (
	EnGB Locale = iota
	EsES
	CaES
)

// DefaultLocale is the mandatory final link of every content fallback
// chain (see FallbackChain) - every Power/Character must have an es-ES
// translation, enforced in the application layer. Also the default UI
// language for a new user and an unresolved Accept-Language/?lang (see
// endpoints/locale.go, game_endpoints.go, entities/user.go). Changed from
// EnGB to EsES on 2026-09-27 - see ObsidianVault/i18n-multi-language.md.
const DefaultLocale = EsES

func (l Locale) String() string {
	switch l {
	case EnGB:
		return "en-GB"
	case EsES:
		return "es-ES"
	case CaES:
		return "ca-ES"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidLocale = errors.New("invalid locale")

func (l Locale) IsValid() bool {
	switch l {
	case EnGB, EsES, CaES:
		return true
	default:
		return false
	}
}

func ParseLocale(str string) (Locale, error) {
	switch str {
	case "en-GB":
		return EnGB, nil
	case "es-ES":
		return EsES, nil
	case "ca-ES":
		return CaES, nil
	default:
		return EnGB, ErrInvalidLocale
	}
}

// Locales lists every supported locale, in fallback-chain order (most
// specific to least), used to build the fallback chain for a given locale.
func Locales() []Locale {
	return []Locale{EnGB, EsES, CaES}
}

// FallbackChain returns the ordered list of locales to try when resolving
// content for l, starting with l itself and always ending in DefaultLocale
// (es-ES) - the one locale every Power/Character/Stage is guaranteed to
// have. en-GB is no longer the final link: an en-GB request now falls back
// to es-ES too, since en-GB translations aren't guaranteed to exist.
func FallbackChain(l Locale) []Locale {
	switch l {
	case CaES:
		return []Locale{CaES, EsES}
	case EnGB:
		return []Locale{EnGB, EsES}
	default:
		return []Locale{EsES}
	}
}
