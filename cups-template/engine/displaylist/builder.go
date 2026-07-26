package displaylist

import (
	"image/color"
	"math"

	"cups-template/engine/dom"
	"cups-template/engine/layout"
	"cups-template/engine/units"
)

func Build(tree *layout.Tree) []Command {
	if tree == nil || tree.Root == nil {
		return nil
	}
	var cmds []Command
	// Emit an invisible page rect so renderers keep the full page size
	// (right/bottom padding would otherwise be cropped by content bounds).
	box := tree.Root.BorderBox()
	cmds = append(cmds, DrawRect{
		X:      box.X,
		Y:      box.Y,
		Width:  box.Width,
		Height: box.Height,
	})
	walk(tree.Root, tree.DPI, &cmds)
	return cmds
}

func walk(n *layout.Node, dpi float64, cmds *[]Command) {
	emitBackground(n, dpi, cmds)

	switch n.DOM.Type {
	case dom.TypeText:
		emitText(n, dpi, cmds)
	case dom.TypeImage:
		emitImage(n, cmds)
	case dom.TypeDivider:
		emitDivider(n, dpi, cmds)
	case dom.TypeQRCode:
		emitQRCode(n, cmds)
	case dom.TypeBarcode:
		emitBarcode(n, cmds)
	}

	emitBorder(n, dpi, cmds)

	for _, child := range n.Children {
		walk(child, dpi, cmds)
	}

	if n.DOM.Type == dom.TypeRow {
		emitRowLeaders(n, cmds)
	}
}

func emitRowLeaders(n *layout.Node, cmds *[]Command) {
	if len(n.Leaders) == 0 {
		return
	}
	c := n.DOM.Style.Color
	if c == nil {
		c = color.Black
	}
	for _, leader := range n.Leaders {
		*cmds = append(*cmds, DrawText{
			X:          leader.X,
			Y:          leader.Y,
			Width:      leader.Width,
			Height:     leader.Height,
			Text:       leader.Text,
			FontFamily: leader.FontFamily,
			FontSize:   leader.FontSize,
			FontWeight: leader.FontWeight,
			Color:      c,
			Align:      int(0), // start
			Opacity:    n.DOM.Style.Opacity,
		})
	}
}

func emitBackground(n *layout.Node, dpi float64, cmds *[]Command) {
	bg := n.DOM.Style.Background
	if bg == nil || isTransparent(bg) {
		return
	}
	box := n.BorderBox()
	*cmds = append(*cmds, DrawRect{
		X:       box.X,
		Y:       box.Y,
		Width:   box.Width,
		Height:  box.Height,
		Fill:    bg,
		Radius:  resolve(n.DOM.Style.BorderRadius, dpi, box.Width),
		Opacity: n.DOM.Style.Opacity,
		Rotate:  n.DOM.Style.Rotate,
	})
}

func emitBorder(n *layout.Node, dpi float64, cmds *[]Command) {
	stroke := resolve(n.DOM.Style.Border.Width, dpi, 0)
	if stroke <= 0 {
		return
	}
	box := n.BorderBox()
	*cmds = append(*cmds, DrawRect{
		X:           box.X,
		Y:           box.Y,
		Width:       box.Width,
		Height:      box.Height,
		Stroke:      n.DOM.Style.Border.Color,
		StrokeWidth: stroke,
		Radius:      resolve(n.DOM.Style.BorderRadius, dpi, box.Width),
		Opacity:     n.DOM.Style.Opacity,
		Rotate:      n.DOM.Style.Rotate,
	})
}

func emitText(n *layout.Node, dpi float64, cmds *[]Command) {
	box := n.ContentBox()
	fontSize := resolve(n.DOM.Style.FontSize, dpi, 0)
	if fontSize <= 0 {
		fontSize = 14
	}
	*cmds = append(*cmds, DrawText{
		X:          box.X,
		Y:          box.Y,
		Width:      box.Width,
		Height:     box.Height,
		Text:       n.Text,
		FontFamily: n.DOM.Style.Font,
		FontSize:   fontSize,
		FontWeight: int(n.DOM.Style.FontWeight),
		Color:      n.DOM.Style.Color,
		Align:      int(n.DOM.Style.Align),
		Opacity:    n.DOM.Style.Opacity,
		Rotate:     n.DOM.Style.Rotate,
	})
}

func emitImage(n *layout.Node, cmds *[]Command) {
	if n.Image == nil {
		return
	}
	box := n.ContentBox()
	*cmds = append(*cmds, DrawImage{
		X:       box.X,
		Y:       box.Y,
		Width:   box.Width,
		Height:  box.Height,
		Image:   n.Image,
		Src:     n.DOM.Style.Src,
		Opacity: n.DOM.Style.Opacity,
		Rotate:  n.DOM.Style.Rotate,
	})
}

func emitDivider(n *layout.Node, dpi float64, cmds *[]Command) {
	box := n.ContentBox()

	if n.DOM.Style.Char != "" {
		fontSize := resolve(n.DOM.Style.FontSize, dpi, 0)
		if fontSize <= 0 {
			fontSize = 14
		}
		c := n.DOM.Style.Color
		if c == nil {
			c = color.Black
		}
		*cmds = append(*cmds, DrawText{
			X:          box.X,
			Y:          box.Y,
			Width:      box.Width,
			Height:     box.Height,
			Text:       n.Text,
			FontFamily: n.DOM.Style.Font,
			FontSize:   fontSize,
			FontWeight: int(n.DOM.Style.FontWeight),
			Color:      c,
			Align:      int(n.DOM.Style.Align),
			Opacity:    n.DOM.Style.Opacity,
			Rotate:     n.DOM.Style.Rotate,
		})
		return
	}

	stroke := resolve(n.DOM.Style.Border.Width, dpi, 0)
	if stroke <= 0 {
		stroke = 1
	}
	c := n.DOM.Style.Color
	if c == nil {
		c = color.RGBA{R: 180, G: 180, B: 180, A: 255}
	}
	y := box.Y + box.Height/2
	*cmds = append(*cmds, DrawLine{
		X1:          box.X,
		Y1:          y,
		X2:          box.X + box.Width,
		Y2:          y,
		Color:       c,
		StrokeWidth: stroke,
		Opacity:     n.DOM.Style.Opacity,
	})
}

func emitQRCode(n *layout.Node, cmds *[]Command) {
	box := n.ContentBox()
	data := n.DOM.Style.Data
	if data == "" {
		data = n.Text
	}
	*cmds = append(*cmds, DrawQRCode{
		X:       box.X,
		Y:       box.Y,
		Width:   box.Width,
		Height:  box.Height,
		Data:    data,
		Opacity: n.DOM.Style.Opacity,
	})
}

func emitBarcode(n *layout.Node, cmds *[]Command) {
	box := n.ContentBox()
	data := n.DOM.Style.Data
	if data == "" {
		data = n.Text
	}
	*cmds = append(*cmds, DrawBarcode{
		X:       box.X,
		Y:       box.Y,
		Width:   box.Width,
		Height:  box.Height,
		Data:    data,
		Format:  "code128",
		Opacity: n.DOM.Style.Opacity,
	})
}

func resolve(l units.Length, dpi, parent float64) float64 {
	if l.IsAuto() {
		return 0
	}
	v := l.ToPixels(dpi, parent)
	if math.IsNaN(v) {
		return 0
	}
	return v
}

func isTransparent(c color.Color) bool {
	_, _, _, a := c.RGBA()
	return a == 0
}
