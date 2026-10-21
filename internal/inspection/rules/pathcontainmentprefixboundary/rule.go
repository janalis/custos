// Package pathcontainmentprefixboundary implements the native PathContainmentPrefixBoundary inspection.
package pathcontainmentprefixboundary

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Require a directory boundary in the containment check."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PathContainmentPrefixBoundary" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	// Each accepted form yields the tested path, the prefix expression and the
	// node to highlight; sized forms must compare exactly the prefix length.
	var subject, needle, length syntax.Expr
	var site syntax.Node = c
	sized := false
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "str_starts_with":
		subject, needle = semanticquery.CallArgument(c.Args, 0, "haystack"), semanticquery.CallArgument(c.Args, 1, "needle")
	case "strpos":
		// Only an identity test against 0 is a prefix test: a loose one also
		// accepts a missing needle.
		cmp, other, strict := pathPrefixComparison(c)
		offset := semanticquery.CallArgument(c.Args, 2, "offset")
		if !strict || !pathPrefixZero(ctx, other) || (offset != nil && !pathPrefixZero(ctx, offset)) {
			return
		}
		subject, needle, site = semanticquery.CallArgument(c.Args, 0, "haystack"), semanticquery.CallArgument(c.Args, 1, "needle"), cmp
	case "strncmp":
		cmp, other, _ := pathPrefixComparison(c)
		if !pathPrefixZero(ctx, other) {
			return
		}
		subject, needle, site = semanticquery.CallArgument(c.Args, 0, "string1"), semanticquery.CallArgument(c.Args, 1, "string2"), cmp
		if _, literal := semanticquery.NativeString(ctx, subject); literal {
			subject, needle = needle, subject
		}
		length, sized = semanticquery.CallArgument(c.Args, 2, "length"), true
	case "substr":
		cmp, other, _ := pathPrefixComparison(c)
		if cmp == nil || !pathPrefixZero(ctx, semanticquery.CallArgument(c.Args, 1, "offset")) {
			return
		}
		subject, needle, site = semanticquery.CallArgument(c.Args, 0, "string"), other, cmp
		length, sized = semanticquery.CallArgument(c.Args, 2, "length"), true
	default:
		return
	}
	prefix, known := semanticquery.NativeString(ctx, needle)
	if !known || prefix == "" || strings.HasSuffix(prefix, "/") || strings.HasSuffix(prefix, `\`) {
		return
	}
	if sized && !pathPrefixLength(ctx, length, prefix) {
		return
	}
	origin, ok := semanticquery.NativeValue(ctx, subject).(*syntax.FuncCall)
	if ok && semanticquery.NativeBuiltinName(ctx, origin) == "realpath" {
		ctx.ReportNode(site, message)
	}
}

// pathPrefixComparison returns the equality or identity comparison whose
// operand is the call, the opposite operand and whether the test is strict.
func pathPrefixComparison(c *syntax.FuncCall) (cmp *syntax.Binary, other syntax.Expr, strict bool) {
	var operand syntax.Expr = c
	for {
		paren, ok := operand.Parent().(*syntax.Paren)
		if !ok {
			break
		}
		operand = paren
	}
	cmp, ok := operand.Parent().(*syntax.Binary)
	if !ok {
		return nil, nil, false
	}
	switch cmp.Op.Kind {
	case syntax.TIsIdentical, syntax.TIsNotIdentical:
		strict = true
	case syntax.TIsEqual, syntax.TIsNotEqual:
	default:
		return nil, nil, false
	}
	other = cmp.Right
	if operand == cmp.Right {
		other = cmp.Left
	}
	return cmp, other, strict
}

// pathPrefixLength proves that a length argument is the byte length of
// the prefix, written as an integer or as strlen() of the same text.
func pathPrefixLength(ctx *analysis.Context, e syntax.Expr, prefix string) bool {
	if n, known := semanticquery.NativeInt(ctx, e); known {
		return n == int64(len(prefix))
	}
	measured, ok := semanticquery.NativeValue(ctx, e).(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, measured) != "strlen" {
		return false
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(measured.Args, 0, "string"))
	return known && text == prefix
}

func pathPrefixZero(ctx *analysis.Context, e syntax.Expr) bool {
	n, known := semanticquery.NativeInt(ctx, e)
	return known && n == 0
}
