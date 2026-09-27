package enums_test

import (
	"reflect"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestFallbackChain_AlwaysEndsInDefaultLocale(t *testing.T) {
	for _, l := range enums.Locales() {
		chain := enums.FallbackChain(l)
		if len(chain) == 0 {
			t.Fatalf("FallbackChain(%s) is empty", l)
		}
		if got := chain[len(chain)-1]; got != enums.DefaultLocale {
			t.Errorf("FallbackChain(%s) = %v, last link = %s, want %s", l, chain, got, enums.DefaultLocale)
		}
	}
}

func TestFallbackChain_StartsWithTheRequestedLocale(t *testing.T) {
	for _, l := range enums.Locales() {
		chain := enums.FallbackChain(l)
		if chain[0] != l {
			t.Errorf("FallbackChain(%s)[0] = %s, want %s", l, chain[0], l)
		}
	}
}

func TestFallbackChain_ExactChains(t *testing.T) {
	cases := []struct {
		locale enums.Locale
		want   []enums.Locale
	}{
		{enums.CaES, []enums.Locale{enums.CaES, enums.EsES}},
		{enums.EnGB, []enums.Locale{enums.EnGB, enums.EsES}},
		{enums.EsES, []enums.Locale{enums.EsES}},
	}
	for _, c := range cases {
		if got := enums.FallbackChain(c.locale); !reflect.DeepEqual(got, c.want) {
			t.Errorf("FallbackChain(%s) = %v, want %v", c.locale, got, c.want)
		}
	}
}

func TestDefaultLocale_IsEsES(t *testing.T) {
	if enums.DefaultLocale != enums.EsES {
		t.Errorf("DefaultLocale = %s, want es-ES", enums.DefaultLocale)
	}
}
