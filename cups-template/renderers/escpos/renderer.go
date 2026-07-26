package escpos

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"

	drivers "cups-drivers"
	_ "cups-drivers/models"
	"cups-template/engine/displaylist"
	"cups-template/engine/font"
	pngrenderer "cups-template/renderers/png"
)

type Options struct {
	Driver    drivers.Driver
	Threshold uint8
	FeedLines int
}

func DefaultOptions() Options {
	return Options{
		Threshold: 128,
		FeedLines: 3,
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
	d := opts.Driver
	if d == nil {
		d = drivers.MustGet("generic")
	}

	aligned := alignWidth(img)
	mono := toMonoBytes(aligned, opts.Threshold)

	var buf bytes.Buffer
	buf.Write(d.Init())
	if clear := d.ClearBuffer(); !bytes.Equal(clear, d.Init()) {
		buf.Write(clear)
	}
	buf.Write(d.AlignCenter())

	raster, err := d.Raster(aligned.Bounds().Dx(), aligned.Bounds().Dy(), mono)
	if err != nil {
		return nil, err
	}
	buf.Write(raster)

	if feed := d.Feed(opts.FeedLines); len(feed) > 0 {
		buf.Write(feed)
	}

	return buf.Bytes(), nil
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
