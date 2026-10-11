// Package sqlitebindingindexzero implements SqliteBindingIndexZero.
package sqlitebindingindexzero

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Start SQLite binding positions at one."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteBindingIndexZero" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "SQLite3Stmt", "bindValue") && !semanticquery.NativeMethod(ctx, c, "SQLite3Stmt", "bindParam") {
		return
	}
	param := semanticquery.CallArgument(c.Args, 0, "param")
	v, known := semanticquery.NativeInt(ctx, param)
	if !known || v != 0 {
		return
	}
	if variable, ok := syntax.UnwrapParens(param).(*syntax.Variable); ok && steppedBefore(ctx, variable) {
		return
	}
	ctx.ReportNode(c, message)
}

// steppedBefore reports whether use may observe an increment, a decrement or
// a write through a by-reference closure capture. Local flow facts keep the
// assigned literal across those constructs, so the position is no longer a
// proven zero. A later change still counts when a loop repeats the use.
func steppedBefore(ctx *analysis.Context, use *syntax.Variable) bool {
	for _, access := range flowquery.VarAccesses(ctx.File, syntax.EnclosingFuncLike(use), use.Name) {
		switch by := access.By.(type) {
		case *syntax.IncDec:
		case *syntax.ClosureUse:
			if !by.ByRef {
				continue
			}
		default:
			continue
		}
		if access.Var.Span().Start < use.Span().Start || sharesLoop(use, access.Var) {
			return true
		}
	}
	return false
}

// sharesLoop reports whether a loop around use also contains other.
func sharesLoop(use, other syntax.Node) bool {
	for p := use.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if span := p.Span(); span.Start <= other.Span().Start && other.Span().End <= span.End {
				return true
			}
		}
	}
	return false
}
