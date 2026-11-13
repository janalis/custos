package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// refineFuncCall refines what a PHP predicate or check settles about its
// arguments. Checks of resources and of what an argument held come first;
// then the predicate tests a variable or the assignment written in its place
// (`is_string($n = f())`).
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
	if truth {
		refineResourceCheck(n, name, args[0], f)
		refineCheckedArguments(name, args, f)
	}
	variable := guarded(args[0].Value)
	if variable == nil {
		return
	}
	refineTypePredicate(name, variable, f, truth)
	if truth && name == "in_array" && len(args) == 3 {
		refineInArray(args[1], args[2], variable, f)
	}
	if !truth && name == "str_contains" && len(args) == 2 {
		refineStrContains(args[1], variable, f)
	}
}

// refineResourceCheck marks the handle a successful flock() or fclose() was
// given as successfully operated on.
func refineResourceCheck(n *syntax.FuncCall, name string, first *syntax.Arg, f *frame) {
	if name != "flock" && name != "fclose" {
		return
	}
	if variable, ok := first.Value.(*syntax.Variable); ok {
		value := f.vars.value(variable.Name)
		if state, ok := f.states[value.Identity]; ok {
			state.Successful = true
			state.closeChecked = name == "fclose" && state.Operation == "fclose" && state.Span == n.Span()
			f.states[value.Identity] = state
		}
	}
}

// refineCheckedArguments excludes false and null from an argument a check
// that held proves to be an array or text that is not empty.
func refineCheckedArguments(name string, args []*syntax.Arg, f *frame) {
	switch name {
	case "array_key_exists", "key_exists":
		if len(args) == 2 && args[1] != nil {
			// Only an array (or an object) has keys.
			exclude(args[1].Value, f, true, true)
		}
	case "preg_match", "preg_match_all":
		// A match of a pattern that needs at least one character: the
		// subject was text that is not empty, so neither false nor null.
		if len(args) >= 2 && args[1] != nil && EmptylessPattern(args[0].Value) {
			exclude(args[1].Value, f, true, true)
		}
	case "json_validate":
		// Valid JSON is never the empty text false and null read as.
		exclude(args[0].Value, f, true, true)
	}
}

// refineTypePredicate refines a variable an is_* or ctype_digit() predicate
// tested.
func refineTypePredicate(name string, variable *syntax.Variable, f *frame, truth bool) {
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
	if truth && (name == "ctype_digit" || name == "is_int" || name == "is_integer") {
		// Digits alone, or a number, are safe in every text context.
		v := f.vars.value(variable.Name)
		v.Safe |= everyContext
		f.setVar(variable.Name, v)
	}
}

// refineInArray refines a variable found by a strict in_array() among a
// list of literals: it is one of them.
func refineInArray(list, strict *syntax.Arg, variable *syntax.Variable, f *frame) {
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
	v := f.vars.value(variable.Name)
	v.Sources = nil
	v.Safe = everyContext | URL | Path
	f.setVar(variable.Name, v)
	// The value is one of the literals: neither false nor null.
	exclude(variable, f, true, true)
}

// refineStrContains refines a variable str_contains() found without a
// carriage return or a line feed: lacking both, it cannot end a header.
func refineStrContains(arg *syntax.Arg, variable *syntax.Variable, f *frame) {
	needle, ok := literalText(arg.Value)
	if !ok {
		return
	}
	v := f.vars.value(variable.Name)
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
