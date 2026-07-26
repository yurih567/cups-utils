package units

import (
	"math"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		unit Unit
		val  float64
	}{
		{"10px", UnitPx, 10},
		{"80mm", UnitMm, 80},
		{"50%", UnitPercent, 50},
		{"12", UnitPx, 12},
		{"auto", UnitAuto, 0},
	}

	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.in, err)
		}
		if got.Unit != tc.unit || got.Value != tc.val {
			t.Fatalf("Parse(%q) = %+v, want unit=%v val=%v", tc.in, got, tc.unit, tc.val)
		}
	}
}

func TestToPixels(t *testing.T) {
	got := Mm(25.4).ToPixels(96, 0)
	if math.Abs(got-96) > 0.001 {
		t.Fatalf("25.4mm @96dpi = %v, want 96", got)
	}
	if got := Percent(50).ToPixels(96, 200); got != 100 {
		t.Fatalf("50%% of 200 = %v, want 100", got)
	}
}
