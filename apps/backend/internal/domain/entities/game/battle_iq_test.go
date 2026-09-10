package game

import "testing"

func TestNoBattleIQ_IsAbsent(t *testing.T) {
	b := NoBattleIQ()
	if b.Present() {
		t.Fatalf("NoBattleIQ() should not be present")
	}
}

func TestBattleIQ_ZeroValueIsAbsent(t *testing.T) {
	var b BattleIQ
	if b.Present() {
		t.Fatalf("zero value BattleIQ should not be present")
	}
}

func TestNewBattleIQ_IsPresentIncludingZero(t *testing.T) {
	b := NewBattleIQ(0)
	if !b.Present() {
		t.Fatalf("NewBattleIQ(0) should be present")
	}
	if b.Value() != 0 {
		t.Fatalf("Value() = %d, want 0", b.Value())
	}

	b255 := NewBattleIQ(255)
	if !b255.Present() || b255.Value() != 255 {
		t.Fatalf("NewBattleIQ(255) = %+v, want present with value 255", b255)
	}
}

func TestBattleIQ_Band(t *testing.T) {
	tests := []struct {
		value byte
		want  BattleIQBand
	}{
		{0, BattleIQExtremelyLow},
		{69, BattleIQExtremelyLow},
		{70, BattleIQBorderline},
		{79, BattleIQBorderline},
		{80, BattleIQLowAverage},
		{89, BattleIQLowAverage},
		{90, BattleIQAverage},
		{109, BattleIQAverage},
		{110, BattleIQHighAverage},
		{119, BattleIQHighAverage},
		{120, BattleIQSuperior},
		{129, BattleIQSuperior},
		{130, BattleIQVerySuperior},
		{255, BattleIQVerySuperior},
	}
	for _, tt := range tests {
		if got := NewBattleIQ(tt.value).Band(); got != tt.want {
			t.Errorf("NewBattleIQ(%d).Band() = %v, want %v", tt.value, got, tt.want)
		}
	}
}
