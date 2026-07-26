package escpos

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"math"

	drivers "cups-drivers"
	_ "cups-drivers/models"
	"cups-template/engine/displaylist"
	"cups-template/engine/font"
	pngrenderer "cups-template/renderers/png"

	xdraw "golang.org/x/image/draw"
)

// DefaultMaxWidthDots is the printable width of most 80mm thermal heads
// at 203 DPI (72mm). Full 80mm paper is wider than the print head, so
// rasters must be capped or the right edge is clipped on the printer.
const DefaultMaxWidthDots = 576

type Options struct {
	Driver    drivers.Driver
	Threshold uint8
	FeedLines int
	// MaxWidthDots caps raster width (0 = DefaultMaxWidthDots).
	MaxWidthDots int
}

func DefaultOptions() Options {
	return Options{
		Threshold:    128,
		FeedLines:    3,
		MaxWidthDots: DefaultMaxWidthDots,
	}
}

type Renderer struct {
	Fonts   *font.Manager
	Options Options
}

func NewRenderer(fonts *font.Manager, opts Options) *Renderer {
	if opts.Threshold == 0 {
		opts.Threshold = 128
	}
	if opts.MaxWidthDots == 0 {
		opts.MaxWidthDots = DefaultMaxWidthDots
	}
	if opts.Driver == nil {
		opts.Driver = drivers.MustGet("generic")
	}
	return &Renderer{Fonts: fonts, Options: opts}
}

func (r *Renderer) Render(commands []displaylist.Command) ([]byte, error) {
	img := pngrenderer.NewRenderer(r.Fonts).Render(commands)
	return EncodeImage(img, r.Options)
}

func EncodeImage(img image.Image, opts Options) ([]byte, error) {
	if opts.Threshold == 0 {
		opts.Threshold = 128
	}
	if opts.MaxWidthDots == 0 {
		opts.MaxWidthDots = DefaultMaxWidthDots
	}
	d := opts.Driver
	if d == nil {
		d = drivers.MustGet("generic")
	}

	fitted := fitWidth(img, opts.MaxWidthDots)
	aligned := alignWidth(fitted)
	mono := toMonoBytes(aligned, opts.Threshold)

	var buf bytes.Buffer
	buf.Write(d.Init())
	if clear := d.ClearBuffer(); !bytes.Equal(clear, d.Init()) {
		buf.Write(clear)
	}
	buf.Write(d.AlignLeft())

	raster, err := d.Raster(aligned.Bounds().Dx(), aligned.Bounds().Dy(), mono)
	if err != nil {
		return nil, err
	}
	buf.Write(raster)

	return buf.Bytes(), nil
}

func fitWidth(src image.Image, maxW int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if maxW <= 0 || w <= maxW {
		return src
	}
	newH := int(math.Round(float64(h) * float64(maxW) / float64(w)))
	if newH < 1 {
		newH = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, maxW, newH))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, b, xdraw.Over, nil)
	return dst
}

func alignWidth(src image.Image) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	alignedW := (w + 7) &^ 7
	dst := image.NewRGBA(image.Rect(0, 0, alignedW, h))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(0, 0, w, h), src, b.Min, draw.Src)
	return dst
}

func toMonoBytes(img image.Image, threshold uint8) []byte {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	rowBytes := (w + 7) / 8
	out := make([]byte, rowBytes*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			lum := (r*299 + g*587 + bl*114) / 1000 >> 8
			if uint8(lum) < threshold {
				byteIndex := y*rowBytes + x/8
				bit := 7 - (x % 8)
				out[byteIndex] |= 1 << bit
			}
		}
	}
	return out
}
