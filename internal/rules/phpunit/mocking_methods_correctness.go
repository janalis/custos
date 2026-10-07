package phpunit

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// mockingMethodsCorrectness reports `willReturn($this->returnValue(..))`
// and mocked methods that do not exist or are final.
type mockingMethodsCorrectness struct{}

func init() { register(mockingMethodsCorrectness{}) }

func (mockingMethodsCorrectness) ID() string { return "MockingMethodsCorrectness" }

func (mockingMethodsCorrectness) Semantic() {}

func (mockingMethodsCorrectness) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall}
}

const (
	mmcWillMsg    = "The stub object is returned as-is here; use '->will(...)'."
	mmcMissingMsg = "The mocked class has no such method."
	mmcFinalMsg   = "Final methods cannot be mocked."
)

func (mockingMethodsCorrectness) Check(ctx *analysis.Context, n syntax.Node) {
	c, ok := asPuCall(n)
	if !ok || n.Span().Len() == 0 {
		return
	}
	switch c.Name {
	case "willReturn", "method":
	default:
		return
	}
	args, ok := c.args()
	if !ok || len(args) != 1 {
		return
	}
	if !util.InTestContext(ctx, n) { // D0
		return
	}
	if c.Name == "willReturn" {
		if _, ok := puMethodCallNamed(args[0], "returnCallback", "returnValue"); ok { // D1, D2
			id := c.Ident
			ctx.ReportNode(id, mmcWillMsg, analysis.Fix{
				Title: "Use ->will()",
				Edits: func() []analysis.TextEdit {
					return []analysis.TextEdit{{Span: id.Span(), NewText: "will"}}
				},
			})
		}
		return
	}
	mmcCheckMethod(ctx, c, args[0])
}

func mmcCheckMethod(ctx *analysis.Context, c puCall, arg syntax.Expr) {
	lit, ok := arg.(*syntax.Literal) // D3
	if !ok || lit.LitKind != syntax.LitString {
		return
	}
	recv := c.Recv // D4
	if ex, ok := puMethodCallNamed(recv, "expects"); ok {
		recv = ex.Recv
	}
	vals := util.PossibleValues(ctx.File, recv) // D5
	if len(vals) != 1 {
		return
	}
	mock, ok := puMethodCallNamed(vals[0], "getMock")
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
		if s == nil || len(s.Args.Args) == 0 {
			continue
		}
		a, ok := s.Args.Args[0].(*syntax.Arg)
		if !ok {
			continue
		}
		arr, ok := a.Value.(*syntax.Array)
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
	bargs, ok := builder.args() // D8
	if !ok || len(bargs) != 1 {
		return
	}
	cn, ok := puClassConstClass(bargs[0])
	if !ok {
		return
	}
	fqn := puResolveClassName(ctx, cn)
	if fqn == "" || ctx.Index().Class(fqn, ctx.PHP) == nil {
		return
	}
	name := mmcLiteralContent(lit.Raw) // D9
	m := ctx.Index().FindMethod(fqn, name, ctx.PHP)
	switch {
	case m == nil:
		ctx.ReportSeverity(lit.Span(), meta.SeverityError, mmcMissingMsg)
	case m.Final:
		ctx.ReportSeverity(lit.Span(), meta.SeverityError, mmcFinalMsg)
	}
}

// mmcFirstCall returns the first (pre-order) method call named name among
// the descendants of root.
func mmcFirstCall(root syntax.Node, name string) *puCall {
	var found *puCall
	syntax.Inspect(root, func(x syntax.Node) bool {
		if found != nil {
			return false
		}
		if x != root {
			if c, ok := puMethodCallNamed(x, name); ok {
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
	if c, _, ok := util.QuotedStringRaw(&syntax.Literal{LitKind: syntax.LitString, Raw: raw}); ok {
		return c
	}
	if strings.HasPrefix(raw, "<<<") {
		if i := strings.IndexByte(raw, '\n'); i >= 0 {
			body := raw[i+1:]
			if j := strings.LastIndexByte(body, '\n'); j >= 0 {
				return body[:j]
			}
		}
	}
	return raw
}
