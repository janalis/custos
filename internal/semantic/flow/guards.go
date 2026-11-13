package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
)

func literalText(x syntax.Expr) (string, bool) {
	n, ok := syntax.UnwrapParens(x).(*syntax.Literal)
	if !ok {
		return "", false
	}
	if n.LitKind != syntax.LitString {
		return n.Raw, true
	}
	return syntax.LiteralStringValue(n)
}

// looseFalsy reports what a value loosely different from literal x is not:
// false for ”, '0', 0 and 0.0 (each loosely equals false), and null too for
// all of them but '0'.
func looseFalsy(x syntax.Expr) (nonFalse, nonNull bool) {
	n, ok := syntax.UnwrapParens(x).(*syntax.Literal)
	if !ok {
		return false, false
	}
	if n.LitKind == syntax.LitString {
		v, ok := syntax.LiteralStringValue(n)
		return ok && (v == "" || v == "0"), ok && v == ""
	}
	f, err := strconv.ParseFloat(strings.ReplaceAll(n.Raw, "_", ""), 64) // int or float
	return err == nil && f == 0, err == nil && f == 0
}

// refine narrows the frame by what holds where a guard evaluated to truth:
// short forms stay inline, the others go to the handler of their form.
func (e *Env) refine(expr syntax.Expr, f *frame, truth bool) {
	if expr == nil {
		return
	}
	ran(expr, f, truth)
	switch n := syntax.UnwrapParens(expr).(type) {
	case *syntax.Variable, *syntax.Assign:
		e.refineVariable(n, f, truth)
	case *syntax.Empty:
		if !truth {
			exclude(n.Expr, f, true, true)
			exclude(indexed(n.Expr), f, true, true)
		}
	case *syntax.Isset:
		if truth {
			for _, x := range n.Vars {
				exclude(x, f, false, true)
				exclude(indexed(x), f, true, true)
			}
		}
	case *syntax.Instanceof:
		if truth {
			exclude(n.Expr, f, true, true)
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TExclaim {
			e.refine(n.Expr, f, !truth)
		}
	case *syntax.Binary:
		e.refineBinary(n, f, truth)
	case *syntax.StaticCall:
		if truth {
			refineIndirectClose(n, f)
		}
	case *syntax.MethodCall:
		e.refineMethodCall(n, f, truth)
	case *syntax.FuncCall:
		e.refineFuncCall(n, f, truth)
	}
}

// boolean reports an operand that yields true or false: a comparison, a
// negation, a boolean combination, a test of existence or of class, a PHP
// function declared to return a boolean, or a bitwise combination of those.
func (e *Env) boolean(x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Isset, *syntax.Empty, *syntax.Instanceof:
		return true
	case *syntax.Unary:
		return n.Op.Kind == syntax.TExclaim || n.Op.Kind == syntax.TBoolCast
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual,
			syntax.TLess, syntax.TGreater, syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor:
			return true
		case syntax.TBar, syntax.TAmpersand:
			return e.boolean(n.Left) && e.boolean(n.Right)
		}
	case *syntax.FuncCall:
		_, builtin, result, _ := e.callee(n)
		return builtin && result == "bool"
	}
	return false
}

// guarded names the variable a condition operand tests: the variable itself or
// the target of a plain assignment written inside the condition.
func guarded(x syntax.Expr) *syntax.Variable {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		if n.NameExpr == nil {
			return n
		}
	case *syntax.Assign:
		if target, ok := n.Var.(*syntax.Variable); ok && target.NameExpr == nil && !n.ByRef && n.Op.Kind == syntax.TEqual {
			return target
		}
	}
	return nil
}

// exclude records that the tested variable is not false and/or not null.
func exclude(x syntax.Expr, f *frame, nonFalse, nonNull bool) {
	variable := guarded(x)
	if variable == nil || !(nonFalse || nonNull) {
		return
	}
	// A variable the frame does not track has no value to qualify.
	if v, tracked := f.vars.lookup(variable.Name); tracked {
		f.excluding(variable.Name, v, nonFalse, nonNull)
	}
}

// excludeValue is exclude for a value a call accepted: the call says what
// the variable holds, not whether an operation made on it before succeeded
// (its state).
func excludeValue(x syntax.Expr, f *frame, nonFalse, nonNull bool) {
	variable := guarded(x)
	if variable == nil {
		return
	}
	if v, tracked := f.vars.lookup(variable.Name); tracked {
		v.NonFalse = v.NonFalse || nonFalse
		v.NonNull = v.NonNull || nonNull
		v.origins = v.origins.excluding(nonFalse, nonNull, f.owner)
		f.setVar(variable.Name, v)
	}
}

// excluding records that the value a name holds is not false and/or not null.
func (f *frame) excluding(name string, v Value, nonFalse, nonNull bool) {
	if nonFalse {
		v.NonFalse = true
		if s, exists := f.states[v.Identity]; exists {
			s.Successful = true
			f.states[v.Identity] = s
		}
	}
	if nonNull {
		v.NonNull = true
	}
	v.origins = v.origins.excluding(nonFalse, nonNull, f.owner)
	f.setVar(name, v)
}

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
			if v, tracked := f.vars.lookup(n.Name); tracked && v.tested != nil && !strict(v.tested) {
				v.tested = nil
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
			v.tested = n
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
	by := f.vars.value(plainVariable(c.left)).tested
	if by == nil {
		by = f.vars.value(plainVariable(c.right)).tested
	}
	if by == nil {
		return
	}
	left, _ := compares(by.Left)
	if !sameOperand(left.left, c.left) || !sameOperand(left.right, c.right) {
		return
	}
	for _, operand := range [2]syntax.Expr{left.left, left.right} {
		if name := plainVariable(operand); name != "" && f.vars.value(name).tested != by {
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
		if v.unless == by {
			nonFalse, nonNull := v.unlessFalse, v.unlessNull
			v.unless, v.unlessFalse, v.unlessNull, v.skippedFalse, v.skippedNull = nil, false, false, false, false
			f.excluding(name, v, nonFalse, nonNull)
		}
	}
}

// indexed returns the array variable an element access reads, at any depth:
// an element that is set, or a key that exists, proves it holds an array.
func indexed(x syntax.Expr) syntax.Expr {
	fetch, ok := syntax.UnwrapParens(x).(*syntax.ArrayDimFetch)
	if !ok {
		return nil
	}
	for {
		inner, nested := syntax.UnwrapParens(fetch.Var).(*syntax.ArrayDimFetch)
		if !nested {
			return fetch.Var
		}
		fetch = inner
	}
}

func constantOperand(x syntax.Expr) bool {
	switch syntax.UnwrapParens(x).(type) {
	case *syntax.ConstFetch, *syntax.Literal:
		return true
	}
	return false
}

// constantName returns the folded name of false, null or true.
func constantName(x syntax.Expr) string {
	if c, ok := syntax.UnwrapParens(x).(*syntax.ConstFetch); ok && c.Name != nil {
		switch name := strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`)); name {
		case "false", "null", "true":
			return name
		}
	}
	return ""
}

// A saved native result can prove success only while its exact pending
// operation is still current; later operations and escapes remove that state.
func refineIndirectClose(expr syntax.Expr, f *frame) {
	for range MaxDepth {
		switch node := syntax.UnwrapParens(expr).(type) {
		case *syntax.Variable:
			value := f.vars.value(node.Name)
			if node.NameExpr != nil || !value.Complete || value.replaced || value.Invalidated != 0 {
				return
			}
			expr = value.Expr
		case *syntax.Assign:
			if node.Op.Kind != syntax.TEqual || node.ByRef {
				return
			}
			expr = node.Value
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
			for id, state := range f.states {
				if state.Span == node.Span() && state.Operation == "@indirect:fclose" {
					state.Operation = "fclose"
					state.Successful, state.closeChecked = true, true
					f.setState(id, state)
				}
			}
			return
		default:
			return
		}
	}
}
