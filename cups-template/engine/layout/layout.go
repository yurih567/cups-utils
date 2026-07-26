package layout

import (
	"fmt"
	"image"
	"math"
	"strings"

	"cups-template/engine/dom"
	"cups-template/engine/font"
	imagecodec "cups-template/engine/image"
	"cups-template/engine/style"
	"cups-template/engine/units"
)

type Options struct {
	DPI           float64
	Fonts         *font.Manager
	Images        *imagecodec.Registry
	AssetBasePath string
}

type Box struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type EdgesPx struct {
	Top, Right, Bottom, Left float64
}

type Node struct {
	DOM      *dom.Node
	Box      Box
	Padding  EdgesPx
	Margin   EdgesPx
	Children []*Node

	Text    string
	Image   image.Image
	Leaders []Leader
}

// Leader is a fill sequence (e.g. dots) drawn between row children.
type Leader struct {
	X, Y, Width, Height float64
	Text                string
	FontFamily          string
	FontSize            float64
	FontWeight          int
}

type Tree struct {
	Root *Node
	DPI  float64
}

// PaperSideInsetPx reserves space on each side of the page so content stays
// visible when the paper roll is slightly off-center in the printer.
const PaperSideInsetPx = 20

func Compute(doc *dom.Document, opts Options) (*Tree, error) {
	if doc == nil || doc.Root == nil {
		return nil, fmt.Errorf("document is empty")
	}
	if opts.DPI <= 0 {
		opts.DPI = 96
	}
	if opts.Fonts == nil {
		return nil, fmt.Errorf("font manager is required")
	}
	if opts.Images == nil {
		opts.Images = imagecodec.DefaultRegistry()
	}

	root, err := buildNode(doc.Root, opts)
	if err != nil {
		return nil, err
	}

	pageWidth := resolveLength(root.DOM.Style.Width, opts.DPI, 0)
	if pageWidth <= 0 {
		pageWidth = units.Mm(80).ToPixels(opts.DPI, 0)
	}

	if err := measure(root, pageWidth, true, opts); err != nil {
		return nil, err
	}
	position(root, 0, 0, opts)

	return &Tree{Root: root, DPI: opts.DPI}, nil
}

func buildNode(n *dom.Node, opts Options) (*Node, error) {
	out := &Node{
		DOM:  n,
		Text: strings.TrimSpace(n.Text),
	}

	if n.Type == dom.TypeImage && n.Style.Src != "" {
		img, err := opts.Images.DecodeSource(n.Style.Src, opts.AssetBasePath)
		if err != nil {
			return nil, fmt.Errorf("image %q: %w", n.Style.Src, err)
		}
		out.Image = img
	}

	for _, child := range n.Children {
		cn, err := buildNode(child, opts)
		if err != nil {
			return nil, err
		}
		out.Children = append(out.Children, cn)
	}
	return out, nil
}

func measure(n *Node, availableWidth float64, stretch bool, opts Options) error {
	st := n.DOM.Style
	n.Margin = resolveEdges(st.Margin, opts.DPI, availableWidth)
	innerAvail := math.Max(0, availableWidth-n.Margin.Left-n.Margin.Right)
	n.Padding = resolveEdges(st.Padding, opts.DPI, innerAvail)
	if n.DOM.Type == dom.TypePage {
		n.Padding.Left += PaperSideInsetPx
		n.Padding.Right += PaperSideInsetPx
	}

	contentAvail := math.Max(0, innerAvail-n.Padding.Left-n.Padding.Right)

	explicitW := resolveLength(st.Width, opts.DPI, innerAvail)
	explicitH := resolveLength(st.Height, opts.DPI, 0)

	switch n.DOM.Type {
	case dom.TypePage:
		n.Box.Width = innerAvail
		if err := measureChildrenColumn(n, contentAvail, opts); err != nil {
			return err
		}
		contentH := childrenColumnHeight(n, opts)
		if explicitH > 0 {
			n.Box.Height = explicitH
		} else {
			n.Box.Height = n.Padding.Top + n.Padding.Bottom + contentH
		}

	case dom.TypeColumn:
		width := contentAvail
		if explicitW > 0 {
			width = explicitW
		}
		n.Box.Width = width + n.Padding.Left + n.Padding.Right
		if err := measureChildrenColumn(n, width, opts); err != nil {
			return err
		}
		contentH := childrenColumnHeight(n, opts)
		if explicitH > 0 {
			n.Box.Height = explicitH
		} else {
			n.Box.Height = n.Padding.Top + n.Padding.Bottom + contentH
		}

	case dom.TypeRow:
		width := contentAvail
		if explicitW > 0 {
			width = explicitW
		}
		n.Box.Width = width + n.Padding.Left + n.Padding.Right
		if err := measureChildrenRow(n, width, opts); err != nil {
			return err
		}
		contentH := childrenRowHeight(n)
		if explicitH > 0 {
			n.Box.Height = explicitH
		} else {
			n.Box.Height = n.Padding.Top + n.Padding.Bottom + contentH
		}

	case dom.TypeText:
		fontSize := resolveLength(st.FontSize, opts.DPI, 0)
		if fontSize <= 0 {
			fontSize = 14
		}
		maxContentW := contentAvail
		if explicitW > 0 {
			maxContentW = explicitW
		} else if stretch && contentAvail > 0 {
			maxContentW = contentAvail
		}

		lines, lineH, err := wrapText(opts.Fonts, st.Font, int(st.FontWeight), fontSize, n.Text, maxContentW)
		if err != nil {
			return err
		}
		n.Text = strings.Join(lines, "\n")

		tw := 0.0
		for _, line := range lines {
			lw, _, err := opts.Fonts.Measure(st.Font, int(st.FontWeight), fontSize, line)
			if err != nil {
				return err
			}
			if lw > tw {
				tw = lw
			}
		}
		th := lineH * float64(len(lines))

		switch {
		case explicitW > 0:
			n.Box.Width = explicitW + n.Padding.Left + n.Padding.Right
		case stretch && contentAvail > 0:
			n.Box.Width = contentAvail + n.Padding.Left + n.Padding.Right
		case contentAvail > 0 && tw > contentAvail:
			n.Box.Width = contentAvail + n.Padding.Left + n.Padding.Right
		default:
			n.Box.Width = tw + n.Padding.Left + n.Padding.Right
		}
		if explicitH > 0 {
			n.Box.Height = explicitH
		} else {
			n.Box.Height = th + n.Padding.Top + n.Padding.Bottom
		}

	case dom.TypeImage:
		iw, ih := 0.0, 0.0
		if n.Image != nil {
			b := n.Image.Bounds()
			iw = float64(b.Dx())
			ih = float64(b.Dy())
		}
		w := explicitW
		h := explicitH
		if w <= 0 && h <= 0 {
			w = iw
			h = ih
		} else if w > 0 && h <= 0 && iw > 0 {
			h = w * ih / iw
		} else if h > 0 && w <= 0 && ih > 0 {
			w = h * iw / ih
		}
		if w <= 0 {
			w = 64
		}
		if h <= 0 {
			h = 64
		}
		n.Box.Width = w + n.Padding.Left + n.Padding.Right
		n.Box.Height = h + n.Padding.Top + n.Padding.Bottom

	case dom.TypeDivider:
		n.Box.Width = contentAvail + n.Padding.Left + n.Padding.Right
		if st.Char != "" {
			fontSize := resolveLength(st.FontSize, opts.DPI, 0)
			if fontSize <= 0 {
				fontSize = 14
			}
			_, th, err := opts.Fonts.Measure(st.Font, int(st.FontWeight), fontSize, st.Char)
			if err != nil {
				return err
			}
			n.Text = repeatChar(st.Char, contentAvail, st.Font, int(st.FontWeight), fontSize, opts.Fonts)
			if explicitH > 0 {
				n.Box.Height = explicitH
			} else {
				n.Box.Height = th + n.Padding.Top + n.Padding.Bottom
			}
			break
		}
		stroke := resolveLength(st.Border.Width, opts.DPI, 0)
		if stroke <= 0 {
			stroke = 1
		}
		if explicitH > 0 {
			n.Box.Height = explicitH
		} else {
			n.Box.Height = stroke + n.Padding.Top + n.Padding.Bottom
		}

	case dom.TypeSpacer:
		h := explicitH
		if h <= 0 {
			h = 8
		}
		w := contentAvail
		if explicitW > 0 {
			w = explicitW
		}
		n.Box.Width = w + n.Padding.Left + n.Padding.Right
		n.Box.Height = h + n.Padding.Top + n.Padding.Bottom

	case dom.TypeQRCode:
		size := explicitW
		if size <= 0 {
			size = explicitH
		}
		if size <= 0 {
			size = 96
		}
		n.Box.Width = size + n.Padding.Left + n.Padding.Right
		n.Box.Height = size + n.Padding.Top + n.Padding.Bottom

	case dom.TypeBarcode:
		w := explicitW
		h := explicitH
		if w <= 0 {
			w = contentAvail
			if w <= 0 {
				w = 160
			}
		}
		if h <= 0 {
			h = 48
		}
		n.Box.Width = w + n.Padding.Left + n.Padding.Right
		n.Box.Height = h + n.Padding.Top + n.Padding.Bottom

	default:
		return fmt.Errorf("unsupported node type %s", n.DOM.Type)
	}

	n.Box.Width += n.Margin.Left + n.Margin.Right
	n.Box.Height += n.Margin.Top + n.Margin.Bottom
	return nil
}

func measureChildrenColumn(n *Node, contentAvail float64, opts Options) error {
	for _, child := range n.Children {
		if err := measure(child, contentAvail, true, opts); err != nil {
			return err
		}
	}
	return nil
}

func measureChildrenRow(n *Node, contentAvail float64, opts Options) error {
	gap := resolveLength(n.DOM.Style.Gap, opts.DPI, contentAvail)
	count := len(n.Children)
	if count == 0 {
		return nil
	}

	fixed := 0.0
	var flexChildren []*Node
	for _, child := range n.Children {
		if !child.DOM.Style.Width.IsAuto() && !child.DOM.Style.Width.IsPercent() {
			if err := measure(child, contentAvail, false, opts); err != nil {
				return err
			}
			fixed += child.Box.Width
		} else {
			flexChildren = append(flexChildren, child)
		}
	}

	gaps := 0.0
	if count > 1 {
		gaps = gap * float64(count-1)
	}
	remaining := math.Max(0, contentAvail-fixed-gaps)

	for _, child := range flexChildren {
		if err := measure(child, remaining, false, opts); err != nil {
			return err
		}
	}

	flexTotal := 0.0
	for _, child := range flexChildren {
		flexTotal += child.Box.Width
	}
	if flexTotal > remaining && len(flexChildren) > 0 && remaining > 0 {
		share := remaining / float64(len(flexChildren))
		for _, child := range flexChildren {
			if err := measure(child, share, true, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func childrenColumnHeight(n *Node, opts Options) float64 {
	if len(n.Children) == 0 {
		return 0
	}
	gap := resolveLength(n.DOM.Style.Gap, opts.DPI, 0)
	total := 0.0
	for i, child := range n.Children {
		total += child.Box.Height
		if i < len(n.Children)-1 {
			total += gap
		}
	}
	return total
}

func childrenRowHeight(n *Node) float64 {
	maxH := 0.0
	for _, child := range n.Children {
		if child.Box.Height > maxH {
			maxH = child.Box.Height
		}
	}
	return maxH
}

func position(n *Node, x, y float64, opts Options) {
	n.Box.X = x
	n.Box.Y = y

	contentX := x + n.Margin.Left + n.Padding.Left
	contentY := y + n.Margin.Top + n.Padding.Top
	contentW := n.Box.Width - n.Margin.Left - n.Margin.Right - n.Padding.Left - n.Padding.Right
	contentH := n.Box.Height - n.Margin.Top - n.Margin.Bottom - n.Padding.Top - n.Padding.Bottom

	switch n.DOM.Type {
	case dom.TypePage, dom.TypeColumn:
		positionColumnChildren(n, contentX, contentY, contentW, opts)
	case dom.TypeRow:
		positionRowChildren(n, contentX, contentY, contentW, contentH, opts)
	}
}

func positionColumnChildren(n *Node, x, y, contentW float64, opts Options) {
	gap := resolveLength(n.DOM.Style.Gap, opts.DPI, contentW)
	cy := y
	for i, child := range n.Children {
		cx := x
		childContentW := child.Box.Width - child.Margin.Left - child.Margin.Right
		switch n.DOM.Style.Align {
		case style.AlignCenter:
			cx = x + (contentW-childContentW)/2 - child.Margin.Left
		case style.AlignEnd:
			cx = x + contentW - childContentW - child.Margin.Left
		}
		position(child, cx, cy, opts)
		cy += child.Box.Height
		if i < len(n.Children)-1 {
			cy += gap
		}
	}
}

func positionRowChildren(n *Node, x, y, contentW, contentH float64, opts Options) {
	gap := resolveLength(n.DOM.Style.Gap, opts.DPI, contentW)
	count := len(n.Children)
	if count == 0 {
		return
	}

	totalW := 0.0
	for _, child := range n.Children {
		totalW += child.Box.Width
	}
	if count > 1 {
		totalW += gap * float64(count-1)
	}

	startX := x
	spacing := gap

	switch n.DOM.Style.Justify {
	case style.JustifyCenter:
		startX = x + (contentW-totalW)/2
	case style.JustifyEnd:
		startX = x + contentW - totalW
	case style.JustifyBetween:
		if count > 1 {
			childrenW := 0.0
			for _, child := range n.Children {
				childrenW += child.Box.Width
			}
			spacing = (contentW - childrenW) / float64(count-1)
			if spacing < 0 {
				spacing = gap
			}
			startX = x
		}
	case style.JustifyAround:
		childrenW := 0.0
		for _, child := range n.Children {
			childrenW += child.Box.Width
		}
		spacing = (contentW - childrenW) / float64(count)
		startX = x + spacing/2
	}

	cx := startX
	for i, child := range n.Children {
		cy := y
		switch n.DOM.Style.Align {
		case style.AlignCenter:
			cy = y + (contentH-child.Box.Height)/2
		case style.AlignEnd:
			cy = y + contentH - child.Box.Height
		}
		position(child, cx, cy, opts)
		cx += child.Box.Width
		if i < count-1 {
			cx += spacing
		}
	}

	if fill := n.DOM.Style.Char; fill != "" && count >= 2 {
		n.Leaders = buildRowLeaders(n, fill, opts)
	}
}

func buildRowLeaders(n *Node, fill string, opts Options) []Leader {
	if opts.Fonts == nil || fill == "" || len(n.Children) < 2 {
		return nil
	}

	fontFamily := n.DOM.Style.Font
	if fontFamily == "" {
		fontFamily = "Arial"
	}
	fontSize := resolveLength(n.DOM.Style.FontSize, opts.DPI, 0)
	fontWeight := int(n.DOM.Style.FontWeight)
	if n.DOM.Style.FillWeight != 0 {
		fontWeight = int(n.DOM.Style.FillWeight)
	}
	if fontSize <= 0 {
		for _, child := range n.Children {
			if child.DOM.Type == dom.TypeText {
				fontSize = resolveLength(child.DOM.Style.FontSize, opts.DPI, 0)
				fontFamily = child.DOM.Style.Font
				if n.DOM.Style.FillWeight == 0 {
					fontWeight = int(child.DOM.Style.FontWeight)
				}
				if fontSize > 0 {
					break
				}
			}
		}
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	if fontFamily == "" {
		fontFamily = "Arial"
	}
	if n.DOM.Style.FillWeight != 0 {
		fontWeight = int(n.DOM.Style.FillWeight)
	}

	_, lineH, err := opts.Fonts.Measure(fontFamily, fontWeight, fontSize, "Mg")
	if err != nil || lineH <= 0 {
		lineH = fontSize
	}

	var leaders []Leader
	for i := 0; i < len(n.Children)-1; i++ {
		left := n.Children[i]
		right := n.Children[i+1]
		gapStart := left.Box.X + left.Box.Width
		gapEnd := right.Box.X
		gapW := gapEnd - gapStart
		if gapW <= 1 {
			continue
		}

		spaceW, _, err := opts.Fonts.Measure(fontFamily, fontWeight, fontSize, " ")
		if err != nil || spaceW <= 0 {
			spaceW = fontSize * 0.3
		}
		innerStart := gapStart + spaceW
		innerW := gapW - 2*spaceW
		if innerW <= spaceW {
			continue
		}

		text := repeatChar(fill, innerW, fontFamily, fontWeight, fontSize, opts.Fonts)
		if text == "" {
			continue
		}
		leftContent := left.ContentBox()
		leaders = append(leaders, Leader{
			X:          innerStart,
			Y:          leftContent.Y,
			Width:      innerW,
			Height:     lineH,
			Text:       text,
			FontFamily: fontFamily,
			FontSize:   fontSize,
			FontWeight: fontWeight,
		})
	}
	return leaders
}

func resolveLength(l units.Length, dpi, parent float64) float64 {
	if l.IsAuto() {
		return 0
	}
	return l.ToPixels(dpi, parent)
}

func resolveEdges(e style.Edges, dpi, parent float64) EdgesPx {
	return EdgesPx{
		Top:    resolveLength(e.Top, dpi, parent),
		Right:  resolveLength(e.Right, dpi, parent),
		Bottom: resolveLength(e.Bottom, dpi, parent),
		Left:   resolveLength(e.Left, dpi, parent),
	}
}

func (n *Node) ContentBox() Box {
	return Box{
		X:      n.Box.X + n.Margin.Left + n.Padding.Left,
		Y:      n.Box.Y + n.Margin.Top + n.Padding.Top,
		Width:  n.Box.Width - n.Margin.Left - n.Margin.Right - n.Padding.Left - n.Padding.Right,
		Height: n.Box.Height - n.Margin.Top - n.Margin.Bottom - n.Padding.Top - n.Padding.Bottom,
	}
}

func (n *Node) BorderBox() Box {
	return Box{
		X:      n.Box.X + n.Margin.Left,
		Y:      n.Box.Y + n.Margin.Top,
		Width:  n.Box.Width - n.Margin.Left - n.Margin.Right,
		Height: n.Box.Height - n.Margin.Top - n.Margin.Bottom,
	}
}

func repeatChar(char string, maxWidth float64, family string, weight int, size float64, fonts *font.Manager) string {
	if char == "" || maxWidth <= 0 {
		return ""
	}
	cw, _, err := fonts.Measure(family, weight, size, char)
	if err != nil || cw <= 0 {
		return char
	}
	count := int(maxWidth / cw)
	if count < 1 {
		count = 1
	}
	return strings.Repeat(char, count)
}

func wrapText(fonts *font.Manager, family string, weight int, size float64, text string, maxWidth float64) ([]string, float64, error) {
	_, lineH, err := fonts.Measure(family, weight, size, "Mg")
	if err != nil {
		return nil, 0, err
	}
	if lineH <= 0 {
		lineH = size
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{""}, lineH, nil
	}
	if maxWidth <= 0 {
		return []string{text}, lineH, nil
	}

	tw, _, err := fonts.Measure(family, weight, size, text)
	if err != nil {
		return nil, 0, err
	}
	if tw <= maxWidth {
		return []string{text}, lineH, nil
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}, lineH, nil
	}

	var lines []string
	current := words[0]
	for _, word := range words[1:] {
		candidate := current + " " + word
		cw, _, err := fonts.Measure(family, weight, size, candidate)
		if err != nil {
			return nil, 0, err
		}
		if cw <= maxWidth {
			current = candidate
			continue
		}
		lines = append(lines, breakLongWord(fonts, family, weight, size, current, maxWidth)...)
		current = word
	}
	lines = append(lines, breakLongWord(fonts, family, weight, size, current, maxWidth)...)
	if len(lines) == 0 {
		lines = []string{text}
	}
	return lines, lineH, nil
}

func breakLongWord(fonts *font.Manager, family string, weight int, size float64, word string, maxWidth float64) []string {
	tw, _, err := fonts.Measure(family, weight, size, word)
	if err != nil || tw <= maxWidth || maxWidth <= 0 {
		return []string{word}
	}

	var lines []string
	runes := []rune(word)
	start := 0
	for start < len(runes) {
		end := start + 1
		for end <= len(runes) {
			part := string(runes[start:end])
			pw, _, err := fonts.Measure(family, weight, size, part)
			if err != nil || pw > maxWidth {
				break
			}
			end++
		}
		if end == start+1 {
			lines = append(lines, string(runes[start:end]))
			start = end
			continue
		}
		lines = append(lines, string(runes[start:end-1]))
		start = end - 1
	}
	return lines
}
