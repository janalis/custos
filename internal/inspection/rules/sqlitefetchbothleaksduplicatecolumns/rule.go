// Package sqlitefetchbothleaksduplicatecolumns implements SqliteFetchBothLeaksDuplicateColumns.
package sqlitefetchbothleaksduplicatecolumns

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Fetch named SQLite columns before JSON serialization."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteFetchBothLeaksDuplicateColumns" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "json_encode") {
		return
	}
	q, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "value")).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, q, "SQLite3Result", "fetchArray") {
		return
	}
	mode := semanticquery.CallArgument(q.Args, 0, "mode")
	if mode != nil && !semanticquery.ExpansionGlobalConstant(ctx, mode, "SQLITE3_BOTH") {
		return
	}
	if !jsonOnly(ctx, q) {
		return
	}
	var fixes []diagnostic.Fix
	if mode == nil && len(q.Args.Args) == 0 {
		s := syntax.Span{Start: q.Args.Span().End - 1, End: q.Args.Span().End - 1}
		fixes = append(fixes, edit(s, semanticquery.QualifiedGlobalConst(ctx, "SQLITE3_ASSOC", q.Span().Start)))
	} else if cf, ok := mode.(*syntax.ConstFetch); ok {
		fixes = append(fixes, edit(cf.Span(), semanticquery.QualifiedGlobalConst(ctx, "SQLITE3_ASSOC", cf.Span().Start)))
	}
	ctx.ReportNode(q, message, fixes...)
}

func edit(span syntax.Span, text string) diagnostic.Fix {
	return diagnostic.Fix{Title: message, Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: text}} }}
}

func jsonOnly(ctx *analysis.Context, producer *syntax.MethodCall) bool {
	parent, _ := astquery.ParentSkipParens(producer)
	assignment, ok := parent.(*syntax.Assign)
	if !ok {
		return true
	}
	variable, ok := assignment.Var.(*syntax.Variable)
	if !ok {
		return false
	}
	safe, budget := true, 256
	visit := func(n syntax.Node) bool {
		budget--
		if budget < 0 {
			safe = false
			return false
		}

		if assignment, ok := n.(*syntax.Assign); ok && assignment.ByRef {
			safe = false
			return false
		}
		v, ok := n.(*syntax.Variable)
		if ok && v.Name == variable.Name && syntax.EnclosingVariableScope(v) != syntax.EnclosingVariableScope(producer) {
			safe = false
			return false
		}
		if !ok || v.Name != variable.Name || v == variable || v.Span().Start < producer.Span().End || semanticquery.NativeValue(ctx, v) != producer {
			return true
		}
		arg, ok := v.Parent().(*syntax.Arg)
		if ok {
			list, ok := arg.Parent().(*syntax.ArgList)
			if ok {
				if call, ok := list.Parent().(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, call, "json_encode") && semanticquery.CallArgument(call.Args, 0, "value") == v {
					return true
				}
			}
		}
		safe = false
		return false
	}
	if scope := syntax.EnclosingVariableScope(producer); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return safe
}
