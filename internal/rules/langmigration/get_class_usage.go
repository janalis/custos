package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// getClassUsage reports get_class() calls whose argument may be null and is
// not null-checked earlier in the enclosing function.
type getClassUsage struct{}

func init() { register(getClassUsage{}) }

func (getClassUsage) ID() string { return "GetClassUsage" }

func (getClassUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (getClassUsage) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP71 { // D3
		return
	}
	call := n.(*syntax.FuncCall)
	if _, part, ok := util.FuncNamePart(call); !ok || !strings.EqualFold(part, "get_class") { // D1 (any case)
		return
	}
	if !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, "get_class") { // D1
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 1 { // D2
		return
	}
	arg := args[0]
	if !ctx.TypeOf(arg).Has("null") && !isNullDefaultedParam(arg) { // D4
		return
	}
	if scope := syntax.EnclosingFuncLike(call); scope != nil && nullCheckedBefore(ctx.File, scope, arg) { // D5
		return
	}
	ctx.ReportNode(call, "get_class() rejects null on PHP 7.2+; guard the argument.")
}

// isNullDefaultedParam implements D4b.
func isNullDefaultedParam(arg syntax.Expr) bool {
	v, ok := arg.(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return false
	}
	scope := syntax.EnclosingFuncLike(arg)
	if scope == nil {
		return false
	}
	if _, ok := scope.(*syntax.ArrowFunction); ok {
		return false
	}
	for _, p := range syntax.FuncLikeParams(scope) {
		if p.Var != nil && p.Var.Name == v.Name && p.Default != nil && syntax.IsNullConst(syntax.UnwrapParens(p.Default)) {
			return true
		}
	}
	return false
}

// nullCheckedBefore implements the null-check search of the spec.
func nullCheckedBefore(f *syntax.File, scope syntax.Node, arg syntax.Expr) bool {
	body := syntax.FuncLikeBody(scope)
	if body == nil {
		return false
	}
	checked := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if checked || n.Span().Start >= arg.Span().Start {
			return false
		}
		if n.Kind() == arg.Kind() && util.EquivalentFoldNames(f, n, arg) && isNullCheck(n) {
			checked = true
			return false
		}
		return true
	})
	return checked
}

func isNullCheck(c syntax.Node) bool {
	switch p := c.Parent().(type) {
	case *syntax.Isset: // C1
		return true
	case *syntax.Empty:
		return true
	case *syntax.Instanceof: // C2
		return true
	case *syntax.Binary: // C3
		switch p.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
			other := p.Right
			if p.Right == c {
				other = p.Left
			}
			if syntax.IsNullConst(syntax.UnwrapParens(other)) {
				return true
			}
		}
	}
	parent, child := util.ParentSkipParens(c) // C4
	switch p := parent.(type) {
	case *syntax.If:
		return p.Cond == child
	case *syntax.ElseIf:
		return p.Cond == child
	case *syntax.While:
		return p.Cond == child
	case *syntax.DoWhile:
		return p.Cond == child
	case *syntax.Unary:
		return p.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return true
		}
	case *syntax.Ternary:
		return p.Then != nil && p.Cond == child
	}
	return false
}
