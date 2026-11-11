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

// refineFuncCall refines what a PHP predicate or check settles about its
// arguments.
func (e *Env) refineFuncCall(n *syntax.FuncCall, f *frame, truth bool) {
	if truth {
		refineIndirectClose(n, f)
	}
	name, builtin := e.callName(n)
	if !builtin {
		return
	}
	args, valid := e.guardArguments(n)
	if !valid || len(args) == 0 || args[0] == nil {
		return
	}
	if truth && (name == "flock" || name == "fclose") {
		first := args[0]
		if variable, ok := first.Value.(*syntax.Variable); ok {
			value := f.vars.value(variable.Name)
			if state, ok := f.states[value.Identity]; ok {
				state.Successful = true
				state.closeChecked = name == "fclose" && state.Operation == "fclose" && state.Span == n.Span()
				f.states[value.Identity] = state
			}
		}

	}
	if truth && (name == "array_key_exists" || name == "key_exists") && len(args) == 2 && args[1] != nil {
		// Only an array (or an object) has keys.
		exclude(args[1].Value, f, true, true)
	}
	if truth && (name == "preg_match" || name == "preg_match_all") && len(args) >= 2 && args[1] != nil {
		// A match of a pattern that needs at least one character: the
		// subject was text that is not empty, so neither false nor null.
		if EmptylessPattern(args[0].Value) {
			exclude(args[1].Value, f, true, true)
		}
	}
	if truth && name == "json_validate" {
		// Valid JSON is never the empty text false and null read as.
		exclude(args[0].Value, f, true, true)
	}
	// The predicate tests a variable or the assignment written in its
	// place (`is_string($n = f())`).
	arg := args[0]
	variable := guarded(arg.Value)
	if variable == nil {
		return
	}
	switch name {
	case "is_string", "is_array", "is_int", "is_integer", "is_long", "is_float", "is_double", "is_numeric",
		"is_object", "is_resource", "is_iterable", "is_countable", "is_callable":
		if truth {
			exclude(variable, f, true, true)
		}
	case "is_bool":
		if !truth {
			exclude(variable, f, true, false)
		}
	case "is_null":
		if !truth {
			exclude(variable, f, false, true)
		}
	}
	if held, tracked := f.vars.lookup(variable.Name); truth && tracked {
		switch name {
		case "is_string", "is_array", "is_int", "is_integer", "is_long", "is_float", "is_double", "is_numeric", "is_resource", "is_bool", "is_null", "is_scalar":
			// None of these is an object, and only an array holds any.
			held.plain, held.deep = true, held.deep || name != "is_array"
			f.setVar(variable.Name, held)
		}
	}
	v := f.vars.value(variable.Name)
	if truth && (name == "ctype_digit" || name == "is_int" || name == "is_integer") {
		// Digits alone, or a number, are safe in every text context.
		v.Safe |= everyContext
		f.setVar(variable.Name, v)
	}
	if truth && name == "in_array" && len(args) == 3 {
		list, strict := args[1], args[2]
		if list == nil || strict == nil {
			return
		}
		flag, ok := strict.Value.(*syntax.ConstFetch)
		if !ok || !strings.EqualFold(flag.Name.Value, "true") {
			return
		}
		array, ok := list.Value.(*syntax.Array)
		if !ok || len(array.Items) == 0 {
			return
		}
		for _, item := range array.Items {
			if item == nil || item.Unpack || item.ByRef {
				return
			}
			if _, ok := literalText(item.Value); !ok {
				return
			}
		}
		v.Sources = nil
		v.Safe = everyContext | URL | Path
		f.setVar(variable.Name, v)
		// The value is one of the literals: neither false nor null.
		exclude(variable, f, true, true)
	}
	if !truth && name == "str_contains" && len(args) == 2 {
		arg := args[1]
		needle, ok := literalText(arg.Value)
		if !ok {
			return
		}
		if needle == "\r" {
			v.headerMask |= 1
		}
		if needle == "\n" {
			v.headerMask |= 2
		}
		if v.headerMask == 3 {
			v.Safe |= Header
		}
		f.setVar(variable.Name, v)
	}
}
