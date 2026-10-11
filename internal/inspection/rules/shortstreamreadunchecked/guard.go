package shortstreamreadunchecked

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// Facts a condition establishes about the read result on the current path.
const (
	factComplete uint8 = 1 << iota // the byte length equals the requested size
	factNonEmpty                   // the value is not the empty string
	factNonFalse                   // the value is not the failure sentinel
)

// guard carries the read result and the size fread was asked for. fread never
// returns more than that, so a lower bound on the length is an exact length.
type guard struct {
	ctx   *analysis.Context
	input syntax.Expr
	size  int64
}

// lengthChecked reports whether every path reaching input passed a condition
// that rules out a short read of the same reaching value.
func lengthChecked(ctx *analysis.Context, input syntax.Expr, size int64) bool {
	g := guard{ctx: ctx, input: input, size: size}
	facts := uint8(0)
	for p := input.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.If:
			facts |= g.branchFacts(x)
		case *syntax.Ternary:
			if x.Then != nil && x.Then.Span().Contains(input.Span()) {
				facts |= g.facts(x.Cond, true)
			} else if x.Else.Span().Contains(input.Span()) {
				facts |= g.facts(x.Cond, false)
			}
		case *syntax.Binary:
			// The right operand of a short-circuit operator runs only after the left decided.
			if x.Right.Span().Contains(input.Span()) && (isAnd(x) || isOr(x)) {
				facts |= g.facts(x.Left, isAnd(x))
			}
		}
		if st, ok := p.(syntax.Stmt); ok {
			facts |= g.earlierExits(st)
		}
	}
	return facts&factComplete != 0 || size == 1 && facts&factNonEmpty != 0 && facts&factNonFalse != 0
}

// branchFacts collects what the conditions of an if chain say about the branch holding the input.
func (g guard) branchFacts(branch *syntax.If) uint8 {
	at := g.input.Span()
	if branch.Body != nil && branch.Body.Span().Contains(at) {
		return g.facts(branch.Cond, true)
	}
	facts := g.facts(branch.Cond, false)
	for _, alternative := range branch.ElseIfs {
		if alternative.Body != nil && alternative.Body.Span().Contains(at) {
			return facts | g.facts(alternative.Cond, true)
		}
		facts |= g.facts(alternative.Cond, false)
	}
	if branch.Else != nil && branch.Else.Body != nil && branch.Else.Body.Span().Contains(at) {
		return facts
	}
	return 0
}

// earlierExits collects the negated conditions of preceding sibling guards whose body always leaves.
func (g guard) earlierExits(st syntax.Stmt) uint8 {
	list, index, ok := astquery.StmtList(g.ctx.File, st)
	if !ok {
		return 0
	}
	facts := uint8(0)
	for _, previous := range list[:index] {
		if exit, ok := previous.(*syntax.If); ok && exit.Else == nil && len(exit.ElseIfs) == 0 && syntax.Terminates(exit.Body) {
			facts |= g.facts(exit.Cond, false)
		}
	}
	return facts
}

// facts returns what is known once condition evaluated to truth.
func (g guard) facts(condition syntax.Expr, truth bool) uint8 {
	switch x := syntax.UnwrapParens(condition).(type) {
	case *syntax.Unary:
		if x.Op.Kind == syntax.TExclaim {
			return g.facts(x.Expr, !truth)
		}
	case *syntax.Binary:
		// A true conjunction and a false disjunction fix both operands.
		if truth && isAnd(x) || !truth && isOr(x) {
			return g.facts(x.Left, truth) | g.facts(x.Right, truth)
		}
		return g.comparisonFacts(x, truth)
	case *syntax.Isset:
		// A readable last offset means every requested byte is present.
		if !truth || len(x.Vars) != 1 {
			return 0
		}
		if access, ok := syntax.UnwrapParens(x.Vars[0]).(*syntax.ArrayDimFetch); ok && g.same(access.Var) {
			if offset, known := semanticquery.NativeInt(g.ctx, access.Dim); known && offset == g.size-1 {
				return factComplete
			}
		}
	case *syntax.Empty:
		if !truth && g.same(x.Expr) {
			return factNonEmpty | factNonFalse
		}
	default:
		if truth && g.same(condition) {
			return factNonEmpty | factNonFalse
		}
	}
	return 0
}

// comparisonFacts reads a length comparison or a comparison against ” or false.
func (g guard) comparisonFacts(b *syntax.Binary, truth bool) uint8 {
	op := b.Op.Kind
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if g.byteLength(pair[0]) {
			if bound, known := semanticquery.NativeInt(g.ctx, pair[1]); known && g.completes(normalize(op, truth), bound) {
				return factComplete
			}
			return 0
		}
		if g.same(pair[0]) {
			return sentinelFacts(normalize(op, truth), pair[1])
		}
		op = mirror(op)
	}
	return 0
}

// completes decides whether `length op bound` leaves only the full size.
// Lengths are integers, so loose and strict equality agree here.
func (g guard) completes(op syntax.TokenKind, bound int64) bool {
	switch op {
	case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsGreaterOrEqual:
		return bound == g.size
	case syntax.TGreater:
		return bound == g.size-1
	case syntax.TIsNotIdentical, syntax.TIsNotEqual:
		return g.size == 1 && bound == 0
	}
	return false
}

// byteLength recognises strlen, or mb_strlen pinned to the 8bit encoding, of the read result.
func (g guard) byteLength(e syntax.Expr) bool {
	call, ok := syntax.UnwrapParens(e).(*syntax.FuncCall)
	if !ok {
		return false
	}
	switch semanticquery.NativeBuiltinName(g.ctx, call) {
	case "strlen":
		return g.same(semanticquery.CallArgument(call.Args, 0, "string"))
	case "mb_strlen":
		encoding, known := semanticquery.NativeString(g.ctx, semanticquery.CallArgument(call.Args, 1, "encoding"))
		return known && strings.EqualFold(encoding, "8bit") && g.same(semanticquery.CallArgument(call.Args, 0, "string"))
	}
	return false
}

// same ties a guard operand to the reaching read: a repeated call or a
// reassignment is a new value even when its text agrees.
func (g guard) same(e syntax.Expr) bool {
	if e == nil || !astquery.Equivalent(g.ctx.File, syntax.UnwrapParens(e), syntax.UnwrapParens(g.input)) {
		return false
	}
	a, b := g.ctx.Flow().Value(e), g.ctx.Flow().Value(g.input)
	return a.Expr == b.Expr && a.Invalidated == b.Invalidated
}

// sentinelFacts reads `value op other` where other is ” or false. Loose
// comparison treats both sentinels alike; strict comparison excludes one.
func sentinelFacts(op syntax.TokenKind, other syntax.Expr) uint8 {
	other = syntax.UnwrapParens(other)
	text, isString := astquery.QuotedStringValue(other)
	failure, isBool := astquery.BoolConst(other)
	empty, sentinel := isString && text == "", isBool && !failure
	switch {
	case op == syntax.TIsNotEqual && (empty || sentinel):
		return factNonEmpty | factNonFalse
	case op == syntax.TIsNotIdentical && empty:
		return factNonEmpty
	case op == syntax.TIsNotIdentical && sentinel:
		return factNonFalse
	}
	return 0
}

// normalize rewrites a comparison known to be false as the one known to be true.
func normalize(op syntax.TokenKind, truth bool) syntax.TokenKind {
	if truth {
		return op
	}
	switch op {
	case syntax.TIsIdentical:
		return syntax.TIsNotIdentical
	case syntax.TIsNotIdentical:
		return syntax.TIsIdentical
	case syntax.TIsEqual:
		return syntax.TIsNotEqual
	case syntax.TIsNotEqual:
		return syntax.TIsEqual
	case syntax.TLess:
		return syntax.TIsGreaterOrEqual
	case syntax.TIsGreaterOrEqual:
		return syntax.TLess
	case syntax.TGreater:
		return syntax.TIsSmallerOrEqual
	case syntax.TIsSmallerOrEqual:
		return syntax.TGreater
	}
	return op
}

// mirror swaps the operands of an ordering comparison.
func mirror(op syntax.TokenKind) syntax.TokenKind {
	switch op {
	case syntax.TLess:
		return syntax.TGreater
	case syntax.TGreater:
		return syntax.TLess
	case syntax.TIsSmallerOrEqual:
		return syntax.TIsGreaterOrEqual
	case syntax.TIsGreaterOrEqual:
		return syntax.TIsSmallerOrEqual
	}
	return op
}

func isAnd(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanAnd || b.Op.Kind == syntax.TAnd
}

func isOr(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanOr || b.Op.Kind == syntax.TOr
}
