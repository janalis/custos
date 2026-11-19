package flow

import "github.com/janalis/custos/internal/php/syntax"

// comparison is a test without effects that decides the same way again
// while the variables it reads are not written: the truth of a variable, or
// a strict comparison of two variables, constants or literals, under any
// number of negations. A loose comparison is not one: it reads the
// properties of the objects it compares.
type comparison struct {
	left, right syntax.Expr
	negated     bool
}

// compares reads a test of that form.
func compares(x syntax.Expr) (comparison, bool) {
	var c comparison
	for {
		x = syntax.UnwrapParens(x)
		not, ok := x.(*syntax.Unary)
		if !ok || not.Op.Kind != syntax.TExclaim {
			break
		}
		c.negated, x = !c.negated, not.Expr
	}
	switch n := x.(type) {
	case *syntax.Variable:
		c.left = n
		return c, n.NameExpr == nil
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TIsIdentical:
		case syntax.TIsNotIdentical:
			c.negated = !c.negated
		default:
			return c, false
		}
		c.left, c.right = syntax.UnwrapParens(n.Left), syntax.UnwrapParens(n.Right)
		return c, plainVariable(c.left) != "" && constantOperand(c.right) || constantOperand(c.left) && plainVariable(c.right) != "" || plainVariable(c.left) != "" && plainVariable(c.right) != ""
	}
	return c, false
}

// sameOperand reports two operands of a comparison that read the same thing.
func sameOperand(a, b syntax.Expr) bool {
	switch a := a.(type) {
	case nil:
		return b == nil
	case *syntax.Variable:
		return plainVariable(b) == a.Name
	case *syntax.Literal:
		other, ok := b.(*syntax.Literal)
		return ok && other.LitKind == a.LitKind && other.Raw == a.Raw
	}
	// Only true, false and null are spelled in any case.
	constant, _ := a.(*syntax.ConstFetch)
	other, ok := b.(*syntax.ConstFetch)
	if !ok || constant.Name == nil || other.Name == nil {
		return false
	}
	// A name written with a leading backslash is the global constant, which
	// the same name without it need not be inside a namespace.
	return constant.Name.Value == other.Name.Value || constantName(a) != "" && constantName(a) == constantName(b)
}

// untest forgets that a short-circuit operation tested the variable an
// access is rooted in: a write to a property of it, a method called on it or
// a reference taken inside it can change what it compares or converts to
// (an object's content, a value PHP turns into an object or an array). With
// identity, the access leaves the variable the value it is: a method call,
// or a property write where writing one to a value that is no object is an
// error. What a strict comparison decided of it then stands.
func untest(x syntax.Expr, f *frame, identity bool) {
	for {
		switch n := syntax.UnwrapParens(x).(type) {
		case *syntax.PropertyFetch:
			x = n.Var
			continue
		case *syntax.ArrayDimFetch:
			x = n.Var
			continue
		case *syntax.MethodCall:
			x = n.Var
			continue
		case *syntax.Variable:
			strict := func(by *syntax.Binary) bool {
				c, _ := compares(by.Left)
				return identity && c.right != nil
			}
			if v, tracked := f.vars.lookup(n.Name); tracked && v.tested() != nil && !strict(v.tested()) {
				v.setTested(nil)
				f.setVar(n.Name, v)
			}
			// What a left operand decided from it may no longer hold.
			for _, known := range f.related {
				if c, _ := compares(known.by.Left); (plainVariable(c.left) == n.Name || plainVariable(c.right) == n.Name) && !strict(known.by) {
					f.relate(known.by, known.bits&f.settled(known.by))
				}
			}
		}
		return
	}
}

// witness marks the variables the left operand of a short-circuit operation
// reads as the ones it tested. Only an operand that is a comparison is
// followed: with the same values it decides the same way again.
func (e *Env) witness(n *syntax.Binary, f *frame) {
	c, ok := compares(n.Left)
	if !ok {
		return
	}
	if c.right == nil && len(e.types.TypeOf(c.left).Classes()) > 0 {
		// The truth of an object can follow its content.
		return
	}
	for _, operand := range [2]syntax.Expr{c.left, c.right} {
		if v, tracked := f.vars.lookup(plainVariable(operand)); tracked {
			v.setTested(n)
			f.setVar(plainVariable(operand), v)
		}
	}
}

// ran applies what a test with a known outcome says of a short-circuit
// operation evaluated before it: when the test is the left operand of that
// operation, or its negation, read from the values that operand read, its
// outcome tells whether the right operand ran. Where it did, what the right
// operand established of the value it assigns in place holds.
func ran(cond syntax.Expr, f *frame, truth bool) {
	c, ok := compares(cond)
	if !ok {
		return
	}
	by := f.vars.value(plainVariable(c.left)).tested()
	if by == nil {
		by = f.vars.value(plainVariable(c.right)).tested()
	}
	if by == nil {
		return
	}
	left, _ := compares(by.Left)
	if !sameOperand(left.left, c.left) || !sameOperand(left.right, c.right) {
		return
	}
	for _, operand := range [2]syntax.Expr{left.left, left.right} {
		if name := plainVariable(operand); name != "" && f.vars.value(name).tested() != by {
			return
		}
	}
	// The right operand of `&&` runs when the left one is true, of `||`
	// when it is false.
	and := by.Op.Kind == syntax.TBooleanAnd || by.Op.Kind == syntax.TAnd
	if (truth == (left.negated == c.negated)) != and {
		return
	}
	for name, vCell := range f.vars {
		v := vCell.Value
		if v.unless() == by {
			nonFalse, nonNull := v.unlessFalse(), v.unlessNull()
			v.setUnlessFacts(nil, false, false, false, false)
			f.excluding(name, v, nonFalse, nonNull)
		}
	}
}
