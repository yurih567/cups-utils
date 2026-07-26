package units

import (
	"fmt"
	"strconv"
	"strings"
)

type Unit int

const (
	UnitPx Unit = iota
	UnitMm
	UnitPercent
	UnitAuto
)

type Length struct {
	Value float64
	Unit  Unit
}

func Auto() Length {
	return Length{Unit: UnitAuto}
}

func Px(v float64) Length {
	return Length{Value: v, Unit: UnitPx}
}

func Mm(v float64) Length {
	return Length{Value: v, Unit: UnitMm}
}

func Percent(v float64) Length {
	return Length{Value: v, Unit: UnitPercent}
}

func (l Length) IsAuto() bool {
	return l.Unit == UnitAuto
}

func (l Length) IsPercent() bool {
	return l.Unit == UnitPercent
}

func (l Length) IsZero() bool {
	return l.Value == 0 && l.Unit != UnitAuto
}

func (l Length) ToPixels(dpi, parentPx float64) float64 {
	switch l.Unit {
	case UnitPx:
		return l.Value
	case UnitMm:
		return l.Value * dpi / 25.4
	case UnitPercent:
		return parentPx * l.Value / 100
	default:
		return 0
	}
}

func Parse(s string) (Length, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "auto") {
		return Auto(), nil
	}

	switch {
	case strings.HasSuffix(s, "%"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if err != nil {
			return Length{}, fmt.Errorf("invalid percent length %q: %w", s, err)
		}
		return Percent(v), nil
	case strings.HasSuffix(s, "mm"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "mm"), 64)
		if err != nil {
			return Length{}, fmt.Errorf("invalid mm length %q: %w", s, err)
		}
		return Mm(v), nil
	case strings.HasSuffix(s, "px"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "px"), 64)
		if err != nil {
			return Length{}, fmt.Errorf("invalid px length %q: %w", s, err)
		}
		return Px(v), nil
	default:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return Length{}, fmt.Errorf("invalid length %q: %w", s, err)
		}
		return Px(v), nil
	}
}

func MustParse(s string) Length {
	l, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return l
}
