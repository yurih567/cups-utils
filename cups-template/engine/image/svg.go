package imagecodec

import (
	"fmt"
	"image"
	"io"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

type SVGDecoder struct{}

func (SVGDecoder) Extensions() []string {
	return []string{"svg"}
}

func (SVGDecoder) Decode(r io.Reader) (image.Image, error) {
	icon, err := oksvg.ReadIconStream(r)
	if err != nil {
		return nil, fmt.Errorf("decode svg: %w", err)
	}

	w := int(icon.ViewBox.W)
	h := int(icon.ViewBox.H)
	if w <= 0 {
		w = 128
	}
	if h <= 0 {
		h = 128
	}

	icon.SetTarget(0, 0, float64(w), float64(h))
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.Draw(raster, 1)
	return rgba, nil
}
