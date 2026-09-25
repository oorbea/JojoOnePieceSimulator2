package dto

import "time"

// wireTimeLayout formats every wire timestamp (closesAt, *EndsAt,
// serverTime) with millisecond precision instead of time.RFC3339's
// whole-second truncation. A deadline or clock sample rounded down to the
// second can be off by up to 999ms - fine for a human reading a countdown,
// not fine for a client trying to measure its own clock's offset from the
// server's (see shared/lib/server-clock.ts on the frontend). Fixed-width
// (always exactly 3 fractional digits, unlike time.RFC3339Nano which trims
// trailing zeros) so every wire timestamp has the same shape.
const wireTimeLayout = "2006-01-02T15:04:05.000Z07:00"

// FormatWireTime renders t on the wire. See wireTimeLayout's doc for why
// this isn't time.RFC3339. Exported: infrastructure/api/endpoints stamps
// frame deadlines and ServerFrame.ServerTime with it too.
func FormatWireTime(t time.Time) string {
	return t.UTC().Format(wireTimeLayout)
}
