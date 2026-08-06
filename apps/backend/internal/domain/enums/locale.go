package enums

import "errors"

type Locale byte

// EnGB is the default and the mandatory final link of every content
// fallback chain (ca-ES -> es-ES -> en-GB): every Power must have an en-GB
// translation, enforced in the application layer.
const (
	EnGB Locale = iota
	EsES
	CaES
)

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
// content for l, starting with l itself and always ending in EnGB.
func FallbackChain(l Locale) []Locale {
	switch l {
	case CaES:
		return []Locale{CaES, EsES, EnGB}
	case EsES:
		return []Locale{EsES, EnGB}
	default:
		return []Locale{EnGB}
	}
}
