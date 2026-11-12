package flow

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// calm reports an expression that cannot end the function by throwing, its
// operands aside: a variable read, a constant, a plain assignment to a
// variable, an identity test or a boolean combination, an array literal, a
// choice between two values.
func calm(x syntax.Expr) bool {
	switch n := x.(type) {
	case *syntax.Literal, *syntax.ConstFetch, *syntax.Name, *syntax.Identifier, *syntax.Paren, *syntax.Array, *syntax.ArrayItem, *syntax.Ternary:
		return true
	case *syntax.Variable:
		return n.NameExpr == nil
	case *syntax.Assign:
		return n.Op.Kind == syntax.TEqual && !n.ByRef && plainVariable(n.Var) != ""
	case *syntax.Unary:
		return n.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return true
		}
	}
	return false
}

// visible reports that a method is the one a call written here runs: a
// method that is not public is only reached from its own class or, when it
// is protected, from a class that extends it; anywhere else the call goes to
// `__call` or `__callStatic`, or fails. Each class on the way up from the
// one the call is written in must be declared once, with its parent.
func (e *Env) visible(call syntax.Node, m *index.Method) bool {
	within := strings.TrimPrefix(e.types.ClassFQN(syntax.EnclosingClass(call)), `\`)
	declaring := strings.TrimPrefix(m.Class, `\`)
	if m.Visibility == index.Public || strings.EqualFold(within, declaring) {
		return true
	}
	ix := e.types.Index
	for steps := 0; m.Visibility == index.Protected && steps < MaxDepth*4; steps++ {
		class := ix.Class(within, e.types.PHP)
		if class == nil || ix.ClassCount(within) != 1 || class.Parent == "" {
			break
		}
		if within = strings.TrimPrefix(class.Parent, `\`); strings.EqualFold(within, declaring) {
			return ix.ClassCount(within) == 1
		}
	}
	return false
}

// quiet reports a call whose callee is exactly known and runs no code of the
// project but its own. Binding the arguments runs none either: no object is
// converted to text by a `string` parameter, no parameter looks a callable
// up, and no default value loads a class or makes an object.
func (e *Env) quiet(call syntax.Node, list *syntax.ArgList, values []Value) bool {
	summary, ok := e.contract(e.summaryName(call))
	if !ok || !summary.Quiet {
		return false
	}
	params, _, _ := e.declaration(call)
	var few [8]int
	positions := few[:0]
	if list != nil {
		positions = parameters(list, params, positions)
	}
	for at, p := range params {
		text, callable := coerced(p.Type)
		if callable || strings.Contains(p.Default, "::") || strings.Contains(strings.ToLower(p.Default), "new") {
			return false
		}
		for i, bound := range positions {
			if text && bound == at && !values[i].plain {
				return false
			}
		}
	}
	switch n := call.(type) {
	case *syntax.FuncCall:
		return true
	case *syntax.StaticCall:
		class, named := n.Class.(*syntax.Name)
		m := e.method(call)
		return named && m != nil && !strings.EqualFold(class.Value, "static") && e.visible(call, m)
	}
	m := e.method(call)
	return m != nil && (m.Final || m.Visibility == index.Private) && e.visible(call, m)
}

// coerced reports a declared parameter type that makes PHP run code of the
// project to bind an argument: `string` calls the `__toString` method of an
// object, `callable` loads the class a callable names.
func coerced(declared string) (text, callable bool) {
	declared = strings.TrimPrefix(declared, "?")
	for declared != "" {
		var part string
		part, declared, _ = strings.Cut(declared, "|")
		switch strings.ToLower(part) {
		case "string":
			text = true
		case "callable":
			callable = true
		}
	}
	return text, callable
}

// What a declared type admits, as far as running code of the project goes.
const (
	// kindKnown: the type was read.
	kindKnown uint8 = 1 << iota
	// kindText: it admits text, which an object is converted to.
	kindText
	// kindObject: it admits objects.
	kindObject
	// kindOpaque: it admits anything, or a callable.
	kindOpaque
	// kindArray: it admits arrays, which may hold objects.
	kindArray
)

// kinds reads a declared type, or the documented one where PHP declares
// none. Numbers, booleans, null, arrays and resources are none of the kinds.
func (e *Env) kinds(declared, documented string) uint8 {
	if declared == "" {
		declared = documented
	}
	if kind, read := e.kindOf[declared]; read {
		return kind
	}
	var kind uint8
	for _, atom := range types.FromDoc(declared, nil).Atoms() {
		kind |= kindKnown
		switch atom {
		case "int", "float", "bool", "true", "false", "null", "void", "never", "resource":
		case "string":
			kind |= kindText
		case "array":
			kind |= kindArray
		case "mixed", "callable":
			kind |= kindOpaque
		default:
			if strings.HasSuffix(atom, "[]") {
				kind |= kindArray
			} else {
				kind |= kindObject
			}
		}
	}
	if e.kindOf == nil {
		e.kindOf = map[string]uint8{}
	}
	e.kindOf[declared] = kind
	return kind
}

// handed reports a call of PHP itself, or the construction of one of its
// classes, that may run code of the project with what it is given: an
// object where it converts to text, counts, iterates, serializes or
// compares, a class it loads, an error it raises, a generator or a fiber it
// resumes, a method a class of the project replaces, a handler or a class
// it is given the name of (see runner, loads, invokes and builds). A
// constructor only keeps the objects it is given, and a type test only looks
// at its argument.
func (e *Env) handed(call syntax.Node, name string, list *syntax.ArgList, values []Value) bool {
	params, qualified, _ := e.declaration(call)
	_, constructed := call.(*syntax.New)
	if _, function := call.(*syntax.FuncCall); function {
		if runner(name) || loads(name, list) {
			return true
		}
		switch name {
		case "is_array", "is_string", "is_int", "is_integer", "is_long", "is_float", "is_double", "is_bool", "is_null", "is_numeric", "is_scalar",
			"is_object", "is_resource", "is_iterable", "is_countable", "gettype", "get_debug_type", "get_class", "spl_object_id", "spl_object_hash",
			"array_key_exists", "key_exists":
			return false
		case "array_shift", "array_pop":
			// The element taken out is destroyed when nothing keeps it.
			if _, dropped := call.Parent().(*syntax.ExprStmt); dropped && len(values) > 0 && !values[0].deep {
				return true
			}
		}
	} else if m := e.method(call); !constructed && m != nil && (m.Abstract || !m.Final && e.replaced(m.Class, m)) {
		return true
	}
	class, _, _ := strings.Cut(qualified, "::")
	if strings.HasPrefix(qualified, "generator::") || strings.HasPrefix(qualified, "fiber::") || !constructed && invokes(qualified, list) || constructed && builds(class, list) {
		return true
	}
	if list == nil {
		return false
	}
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		kind := uint8(0)
		if at >= 0 {
			kind = e.kinds(params[at].Type, params[at].DocType)
		}
		switch given := values[i]; {
		case given.deep:
		case kind&kindKnown == 0 || kind&kindOpaque != 0:
			return true
		case !given.plain && (kind&kindText != 0 || kind&kindObject != 0 && !constructed):
			return true
		case kind&kindArray != 0 && !constructed && !structural(name, len(values)):
			// The objects an array holds are converted or compared.
			return true
		}
	}
	return false
}

// converted reports a binary operation that calls the `__toString` method
// of an operand that is an object: a concatenation, or a loose comparison
// with a value that may be text.
func converted(n *syntax.Binary, left, right Value) bool {
	switch n.Op.Kind {
	case syntax.TDot:
		return !left.plain || !right.plain
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual, syntax.TSpaceship:
		return !left.plain && !textless(n.Right) || !right.plain && !textless(n.Left)
	}
	return false
}

// textless reports an operand written as a number, a boolean or null:
// comparing an object with it converts nothing to text.
func textless(x syntax.Expr) bool {
	x = syntax.UnwrapParens(x)
	if sign, ok := x.(*syntax.Unary); ok && (sign.Op.Kind == syntax.TMinus || sign.Op.Kind == syntax.TPlus) {
		x = syntax.UnwrapParens(sign.Expr)
	}
	literal, ok := x.(*syntax.Literal)
	return constantName(x) != "" || ok && literal.LitKind != syntax.LitString
}

// returnsText reports a function whose declared return type admits text:
// an object returned is converted to it.
func returnsText(scope syntax.Node) bool {
	var declared syntax.Expr
	switch n := scope.(type) {
	case *syntax.Function:
		declared = n.ReturnType
	case *syntax.Method:
		declared = n.ReturnType
	case *syntax.Closure:
		declared = n.ReturnType
	}
	return types.FromNode(declared, nil).Has("string")
}

// callbackParameter reports a resolved builtin given a function it may
// call. None is given where the parameter is left out. A PHP function named
// by a literal, that only computes a number, a boolean or text, runs no code
// of the project when no argument of the call can hold an object for it to
// convert.
func (e *Env) callbackParameter(n syntax.Node, list *syntax.ArgList, values []Value) bool {
	params, _, _ := e.declaration(n)
	if list == nil || !slices.ContainsFunc(params, func(p index.Param) bool { return strings.Contains(p.Type, "callable") }) {
		return false
	}
	var few [8]int
	given, converts := false, false
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		switch {
		case !ok || at < 0 || arg.Unpack || i >= len(values):
			return true
		case !strings.Contains(params[at].Type, "callable"):
			converts = converts || !values[i].deep
		case constantName(arg.Value) == "null":
		case !e.harmless(arg.Value):
			return true
		default:
			given = true
		}
	}
	return given && converts
}

// scalarResult reports a native type that only admits numbers, booleans and
// null. Such a value cannot carry markup, a query fragment or a command.
func scalarResult(declared string) bool {
	if declared == "" {
		return false
	}
	for declared != "" {
		var part string
		part, declared, _ = strings.Cut(declared, "|")
		switch strings.ToLower(strings.TrimPrefix(part, "?")) {
		case "int", "float", "bool", "false", "true", "null", "void", "never":
		default:
			return false
		}
	}
	return true
}

// nativeClass reports the construction of a class PHP itself declares, whose
// constructor runs no project code.
func (e *Env) nativeClass(n *syntax.New) bool {
	name, ok := n.Class.(*syntax.Name)
	if !ok {
		return false
	}
	class := e.types.Index.Class(e.types.Names.Class(name.Value, name.Span().Start), e.types.PHP)
	return class != nil && strings.HasPrefix(class.File, "stubs/")
}

// invoked reports a parameter of a PHP function that takes a function to
// call.
func invoked(p index.Param) bool {
	return strings.Contains(strings.ToLower(p.Type), "callable") || strings.Contains(strings.ToLower(p.DocType), "callable")
}

// harmless reports a function to call, named by a string literal, that is a
// PHP function which only computes on text and numbers (see pure): it makes
// no object, hands none on, and runs no code of the project.
func (e *Env) harmless(callback syntax.Expr) bool {
	name, known := literalText(callback)
	if !known {
		return false
	}
	name = strings.ToLower(strings.TrimPrefix(name, `\`))
	f := e.types.Index.Function(name, e.types.PHP)
	return f != nil && f.Builtin && pure(name)
}

// ends reports an expression that never completes: a throw, an exit or a
// call that does not return.
func (e *Env) ends(x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Throw, *syntax.Exit:
		return true
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		return e.never(n)
	}
	return false
}
