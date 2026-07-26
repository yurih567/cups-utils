package engine

import (
	"cups-template/engine/displaylist"
	"cups-template/engine/dom"
	"cups-template/engine/font"
	imagecodec "cups-template/engine/image"
	"cups-template/engine/layout"
)

type Options struct {
	DPI           float64
	Fonts         *font.Manager
	Images        *imagecodec.Registry
	AssetBasePath string
}

func Layout(doc *dom.Document, opts Options) (*layout.Tree, error) {
	return layout.Compute(doc, layout.Options{
		DPI:           opts.DPI,
		Fonts:         opts.Fonts,
		Images:        opts.Images,
		AssetBasePath: opts.AssetBasePath,
	})
}

func BuildDisplayList(tree *layout.Tree) []displaylist.Command {
	return displaylist.Build(tree)
}
