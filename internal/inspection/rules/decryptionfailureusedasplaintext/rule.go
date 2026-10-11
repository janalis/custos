// Package decryptionfailureusedasplaintext implements the native DecryptionFailureUsedAsPlaintext inspection.
package decryptionfailureusedasplaintext

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check decryption success before using the plaintext."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DecryptionFailureUsedAsPlaintext" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	sink := n.(*syntax.FuncCall)
	var data syntax.Expr
	switch semanticquery.NativeBuiltinName(ctx, sink) {
	case "strlen":
		data = semanticquery.CallArgument(sink.Args, 0, "string")
	case "file_put_contents":
		data = semanticquery.CallArgument(sink.Args, 1, "data")
	default:
		return
	}
	call, ok := semanticquery.NativeValue(ctx, data).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, call, "openssl_decrypt") || ctx.Flow().Excludes(data, "false") || semanticquery.NativeSentinelGuard(ctx, data, "false") || guarded(ctx, data, call) {
		return
	}
	ctx.ReportNode(data, message)
}

// guard ties condition operands to one local holding one decryption result.
type guard struct {
	ctx  *analysis.Context
	name string
	call *syntax.FuncCall
}

// guarded proves that a condition on the path to data rejects a false result
// of the same reaching call: an enclosing branch, a short-circuit operand, a
// ternary arm or an earlier terminating if in an enclosing statement list.
func guarded(ctx *analysis.Context, data syntax.Expr, call *syntax.FuncCall) bool {
	v, ok := syntax.UnwrapParens(data).(*syntax.Variable)
	if !ok || v.Name == "" {
		return false
	}
	g := guard{ctx, v.Name, call}
	at := data.Span()
	in := func(n syntax.Node) bool { return n.Span().Contains(at) }
	for p := data.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch p := p.(type) {
		case *syntax.Binary:
			if in(p.Right) && (isAnd(p) && g.holds(p.Left, true) || isOr(p) && g.holds(p.Left, false)) {
				return true
			}
		case *syntax.Ternary:
			if p.Then != nil && in(p.Then) && g.holds(p.Cond, true) || in(p.Else) && g.holds(p.Cond, false) {
				return true
			}
		case *syntax.ElseIf:
			if in(p.Body) && g.holds(p.Cond, true) {
				return true
			}
		case *syntax.If:
			if in(p.Cond) {
				break
			}
			if in(p.Body) {
				if g.holds(p.Cond, true) {
					return true
				}
				break
			}
			// A later arm runs only after every earlier condition failed.
			if g.holds(p.Cond, false) {
				return true
			}
			for _, arm := range p.ElseIfs {
				if arm.Span().End <= at.Start && g.holds(arm.Cond, false) {
					return true
				}
			}
		}
		st, ok := p.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, _ := astquery.StmtList(ctx.File, st)
		for _, previous := range list[:max(i, 0)] {
			if exit, ok := previous.(*syntax.If); ok && exit.Else == nil && len(exit.ElseIfs) == 0 && syntax.Terminates(exit.Body) && g.holds(exit.Cond, false) {
				return true
			}
		}
	}
	return false
}

func isAnd(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanAnd || b.Op.Kind == syntax.TAnd
}

func isOr(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanOr || b.Op.Kind == syntax.TOr
}

// holds reports whether cond evaluating to truth implies the result is not false.
func (g guard) holds(cond syntax.Expr, truth bool) bool {
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && g.holds(c.Expr, !truth)
	case *syntax.Binary:
		switch {
		case isAnd(c):
			return truth && (g.holds(c.Left, true) || g.holds(c.Right, true))
		case isOr(c):
			return !truth && (g.holds(c.Left, false) || g.holds(c.Right, false))
		case c.Op.Kind == syntax.TIsNotIdentical || c.Op.Kind == syntax.TIsNotEqual:
			return truth && g.comparesFalse(c)
		case c.Op.Kind == syntax.TIsIdentical || c.Op.Kind == syntax.TIsEqual:
			return !truth && g.comparesFalse(c)
		}
		return false
	case *syntax.FuncCall:
		return truth && semanticquery.NativeBuiltin(g.ctx, c, "is_string") && g.same(semanticquery.CallArgument(c.Args, 0, "value"))
	case *syntax.Empty:
		return !truth && g.same(c.Expr)
	}
	return truth && g.same(cond)
}

// comparesFalse matches the result against the false constant on either side.
func (g guard) comparesFalse(b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
		if ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "false") && g.same(pair[0]) {
			return true
		}
	}
	return false
}

// same accepts the local read after the call that still resolves to it, or
// the assignment storing the call. Another producer is a different value.
func (g guard) same(e syntax.Expr) bool {
	switch e := syntax.UnwrapParens(e).(type) {
	case *syntax.Assign:
		target, ok := syntax.UnwrapParens(e.Var).(*syntax.Variable)
		return ok && target.Name == g.name && e.Op.Kind == syntax.TEqual && !e.ByRef && syntax.UnwrapParens(e.Value) == syntax.Expr(g.call)
	case *syntax.Variable:
		return e.Name == g.name && e.Span().Start >= g.call.Span().End && semanticquery.NativeValue(g.ctx, e) == syntax.Expr(g.call)
	}
	return false
}
