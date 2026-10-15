package opensslsigningfailureignored

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// proof ranks what a test of the signature variable establishes about the
// preceding openssl_sign call. A failed call leaves the variable untouched, so
// an emptiness test only proves success when the earlier content is known.
type proof uint8

const (
	unproven  proof = iota
	whenUnset       // isset(): valid when the variable was unset or null before signing
	whenEmpty       // empty()/truthiness: valid when it was unset, null or ''
	always          // openssl_verify() === 1: valid whatever the variable held
)

type guard struct {
	ctx      *analysis.Context
	name     string
	scope    syntax.Node
	stmt     *syntax.ExprStmt
	required proof
}

// signatureChecked reports whether every later mention of the signature
// variable of a status-discarding openssl_sign call is reached only after the
// signature itself proved that signing succeeded.
func signatureChecked(ctx *analysis.Context, call *syntax.FuncCall) bool {
	if !semanticquery.NativeBuiltin(ctx, call, "openssl_sign") {
		return false
	}
	stmt, ok := call.Parent().(*syntax.ExprStmt)
	signature, isVar := semanticquery.CallArgument(call.Args, 1, "signature").(*syntax.Variable)
	if !ok || !isVar || signature.Name == "" {
		return false
	}
	g := guard{ctx: ctx, name: signature.Name, scope: syntax.EnclosingVariableScope(call), stmt: stmt}
	earlier := false
	var later []*syntax.Variable
	record := func(v *syntax.Variable) {
		switch {
		case v == signature || (v.Name != g.name && v.Name != ""):
		case v.Span().End <= stmt.Span().End:
			earlier = true
		default:
			later = append(later, v)
		}
	}
	visit := func(n syntax.Node) bool {
		if n != g.scope && syntax.IsVariableScope(n) {
			// A nested closure shares the variable only through its use list.
			if closure, ok := n.(*syntax.Closure); ok {
				for _, use := range closure.Uses {
					record(use.Var)
				}
			}
			return false
		}
		switch n := n.(type) {
		case *syntax.Variable:
			record(n)
		case *syntax.Label, *syntax.Include, *syntax.Eval:
			// Backward jumps and included code may fill the variable first.
			earlier = true
		}
		return true
	}
	if g.scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(g.scope, visit)
	}
	g.required = g.initialState(earlier)
	for _, v := range later {
		if !g.safe(v) {
			return false
		}
	}
	return true
}

// initialState returns the weakest proof that is sound for the content the
// signature variable can hold immediately before the call.
func (g guard) initialState(earlier bool) proof {
	if prev, ok := astquery.PrevStmt(g.ctx.File, g.stmt); ok {
		if es, ok := prev.(*syntax.ExprStmt); ok {
			if a, ok := es.Expr.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef && g.same(a.Var) {
				if c, ok := a.Value.(*syntax.ConstFetch); ok && strings.EqualFold(c.Name.Value, "null") {
					return whenUnset
				}
				if s, ok := semanticquery.NativeString(g.ctx, a.Value); ok && s == "" {
					return whenEmpty
				}
			}
		}
	}
	if earlier || g.scope == nil {
		return always
	}
	for p := g.stmt.Parent(); p != g.scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			// A previous iteration may have left a signature behind.
			return always
		}
	}
	return whenUnset
}

func (g guard) same(e syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return ok && v.Name == g.name
}

// implies returns how strongly cond evaluating to truth proves that the
// signature was written.
func (g guard) implies(cond syntax.Expr, truth bool) proof {
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Variable:
		if truth && c.Name == g.name {
			return whenEmpty
		}
	case *syntax.Isset:
		for _, v := range c.Vars {
			if truth && g.same(v) {
				return whenUnset
			}
		}
	case *syntax.Empty:
		if !truth && g.same(c.Expr) {
			return whenEmpty
		}
	case *syntax.Unary:
		if c.Op.Kind == syntax.TExclaim {
			return g.implies(c.Expr, !truth)
		}
	case *syntax.Binary:
		switch {
		case (c.Op.Kind == syntax.TBooleanAnd && truth) || (c.Op.Kind == syntax.TBooleanOr && !truth):
			return max(g.implies(c.Left, truth), g.implies(c.Right, truth))
		case (c.Op.Kind == syntax.TIsIdentical && truth) || (c.Op.Kind == syntax.TIsNotIdentical && !truth):
			if g.verified(c.Left, c.Right) || g.verified(c.Right, c.Left) {
				return always
			}
		}
	}
	return unproven
}

// verified matches openssl_verify(..., $signature, ...) compared with 1.
func (g guard) verified(call, one syntax.Expr) bool {
	c, ok := syntax.UnwrapParens(call).(*syntax.FuncCall)
	if n, isInt := semanticquery.NativeInt(g.ctx, one); !ok || !isInt || n != 1 || !semanticquery.NativeBuiltin(g.ctx, c, "openssl_verify") {
		return false
	}
	checked := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 1, "signature"))
	if cast, ok := checked.(*syntax.Unary); ok && cast.Op.Kind == syntax.TStringCast {
		checked = cast.Expr
	}
	return g.same(checked)
}

// safe reports whether the mention v is itself part of a test of the
// signature, or is only reached once such a test proved signing succeeded.
func (g guard) safe(v *syntax.Variable) bool {
	// boolean stays true while v is only an operand of a truthiness test:
	// `!send($signature)` or `empty(send($signature))` uses the value, it
	// does not test the variable.
	boolean := true
	var child syntax.Node = v
	for p := v.Parent(); ; child, p = p, p.Parent() {
		if p.Span().Start < g.stmt.Span().End {
			// Constructs enclosing the call were evaluated before it ran.
			// The body or File root starts before the call, ending every ascent.
			return false
		}
		got := unproven
		switch p := p.(type) {
		case *syntax.Isset, *syntax.Empty:
			if boolean {
				return true
			}
		case *syntax.Paren:
			continue
		case *syntax.Unary:
			if boolean && p.Op.Kind == syntax.TExclaim {
				return true
			}
		case *syntax.FuncCall:
			if semanticquery.NativeBuiltin(g.ctx, p, "openssl_verify") {
				return true
			}
		case *syntax.Binary:
			if p.Op.Kind == syntax.TBooleanAnd || p.Op.Kind == syntax.TBooleanOr {
				if p.Op.Kind == syntax.TBooleanAnd && child == syntax.Node(p.Right) {
					got = g.implies(p.Left, true)
				}
				if got < g.required {
					continue
				}
			}
		case *syntax.Ternary:
			switch {
			case p.Then != nil && child == syntax.Node(p.Then):
				got = g.implies(p.Cond, true)
			case child == syntax.Node(p.Else):
				got = g.implies(p.Cond, false)
			case boolean && p.Then != nil:
				return true
			}
		case *syntax.If:
			switch {
			case child == syntax.Node(p.Body):
				got = g.implies(p.Cond, true)
			case boolean && child == syntax.Node(p.Cond):
				return true
			}
		case *syntax.Else:
			if branch := p.Parent().(*syntax.If); len(branch.ElseIfs) == 0 {
				got = g.implies(branch.Cond, false)
			}
		}
		if got >= g.required {
			return true
		}
		boolean = false
		if st, ok := p.(syntax.Stmt); ok && g.rejectedBefore(st) {
			return true
		}
	}
}

// rejectedBefore reports whether a statement between the signing call and st,
// in the statement list of st, leaves the path unless signing succeeded.
func (g guard) rejectedBefore(st syntax.Stmt) bool {
	for {
		prev, ok := astquery.PrevStmt(g.ctx.File, st)
		if !ok || prev.Span().Start < g.stmt.Span().End {
			return false
		}
		if branch, ok := prev.(*syntax.If); ok && branch.Else == nil && len(branch.ElseIfs) == 0 && syntax.Terminates(branch.Body) && g.implies(branch.Cond, false) >= g.required {
			return true
		}
		st = prev
	}
}
