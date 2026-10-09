package mockingmethodscorrectness

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/phpunit"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// mockingMethodsCorrectness reports `willReturn($this->returnValue(..))`
// and mocked methods that do not exist or are final.
type mockingMethodsCorrectness struct{}

func (mockingMethodsCorrectness) ID() string { return "MockingMethodsCorrectness" }
func (mockingMethodsCorrectness) Semantic()  {}
func (mockingMethodsCorrectness) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall}
}

const (
	mmcWillMsg    = "The stub object is returned as-is here; use '->will(...)'."
	mmcMissingMsg = "The mocked class has no such method."
	mmcFinalMsg   = "Final methods cannot be mocked."
)

func (mockingMethodsCorrectness) Check(ctx *analysis.Context, n syntax.Node) {
	c, ok := phpunit.AsPuCall(n)
	if !ok {
		return
	}
	switch c.Name {
	case "willReturn", "method":
	default:
		return
	}
	args, ok := c.Values()
	if !ok || len(args) != 1 {
		return
	}
	if !semanticquery.InTestContext(ctx, n) { // D0
		return
	}
	if c.Name == "willReturn" {
		if _, ok := phpunit.PuMethodCallNamed(args[0], "returnCallback", "returnValue"); ok { // D1, D2
			id := c.Ident
			ctx.ReportNode(id, mmcWillMsg, diagnostic.Fix{
				Title: "Use ->will()",
				Edits: func() []diagnostic.TextEdit {
					return []diagnostic.TextEdit{{Span: id.Span(), NewText: "will"}}
				},
			})
		}
		return
	}
	mmcCheckMethod(ctx, c, args[0])
}

func mmcCheckMethod(ctx *analysis.Context, c phpunit.PuCall, arg syntax.Expr) {
	lit, ok := arg.(*syntax.Literal) // D3
	if !ok || lit.LitKind != syntax.LitString {
		return
	}
	recv := c.Recv // D4
	if ex, ok := phpunit.PuMethodCallNamed(recv, "expects"); ok {
		recv = ex.Recv
	}
	vals := flowquery.PossibleValues(ctx.File, recv) // D5
	if len(vals) != 1 {
		return
	}
	mock, ok := phpunit.PuMethodCallNamed(vals[0], "getMock")
	if !ok {
		return
	}
	builder := mmcFirstCall(mock.Node, "getMockBuilder") // D6
	if builder == nil {
		return
	}
	// D7: methods explicitly added to the double. setMethods() compares the
	// literal text (upstream); addMethods() (PHPUnit 8.3+, custos) compares
	// the method names case-insensitively.
	for _, adder := range []string{"setMethods", "addMethods"} {
		s := mmcFirstCall(mock.Node, adder)
		if s == nil {
			continue
		}
		sargs, _ := s.Values() // nil with spreads, named arguments or placeholders
		if len(sargs) == 0 {
			continue
		}
		arr, ok := sargs[0].(*syntax.Array)
		if !ok {
			continue
		}
		listed := false
		syntax.Inspect(arr, func(x syntax.Node) bool {
			if l, ok := x.(*syntax.Literal); ok && l.LitKind == syntax.LitString {
				if adder == "setMethods" && l.Raw == lit.Raw ||
					adder == "addMethods" && strings.EqualFold(mmcLiteralContent(l.Raw), mmcLiteralContent(lit.Raw)) {
					listed = true
				}
			}
			return !listed
		})
		if listed {
			return
		}
	}
	bargs, ok := builder.Values() // D8
	if !ok || len(bargs) != 1 {
		return
	}
	cn, ok := phpunit.PuClassConstClass(bargs[0])
	if !ok {
		return
	}
	fqn := phpunit.PuResolveClassName(ctx, cn)
	if fqn == "" || ctx.Index().Class(fqn, ctx.PHP) == nil {
		return
	}
	name := mmcLiteralContent(lit.Raw) // D9
	m := ctx.Index().FindMethod(fqn, name, ctx.PHP)
	switch {
	case m == nil:
		if !semanticquery.HierarchyResolved(ctx.Index(), fqn, ctx.PHP) {
			return // an unresolvable ancestor may declare it (custos)
		}
		ctx.ReportSeverity(lit.Span(), diagnostic.SeverityError, mmcMissingMsg)
	case m.Final:
		ctx.ReportSeverity(lit.Span(), diagnostic.SeverityError, mmcFinalMsg)
	}
}

// mmcFirstCall returns the first (pre-order) method call named name among
// the descendants of root.
func mmcFirstCall(root syntax.Node, name string) *phpunit.PuCall {
	var found *phpunit.PuCall
	syntax.Inspect(root, func(x syntax.Node) bool {
		if found != nil {
			return false
		}
		if x != root {
			if c, ok := phpunit.PuMethodCallNamed(x, name); ok {
				found = &c
				return false
			}
		}
		return true
	})
	return found
}

// mmcLiteralContent returns the raw text between the quotes of a string
// literal (heredoc/nowdoc: the body lines).
func mmcLiteralContent(raw string) string {
	if c, _, ok := astquery.QuotedStringRaw(&syntax.Literal{LitKind: syntax.LitString, Raw: raw}); ok {
		return c
	}
	// Otherwise a heredoc/nowdoc: `<<<ID` line, body lines, closing marker
	// line whose indentation (PHP 7.3+) is removed from every body line.
	body := raw[strings.IndexByte(raw, '\n')+1:]
	j := strings.LastIndexByte(body, '\n')
	if j < 0 {
		return "" // empty body
	}
	closing := body[j+1:]
	indent := closing[:len(closing)-len(strings.TrimLeft(closing, " \t"))]
	lines := strings.Split(body[:j], "\n")
	for k := range lines {
		lines[k] = strings.TrimPrefix(lines[k], indent)
	}
	return strings.Join(lines, "\n")
}
