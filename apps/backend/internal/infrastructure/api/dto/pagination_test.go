package dto

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

func TestPageParamsFromQuery_AbsentParams_NotRequested(t *testing.T) {
	params, err := PageParamsFromQuery(url.Values{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Requested {
		t.Error("Requested = true, want false when neither limit nor cursor is present")
	}
	if params.Limit != defaultPageSize {
		t.Errorf("Limit = %d, want default %d", params.Limit, defaultPageSize)
	}
}

func TestPageParamsFromQuery_InvalidLimit_Returns400(t *testing.T) {
	for _, v := range []string{"abc", "0", "-1"} {
		q := url.Values{"limit": {v}}
		if _, err := PageParamsFromQuery(q); err == nil {
			t.Errorf("limit=%q: expected a validation error, got nil", v)
		}
	}
}

func TestPageParamsFromQuery_LimitOverMax_ClampedSilently(t *testing.T) {
	q := url.Values{"limit": {"500"}}
	params, err := PageParamsFromQuery(q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Limit != maxPageSize {
		t.Errorf("Limit = %d, want clamped to %d", params.Limit, maxPageSize)
	}
}

func TestPageParamsFromQuery_CursorAlonePresent_Requested(t *testing.T) {
	q := url.Values{"cursor": {"xyz"}}
	params, err := PageParamsFromQuery(q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !params.Requested || !params.HasCursor {
		t.Error("expected Requested and HasCursor both true when only ?cursor= is set")
	}
}

func TestPageParamsFromQuery_TotalFalse_OptsOut(t *testing.T) {
	q := url.Values{"limit": {"10"}, "total": {"false"}}
	params, err := PageParamsFromQuery(q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.WithTotal {
		t.Error("WithTotal = true, want false with ?total=false")
	}
}

func TestCursor_RoundTrip(t *testing.T) {
	fp := FilterFingerprint("some-canonical-string")
	encoded := EncodeCursor(StandCursor{Name: "Star Platinum"}, fp)

	decoded, err := DecodeCursor[StandCursor](encoded, fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded.Name != "Star Platinum" {
		t.Errorf("Name = %q, want %q", decoded.Name, "Star Platinum")
	}
}

func TestDecodeCursor_TruncatedBase64_Returns400(t *testing.T) {
	if _, err := DecodeCursor[StandCursor]("not-valid-base64-!!!", "fp"); err == nil {
		t.Error("expected an error for malformed base64, got nil")
	}
}

func TestDecodeCursor_WrongFingerprint_Returns400(t *testing.T) {
	encoded := EncodeCursor(StandCursor{Name: "Star Platinum"}, FilterFingerprint("filters-a"))
	if _, err := DecodeCursor[StandCursor](encoded, FilterFingerprint("filters-b")); err == nil {
		t.Error("expected an error when the cursor's fingerprint doesn't match, got nil")
	}
}

func TestDecodeCursor_WrongVersion_Returns400(t *testing.T) {
	fp := FilterFingerprint("filters")
	env := cursorEnvelope[StandCursor]{V: 2, F: fp, K: StandCursor{Name: "x"}}
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if _, err := DecodeCursor[StandCursor](base64.RawURLEncoding.EncodeToString(data), fp); err == nil {
		t.Error("expected an error for an unsupported cursor version, got nil")
	}
}

// StandFiltersFingerprint: reproducing a cursor under a different filter
// combination must fail (the whole point of binding the fingerprint) - a
// cursor issued unfiltered must be rejected once a filter is later applied,
// even though the underlying cursor key (a stand name) is still valid data.
func TestStandFiltersFingerprint_DiffersAcrossFilterSets(t *testing.T) {
	rarity := enums.Common
	unfiltered := StandFiltersFingerprint(ports.StandFilters{}, enums.EnGB)
	filtered := StandFiltersFingerprint(ports.StandFilters{Rarity: &rarity}, enums.EnGB)
	if unfiltered == filtered {
		t.Error("fingerprint must differ between an unfiltered and a filtered request")
	}

	encoded := EncodeCursor(StandCursor{Name: "A"}, unfiltered)
	if _, err := DecodeCursor[StandCursor](encoded, filtered); err == nil {
		t.Error("a cursor issued unfiltered must be rejected when replayed against a filtered request")
	}
}

func TestStandFiltersFingerprint_SameFiltersDifferentOrder_SameFingerprint(t *testing.T) {
	rarity := enums.Common
	search := "foo"
	a := StandFiltersFingerprint(ports.StandFilters{Rarity: &rarity, Search: &search}, enums.EnGB)
	b := StandFiltersFingerprint(ports.StandFilters{Search: &search, Rarity: &rarity}, enums.EnGB)
	if a != b {
		t.Error("fingerprint must not depend on the order filter fields were set in Go, only their values")
	}
}

// TestJojoCharacterFiltersFingerprint_DiffersAcrossFilterSets - see
// TestStandFiltersFingerprint_DiffersAcrossFilterSets's doc.
func TestJojoCharacterFiltersFingerprint_DiffersAcrossFilterSets(t *testing.T) {
	hamon := enums.HamonAdvanced
	unfiltered := JojoCharacterFiltersFingerprint(ports.JojoCharacterFilters{}, enums.EnGB)
	filtered := JojoCharacterFiltersFingerprint(ports.JojoCharacterFilters{Hamon: &hamon}, enums.EnGB)
	if unfiltered == filtered {
		t.Error("fingerprint must differ between an unfiltered and a filtered request")
	}

	encoded := EncodeCursor(JojoCharacterCursor{Name: "A"}, unfiltered)
	if _, err := DecodeCursor[JojoCharacterCursor](encoded, filtered); err == nil {
		t.Error("a cursor issued unfiltered must be rejected when replayed against a filtered request")
	}
}

// TestOnePieceCharacterFiltersFingerprint_DiffersAcrossFilterSets - see
// TestStandFiltersFingerprint_DiffersAcrossFilterSets's doc.
func TestOnePieceCharacterFiltersFingerprint_DiffersAcrossFilterSets(t *testing.T) {
	mastery := enums.FruitMasteryAdvanced
	unfiltered := OnePieceCharacterFiltersFingerprint(ports.OnePieceCharacterFilters{}, enums.EnGB)
	filtered := OnePieceCharacterFiltersFingerprint(ports.OnePieceCharacterFilters{FruitMastery: &mastery}, enums.EnGB)
	if unfiltered == filtered {
		t.Error("fingerprint must differ between an unfiltered and a filtered request")
	}

	encoded := EncodeCursor(OnePieceCharacterCursor{Name: "A"}, unfiltered)
	if _, err := DecodeCursor[OnePieceCharacterCursor](encoded, filtered); err == nil {
		t.Error("a cursor issued unfiltered must be rejected when replayed against a filtered request")
	}
}

// TestCharacterFiltersFingerprint_DifferentKindsDoNotCollide proves the two
// kinds' fingerprints are namespaced separately: a JoJo cursor must never
// decode successfully against a One Piece fingerprint even when both sides
// happen to have no filters set (the only field distinguishing the two
// FiltersFingerprint funcs then would otherwise be silent).
func TestCharacterFiltersFingerprint_DifferentKindsDoNotCollide(t *testing.T) {
	jojoFp := JojoCharacterFiltersFingerprint(ports.JojoCharacterFilters{}, enums.EnGB)
	opFp := OnePieceCharacterFiltersFingerprint(ports.OnePieceCharacterFilters{}, enums.EnGB)
	if jojoFp == opFp {
		t.Error("JojoCharacterFiltersFingerprint and OnePieceCharacterFiltersFingerprint must not collide when both are unfiltered")
	}
}
