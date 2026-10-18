package infer

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// BuiltinAssert identifies a direct named invocation of PHP's assertion
// construct. Dynamic callables evaluate their arguments as ordinary calls.
func (e *Env) BuiltinAssert(call *syntax.FuncCall) bool {
	name, named := call.Name.(*syntax.Name)
	if !named {
		return false
	}
	fqn, fallback := e.Names.Function(name.Value, name.Span().Start)
	if !strings.EqualFold(fqn, "assert") && !strings.EqualFold(fallback, "assert") {
		return false
	}
	f := e.ResolveFunction(call)
	if f == nil || !f.Builtin || !strings.EqualFold(f.FQN, "assert") {
		return false
	}
	if call.Args != nil && len(call.Args.Args) == 1 {
		if _, closure := call.Args.Args[0].(*syntax.VariadicPlaceholder); closure {
			return false
		}
	}
	return true
}

// AssertGuard returns a bound assertion condition only when the resolved
// environment assumes assertions execute and failures terminate.
func (e *Env) AssertGuard(call *syntax.FuncCall) syntax.Expr {
	if !e.options.Assertions || e.LegacyAssertCode(call) {
		return nil
	}
	condition := e.AssertCondition(call)
	if condition == nil {
		return nil
	}
	// A direct operation result may establish its own success. Its arguments
	// still pass the same effect-free traversal; composite predicates do not
	// gain permission to execute arbitrary calls between their tests.
	operation := false
	if root, ok := syntax.UnwrapParens(condition).(*syntax.FuncCall); ok {
		f := e.ResolveFunction(root)
		operation = f != nil && f.Builtin && strings.EqualFold(f.FQN, "flock")
	}
	safe := true
	syntax.Inspect(condition, func(n syntax.Node) bool {
		if !safe {
			return false
		}
		switch n := n.(type) {
		case *syntax.Literal, *syntax.Paren, *syntax.Name, *syntax.ArgList, *syntax.Instanceof, *syntax.Isset, *syntax.Empty, *syntax.Identifier:
		case *syntax.Variable:
			safe = n.NameExpr == nil
		case *syntax.ConstFetch:
			safe = strings.EqualFold(n.Name.Value, "true") || strings.EqualFold(n.Name.Value, "false") || strings.EqualFold(n.Name.Value, "null")
			if operation {
				switch strings.ToUpper(strings.TrimPrefix(n.Name.Value, `\`)) {
				case "LOCK_EX", "LOCK_SH", "LOCK_NB", "LOCK_UN":
					safe = true
				}
			}
		case *syntax.Arg:
			safe = !n.Unpack && !n.ByRef
		case *syntax.Unary:
			safe = n.Op.Kind == syntax.TExclaim
		case *syntax.Binary:
			switch n.Op.Kind {
			case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor:
			case syntax.TBar:
				safe = operation
			default:
				safe = false
			}
		case *syntax.FuncCall:
			f := e.ResolveFunction(n)
			safe = f != nil && f.Builtin
			if safe {
				switch strings.ToLower(f.FQN) {
				case "flock":
					safe = operation && syntax.UnwrapParens(condition) == n
				case "is_string", "is_int", "is_integer", "is_long", "is_float", "is_double", "is_real", "is_bool", "is_array", "is_object", "is_resource", "is_null", "is_scalar", "is_numeric", "is_countable", "is_iterable":
				default:
					safe = false
				}
			}
		default:
			safe = false
		}

		return safe
	})
	if !safe {
		return nil
	}
	return condition
}

// AssertCondition binds the first native parameter without accepting spread,
// duplicate, unknown or position-after-name arguments as a guard.
func (e *Env) AssertCondition(call *syntax.FuncCall) syntax.Expr {
	if !e.BuiltinAssert(call) || call.Args == nil {
		return nil
	}
	f := e.ResolveFunction(call)
	seen := map[int]bool{}
	var condition syntax.Expr
	named := false
	for i, x := range call.Args.Args {
		a, ok := x.(*syntax.Arg)
		if !ok || a.Unpack || a.ByRef {
			return nil
		}
		at := i
		if a.Name != nil {
			named = true
			at = -1
			for j, p := range f.Params {
				if a.Name.Value == strings.TrimPrefix(p.Name, "$") {
					at = j
					break
				}
			}
		} else if named {
			return nil
		}
		if at < 0 || at >= len(f.Params) || seen[at] {
			return nil
		}
		seen[at] = true
		if at == 0 {
			condition = a.Value
		} else if condition != nil && !assertDescriptionInert(a.Value) {
			return nil
		}
	}
	return condition
}

func (e *Env) optionalAssert(call *syntax.FuncCall) bool {
	return !e.options.Assertions && e.PHP.AtLeast(phpversion.PHP70) && e.BuiltinAssert(call)
}

// optionalAssertion stops at nested local scopes: executing a body later is
// independent of whether the enclosing assert created its closure object.
func (e *Env) optionalAssertion(n syntax.Node) bool {
	if e.options.Assertions || !e.PHP.AtLeast(phpversion.PHP70) {
		return false
	}
	for child, p := n, n.Parent(); p != nil; child, p = p, p.Parent() {
		if syntax.IsVariableScope(p) {
			return false
		}
		if call, ok := p.(*syntax.FuncCall); ok && call.Args == child && e.optionalAssert(call) {
			return true
		}
	}
	return false
}

// A later description must not rewrite what the assertion already tested.
// Unknown calls and magic-capable expressions may enter reference holders.
func assertDescriptionInert(x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal:
		return true
	case *syntax.ConstFetch:
		return strings.EqualFold(n.Name.Value, "true") || strings.EqualFold(n.Name.Value, "false") || strings.EqualFold(n.Name.Value, "null")
	}
	return false
}

// LegacyAssertCode identifies pre-PHP8 string assertions that can execute PHP
// code in the current scope. The analyzer never interprets that code.
func (e *Env) LegacyAssertCode(call *syntax.FuncCall) bool {
	if e.PHP.AtLeast(phpversion.PHP80) || !e.BuiltinAssert(call) {
		return false
	}
	f := e.ResolveFunction(call)
	value := e.tplArg(call.Args, f.Params, 0)
	if value == nil {
		return true
	}
	// This runs while definitions are being collected. TypeOf here can observe
	// a busy argument and permanently classify the same call differently by
	// query order. Only syntax whose result cannot be a string is safe.
	if nested, ok := syntax.UnwrapParens(value).(*syntax.FuncCall); ok {
		if f := e.ResolveFunction(nested); f != nil && f.Builtin && (f.Return == "bool" || f.Return == "int" || f.Return == "float") {
			return false
		}
	}
	return !assertNonString(value)
}

func assertNonString(x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal:
		return n.LitKind != syntax.LitString
	case *syntax.ConstFetch:
		name := n.Name
		return name != nil && (strings.EqualFold(name.Value, "true") || strings.EqualFold(name.Value, "false") || strings.EqualFold(name.Value, "null"))
	case *syntax.Unary:
		switch n.Op.Kind {
		case syntax.TExclaim, syntax.TIntCast, syntax.TDoubleCast, syntax.TArrayCast, syntax.TBoolCast, syntax.TObjectCast, syntax.TUnsetCast:
			return true
		}
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
			syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual,
			syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor,
			syntax.TSpaceship:
			return true
		}
	case *syntax.Array, *syntax.New, *syntax.Isset, *syntax.Empty, *syntax.Instanceof:
		return true
	}
	return false
}
