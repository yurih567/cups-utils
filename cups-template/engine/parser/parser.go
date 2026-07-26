package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"cups-template/engine/dom"
	"cups-template/engine/style"
)

func ParseFile(path string) (*dom.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func Parse(r io.Reader) (*dom.Document, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false

	var root *dom.Node
	var stack []*dom.Node

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml parse error: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			node, err := elementToNode(t)
			if err != nil {
				return nil, err
			}
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)

		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("unexpected closing tag </%s>", t.Name.Local)
			}
			stack = stack[:len(stack)-1]

		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			text := strings.TrimSpace(string(t))
			if text == "" {
				continue
			}
			cur := stack[len(stack)-1]
			if cur.Type == dom.TypeText {
				if cur.Text != "" {
					cur.Text += " "
				}
				cur.Text += text
			} else {
				child := &dom.Node{
					Type:  dom.TypeText,
					Style: style.Default(),
					Text:  text,
				}
				cur.Children = append(cur.Children, child)
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("empty template")
	}
	if root.Type != dom.TypePage {
		return nil, fmt.Errorf("root element must be <Page>, got <%s>", root.Type)
	}

	return &dom.Document{Root: root}, nil
}

func elementToNode(el xml.StartElement) (*dom.Node, error) {
	typ, err := parseType(el.Name.Local)
	if err != nil {
		return nil, err
	}

	attrs := make(map[string]string, len(el.Attr))
	for _, a := range el.Attr {
		attrs[a.Name.Local] = a.Value
	}

	st, err := style.ParseAttrs(attrs)
	if err != nil {
		return nil, fmt.Errorf("<%s>: %w", typ, err)
	}

	return &dom.Node{
		Type:  typ,
		Style: st,
	}, nil
}

func parseType(name string) (dom.NodeType, error) {
	switch name {
	case "Page":
		return dom.TypePage, nil
	case "Column":
		return dom.TypeColumn, nil
	case "Row":
		return dom.TypeRow, nil
	case "Text":
		return dom.TypeText, nil
	case "Image":
		return dom.TypeImage, nil
	case "Divider", "Divisor":
		return dom.TypeDivider, nil
	case "Spacer":
		return dom.TypeSpacer, nil
	case "QRCode":
		return dom.TypeQRCode, nil
	case "Barcode":
		return dom.TypeBarcode, nil
	default:
		return "", fmt.Errorf("unknown component <%s>", name)
	}
}
