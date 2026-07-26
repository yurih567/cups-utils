package imagecodec

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Decoder interface {
	Extensions() []string
	Decode(r io.Reader) (image.Image, error)
}

type Registry struct {
	decoders []Decoder
	client   *http.Client
}

func NewRegistry(decoders ...Decoder) *Registry {
	return &Registry{
		decoders: decoders,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (r *Registry) Register(d Decoder) {
	r.decoders = append(r.decoders, d)
}

func (r *Registry) DecodeFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return r.Decode(filepath.Ext(path), f)
}

func (r *Registry) DecodeURL(rawURL string) (image.Image, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch image: %w", err)
	}
	req.Header.Set("User-Agent", "cups-template/1.0")
	req.Header.Set("Accept", "image/svg+xml,image/webp,image/*;q=0.8,*/*;q=0.5")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch image: status %d", resp.StatusCode)
	}

	ext := extensionFromURL(rawURL)
	if ext == "" {
		ext = extensionFromContentType(resp.Header.Get("Content-Type"))
	}
	if ext == "" {
		return nil, fmt.Errorf("could not determine image format for %q", rawURL)
	}

	return r.Decode(ext, resp.Body)
}

func (r *Registry) DecodeDataURI(src string) (image.Image, error) {
	if !strings.HasPrefix(src, "data:") {
		return nil, fmt.Errorf("not a data URI")
	}

	payload := strings.TrimPrefix(src, "data:")
	comma := strings.IndexByte(payload, ',')
	if comma < 0 {
		return nil, fmt.Errorf("invalid data URI")
	}

	meta := payload[:comma]
	dataPart := payload[comma+1:]

	ext := extensionFromContentType(strings.Split(meta, ";")[0])
	if ext == "" {
		return nil, fmt.Errorf("unsupported data URI media type %q", meta)
	}

	var reader io.Reader
	if strings.Contains(meta, ";base64") {
		raw, err := base64.StdEncoding.DecodeString(dataPart)
		if err != nil {
			raw, err = base64.RawStdEncoding.DecodeString(dataPart)
			if err != nil {
				return nil, fmt.Errorf("decode data URI base64: %w", err)
			}
		}
		reader = bytes.NewReader(raw)
	} else {
		decoded, err := url.QueryUnescape(dataPart)
		if err != nil {
			return nil, fmt.Errorf("decode data URI payload: %w", err)
		}
		reader = strings.NewReader(decoded)
	}

	return r.Decode(ext, reader)
}

func (r *Registry) DecodeSource(src, assetBasePath string) (image.Image, error) {
	switch {
	case strings.HasPrefix(src, "data:"):
		return r.DecodeDataURI(src)
	case isRemoteURL(src):
		return r.DecodeURL(src)
	default:
		path := src
		if !filepath.IsAbs(path) && assetBasePath != "" {
			path = filepath.Join(assetBasePath, path)
		}
		return r.DecodeFile(path)
	}
}

func (r *Registry) Decode(ext string, reader io.Reader) (image.Image, error) {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	for _, d := range r.decoders {
		for _, supported := range d.Extensions() {
			if supported == ext {
				return d.Decode(reader)
			}
		}
	}
	return nil, fmt.Errorf("unsupported image format %q", ext)
}

func DefaultRegistry() *Registry {
	return NewRegistry(StdDecoder{}, SVGDecoder{}, WebPDecoder{})
}

func isRemoteURL(src string) bool {
	u, err := url.Parse(src)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func extensionFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(u.Path), "."))
}

func extensionFromContentType(ct string) string {
	ct = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	switch ct {
	case "image/svg+xml":
		return "svg"
	case "image/webp":
		return "webp"
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/gif":
		return "gif"
	default:
		return ""
	}
}
