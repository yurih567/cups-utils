package dom

import "cups-template/engine/style"

type NodeType string

const (
	TypePage    NodeType = "Page"
	TypeColumn  NodeType = "Column"
	TypeRow     NodeType = "Row"
	TypeText    NodeType = "Text"
	TypeImage   NodeType = "Image"
	TypeDivider NodeType = "Divider"
	TypeSpacer  NodeType = "Spacer"
	TypeQRCode  NodeType = "QRCode"
	TypeBarcode NodeType = "Barcode"
)

type Node struct {
	Type     NodeType
	Style    style.Style
	Text     string
	Children []*Node
}

type Document struct {
	Root *Node
}

func (n *Node) IsContainer() bool {
	switch n.Type {
	case TypePage, TypeColumn, TypeRow:
		return true
	default:
		return false
	}
}
