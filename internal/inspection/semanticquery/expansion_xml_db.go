package semanticquery

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// ExpansionDOMOwner follows locally proven node creation to its owning document.
// It deliberately excludes arbitrary DOM traversal and unknown adopted nodes.
func ExpansionDOMOwner(ctx *analysis.Context, e syntax.Expr) *syntax.New {
	return expansionDOMOwner(ctx, e, 0)
}

func expansionDOMOwner(ctx *analysis.Context, e syntax.Expr, depth int) *syntax.New {
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
			return expansionDOMOwner(ctx, x.Var, depth+1)
		}
	case *syntax.MethodCall:
		for _, name := range []string{"createElement", "createTextNode", "createElementNS", "importNode", "adoptNode"} {
			if NativeMethod(ctx, x, "DOMDocument", name) {
				return expansionDOMOwner(ctx, x.Var, depth+1)
			}
		}
		if NativeMethod(ctx, x, "DOMNode", "cloneNode") {
			return expansionDOMOwner(ctx, x.Var, depth+1)
		}
	}
	return nil
}

// ExpansionDOMElementChildren proves existing element children from literal XML
// or a preceding append of a locally created element. Unknown documents stay unknown.
func ExpansionDOMElementChildren(ctx *analysis.Context, e syntax.Expr, at syntax.Node) bool {
	v := NativeValue(ctx, e)
	if p, ok := syntax.UnwrapParens(e).(*syntax.PropertyFetch); ok {
		v = p
	}
	if p, ok := v.(*syntax.PropertyFetch); ok {
		if id, ok := p.Name.(*syntax.Identifier); ok && id.Value == "documentElement" {
			prior := NativePriorCalls(ctx, at, p.Var, "loadXML", "load", "appendChild", "removeChild")
			if len(prior) == 0 {
				return false
			}
			c, ok := prior[len(prior)-1].(*syntax.MethodCall)
			if !ok || !NativeMethod(ctx, c, "DOMDocument", "loadXML") {
				return false
			}
			s, known := NativeString(ctx, CallArgument(c.Args, 0, "source"))
			return known && expansionXMLChildren(s)
		}
	}
	found := false
	for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		c, ok := record.Node.(*syntax.MethodCall)
		if !ok || !NativeDominates(c, at) || !ExpansionSameObject(ctx, c.Var, e) {
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

func expansionXMLChildren(s string) bool {
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

// ExpansionSQLParts splits a constant SQL skeleton around dynamic PHP operands.
// It accepts only literal concatenation and ordinary interpolated strings.
func ExpansionSQLParts(ctx *analysis.Context, e syntax.Expr) ([]string, []syntax.Expr, bool) {
	var text strings.Builder
	var parts []string
	var values []syntax.Expr
	var visit func(syntax.Expr) bool
	visit = func(e syntax.Expr) bool {
		e = syntax.UnwrapParens(e)
		if s, ok := NativeString(ctx, e); ok {
			text.WriteString(s)
			return true
		}
		if b, ok := e.(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
			return visit(b.Left) && visit(b.Right)
		}
		if s, ok := e.(*syntax.InterpolatedString); ok {
			if s.Backtick {
				return false
			}
			for _, p := range s.Parts {
				if literal, ok := p.(*syntax.StringPart); ok {
					text.WriteString(literal.Raw)
				} else {
					parts = append(parts, text.String())
					text.Reset()
					values = append(values, p)
				}
			}
			return true
		}
		parts = append(parts, text.String())
		text.Reset()
		values = append(values, e)
		return true
	}
	if !visit(e) {
		return nil, nil, false
	}
	parts = append(parts, text.String())
	return parts, values, len(values) > 0
}

// ExpansionGlobalConstant recognizes an unshadowed named extension constant.
func ExpansionGlobalConstant(ctx *analysis.Context, e syntax.Expr, name string) bool {
	c, ok := e.(*syntax.ConstFetch)
	return ok && GlobalConstName(ctx, c) == name
}

// ExpansionSameObject compares retained local binding provenance when the flow
// model cannot assign allocation identities to extension-created objects.
func ExpansionSameObject(ctx *analysis.Context, a, b syntax.Expr) bool {
	x := ctx.Flow().Value(a)
	y := ctx.Flow().Value(b)
	if x.Complete && y.Complete && x.Identity != 0 && x.Identity == y.Identity && x.Invalidated == y.Invalidated {
		return true
	}
	left := NativeValue(ctx, a)
	right := NativeValue(ctx, b)
	if left == nil || left != right {
		return false
	}
	switch left.(type) {
	case *syntax.New, *syntax.MethodCall:
		return true
	}
	return false
}

// ExpansionDOMBeforeWrite handles a DOM assignment's left-hand receiver, which
// has no read value in the flow model. Only direct, nonescaping statements prove state.
func ExpansionDOMBeforeWrite(ctx *analysis.Context, e syntax.Expr, at syntax.Node) bool {
	p, ok := e.(*syntax.PropertyFetch)
	if !ok {
		return false
	}
	field, ok := p.Name.(*syntax.Identifier)
	if !ok || field.Value != "documentElement" {
		return false
	}
	v, ok := p.Var.(*syntax.Variable)
	if !ok {
		return false
	}
	statements := flowquery.NativePriorStatements(ctx.File, at)
	if len(statements) > 256 {
		return false
	}
	document, children := false, false
	for _, statement := range statements {
		st, ok := statement.(*syntax.ExprStmt)
		if !ok {
			document, children = false, false
			continue
		}
		if a, ok := st.Expr.(*syntax.Assign); ok {
			if target, ok := a.Var.(*syntax.Variable); ok && target.Name == v.Name {
				creation, ok := a.Value.(*syntax.New)
				document = ok && !a.ByRef && a.Op.Kind == syntax.TEqual && NamesGlobalClass(ctx, creation.Class, "DOMDocument")
				children = false
				continue
			}
		}
		if c, ok := st.Expr.(*syntax.MethodCall); ok {
			if target, ok := c.Var.(*syntax.Variable); ok && target.Name == v.Name && document && NativeMethod(ctx, c, "DOMDocument", "loadXML") {
				s, known := NativeString(ctx, CallArgument(c.Args, 0, "source"))
				children = known && expansionXMLChildren(s)
				continue
			}
		}
		syntax.Inspect(statement, func(n syntax.Node) bool {
			if x, ok := n.(*syntax.Variable); ok && x.Name == v.Name {
				document, children = false, false
			}
			return true
		})
	}
	return document && children
}

// ExpansionDOMCreatedBeforeWrite proves children of a directly constructed
// local element before a write to that element's properties.
func ExpansionDOMCreatedBeforeWrite(ctx *analysis.Context, e syntax.Expr, at syntax.Node) bool {
	variable, ok := e.(*syntax.Variable)
	if !ok {
		return false
	}
	statements := flowquery.NativePriorStatements(ctx.File, at)
	if len(statements) > 256 {
		return false
	}
	element, children := false, false
	for _, statement := range statements {
		st, ok := statement.(*syntax.ExprStmt)
		if !ok {
			element, children = false, false
			continue
		}
		if a, ok := st.Expr.(*syntax.Assign); ok {
			if target, ok := a.Var.(*syntax.Variable); ok && target.Name == variable.Name {
				c, ok := a.Value.(*syntax.MethodCall)
				element = ok && !a.ByRef && a.Op.Kind == syntax.TEqual && (NativeMethod(ctx, c, "DOMDocument", "createElement") || NativeMethod(ctx, c, "DOMDocument", "createElementNS"))
				children = false
				continue
			}
		}
		if c, ok := st.Expr.(*syntax.MethodCall); ok {
			if target, ok := c.Var.(*syntax.Variable); ok && target.Name == variable.Name && element && NativeMethod(ctx, c, "DOMNode", "appendChild") {
				child, ok := NativeValue(ctx, CallArgument(c.Args, 0, "node")).(*syntax.MethodCall)
				children = children || ok && (NativeMethod(ctx, child, "DOMDocument", "createElement") || NativeMethod(ctx, child, "DOMDocument", "createElementNS"))
				continue
			}
		}
		syntax.Inspect(statement, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok && v.Name == variable.Name {
				element, children = false, false
			}
			return true
		})
	}
	return element && children
}
