package dto_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// TestFormatWireTime_MillisecondPrecision guards the reason wireTimeLayout
// replaced time.RFC3339: a deadline/clock-sample rounded down to the whole
// second (time.RFC3339's behavior) can be off by up to 999ms, which is
// exactly the drift a client tries to measure when computing its own clock
// offset from the server's (see shared/lib/server-clock.ts). Also checks the
// fractional part is always exactly 3 digits, never trimmed like
// time.RFC3339Nano.
func TestFormatWireTime_MillisecondPrecision(t *testing.T) {
	moment := time.Date(2026, 9, 25, 10, 30, 0, 123_000_000, time.UTC)
	got := dto.FormatWireTime(moment)
	want := "2026-09-25T10:30:00.123Z"
	if got != want {
		t.Fatalf("FormatWireTime = %q, want %q", got, want)
	}

	// A whole-second instant still renders 3 fractional digits, not zero.
	wholeSecond := time.Date(2026, 9, 25, 10, 30, 0, 0, time.UTC)
	wholeSecondWire := dto.FormatWireTime(wholeSecond)
	if wholeSecondWire != "2026-09-25T10:30:00.000Z" {
		t.Fatalf("FormatWireTime (whole second) = %q, want fixed-width fractional digits", wholeSecondWire)
	}

	parsed, err := time.Parse(time.RFC3339Nano, got)
	if err != nil {
		t.Fatalf("wire format must remain RFC3339Nano-parseable: %v", err)
	}
	if !parsed.Equal(moment) {
		t.Fatalf("round-tripped time = %v, want %v", parsed, moment)
	}
}

// TestNewServerFrame_StampsServerTime guards the reason ServerFrame carries
// a serverTime on every frame (not just the five timed ones): a client
// samples it against its own receipt clock to compute a clock offset (see
// shared/lib/server-clock.ts), so it must be present, parseable, and close
// to "now" for a frame built with no other timestamp involved at all.
func TestNewServerFrame_StampsServerTime(t *testing.T) {
	before := time.Now()
	frame := dto.NewServerFrame(dto.FrameGameStarted, "", struct{}{})
	after := time.Now()

	got, err := time.Parse(time.RFC3339Nano, frame.ServerTime)
	if err != nil {
		t.Fatalf("ServerTime not parseable: %v (%q)", err, frame.ServerTime)
	}
	if got.Before(before.Add(-time.Second)) || got.After(after.Add(time.Second)) {
		t.Fatalf("ServerTime = %v, want within [%v, %v]", got, before, after)
	}

	raw, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var roundTripped dto.ServerFrame
	if err := json.Unmarshal(raw, &roundTripped); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if roundTripped.ServerTime != frame.ServerTime {
		t.Fatalf("serverTime did not round-trip through JSON: got %q, want %q", roundTripped.ServerTime, frame.ServerTime)
	}
}
