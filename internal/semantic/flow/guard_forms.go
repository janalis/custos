package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
)

// refineVariable refines a variable tested for truth, or an assignment
// tested by the value it stores.
func (e *Env) refineVariable(n syntax.Expr, f *frame, truth bool) {
	if truth {
		if assignment, ok := n.(*syntax.Assign); ok && assignment.Op.Kind == syntax.TEqual && !assignment.ByRef {
			refineIndirectClose(syntax.UnwrapParens(assignment.Value), f)
		}
		if variable, ok := n.(*syntax.Variable); ok {
			value := f.vars.value(variable.Name)
			if value.Complete && !value.replaced && value.Invalidated == 0 {
				refineIndirectClose(value.Expr, f)
			}
		}
		exclude(n, f, true, true)
	}
	if flag, ok := n.(*syntax.Variable); ok && flag.NameExpr == nil {
		// A boolean that stores a test stands for it while the values
		// the test read are still the ones in place.
		if test := e.stored(flag, f); test != nil {
			e.refine(test, f, truth)
		}
	}
}

// refineBinary refines the operands of a comparison, a coalescing default,
// or a logical or bitwise combination of tests.
func (e *Env) refineBinary(n *syntax.Binary, f *frame, truth bool) {
	if truth && n.Op.Kind == syntax.TCoalesce && falsyConstant(n.Right) {
		// `$v['k'] ?? null` holds only where $v['k'] is set and holds,
		// as isset() and a truth test of it would.
		exclude(n.Left, f, true, true)
		exclude(indexed(n.Left), f, true, true)
	}
	// A constant-first comparison states the same fact. It only pins the
	// variable to a literal in the written order, as it always did.
	e.refineInts(n, f, truth)
	if text := e.nonEmptyText(n, truth); text != nil {
		// null, whose strlen() PHP deprecates, stays a possible value.
		exclude(text, f, true, false)
	}
	written := n
	mirrored := constantOperand(n.Left) && !constantOperand(n.Right)
	if mirrored {
		copy := *n
		copy.Left, copy.Right = n.Right, n.Left
		n = &copy
	}
	if (truth && (n.Op.Kind == syntax.TBooleanAnd || n.Op.Kind == syntax.TAnd)) || (!truth && (n.Op.Kind == syntax.TBooleanOr || n.Op.Kind == syntax.TOr)) {
		// The frame already holds what the right operand wrote: a fact
		// about the left operand describes a value that may be gone.
		if !e.rebinds(n.Right, n.Left, f) {
			e.refine(n.Left, f, truth)
		}
		e.refine(n.Right, f, truth)
	}
	if (!truth && (n.Op.Kind == syntax.TBooleanAnd || n.Op.Kind == syntax.TAnd)) || (truth && (n.Op.Kind == syntax.TBooleanOr || n.Op.Kind == syntax.TOr)) {
		// Either operand decided. The right one ran only once the left
		// one had not: what it tests of a value it assigns in place is
		// settled for that value, whichever operand decided.
		e.refineAssigned(written, f, truth)
	}
	if (!truth && n.Op.Kind == syntax.TBar) || (truth && n.Op.Kind == syntax.TAmpersand) {
		// A bitwise operation evaluates both operands and combines them
		// as numbers: `|` is zero only when each is, `&` is not zero
		// only when neither is. That settles an operand that is a test.
		if e.boolean(n.Left) && !e.rebinds(n.Right, n.Left, f) {
			e.refine(n.Left, f, truth)
		}
		if e.boolean(n.Right) {
			e.refine(n.Right, f, truth)
		}
	}
	flag := constantName(n.Right)
	if flag == "true" || flag == "false" {
		equal := n.Op.Kind == syntax.TIsEqual || n.Op.Kind == syntax.TIsIdentical
		unequal := n.Op.Kind == syntax.TIsNotEqual || n.Op.Kind == syntax.TIsNotIdentical
		if (equal || unequal) && (truth == equal) == (flag == "true") {
			refineIndirectClose(n.Left, f)
		}
	}
	if (truth && n.Op.Kind == syntax.TIsIdentical) || (!truth && n.Op.Kind == syntax.TIsNotIdentical) {
		if variable, ok := syntax.UnwrapParens(n.Left).(*syntax.Variable); ok && variable.NameExpr == nil && !mirrored {
			if _, literal := literalText(n.Right); literal {
				f.setVar(variable.Name, Value{Complete: true, Expr: n.Right, plain: true, deep: true})
			}
		}
	}
	if (!truth && n.Op.Kind == syntax.TIsIdentical) || (truth && n.Op.Kind == syntax.TIsNotIdentical) {
		exclude(n.Left, f, flag == "false", flag == "null")
	}
	// A value loosely different from false or null is neither of them;
	// so is one loosely different from '', 0 or 0.0, which both equal,
	// and one loosely equal to true. Loosely different from '0', it is
	// not false (`null == '0'` does not hold).
	if (!truth && n.Op.Kind == syntax.TIsEqual) || (truth && n.Op.Kind == syntax.TIsNotEqual) {
		if flag == "false" || flag == "null" {
			exclude(n.Left, f, true, true)
		} else if nonFalse, nonNull := looseFalsy(n.Right); nonFalse {
			exclude(n.Left, f, true, nonNull)
		}
	}
	if ((truth && n.Op.Kind == syntax.TIsEqual) || (!truth && n.Op.Kind == syntax.TIsNotEqual)) && flag == "true" {
		exclude(n.Left, f, true, true)
	}
	if (truth && n.Op.Kind == syntax.TIsIdentical) || (!truth && n.Op.Kind == syntax.TIsNotIdentical) {
		if call, ok := n.Left.(*syntax.FuncCall); ok {
			name, builtin := e.callName(call)
			zero, known := literalText(n.Right)
			if args, valid := e.guardArguments(call); valid && builtin && name == "preg_match" && known && zero == "0" && len(args) >= 2 {
				pattern, subject := args[0], args[1]
				text, known := literalText(pattern.Value)
				variable, vok := subject.Value.(*syntax.Variable)
				if known && vok && (text == `/[\r\n]/` || text == `/[\n\r]/`) {
					v := f.vars.value(variable.Name)
					v.Safe |= Header
					v.headerMask = 3
					f.setVar(variable.Name, v)
				}

			}
		}
		if call, ok := n.Left.(*syntax.FuncCall); ok && e.types.ResolveFunction(call) != nil {
			name, builtin := e.callName(call)
			if args, valid := e.guardArguments(call); valid && builtin && name == "strlen" && len(args) == 1 {
				arg := args[0]
				variable, ok := arg.Value.(*syntax.Variable)
				length, known := literalText(n.Right)
				if ok && known {
					if number, err := strconv.ParseInt(length, 10, 64); err == nil {
						v := f.vars.value(variable.Name)
						v.Length = number
						v.LengthKnown = true
						f.setVar(variable.Name, v)
					}
				}

			}
		}
	}
}

// refineMethodCall refines what a method call that held settles about its
// receiver.
func (e *Env) refineMethodCall(n *syntax.MethodCall, f *frame, truth bool) {
	if truth {
		refineIndirectClose(n, f)
	}
	name, builtin := e.callName(n)
	if truth && builtin && name == "begintransaction" {
		if variable, ok := n.Var.(*syntax.Variable); ok {
			value := f.vars.value(variable.Name)
			if state, ok := f.states[value.Identity]; ok {
				state.Successful = true
				f.states[value.Identity] = state
			}
		}
	}
}

// stored returns the test a flag variable holds (`$ok = $u !== false`) when
// its outcome still describes the frame: every variable the test read has the
// producer it had then. A test that assigns, or reads a value of unknown
// origin, is not followed, nor is a flag that merely copies another one.
func (e *Env) stored(flag *syntax.Variable, f *frame) syntax.Expr {
	test := syntax.UnwrapParens(f.vars.value(flag.Name).Expr)
	switch test.(type) {
	case *syntax.Binary, *syntax.Unary, *syntax.FuncCall, *syntax.Isset, *syntax.Empty, *syntax.Instanceof:
	default:
		return nil
	}
	r := e.scopes[syntax.EnclosingVariableScope(flag)]
	current := true
	syntax.Inspect(test, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Assign, *syntax.IncDec, *syntax.Closure, *syntax.ArrowFunction:
			current = false
		case *syntax.Variable:
			then, now := r.values.at(x), f.vars.value(x.Name)
			current = current && x.NameExpr == nil && now.Expr != nil && then.Expr == now.Expr && then.Invalidated == now.Invalidated
		}
		return current
	})
	if !current {
		return nil
	}
	return test
}

// falsyConstant reports a constant PHP reads as false: null, false, 0,
// 0.0, ” or '0' written as such.
func falsyConstant(x syntax.Expr) bool {
	switch c := syntax.UnwrapParens(x).(type) {
	case *syntax.ConstFetch:
		name := strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`))
		return name == "null" || name == "false"
	case *syntax.Literal:
		switch c.Raw {
		case "0", "0.0", "''", `""`, "'0'", `"0"`:
			return true
		}
	}
	return false
}

// nonEmptyText returns the argument of strlen() or mb_strlen() when
// comparison n, holding as truth says, bounds its length to at least one
// (`strlen($s) > 0`, `48 > strlen($s)` that does not hold): false reads as
// the empty text there, of length zero.
func (e *Env) nonEmptyText(n *syntax.Binary, truth bool) syntax.Expr {
	call, ok := syntax.UnwrapParens(n.Left).(*syntax.FuncCall)
	bound := n.Right
	op := n.Op.Kind
	if !ok {
		call, ok = syntax.UnwrapParens(n.Right).(*syntax.FuncCall)
		bound, op = n.Left, syntax.MirroredComparison[op]
	}
	if !truth {
		op = map[syntax.TokenKind]syntax.TokenKind{
			syntax.TGreater: syntax.TIsSmallerOrEqual, syntax.TIsSmallerOrEqual: syntax.TGreater,
			syntax.TLess: syntax.TIsGreaterOrEqual, syntax.TIsGreaterOrEqual: syntax.TLess,
		}[op]
	}
	text, known := literalText(bound)
	if !ok || !known {
		return nil
	}
	least, err := strconv.ParseInt(text, 10, 64)
	if err != nil || !(op == syntax.TGreater && least >= 0 || op == syntax.TIsGreaterOrEqual && least >= 1) {
		return nil
	}
	name, builtin := e.callName(call)
	args, valid := e.guardArguments(call)
	if !builtin || !valid || len(args) == 0 || name != "strlen" && name != "mb_strlen" {
		return nil
	}
	return args[0].Value
}
