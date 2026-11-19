package flow

import (
	"strconv"
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
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
		v.setOrigins(v.origins().excluding(nonFalse, nonNull, f.owner))
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
	v.setOrigins(v.origins().excluding(nonFalse, nonNull, f.owner))
	f.setVar(name, v)
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
			if node.NameExpr != nil || !value.Complete || value.replaced() || value.Invalidated != 0 {
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
