// Package sqlitefetchbothleaksduplicatecolumns implements SqliteFetchBothLeaksDuplicateColumns.
package sqlitefetchbothleaksduplicatecolumns

import (
	"slices"
	"strings"

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
	for _, a := range q.Args.Args {
		// A placeholder yields a Closure and a spread hides the mode.
		if arg, ok := a.(*syntax.Arg); !ok || arg.Unpack {
			return
		}
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

// jsonOnly proves that the fetched row reaches nothing but json_encode.
// Every other mention of the row variable invalidates the proof, as does any
// construct that reaches locals without naming them.
func jsonOnly(ctx *analysis.Context, producer *syntax.MethodCall) bool {
	parent, _ := astquery.ParentSkipParens(producer)
	assignment, ok := parent.(*syntax.Assign)
	if !ok {
		return true
	}
	variable, ok := assignment.Var.(*syntax.Variable)
	if !ok || variable.NameExpr != nil || assignment.ByRef || assignment.Op.Kind != syntax.TEqual {
		return false
	}
	scope := syntax.EnclosingVariableScope(producer)
	// A loop or a jump can carry the row back to textually earlier code.
	repeated := false
	for p := producer.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			repeated = true
		}
	}
	safe, budget := true, 256
	visit := func(n syntax.Node) bool {
		budget--
		if budget < 0 {
			safe = false
			return false
		}
		switch x := n.(type) {
		case *syntax.Assign:
			safe = safe && !x.ByRef
		case *syntax.Include, *syntax.Eval, *syntax.Goto:
			safe = safe && syntax.EnclosingVariableScope(x) != scope
		case *syntax.FuncCall:
			// These expose every local by name without mentioning it.
			if _, part, ok := astquery.FuncNamePart(x); ok && slices.Contains([]string{"compact", "extract", "get_defined_vars"}, strings.ToLower(part)) {
				safe = safe && syntax.EnclosingVariableScope(x) != scope
			}
		case *syntax.Variable:
			safe = safe && consumer(ctx, x, variable, producer, scope, repeated)
		}
		return safe
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return safe
}

// consumer reports whether one variable mention keeps the json-only proof.
func consumer(ctx *analysis.Context, v, variable *syntax.Variable, producer *syntax.MethodCall, scope syntax.Node, repeated bool) bool {
	if v.NameExpr != nil {
		return syntax.EnclosingVariableScope(v) != scope
	}
	if v.Name == "GLOBALS" && scope == nil {
		return false
	}
	if v.Name != variable.Name || v == variable {
		return true
	}
	if syntax.EnclosingVariableScope(v) != scope {
		return false
	}
	switch p := v.Parent().(type) {
	case *syntax.Global, *syntax.StaticVar:
		return false
	case *syntax.Param:
		return !p.ByRef
	case *syntax.ClosureUse:
		return !p.ByRef
	case *syntax.Assign:
		// A complete overwrite ends the row instead of consuming it.
		if p.Var == syntax.Expr(v) && p.Op.Kind == syntax.TEqual {
			return true
		}
	case *syntax.Arg:
		if list, ok := p.Parent().(*syntax.ArgList); ok {
			if call, ok := list.Parent().(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, call, "json_encode") && semanticquery.CallArgument(call.Args, 0, "value") == syntax.Expr(v) {
				return true
			}
		}
	}
	if !repeated && v.Span().Start < producer.Span().End {
		return true
	}
	// Only a mention proven to hold another complete value is unrelated;
	// an unknown value may still be the row after an element write.
	value := semanticquery.NativeValue(ctx, v)
	return !repeated && value != nil && value != syntax.Expr(producer)
}
