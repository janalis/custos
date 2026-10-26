// Package zipopentruthyresult implements the native ZipOpenTruthyResult inspection.
package zipopentruthyresult

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Accept only true from ZipArchive::open."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipOpenTruthyResult" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ZipArchive", "open") {
		return
	}
	parent, child := astquery.ParentSkipParens(c)
	op := "==="
	switch p := parent.(type) {
	case *syntax.If, *syntax.ElseIf, *syntax.While:
	case *syntax.Unary:
		if p.Op.Kind != syntax.TExclaim {
			return
		}
		op = "!=="
		written := ctx.Text(p)
		condition, _ := astquery.ParentSkipParens(p)
		if branchCondition(condition) && !strings.Contains(written, "/*") && !strings.Contains(written, "//") && !strings.Contains(written, "#") {
			ctx.ReportNode(c, message, astquery.ReplaceFix(p.Span(), strictTrue(ctx, ctx.Text(c), op)))
		} else {
			ctx.ReportNode(c, message)
		}
		return
	default:
		return
	}
	ctx.ReportNode(c, message, astquery.ReplaceFix(child.Span(), strictTrue(ctx, ctx.Text(child), op)))
}

// strictTrue writes the comparison of call with true, in the configured
// operand order. The result is a whole branch condition, so it needs no
// parentheses of its own.
func strictTrue(ctx *analysis.Context, call, op string) string {
	if ctx.ComparisonStyle == analysis.StyleYoda {
		return "true " + op + " " + call
	}
	return call + " " + op + " true"
}

// branchCondition reports whether n owns an if, elseif or while condition.
func branchCondition(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.If, *syntax.ElseIf, *syntax.While:
		return true
	}
	return false
}
