package flow

import "custos/internal/php/syntax"

// evalBinary evaluates a binary operator, logical, null-coalescing or other.
func (e *Env) evalBinary(n *syntax.Binary, f *frame, r *scopeResult, depth int) Value {
	left := e.eval(n.Left, f, r, depth+1)
	var right Value
	switch n.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
		right = e.shortCircuit(n, f, r, depth)
	case syntax.TCoalesce:
		before := clone(f)
		right = e.eval(n.Right, f, r, depth+1)
		*f = rejoin(before, *f, r)
	default:
		right = e.eval(n.Right, f, r, depth+1)
	}
	if converted(n, left, right) {
		e.implicit(n, f, r)
	}
	v := merge(left, right)
	if n.Op.Kind != syntax.TCoalesce {
		v.interval = arithmetic(n.Op.Kind, left, right)
	}
	v.Expr = n
	v.Identity = 0
	// Only a concatenation, or a choice, yields the text it is given.
	v.matched = v.matched && (n.Op.Kind == syntax.TDot || n.Op.Kind == syntax.TCoalesce)
	switch n.Op.Kind {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
		syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual, syntax.TLess, syntax.TGreater,
		syntax.TSpaceship, syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor,
		// These operators yield a number or throw; `+` may also unite arrays.
		syntax.TMinus, syntax.TMul, syntax.TDiv, syntax.TMod, syntax.TPow, syntax.TSl, syntax.TSr:
		v.Safe |= everyContext
	}
	return v
}

// shortCircuit evaluates the right operand of a logical operator, which
// runs only on the left operand's deciding outcome. The path that skips it
// saw the left operand decide the other way.
func (e *Env) shortCircuit(n *syntax.Binary, f *frame, r *scopeResult, depth int) Value {
	deciding := n.Op.Kind == syntax.TBooleanAnd || n.Op.Kind == syntax.TAnd
	skipped := clone(f)
	e.refine(n.Left, &skipped, !deciding)
	for name, refinedCell := range skipped.vars {
		refined := refinedCell.Value
		// Only what the outcome excludes is kept: a value pinned
		// or cleared by the test is another value to later reads.
		if held := f.vars.value(name); refined.Expr != held.Expr || len(refined.Sources) != len(held.Sources) {
			skipped.setVar(name, held)
		}
	}
	e.refine(n.Left, f, deciding)
	if assignedInPlace(n.Right) != nil {
		// A later test of what the left operand read tells which
		// path was taken, while the values it read are in place.
		e.witness(n, &skipped)
		e.witness(n, f)
	}
	right := e.eval(n.Right, f, r, depth+1)
	if stored := assignedInPlace(n.Right); stored != nil {
		// A value the same assignment produced earlier may arrive on
		// the path that skips it: its test then says nothing of it.
		reached := skipped.vars.value(stored.Var.(*syntax.Variable).Name)
		produced := Value{Expr: syntax.UnwrapParens(stored.Value)}
		if reached.Expr == produced.Expr || reached.Expr == nil && reached.origins != nil && (reached.origins.full || reached.origins.holds(&Value{Expr: produced.Expr, NonFalse: true, NonNull: true})) {
			if r.stale == nil {
				r.stale = map[*syntax.Assign]bool{}
			}
			r.stale[stored] = true
		}
		// Where this operand is skipped the variable keeps what it
		// held: a value that is not false or null there, joined with
		// one this operand's test vouches for, is neither (see
		// refineAssigned). A variable nothing assigned is null.
		// Nor is one that passed this test on an earlier turn, where
		// the left operand skips it since (see relation).
		name := stored.Var.(*syntax.Variable).Name
		kept, tracked := skipped.vars.lookup(name)
		held := skipped.relates(n)
		notFalse, notNull := kept.NonFalse || !tracked && skipped.definite || held&relatedFalse != 0, kept.NonNull || held&relatedNull != 0
		ran, left := uint8(0), uint8(0)
		if c, decides := compares(n.Left); decides && (c.right != nil || len(e.types.TypeOf(c.left).Classes()) == 0) {
			// Where the operand ran, the left operand does not skip
			// it while what it read stays in place.
			if !e.rebinds(n.Right, n.Left, f) {
				ran = relatedFalse | relatedNull
			}
			if notFalse {
				left |= relatedFalse
			}
			if notNull {
				left |= relatedNull
			}
		}
		skipped.relate(n, left)
		f.relate(n, ran)
		*f = rejoin(skipped, *f, r)
		if joined := f.vars.value(name); notFalse || notNull {
			joined.unless, joined.unlessFalse, joined.unlessNull = n, false, false
			joined.skippedFalse, joined.skippedNull = notFalse, notNull
			f.setVar(name, joined)
		}
		return right
	}
	*f = rejoin(skipped, *f, r)
	return right
}

// evalUnary evaluates a unary operator or a cast.
func (e *Env) evalUnary(n *syntax.Unary, f *frame, r *scopeResult, depth int) Value {
	v := e.eval(n.Expr, f, r, depth+1)
	v.interval = unaryInts(n.Op.Kind, v)
	if n.Op.Kind == syntax.TStringCast && !v.plain {
		e.implicit(n, f, r)
	}
	v.Expr = n
	v.Identity = 0
	if n.Op.Kind != syntax.TArrayCast {
		v.upload, v.scalars = 0, false
	}
	switch n.Op.Kind {
	case syntax.TExclaim, syntax.TIntCast, syntax.TDoubleCast, syntax.TBoolCast, syntax.TMinus, syntax.TPlus:
		v.Safe |= everyContext
	}
	return v
}
