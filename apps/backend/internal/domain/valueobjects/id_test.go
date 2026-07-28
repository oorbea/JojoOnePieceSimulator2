package valueobjects_test

import (
	"errors"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
)

type testID [16]byte

func TestFormatParseRoundTrip(t *testing.T) {
	id := testID{0x01, 0x92, 0x83, 0xa4, 0xb5, 0xc6, 0xd7, 0xe8, 0xf9, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	str := valueobjects.Format(id)
	want := "019283a4-b5c6-d7e8-f900-112233445566"
	if str != want {
		t.Fatalf("Format() = %q, want %q", str, want)
	}

	got, err := valueobjects.Parse[testID](str)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got != id {
		t.Errorf("Parse() = %v, want %v", got, id)
	}
}

func TestParseBareHex(t *testing.T) {
	id := testID{0x01, 0x92, 0x83, 0xa4, 0xb5, 0xc6, 0xd7, 0xe8, 0xf9, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66}
	bare := "019283a4b5c6d7e8f900112233445566"

	got, err := valueobjects.Parse[testID](bare)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got != id {
		t.Errorf("Parse() = %v, want %v", got, id)
	}
}

func TestParseInvalid(t *testing.T) {
	cases := []string{
		"",
		"not-a-uuid",
		"019283a4-b5c6-d7e8-f900-11223344556",  // one hex digit short
		"019283a4-b5c6-d7e8-f900_112233445566", // wrong separator
		"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", // non-hex
	}
	for _, c := range cases {
		if _, err := valueobjects.Parse[testID](c); !errors.Is(err, valueobjects.ErrInvalidID) {
			t.Errorf("Parse(%q) err = %v, want ErrInvalidID", c, err)
		}
	}
}

func TestIsNil(t *testing.T) {
	var zero testID
	if !valueobjects.IsNil(zero) {
		t.Error("IsNil(zero) = false, want true")
	}

	nonZero := testID{1}
	if valueobjects.IsNil(nonZero) {
		t.Error("IsNil(nonZero) = true, want false")
	}
}
