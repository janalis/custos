package semanticquery

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// ParseLiteralXML parses bounded literal XML without resolving external entities.
// Namespaces and nontrivial declarations remain unknown to these inspections.
func ParseLiteralXML(s string) (children []string, text string, plain, known bool) {
	if len(s) > 65536 {
		return nil, "", false, false
	}
	d := xml.NewDecoder(strings.NewReader(s))
	depth := 0
	roots := 0
	plain = true
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			return children, text, plain, depth == 0 && roots == 1
		}
		if err != nil {
			return nil, "", false, false
		}
		switch x := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 {
				roots++
			}
			if x.Name.Space != "" {
				return nil, "", false, false
			}
			if depth == 2 {
				children = append(children, x.Name.Local)
				plain = false
			}
			if len(x.Attr) > 0 {
				plain = false
				for _, a := range x.Attr {
					if a.Name.Local == "xmlns" || a.Name.Space != "" {
						return nil, "", false, false
					}
				}
			}
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 1 {
				text += string(x)
			}
		case xml.Directive:
			return nil, "", false, false
		}
	}
}

// DOMOwnerDocument follows locally proven node creation to its owning document.
// It deliberately excludes arbitrary DOM traversal and unknown adopted nodes.
func DOMOwnerDocument(ctx *analysis.Context, e syntax.Expr) *syntax.New {
	return domOwnerDocument(ctx, e, 0)
}

func domOwnerDocument(ctx *analysis.Context, e syntax.Expr, depth int) *syntax.New {
	if depth > 16 {
		return nil
	}
	v := NativeValue(ctx, e)
	if p, ok := syntax.UnwrapParens(e).(*syntax.PropertyFetch); ok {
		v = p
	}
	switch x := v.(type) {
	case *syntax.New:
		if NamesGlobalClass(ctx, x.Class, "DOMDocument") {
			return x
		}
	case *syntax.PropertyFetch:
		if id, ok := x.Name.(*syntax.Identifier); ok && id.Value == "documentElement" {
			return domOwnerDocument(ctx, x.Var, depth+1)
		}
	case *syntax.MethodCall:
		for _, name := range []string{"createElement", "createTextNode", "createElementNS", "importNode", "adoptNode"} {
			if NativeMethod(ctx, x, "DOMDocument", name) {
				return domOwnerDocument(ctx, x.Var, depth+1)
			}
		}
		if NativeMethod(ctx, x, "DOMNode", "cloneNode") {
			return domOwnerDocument(ctx, x.Var, depth+1)
		}
	}
	return nil
}

// DOMElementHasChildren proves existing element children from literal XML
// or a preceding append of a locally created element. Unknown documents stay unknown.
func DOMElementHasChildren(ctx *analysis.Context, e syntax.Expr, at syntax.Node) bool {
	v := NativeValue(ctx, e)
	if p, ok := syntax.UnwrapParens(e).(*syntax.PropertyFetch); ok {
		v = p
	}
	if p, ok := v.(*syntax.PropertyFetch); ok {
		if id, ok := p.Name.(*syntax.Identifier); ok && id.Value == "documentElement" {
			history := NativeCallHistory(ctx, at, p.Var, "loadXML", "load", "appendChild", "removeChild")
			if !history.Complete {
				return false
			}
			// Zero is unknown; one and two mean absent and present.
			state := 0
			for _, event := range history.Events {
				next := 0
				if c, ok := event.Node.(*syntax.MethodCall); ok && NativeMethod(ctx, c, "DOMDocument", "loadXML") {
					if source, known := NativeString(ctx, CallArgument(c.Args, 0, "source")); known {
						next = 1
						if XMLChildren(source) {
							next = 2
						}
					}
				}
				if event.Must || state == next {
					state = next
				} else {
					state = 0
				}
			}
			return state == 2
		}
	}
	found := false
	for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		c, ok := record.Node.(*syntax.MethodCall)
		if !ok || !NativeDominates(c, at) || !SameLocalObject(ctx, c.Var, e) {
			continue
		}
		if !NativeMethod(ctx, c, "DOMNode", "appendChild") && !NativeMethod(ctx, c, "DOMNode", "removeChild") && !NativeMethod(ctx, c, "DOMNode", "replaceChild") {
			continue
		}
		if !NativeMethod(ctx, c, "DOMNode", "appendChild") {
			return false
		}
		child := NativeValue(ctx, CallArgument(c.Args, 0, "node"))
		if m, ok := child.(*syntax.MethodCall); ok && (NativeMethod(ctx, m, "DOMDocument", "createElement") || NativeMethod(ctx, m, "DOMDocument", "createElementNS")) {
			found = true
		}
	}
	return found
}

// XMLChildren reports whether the XML document s parses and its root
// element has a child element; false past 64 KiB.
func XMLChildren(s string) bool {
	if len(s) > 65536 {
		return false
	}
	d := xml.NewDecoder(strings.NewReader(s))
	depth := 0
	found := false
	for {
		t, err := d.Token()
		if errors.Is(err, io.EOF) {
			return found
		}
		if err != nil {
			return false
		}
		switch t.(type) {
		case xml.StartElement:
			depth++
			if depth == 2 {
				found = true
			}
		case xml.EndElement:
			depth--
		}
	}
}
