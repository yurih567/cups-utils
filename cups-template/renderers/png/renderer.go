package png

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"cups-template/engine/displaylist"
	"cups-template/engine/font"
	"cups-template/engine/style"

	xdraw "golang.org/x/image/draw"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Renderer struct {
	Fonts         *font.Manager
	Background    color.Color
	PaddingBottom float64
}

func NewRenderer(fonts *font.Manager) *Renderer {
	return &Renderer{
		Fonts:      fonts,
		Background: color.White,
	}
}

func (r *Renderer) Render(commands []displaylist.Command) *image.RGBA {
	bounds := measureBounds(commands)
	w := int(math.Ceil(bounds.Width))
	h := int(math.Ceil(bounds.Height + r.PaddingBottom))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := r.Background
	if bg == nil {
		bg = color.White
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	for _, cmd := range commands {
		r.draw(img, cmd)
	}
	return img
}

func (r *Renderer) draw(dst *image.RGBA, cmd displaylist.Command) {
	switch c := cmd.(type) {
	case displaylist.DrawRect:
		drawRect(dst, c)
	case displaylist.DrawText:
		r.drawText(dst, c)
	case displaylist.DrawImage:
		drawImage(dst, c)
	case displaylist.DrawLine:
		drawLine(dst, c)
	case displaylist.DrawQRCode:
		drawPlaceholder(dst, c.X, c.Y, c.Width, c.Height, "QR", c.Data, c.Opacity)
	case displaylist.DrawBarcode:
		drawPlaceholder(dst, c.X, c.Y, c.Width, c.Height, "BARCODE", c.Data, c.Opacity)
	}
}

func measureBounds(commands []displaylist.Command) displaylist.DrawRect {
	maxX, maxY := 0.0, 0.0
	for _, cmd := range commands {
		var x2, y2 float64
		switch c := cmd.(type) {
		case displaylist.DrawRect:
			x2, y2 = c.X+c.Width, c.Y+c.Height
		case displaylist.DrawText:
			x2, y2 = c.X+c.Width, c.Y+c.Height
		case displaylist.DrawImage:
			x2, y2 = c.X+c.Width, c.Y+c.Height
		case displaylist.DrawLine:
			x2 = math.Max(c.X1, c.X2)
			y2 = math.Max(c.Y1, c.Y2) + c.StrokeWidth
		case displaylist.DrawQRCode:
			x2, y2 = c.X+c.Width, c.Y+c.Height
		case displaylist.DrawBarcode:
			x2, y2 = c.X+c.Width, c.Y+c.Height
		}
		if x2 > maxX {
			maxX = x2
		}
		if y2 > maxY {
			maxY = y2
		}
	}
	return displaylist.DrawRect{Width: maxX, Height: maxY}
}

func drawRect(dst *image.RGBA, c displaylist.DrawRect) {
	if c.Width <= 0 || c.Height <= 0 {
		return
	}
	op := c.Opacity
	if op <= 0 {
		op = 1
	}

	if c.Fill != nil {
		col := applyOpacity(c.Fill, op)
		fillRounded(dst, c.X, c.Y, c.Width, c.Height, c.Radius, col)
	}
	if c.Stroke != nil && c.StrokeWidth > 0 {
		col := applyOpacity(c.Stroke, op)
		strokeRounded(dst, c.X, c.Y, c.Width, c.Height, c.Radius, c.StrokeWidth, col)
	}
}

func (r *Renderer) drawText(dst *image.RGBA, c displaylist.DrawText) {
	if r.Fonts == nil || c.Text == "" {
		return
	}
	face, err := r.Fonts.Face(c.FontFamily, c.FontWeight, c.FontSize)
	if err != nil {
		return
	}

	op := c.Opacity
	if op <= 0 {
		op = 1
	}
	col := applyOpacity(c.Color, op)
	metrics := face.Metrics()
	ascent := float64(metrics.Ascent) / 64
	lineH := float64(metrics.Height) / 64
	if lineH <= 0 {
		lineH = ascent + float64(metrics.Descent)/64
	}

	y := c.Y + ascent
	for i, line := range strings.Split(c.Text, "\n") {
		if i > 0 {
			y += lineH
		}
		textW := float64(xfont.MeasureString(face, line)) / 64
		x := c.X
		switch style.Align(c.Align) {
		case style.AlignCenter:
			x = c.X + (c.Width-textW)/2
		case style.AlignEnd:
			x = c.X + c.Width - textW
		}
		d := &xfont.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(col),
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)},
		}
		d.DrawString(line)
	}
}

func drawImage(dst *image.RGBA, c displaylist.DrawImage) {
	if c.Image == nil || c.Width <= 0 || c.Height <= 0 {
		return
	}
	target := image.Rect(
		int(math.Round(c.X)),
		int(math.Round(c.Y)),
		int(math.Round(c.X+c.Width)),
		int(math.Round(c.Y+c.Height)),
	)
	scaled := image.NewRGBA(image.Rect(0, 0, target.Dx(), target.Dy()))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), c.Image, c.Image.Bounds(), xdraw.Over, nil)

	op := c.Opacity
	if op <= 0 || op >= 1 {
		draw.Draw(dst, target, scaled, image.Point{}, draw.Over)
		return
	}
	draw.DrawMask(dst, target, scaled, image.Point{}, &image.Uniform{C: color.Alpha{A: uint8(op * 255)}}, image.Point{}, draw.Over)
}

func drawLine(dst *image.RGBA, c displaylist.DrawLine) {
	op := c.Opacity
	if op <= 0 {
		op = 1
	}
	col := applyOpacity(c.Color, op)
	thickness := c.StrokeWidth
	if thickness <= 0 {
		thickness = 1
	}
	x0, y0 := c.X1, c.Y1
	x1, y1 := c.X2, c.Y2
	dx := x1 - x0
	dy := y1 - y0
	steps := int(math.Ceil(math.Hypot(dx, dy)))
	if steps < 1 {
		steps = 1
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := x0 + dx*t
		y := y0 + dy*t
		fillCircle(dst, x, y, thickness/2, col)
	}
}

func drawPlaceholder(dst *image.RGBA, x, y, w, h float64, label, data string, opacity float64) {
	op := opacity
	if op <= 0 {
		op = 1
	}
	border := color.RGBA{R: 60, G: 60, B: 60, A: 255}
	fill := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	drawRect(dst, displaylist.DrawRect{
		X: x, Y: y, Width: w, Height: h,
		Fill: applyOpacity(fill, op), Stroke: applyOpacity(border, op), StrokeWidth: 1, Radius: 4,
	})

	pattern := color.RGBA{R: 40, G: 40, B: 40, A: 220}
	cell := math.Max(4, math.Min(w, h)/12)
	for row := 0; float64(row)*cell < h-8; row++ {
		for col := 0; float64(col)*cell < w-8; col++ {
			if (row+col)%2 == 0 {
				px := x + 4 + float64(col)*cell
				py := y + 4 + float64(row)*cell
				fillRounded(dst, px, py, cell*0.7, cell*0.7, 0, applyOpacity(pattern, op))
			}
		}
	}

	_ = label
	_ = data
}

func fillRounded(dst *image.RGBA, x, y, w, h, radius float64, col color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	r := math.Min(radius, math.Min(w, h)/2)
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := int(math.Ceil(x + w))
	y1 := int(math.Ceil(y + h))

	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if insideRounded(float64(px)+0.5, float64(py)+0.5, x, y, w, h, r) {
				dst.Set(px, py, alphaOver(dst.At(px, py), col))
			}
		}
	}
}

func strokeRounded(dst *image.RGBA, x, y, w, h, radius, stroke float64, col color.Color) {
	outer := displaylist.DrawRect{X: x, Y: y, Width: w, Height: h, Radius: radius}
	_ = outer
	inset := stroke
	fillRounded(dst, x, y, w, h, radius, col)
	if w > inset*2 && h > inset*2 {
		innerR := math.Max(0, radius-inset)
		clearRounded(dst, x+inset, y+inset, w-inset*2, h-inset*2, innerR)
	}
}

func clearRounded(dst *image.RGBA, x, y, w, h, radius float64) {
	r := math.Min(radius, math.Min(w, h)/2)
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := int(math.Ceil(x + w))
	y1 := int(math.Ceil(y + h))
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if insideRounded(float64(px)+0.5, float64(py)+0.5, x, y, w, h, r) {
				dst.Set(px, py, color.White)
			}
		}
	}
}

func insideRounded(px, py, x, y, w, h, r float64) bool {
	if px < x || py < y || px > x+w || py > y+h {
		return false
	}
	if r <= 0 {
		return true
	}
	cx := clamp(px, x+r, x+w-r)
	cy := clamp(py, y+r, y+h-r)
	if px >= x+r && px <= x+w-r {
		return true
	}
	if py >= y+r && py <= y+h-r {
		return true
	}
	dx := px - cx
	dy := py - cy
	return dx*dx+dy*dy <= r*r
}

func fillCircle(dst *image.RGBA, cx, cy, radius float64, col color.Color) {
	if radius <= 0 {
		return
	}
	r := int(math.Ceil(radius))
	ix, iy := int(math.Round(cx)), int(math.Round(cy))
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if float64(x*x+y*y) <= radius*radius {
				px, py := ix+x, iy+y
				if image.Pt(px, py).In(dst.Bounds()) {
					dst.Set(px, py, alphaOver(dst.At(px, py), col))
				}
			}
		}
	}
}

func applyOpacity(c color.Color, opacity float64) color.Color {
	if c == nil {
		return color.Transparent
	}
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(float64(a>>8) * clamp(opacity, 0, 1)),
	}
}

func alphaOver(dstC, srcC color.Color) color.Color {
	sr, sg, sb, sa := srcC.RGBA()
	dr, dg, db, da := dstC.RGBA()
	a := sa
	inv := 0xffff - a
	return color.RGBA64{
		R: uint16((sr*a + dr*inv) / 0xffff),
		G: uint16((sg*a + dg*inv) / 0xffff),
		B: uint16((sb*a + db*inv) / 0xffff),
		A: uint16(a + (da*inv)/0xffff),
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
