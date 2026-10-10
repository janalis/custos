// Package nonblockingemptyreadtreatedaseof implements the native NonBlockingEmptyReadTreatedAsEof inspection.
package nonblockingemptyreadtreatedaseof

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check EOF separately from an empty nonblocking read."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonBlockingEmptyReadTreatedAsEof" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsIdentical {
		return
	}
	read, empty := b.Left, b.Right
	if text, known := semanticquery.NativeString(ctx, read); known && text == "" {
		read, empty = empty, read
	}
	text, known := semanticquery.NativeString(ctx, empty)
	if !known || text != "" {
		return
	}
	c, ok := syntax.UnwrapParens(read).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "fread") {
		return
	}
	branch, ok := b.Parent().(*syntax.If)
	if !ok || branch.Cond != b {
		return
	}
	block, ok := branch.Body.(*syntax.Block)
	if !ok || len(block.Stmts) != 1 {
		return
	}
	st, ok := block.Stmts[0].(*syntax.ExprStmt)
	if !ok {
		return
	}
	close, ok := st.Expr.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, close, "fclose") {
		return
	}
	handle := semanticquery.CallArgument(c.Args, 0, "stream")
	identity := ctx.Flow().Value(handle).Identity
	if identity == 0 || ctx.Flow().Value(semanticquery.CallArgument(close.Args, 0, "stream")).Identity != identity {
		return
	}
	prior := semanticquery.NativePriorCalls(ctx, c, handle, "stream_set_blocking")
	if len(prior) == 0 {
		return
	}
	setter, ok := prior[len(prior)-1].(*syntax.FuncCall)
	if ok {
		truth, known := semanticquery.NativeTruth(ctx, semanticquery.CallArgument(setter.Args, 1, "enable"))
		if known && !truth {
			ctx.ReportNode(b, message)
		}
	}
}
