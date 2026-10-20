package opensslsigningfailureignored

import (
	"math"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// proof says what the signature variable held before the openssl_sign call,
// and so which test proves the call wrote it. A failed call leaves the
// variable untouched.
type proof uint8

const (
	whenUnset proof = iota // unset or null: any test ruling null out
	whenEmpty              // '': any test ruling the empty string out
	always                 // unknown: only openssl_verify() === 1
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

// safe reports whether the mention v is itself part of a test of the
// signature, or is only reached once such a test proved signing succeeded.
func (g guard) safe(v *syntax.Variable) bool {
	if use, ok := v.Parent().(*syntax.ClosureUse); ok {
		// A capture copies the value where the closure is created.
		closure := use.Parent().(*syntax.Closure)
		return g.verifiedBefore(closure) || g.writtenAt(closure)
	}
	return g.tested(v) || g.verifiedBefore(v) || g.writtenAt(v)
}

// tested reports whether v is the operand of a test of the signature: of
// isset, empty, `!`, a comparison with a constant or a truthiness condition
// (parentheses, `&&` and `||` aside), or an argument of openssl_verify. A value computed from the
// signature (`!deliver(base64_encode($signature))`) is a use, not a test.
func (g guard) tested(v *syntax.Variable) bool {
	var child syntax.Node = v
	for p := v.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch p := p.(type) {
		case *syntax.Paren:
			continue
		case *syntax.Binary:
			switch p.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr:
				continue
			case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
				// `$signature === ''`, `null !== $signature`: compared with a
				// constant, the value is not used.
				other := p.Right
				if child == syntax.Node(p.Right) {
					other = p.Left
				}
				switch syntax.UnwrapParens(other).(type) {
				case *syntax.Literal, *syntax.ConstFetch:
					return true
				}
			}
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Unary:
			if p.Op.Kind == syntax.TExclaim {
				return true
			}
		case *syntax.If:
			return child == syntax.Node(p.Cond)
		case *syntax.Ternary:
			return p.Then != nil && child == syntax.Node(p.Cond)
		}
		break
	}
	for p := v.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if c, ok := p.(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(g.ctx, c, "openssl_verify") {
			return true
		}
	}
	return false
}

// verifiedBefore reports whether an openssl_verify(..., $signature, ...)
// call made after signing returned 1 where at runs: the signature is valid
// whatever the variable held.
func (g guard) verifiedBefore(at syntax.Expr) bool {
	return semanticquery.NativeEarlierCallProves(g.ctx, at, func(x syntax.Expr) (types.Type, func(types.Type) bool) {
		c, ok := x.(*syntax.FuncCall)
		if !ok || c.Span().Start < g.stmt.Span().End || !semanticquery.NativeBuiltin(g.ctx, c, "openssl_verify") {
			return types.Unknown, nil
		}
		checked := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 1, "signature"))
		if cast, ok := checked.(*syntax.Unary); ok && cast.Op.Kind == syntax.TStringCast {
			checked = cast.Expr
		}
		if !g.same(checked) {
			return types.Unknown, nil
		}
		return types.Ints(types.IntRange{Lo: -1, Hi: 1}), func(t types.Type) bool { return t.IntsOnly(1, 1) }
	})
}

// writtenAt reports whether the conditions guarding at prove the call wrote
// the signature, given what it held before: not null any more when it was
// unset or null, not empty when it was ”.
func (g guard) writtenAt(at syntax.Expr) bool {
	if g.required == always || semanticquery.NativeLabelled(g.ctx, at) {
		return false
	}
	env := g.ctx.Types().Native()
	if g.required == whenUnset {
		t := env.NarrowedVarAt(g.name, at, types.Union(types.Null, types.StringLengths(types.IntRange{Lo: 1, Hi: math.MaxInt64})))
		return t.IsUnknown() || !t.Has("null")
	}
	// No signature is MaxInt64 bytes long: the bound keeps the lengths known,
	// so that truthiness and emptiness tests refine them.
	t := env.NarrowedVarAt(g.name, at, types.StringLengths(types.IntRange{Lo: 0, Hi: math.MaxInt64 - 1}))
	return t.IsUnknown() || t.LengthsOnly(1, math.MaxInt64)
}
