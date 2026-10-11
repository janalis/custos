// Package zipfirstentrytreatedasmissing implements the native ZipFirstEntryTreatedAsMissing inspection.
package zipfirstentrytreatedasmissing

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Distinguish a missing ZIP entry from index zero."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipFirstEntryTreatedAsMissing" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ZipArchive", "locateName") {
		return
	}
	p, child := astquery.ParentSkipParens(c)
	if a, ok := p.(*syntax.Assign); ok {
		// A compound assignment yields a combined value, not the lookup result.
		if a.Op.Kind != syntax.TEqual {
			return
		}
		p, child = astquery.ParentSkipParens(a)
	}
	// A short-circuited nullsafe chain yields null, which is neither index
	// zero nor false: every strict rewrite would move that case to the other
	// branch, so such calls are reported without a fix.
	nullable := c.NullSafe || strings.Contains(ctx.Text(c.Var), "?->")
	if b, ok := p.(*syntax.Binary); ok {
		checkLooseFalse(ctx, b, child, nullable)
		return
	}
	target := child
	neg, negated := p.(*syntax.Unary)
	if negated = negated && neg.Op.Kind == syntax.TExclaim; negated {
		target = neg
		p, _ = astquery.ParentSkipParens(neg)
	}
	if !isCondition(p) {
		return
	}
	text := ctx.Text(child)
	if _, assignment := child.(*syntax.Assign); assignment {
		text = "(" + text + ")"
	}
	suffix := " !== false"
	if negated {
		suffix = " === false"
		prefix := ctx.SpanText(syntax.Span{Start: neg.Span().Start, End: child.Span().Start})
		nullable = nullable || strings.Contains(prefix, "/*") || strings.Contains(prefix, "//") || strings.Contains(prefix, "#")
	}
	if nullable {
		ctx.ReportNode(target, message)
		return
	}
	ctx.ReportNode(target, message, astquery.ReplaceFix(target.Span(), text+suffix))
}

// checkLooseFalse reports a loose comparison of the lookup result with the
// false constant: index zero compares equal to false. Only the operator is
// rewritten, so operands, comments and evaluation order stay untouched.
func checkLooseFalse(ctx *analysis.Context, b *syntax.Binary, operand syntax.Node, nullable bool) {
	strict := "==="
	switch b.Op.Kind {
	case syntax.TIsEqual:
	case syntax.TIsNotEqual:
		strict = "!=="
	default:
		return
	}
	other := b.Right
	if syntax.Node(other) == operand {
		other = b.Left
	}
	if value, ok := astquery.BoolConst(syntax.UnwrapParens(other)); !ok || value {
		return
	}
	if nullable {
		ctx.ReportNode(b, message)
		return
	}
	ctx.ReportNode(b, message, astquery.ReplaceFix(b.Op.Span, strict))
}

// isCondition reports whether parent is a branch or loop statement. An
// expression's direct statement parent of these kinds always holds it as its
// condition: bodies are statements.
func isCondition(parent syntax.Node) bool {
	switch parent.(type) {
	case *syntax.If, *syntax.ElseIf, *syntax.While, *syntax.DoWhile:
		return true
	}
	return false
}
