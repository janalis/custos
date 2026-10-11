// Package connecttimeoutassumedtoboundreads implements the native ConnectTimeoutAssumedToBoundReads inspection.
package connecttimeoutassumedtoboundreads

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Set a separate timeout for stream reads."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ConnectTimeoutAssumedToBoundReads" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "fgets", "fread", "stream_get_contents", "fgetcsv":
	default:
		return
	}
	handle := semanticquery.CallArgument(c.Args, 0, "stream")
	origin, ok := semanticquery.NativeValue(ctx, handle).(*syntax.FuncCall)
	if !ok {
		return
	}
	name := semanticquery.NativeBuiltinName(ctx, origin)
	if name != "fsockopen" && name != "pfsockopen" {
		return
	}
	if !positiveTimeout(ctx, semanticquery.CallArgument(origin.Args, 4, "timeout")) {
		return
	}

	g := guard{ctx: ctx, handle: handle, origin: origin}
	for child := syntax.Node(c); child != nil && !syntax.IsVariableScope(child); child = child.Parent() {
		if g.branch(child.Parent(), child) || g.earlyExit(child) {
			return
		}
	}

	ctx.ReportNode(c, message)
}

// positiveTimeout accepts a known integer or float literal above zero.
func positiveTimeout(ctx *analysis.Context, e syntax.Expr) bool {
	if n, ok := semanticquery.NativeInt(ctx, e); ok {
		return n > 0
	}
	lit, ok := semanticquery.NativeValue(ctx, e).(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitFloat {
		return false
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(lit.Raw, "_", ""), 64)
	return err == nil && v > 0
}

// guard recognizes checked stream_set_timeout calls on one socket handle.
type guard struct {
	ctx    *analysis.Context
	handle syntax.Expr
	origin *syntax.FuncCall
}

// branch reports whether reaching child inside parent proves a successful
// configuration: the matching arm of an if, a ternary or a short-circuit.
func (g guard) branch(parent, child syntax.Node) bool {
	switch p := parent.(type) {
	case *syntax.If:
		if child == syntax.Node(p.Cond) {
			return false
		}
		if child == syntax.Node(p.Body) {
			return g.proves(p.Cond, true)
		}
		// Every elseif and the else are reached only after the if condition failed.
		if g.proves(p.Cond, false) {
			return true
		}
		for _, alt := range p.ElseIfs {
			if child == syntax.Node(alt) {
				break
			}
			if g.proves(alt.Cond, false) {
				return true
			}
		}
	case *syntax.ElseIf:
		return child == syntax.Node(p.Body) && g.proves(p.Cond, true)
	case *syntax.Ternary:
		if child != syntax.Node(p.Cond) {
			return g.proves(p.Cond, child == syntax.Node(p.Then))
		}
	case *syntax.Binary:
		if child != syntax.Node(p.Right) {
			return false
		}
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return g.proves(p.Left, true)
		case syntax.TBooleanOr, syntax.TOr:
			return g.proves(p.Left, false)
		}
	}
	return false
}

// earlyExit reports whether an earlier statement of the same list leaves the
// path whenever the configuration did not succeed.
func (g guard) earlyExit(n syntax.Node) bool {
	st, ok := n.(syntax.Stmt)
	if !ok {
		return false
	}
	list, i, _ := astquery.StmtList(g.ctx.File, st)
	for _, prev := range list[:max(i, 0)] {
		if exit, ok := prev.(*syntax.If); ok && syntax.Terminates(exit.Body) && g.proves(exit.Cond, false) {
			return true
		}
	}
	return false
}

// proves reports whether cond evaluating to truth implies that
// stream_set_timeout returned true for the handle opened by origin.
func (g guard) proves(cond syntax.Expr, truth bool) bool {
	switch e := syntax.UnwrapParens(cond).(type) {
	case *syntax.FuncCall:
		if !truth || !semanticquery.NativeBuiltin(g.ctx, e, "stream_set_timeout") {
			return false
		}
		stream := semanticquery.CallArgument(e.Args, 0, "stream")
		return semanticquery.ExpansionCSame(g.ctx, g.handle, stream) && semanticquery.NativeValue(g.ctx, stream) == syntax.Expr(g.origin)
	case *syntax.Unary:
		return e.Op.Kind == syntax.TExclaim && g.proves(e.Expr, !truth)
	case *syntax.Binary:
		switch e.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (g.proves(e.Left, true) || g.proves(e.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (g.proves(e.Left, false) || g.proves(e.Right, false))
		case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
			if e.Op.Kind == syntax.TIsNotIdentical || e.Op.Kind == syntax.TIsNotEqual {
				truth = !truth
			}
			for _, pair := range [][2]syntax.Expr{{e.Left, e.Right}, {e.Right, e.Left}} {
				if want, ok := astquery.BoolConst(syntax.UnwrapParens(pair[1])); ok {
					return g.proves(pair[0], truth == want)
				}
			}
		}
	}
	return false
}
