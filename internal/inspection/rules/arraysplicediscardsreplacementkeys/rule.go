// Package arraysplicediscardsreplacementkeys implements the native ArraySpliceDiscardsReplacementKeys inspection.
package arraysplicediscardsreplacementkeys

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve replacement keys before reading the inserted key."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArraySpliceDiscardsReplacementKeys" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_splice") {
		return
	}
	local := semanticquery.CallArgument(call.Args, 0, "array")
	if variable(local) == "" || flowquery.MayHaveSideEffects(call.Args) {
		return
	}
	stmt, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	access, ok := astquery.NextEcho(ctx.File, stmt).(*syntax.ArrayDimFetch)
	if !ok || variable(access.Var) != variable(local) {
		return
	}
	original, known := initialEntries(ctx, stmt, variable(local))
	if !known && private(ctx, call, local, access.Var) {
		original, known = semanticquery.NativeArrayEntries(ctx, local)
	}
	replacement, ok := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 3, "replacement"))
	if !known || !ok {
		return
	}
	key, known := semanticquery.NativeArrayKey(ctx, access.Dim)
	if !known || len(key) < 2 || key[:2] != "s:" {
		return
	}
	if _, exists := original[key]; exists {
		return
	}
	if _, exists := replacement[key]; exists {
		ctx.ReportNode(call, message)
	}
}

// initialEntries returns the entries of the literal array assigned to $name by
// the statement directly before s. Requiring adjacency keeps the proof local:
// no call, alias write or rebinding can run between the assignment and s, so a
// reference capture or a global binding of the variable cannot add keys.
func initialEntries(ctx *analysis.Context, s syntax.Stmt, name string) (map[string]syntax.Expr, bool) {
	prev, ok := astquery.PrevStmt(ctx.File, s)
	if !ok {
		return nil, false
	}
	expr, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return nil, false
	}
	assign, ok := syntax.UnwrapParens(expr.Expr).(*syntax.Assign)
	if !ok || assign.ByRef || assign.Op.Kind != syntax.TEqual || variable(assign.Var) != name {
		return nil, false
	}
	return semanticquery.NativeArrayEntries(ctx, assign.Value)
}

// private reports whether the spliced variable is a function local that only
// plain assignments can change: apart from the splice argument and the read
// that follows it, every occurrence in the function is the target of a plain
// `$name = value`. The variable is then never read into an alias, captured,
// declared global or static, or written element by element, so the tracked
// assignment that reaches the splice is the whole content. Top-level code has
// no such scope: its variables are globals any call can change.
func private(ctx *analysis.Context, call *syntax.FuncCall, local, read syntax.Expr) bool {
	scope := syntax.EnclosingVariableScope(call)
	if scope == nil || !syntax.IsFuncLike(scope) {
		return false
	}
	name := variable(local)
	for _, p := range syntax.FuncLikeParams(scope) {
		if p.Var != nil && p.Var.Name == name {
			return false
		}
	}
	if closure, ok := scope.(*syntax.Closure); ok {
		for _, use := range closure.Uses {
			if use.Var != nil && use.Var.Name == name {
				return false
			}
		}
	}
	for _, a := range flowquery.VarAccesses(ctx.File, scope, name) {
		if a.Var == syntax.UnwrapParens(local) || a.Var == syntax.UnwrapParens(read) {
			continue
		}
		assign, ok := a.By.(*syntax.Assign)
		if !ok || assign.ByRef || assign.Op.Kind != syntax.TEqual || syntax.UnwrapParens(assign.Var) != syntax.Expr(a.Var) {
			return false
		}
	}
	rebinds := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Include, *syntax.Eval:
			rebinds = true
		case *syntax.FuncCall:
			_, extract := astquery.IsFuncNamedFold(n, "extract")
			rebinds = rebinds || extract
		}
		return !rebinds
	})
	return !rebinds
}

func variable(e syntax.Expr) string {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		return ""
	}
	return v.Name
}
