package imagecodec

import (
	"fmt"
	"image"
	"io"

	"golang.org/x/image/webp"
)

type WebPDecoder struct{}

func (WebPDecoder) Extensions() []string {
	return []string{"webp"}
}

func (WebPDecoder) Decode(r io.Reader) (image.Image, error) {
	img, err := webp.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode webp: %w", err)
	}
	return img, nil
}
