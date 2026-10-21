// Package datelasterrorsfalseunchecked implements the DateLastErrorsFalseUnchecked inspection.
package datelasterrorsfalseunchecked

import (
	"slices"
	"strings"

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
// returns an array. Two shapes are read: the call in the first statement of
// the body of an `if` that holds when the parse failed, or in the statement
// right after an `if` whose condition holds when it succeeded and whose body
// never falls through.
func afterFailedParse(ctx *analysis.Context, call *syntax.StaticCall) bool {
	stmt, block := enclosingStatement(call)
	if block == nil {
		return false
	}
	if parsesAgain(ctx, stmt, call) {
		return false // the call may report another parse
	}
	i := slices.IndexFunc(block.Stmts, func(s syntax.Stmt) bool { return syntax.Node(s) == stmt })
	if i == 0 {
		guard, ok := block.Parent().(*syntax.If)
		return ok && syntax.Node(guard.Body) == syntax.Node(block) && failed(ctx, guard.Cond, false)
	}
	prev, ok := block.Stmts[i-1].(*syntax.If)
	return ok && len(prev.ElseIfs) == 0 && prev.Else == nil && syntax.Terminates(prev.Body) && failed(ctx, prev.Cond, true)
}

// parsesAgain reports a date parse in stmt other than call: a parsing call,
// or constructing a date (or an object of a class it cannot resolve).
func parsesAgain(ctx *analysis.Context, stmt syntax.Node, call *syntax.StaticCall) bool {
	found := false
	syntax.Inspect(stmt, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.New:
			c := semanticquery.NativeNewClass(ctx, x)
			found = found || c == nil || ctx.Index().IsSubtype(c.FQN, "DateTimeInterface", ctx.PHP)
		case syntax.Expr:
			found = found || x != syntax.Expr(call) && parse(ctx, x)
		}
		return !found
	})
	return found
}

// enclosingStatement returns the statement holding n directly in a block,
// and that block; nil past a function or closure.
func enclosingStatement(n syntax.Node) (syntax.Node, *syntax.Block) {
	for {
		parent := n.Parent()
		switch p := parent.(type) {
		case nil, *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			return nil, nil
		case *syntax.Block:
			return n, p
		}
		n = parent
	}
}

// failed reports whether condition c, holding (or failing, when
// succeeded), means a date parse failed: `!parse`, `false === parse`,
// `parse === false` (or, when succeeded, the parse tested for truth or
// `false !== parse`), where parse is a parsing call returning false on
// failure, possibly assigned.
func failed(ctx *analysis.Context, c syntax.Expr, succeeded bool) bool {
	c = syntax.UnwrapParens(c)
	if !succeeded {
		if u, ok := c.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
			return parse(ctx, u.Expr)
		}
		if b, ok := c.(*syntax.Binary); ok && b.Op.Kind == syntax.TIsIdentical {
			return isFalse(b.Left) && parse(ctx, b.Right) || isFalse(b.Right) && parse(ctx, b.Left)
		}
		return false
	}
	if b, ok := c.(*syntax.Binary); ok && b.Op.Kind == syntax.TIsNotIdentical {
		return isFalse(b.Left) && parse(ctx, b.Right) || isFalse(b.Right) && parse(ctx, b.Left)
	}
	return parse(ctx, c)
}

func isFalse(x syntax.Expr) bool {
	c, ok := syntax.UnwrapParens(x).(*syntax.ConstFetch)
	return ok && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false")
}

// parse reports a date parsing call that returns false on failure, possibly
// as the value of a plain assignment.
func parse(ctx *analysis.Context, x syntax.Expr) bool {
	x = syntax.UnwrapParens(x)
	if a, ok := x.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
		x = syntax.UnwrapParens(a.Value)
	}
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
