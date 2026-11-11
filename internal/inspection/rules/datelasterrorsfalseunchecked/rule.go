// Package datelasterrorsfalseunchecked implements the DateLastErrorsFalseUnchecked inspection.
package datelasterrorsfalseunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DateLastErrorsFalseUnchecked" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }

const message = "Handle the false result before reading parsing diagnostics."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP82 {
		return
	}
	fetch := n.(*syntax.ArrayDimFetch)
	origin := semanticquery.NativeValue(ctx, fetch.Var)
	if ctx.Flow().Excludes(fetch.Var, "false") {
		return
	}
	errors := semanticquery.NativeDateStatic(ctx, origin, "getLastErrors")
	if errors == nil || afterFailedParse(ctx, errors) {
		return
	}
	if silentProbe(fetch) || !semanticquery.NativeLabelled(ctx, fetch.Var) && semanticquery.NativeNarrowedExcludes(ctx, fetch.Var, types.Of("array", "false"), "false") {
		return
	}
	ctx.ReportNode(fetch, message)
}
func (rule) Semantic() {}

// silentProbe recognises the contexts where PHP evaluates an offset of false
// without a diagnostic: isset, empty and the tested side of `??`.
func silentProbe(fetch *syntax.ArrayDimFetch) bool {
	for child, parent := syntax.Node(fetch), fetch.Parent(); ; child, parent = parent, parent.Parent() {
		switch p := parent.(type) {
		case *syntax.ArrayDimFetch:
			if p.Var != child {
				return false
			}
		case *syntax.Paren:
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			return p.Op.Kind == syntax.TCoalesce && p.Left == child
		default:
			return false
		}
	}
}

// afterFailedParse reports a getLastErrors() call that only runs once a date
// parse failed (E8): a failed parse always records an error, so the call
// returns an array. The conditions guarding the call must prove that an
// earlier parse returned false (an if or elseif arm, an early exit, a
// short-circuit operand, or a test of the local holding the result), and no
// other parse may run between that parse and the call.
func afterFailedParse(ctx *analysis.Context, call *syntax.StaticCall) bool {
	stmt := enclosingStatement(call)
	if stmt == nil {
		return false
	}
	return semanticquery.NativeEarlierCallProves(ctx, call, func(x syntax.Expr) (types.Type, func(types.Type) bool) {
		if !parseCall(ctx, x) || parsesAgain(ctx, stmt, x, call) {
			return types.Unknown, nil
		}
		// A parsed date is always truthy: success is modelled as true.
		return types.Bool, func(t types.Type) bool { return t.OnlyOf("false") }
	})
}

// parsesAgain reports a date parse other than first and call that may run
// between them: in stmt, the statement holding call; written between them,
// outside a branch that never falls through to call; or anywhere in a loop
// around call that first is not in. A parse is a parsing call or
// constructing a date (or an object of a class it cannot resolve).
func parsesAgain(ctx *analysis.Context, stmt syntax.Node, first syntax.Expr, call *syntax.StaticCall) bool {
	from, to := first.Span().End, call.Span().Start
	var loops []syntax.Node
	for n := call.Parent(); n != nil && !syntax.IsFuncLike(n); n = n.Parent() {
		switch n.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if !within(first, n) {
				loops = append(loops, n)
			}
		}
	}
	relevant := func(n syntax.Node) bool {
		if within(n, stmt) || n.Span().Start >= from && n.Span().End <= to {
			return true
		}
		for _, l := range loops {
			if within(n, l) {
				return true
			}
		}
		return false
	}
	found := false
	scope := syntax.EnclosingVariableScope(call)
	visit := func(n syntax.Node) bool {
		if found || n == syntax.Node(first) || n == syntax.Node(call) || n != scope && syntax.IsFuncLike(n) {
			return false
		}
		if s, ok := n.(syntax.Stmt); ok && len(loops) == 0 && n.Span().Start >= from && !within(call, n) && leavingBranch(s) {
			return false // never falls through to call
		}
		if !relevant(n) {
			return true
		}
		switch x := n.(type) {
		case *syntax.New:
			c := semanticquery.NativeNewClass(ctx, x)
			found = c == nil || ctx.Index().IsSubtype(c.FQN, "DateTimeInterface", ctx.PHP)
		case syntax.Expr:
			found = parseCall(ctx, x)
		}
		return !found
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}

// within reports whether n lies inside outer (or is outer).
func within(n, outer syntax.Node) bool {
	return n.Span().Start >= outer.Span().Start && n.Span().End <= outer.Span().End
}

// leavingBranch reports whether s is the body of an if, elseif or else arm
// that never falls through.
func leavingBranch(s syntax.Stmt) bool {
	switch s.Parent().(type) {
	case *syntax.If, *syntax.ElseIf, *syntax.Else:
		return syntax.Terminates(s)
	}
	return false
}

// enclosingStatement returns the statement holding n directly in a block;
// nil past a function or closure.
func enclosingStatement(n syntax.Node) syntax.Node {
	for {
		parent := n.Parent()
		switch parent.(type) {
		case nil, *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			return nil
		case *syntax.Block:
			return n
		}
		n = parent
	}
}

// parseCall reports a date parsing call that returns false on failure.
func parseCall(ctx *analysis.Context, x syntax.Expr) bool {
	if semanticquery.NativeDateStatic(ctx, x, "createFromFormat") != nil {
		return true
	}
	if f, ok := x.(*syntax.FuncCall); ok {
		for _, name := range []string{"date_create", "date_create_immutable", "date_create_from_format", "date_create_immutable_from_format"} {
			if semanticquery.NativeBuiltin(ctx, f, name) {
				return true
			}
		}
	}
	return false
}
