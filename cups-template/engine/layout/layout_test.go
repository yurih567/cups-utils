package layout_test

import (
	"strings"
	"testing"

	"cups-template/engine"
	"cups-template/engine/font"
	"cups-template/engine/parser"
)

func TestRowTextWrapsInsteadOfOverlapping(t *testing.T) {
	fonts, err := font.NewDefault(203)
	if err != nil {
		t.Fatal(err)
	}

	xml := `
<Page width="80mm" padding="4">
  <Column>
    <Row justify="between">
      <Text size="10" weight="bold">Data: 25/07/2026 19:15</Text>
      <Text size="10" weight="bold" align="right">Pedido: A4F2</Text>
    </Row>
  </Column>
</Page>`

	doc, err := parser.Parse(strings.NewReader(xml))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := engine.Layout(doc, engine.Options{DPI: 203, Fonts: fonts})
	if err != nil {
		t.Fatal(err)
	}

	row := tree.Root.Children[0].Children[0]
	if len(row.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(row.Children))
	}
	left, right := row.Children[0], row.Children[1]
	contentW := row.Box.Width - row.Padding.Left - row.Padding.Right - row.Margin.Left - row.Margin.Right
	if left.Box.Width+right.Box.Width > contentW+1 {
		t.Fatalf("children overflow row: left=%.1f right=%.1f content=%.1f", left.Box.Width, right.Box.Width, contentW)
	}
	if right.Box.X < left.Box.X+left.Box.Width-0.5 {
		t.Fatalf("overlap: left ends at %.1f right starts at %.1f", left.Box.X+left.Box.Width, right.Box.X)
	}
}
