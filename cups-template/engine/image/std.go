package imagecodec

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

// StdDecoder decodes PNG, JPEG and GIF via the Go standard library.
type StdDecoder struct{}

func (StdDecoder) Extensions() []string {
	return []string{"png", "jpg", "jpeg", "gif"}
}

func (StdDecoder) Decode(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}
