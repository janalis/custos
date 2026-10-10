// Package readdirfalsyfilenameloss implements the native ReaddirFalsyFilenameLoss inspection.
package readdirfalsyfilenameloss

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare the directory entry with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReaddirFalsyFilenameLoss" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "readdir") {
		return
	}
	use := semanticquery.NativeConditionUse(ctx, c, true)
	if b, ok := use.(*syntax.Binary); ok && (b.Op.Kind == syntax.TIsIdentical || b.Op.Kind == syntax.TIsNotIdentical) {
		return
	}
	if use == nil {
		return
	}
	op := " !== false"
	inner := use
	if u, ok := syntax.UnwrapParens(use).(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		inner = u.Expr
		op = " === false"
	} else if b, ok := syntax.UnwrapParens(use).(*syntax.Binary); ok {
		var operand syntax.Expr = c
		for operand.Parent() != b {
			operand = operand.Parent().(syntax.Expr)
		}
		other := b.Right
		if b.Right == operand {
			other = b.Left
		}
		v, _ := semanticquery.NativeTruth(ctx, other)
		equal := b.Op.Kind == syntax.TIsEqual || b.Op.Kind == syntax.TIsIdentical
		if equal == v {
			op = " !== false"
		} else {
			op = " === false"
		}
		inner = operand
	}
	text := ctx.Text(inner)
	if strings.Contains(text, "/*") || strings.Contains(text, "//") || strings.Contains(text, "#") {
		ctx.ReportNode(use, message)
		return
	}
	fix := diagnostic.Fix{Title: "Compare the entry with false", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: use.Span(), NewText: "(" + text + ")" + op}}
	}}
	ctx.ReportNode(use, message, fix)
}
