// Package asyncsignalsetterreturnmisread checks the previous-state return of the signal setter.
package asyncsignalsetterreturnmisread

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Query the signal state after changing it."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AsyncSignalSetterReturnMisread" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pcntl_async_signals") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "enable")
	if arg == nil {
		return
	}
	value := semanticquery.NativeValue(ctx, arg)
	if null, ok := value.(*syntax.ConstFetch); ok && strings.EqualFold(null.Name.Value, "null") {
		return
	}
	if _, known := semanticquery.NativeTruth(ctx, arg); !known {
		return
	}
	if !semanticquery.ExpansionCTruthy(c) {
		return
	}
	parent, child := astquery.ParentSkipParens(c)
	if u, ok := parent.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		condition, _ := astquery.ParentSkipParens(u)
		if branch, ok := condition.(*syntax.If); ok && branch.Cond == u && statementList(branch) {
			name := ctx.Text(c.Name)
			start := branch.Span().Start
			fix := diagnostic.Fix{Title: "Change signal state before querying it", Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: syntax.Span{Start: start, End: start}, NewText: ctx.Text(c) + ";\n"}, {Span: c.Span(), NewText: name + "()"}}
			}}
			ctx.ReportNode(u, message, fix)
			return
		}
	}
	ctx.ReportNode(child, message)
}

// Inserting a preceding statement is safe only in a statement list.
func statementList(branch *syntax.If) bool {
	if branch.Parent() == nil {
		return true
	}
	_, ok := branch.Parent().(*syntax.Block)
	return ok
}
