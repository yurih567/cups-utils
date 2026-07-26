package style

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"cups-template/engine/units"
)

type Align int

const (
	AlignStart Align = iota
	AlignCenter
	AlignEnd
)

type Justify int

const (
	JustifyStart Justify = iota
	JustifyCenter
	JustifyEnd
	JustifyBetween
	JustifyAround
)

type FontWeight int

const (
	FontWeightNormal FontWeight = 400
	FontWeightBold   FontWeight = 700
)

type Edges struct {
	Top    units.Length
	Right  units.Length
	Bottom units.Length
	Left   units.Length
}

type Border struct {
	Width units.Length
	Color color.Color
}

type Style struct {
	Width        units.Length
	Height       units.Length
	Padding      Edges
	Margin       Edges
	Gap          units.Length
	Background   color.Color
	Color        color.Color
	Font         string
	FontSize     units.Length
	FontWeight   FontWeight
	Align        Align
	Justify      Justify
	Border       Border
	BorderRadius units.Length
	Opacity      float64
	Rotate       float64
	Src          string
	Data         string
	Char         string
	FillWeight   FontWeight
}

func Default() Style {
	return Style{
		Width:      units.Auto(),
		Height:     units.Auto(),
		Font:       "Arial",
		FontSize:   units.Px(12),
		FontWeight: FontWeightNormal,
		Align:      AlignStart,
		Justify:    JustifyStart,
		Color:      color.Black,
		Opacity:    1,
		Border: Border{
			Color: color.Black,
		},
	}
}

func ParseAttrs(attrs map[string]string) (Style, error) {
	s := Default()

	for key, raw := range attrs {
		k := strings.ToLower(strings.TrimSpace(key))
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}

		switch k {
		case "width":
			l, err := units.Parse(v)
			if err != nil {
				return s, err
			}
			s.Width = l
		case "height":
			l, err := units.Parse(v)
			if err != nil {
				return s, err
			}
			s.Height = l
		case "padding":
			e, err := parseEdges(v)
			if err != nil {
				return s, err
			}
			s.Padding = e
		case "margin":
			e, err := parseEdges(v)
			if err != nil {
				return s, err
			}
			s.Margin = e
		case "gap":
			l, err := units.Parse(v)
			if err != nil {
				return s, err
			}
			s.Gap = l
		case "background":
			c, err := ParseColor(v)
			if err != nil {
				return s, err
			}
			s.Background = c
		case "color":
			c, err := ParseColor(v)
			if err != nil {
				return s, err
			}
			s.Color = c
		case "font":
			s.Font = v
		case "font-size", "size":
			l, err := units.Parse(v)
			if err != nil {
				return s, err
			}
			s.FontSize = l
		case "font-weight", "weight":
			w, err := parseWeight(v)
			if err != nil {
				return s, err
			}
			s.FontWeight = w
		case "align":
			a, err := parseAlign(v)
			if err != nil {
				return s, err
			}
			s.Align = a
		case "justify":
			j, err := parseJustify(v)
			if err != nil {
				return s, err
			}
			s.Justify = j
		case "border":
			b, err := parseBorder(v)
			if err != nil {
				return s, err
			}
			s.Border = b
		case "border-radius":
			l, err := units.Parse(v)
			if err != nil {
				return s, err
			}
			s.BorderRadius = l
		case "opacity":
			op, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return s, fmt.Errorf("invalid opacity %q: %w", v, err)
			}
			s.Opacity = op
		case "rotate":
			r, err := strconv.ParseFloat(strings.TrimSuffix(v, "deg"), 64)
			if err != nil {
				return s, fmt.Errorf("invalid rotate %q: %w", v, err)
			}
			s.Rotate = r
		case "src":
			s.Src = v
		case "data", "value":
			s.Data = v
		case "char", "fill", "whitespace", "leader":
			s.Char = firstGrapheme(v)
		case "fill-weight", "leader-weight":
			w, err := parseWeight(v)
			if err != nil {
				return s, err
			}
			s.FillWeight = w
		}
	}

	return s, nil
}

func parseEdges(v string) (Edges, error) {
	parts := strings.Fields(v)
	switch len(parts) {
	case 1:
		l, err := units.Parse(parts[0])
		if err != nil {
			return Edges{}, err
		}
		return Edges{Top: l, Right: l, Bottom: l, Left: l}, nil
	case 2:
		vert, err := units.Parse(parts[0])
		if err != nil {
			return Edges{}, err
		}
		horiz, err := units.Parse(parts[1])
		if err != nil {
			return Edges{}, err
		}
		return Edges{Top: vert, Right: horiz, Bottom: vert, Left: horiz}, nil
	case 4:
		top, err := units.Parse(parts[0])
		if err != nil {
			return Edges{}, err
		}
		right, err := units.Parse(parts[1])
		if err != nil {
			return Edges{}, err
		}
		bottom, err := units.Parse(parts[2])
		if err != nil {
			return Edges{}, err
		}
		left, err := units.Parse(parts[3])
		if err != nil {
			return Edges{}, err
		}
		return Edges{Top: top, Right: right, Bottom: bottom, Left: left}, nil
	default:
		return Edges{}, fmt.Errorf("invalid edges value %q", v)
	}
}

func parseWeight(v string) (FontWeight, error) {
	switch strings.ToLower(v) {
	case "normal", "regular", "400":
		return FontWeightNormal, nil
	case "bold", "700":
		return FontWeightBold, nil
	default:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid font-weight %q", v)
		}
		return FontWeight(n), nil
	}
}

func parseAlign(v string) (Align, error) {
	switch strings.ToLower(v) {
	case "start", "left":
		return AlignStart, nil
	case "center":
		return AlignCenter, nil
	case "end", "right":
		return AlignEnd, nil
	default:
		return 0, fmt.Errorf("invalid align %q", v)
	}
}

func parseJustify(v string) (Justify, error) {
	switch strings.ToLower(v) {
	case "start", "left":
		return JustifyStart, nil
	case "center":
		return JustifyCenter, nil
	case "end", "right":
		return JustifyEnd, nil
	case "between", "space-between":
		return JustifyBetween, nil
	case "around", "space-around":
		return JustifyAround, nil
	default:
		return 0, fmt.Errorf("invalid justify %q", v)
	}
}

func parseBorder(v string) (Border, error) {
	parts := strings.Fields(v)
	if len(parts) == 0 {
		return Border{}, fmt.Errorf("invalid border %q", v)
	}
	width, err := units.Parse(parts[0])
	if err != nil {
		return Border{}, err
	}
	b := Border{Width: width, Color: color.Black}
	if len(parts) >= 2 {
		c, err := ParseColor(parts[len(parts)-1])
		if err != nil {
			return Border{}, err
		}
		b.Color = c
	}
	return b, nil
}

func ParseColor(v string) (color.Color, error) {
	v = strings.TrimSpace(strings.ToLower(v))
	switch v {
	case "black":
		return color.Black, nil
	case "white":
		return color.White, nil
	case "transparent", "none":
		return color.Transparent, nil
	case "red":
		return color.RGBA{R: 220, G: 38, B: 38, A: 255}, nil
	case "gray", "grey":
		return color.RGBA{R: 107, G: 114, B: 128, A: 255}, nil
	}

	if strings.HasPrefix(v, "#") {
		hex := strings.TrimPrefix(v, "#")
		var r, g, b, a uint8 = 0, 0, 0, 255
		switch len(hex) {
		case 3:
			var rv, gv, bv uint8
			_, err := fmt.Sscanf(hex, "%1x%1x%1x", &rv, &gv, &bv)
			if err != nil {
				return nil, fmt.Errorf("invalid color %q", v)
			}
			r, g, b = rv*17, gv*17, bv*17
		case 6:
			var rv, gv, bv uint32
			_, err := fmt.Sscanf(hex, "%02x%02x%02x", &rv, &gv, &bv)
			if err != nil {
				return nil, fmt.Errorf("invalid color %q", v)
			}
			r, g, b = uint8(rv), uint8(gv), uint8(bv)
		case 8:
			var rv, gv, bv, av uint32
			_, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &rv, &gv, &bv, &av)
			if err != nil {
				return nil, fmt.Errorf("invalid color %q", v)
			}
			r, g, b, a = uint8(rv), uint8(gv), uint8(bv), uint8(av)
		default:
			return nil, fmt.Errorf("invalid color %q", v)
		}
		return color.RGBA{R: r, G: g, B: b, A: a}, nil
	}

	return nil, fmt.Errorf("invalid color %q", v)
}

func firstGrapheme(v string) string {
	for _, r := range v {
		return string(r)
	}
	return ""
}
