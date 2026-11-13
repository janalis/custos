package flow

import (
	"slices"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// keyed reports a PHP function whose result holds no object whatever it is
// given: the keys of an array, its values turned into keys or counted, or
// what JSON text decodes to when objects are asked for as arrays.
func (e *Env) keyed(call syntax.Node, name string, list *syntax.ArgList) bool {
	if _, function := call.(*syntax.FuncCall); !function {
		return false
	}
	switch name {
	case "array_keys", "array_flip", "array_count_values":
		return true
	case "json_decode":
		if list == nil {
			return false
		}
		position := 0
		for _, x := range list.Args {
			arg, ok := x.(*syntax.Arg)
			if !ok || arg.Unpack {
				return false
			}
			if arg.Name == nil {
				position++
			}
			if arg.Name == nil && position == 2 || arg.Name != nil && arg.Name.Value == "associative" {
				return constantName(arg.Value) == "true"
			}
		}
	}
	return false
}

// structural reports a PHP function that counts, moves or picks the
// elements of an array without looking at what they are, called with that
// many arguments.
func structural(name string, given int) bool {
	switch name {
	case "array_keys":
		// With a value to search for, each element is compared with it.
		return given == 1
	case "count", "sizeof", "array_values", "array_key_first", "array_key_last", "array_merge", "array_slice", "array_reverse",
		"array_pop", "array_shift", "array_push", "array_unshift", "array_chunk", "array_is_list", "end", "reset", "current", "key", "next", "prev":
		return true
	}
	return false
}

// sealed reports a PHP function whose result holds no object: it makes none
// and is given none, in an argument that is an array or may be anything, nor
// a function to call, which can make some.
// The same holds of what it leaves in a reference argument.
func (e *Env) sealed(call syntax.Node, name string, list *syntax.ArgList, r *scopeResult) bool {
	if _, function := call.(*syntax.FuncCall); !function {
		return false
	}
	switch name {
	case "func_get_args", "func_get_arg", "get_defined_vars", "debug_backtrace", "compact", "get_object_vars", "get_class_vars":
		return false
	}
	params, _, _ := e.declaration(call)
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		if !ok || at < 0 || invoked(params[at]) && constantName(arg.Value) != "null" && !e.harmless(arg.Value) {
			return false
		}
		if kind := e.kinds(params[at].Type, params[at].DocType); !params[at].ByRef && !r.values.at(arg.Value).deep && (kind&kindKnown == 0 || kind&(kindObject|kindOpaque|kindArray) != 0) {
			return false
		}
	}
	return true
}

// declared returns the static variables of a scope, each with what the code
// of the scope does to it, read once however often the scope is analysed.
//
// state is staticLoose for a variable the scope may write, itself or one of
// its elements: it is the target of an assignment or a step, an argument (a
// callee may take it by reference), a slot of an array or of a
// destructuring, or named by `unset`, `catch` or `foreach`. Only an
// invocation that does so changes a static variable for the others.
//
// starts tells that its first value is not false: each of its declarations
// is written without an initializer, which starts it null, or with a
// literal, a signed number, an array, `null` or `true`.
//
// plain tells that it never holds an object: it starts from none, and every
// write the scope makes to it gives it a value that is none. Reading it then
// runs no method, whichever invocation wrote it last. deep tells that it
// never holds an array with an object in it either.
func (e *Env) declared(scope syntax.Node) map[string]static {
	if known, read := e.statics[scope]; read {
		e.depend(e.staticReads[scope])
		return known
	}
	reads := e.begin()
	defer e.end(reads)
	body := syntax.VariableScopeBody(scope)
	known := map[string]static{}
	syntax.Inspect(body, func(n syntax.Node) bool {
		if declared, ok := n.(*syntax.StaticVar); ok {
			name := declared.Var.Name
			fact, again := known[name]
			if !again {
				fact = static{starts: true, plain: true, deep: true}
			}
			initial := syntax.UnwrapParens(declared.Default)
			if sign, signed := initial.(*syntax.Unary); signed && (sign.Op.Kind == syntax.TMinus || sign.Op.Kind == syntax.TPlus) {
				initial = syntax.UnwrapParens(sign.Expr)
			}
			switch initial := initial.(type) {
			case nil, *syntax.Literal, *syntax.Array:
			case *syntax.ConstFetch:
				fact.starts = fact.starts && (constantName(initial) == "null" || constantName(initial) == "true")
			default:
				fact.starts = false
			}
			fact.plain, fact.deep = fact.plain && e.inert(declared.Default, name, false), fact.deep && e.inert(declared.Default, name, true)
			known[name] = fact
		}
		return true
	})
	syntax.Inspect(body, func(n syntax.Node) bool {
		v, ok := n.(*syntax.Variable)
		if !ok || v.NameExpr != nil {
			return true
		}
		fact, static := known[v.Name]
		if !static {
			return true
		}
		var written syntax.Expr = v
		parent := v.Parent()
		for {
			element, ok := parent.(*syntax.ArrayDimFetch)
			if !ok || element.Var != written {
				break
			}
			written, parent = element, element.Parent()
		}
		switch p := parent.(type) {
		case *syntax.Assign:
			switch {
			case p.ByRef:
				fact.state, fact.plain = staticLoose, false
			case p.Var != written:
			case p.Op.Kind != syntax.TEqual && p.Op.Kind != syntax.TCoalesceEqual && p.Op.Kind != syntax.TPlusEqual:
				fact.state = staticLoose
			case written == syntax.Expr(v):
				fact.state = staticLoose
				fact.plain, fact.deep = fact.plain && e.inert(p.Value, v.Name, false), fact.deep && e.inert(p.Value, v.Name, true)
			default:
				// An element is written: the variable stays an array.
				fact.state, fact.deep = staticLoose, fact.deep && e.inert(p.Value, v.Name, true)
			}
		case *syntax.IncDec:
			fact.state = staticLoose
		case *syntax.Arg:
			fact.state, fact.plain = staticLoose, fact.plain && e.byValue(p)
		case *syntax.ArrayItem:
			fact.state, fact.plain = staticLoose, fact.plain && !p.ByRef && !destructured(p)
		case *syntax.Foreach:
			fact.state, fact.plain = staticLoose, fact.plain && p.Expr == written
		case *syntax.ClosureUse:
			fact.state, fact.plain = staticLoose, fact.plain && !p.ByRef
		case *syntax.Unset, *syntax.Catch:
			fact.state, fact.plain = staticLoose, false
		case *syntax.Global:
			fact.plain = false
		}
		known[v.Name] = fact
		return true
	})
	// A write through a name chosen at run time, an import or code evaluated
	// in the scope may give any of them any value.
	open := len(known) > 0 && e.imports(body)
	for name, fact := range known {
		if open {
			fact.state, fact.plain = staticLoose, false
		}
		fact.deep = fact.deep && fact.plain
		known[name] = fact
	}
	if e.statics == nil {
		e.statics = map[syntax.Node]map[string]static{}
	}
	e.statics[scope] = known
	if reads != nil {
		if e.staticReads == nil {
			e.staticReads = map[syntax.Node]map[string]bool{}
		}
		e.staticReads[scope] = reads
	}
	return known
}

// inert reports an expression that yields no object, whatever the variables
// hold but the one named: a literal, an array, a computation, a test, a
// function, what a PHP function declares to return, or a choice of those.
// With deep, an array it yields holds no object either.
func (e *Env) inert(x syntax.Expr, name string, deep bool) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case nil, *syntax.Literal, *syntax.InterpolatedString, *syntax.MagicConst, *syntax.Isset, *syntax.Empty, *syntax.Instanceof, *syntax.Closure, *syntax.ArrowFunction:
		return true
	case *syntax.ConstFetch:
		return e.fixed(n)
	case *syntax.IncDec:
		// Before PHP 8.3 a step leaves an object as it is.
		return e.types.PHP.AtLeast(phpversion.PHP83) || e.inert(n.Var, name, deep)
	case *syntax.Array:
		for _, item := range n.Items {
			if deep && item != nil && (item.ByRef || !e.inert(item.Value, name, true)) {
				return false
			}
		}
		return true
	case *syntax.Variable:
		return n.NameExpr == nil && n.Name == name
	case *syntax.Binary:
		if n.Op.Kind == syntax.TCoalesce || deep && n.Op.Kind == syntax.TPlus {
			return e.inert(n.Left, name, deep) && e.inert(n.Right, name, deep)
		}
		return true
	case *syntax.Unary:
		if n.Op.Kind == syntax.TAt || deep && n.Op.Kind == syntax.TArrayCast {
			return e.inert(n.Expr, name, deep)
		}
		return n.Op.Kind != syntax.TObjectCast
	case *syntax.Ternary:
		return (n.Then != nil || e.inert(n.Cond, name, deep)) && e.inert(n.Then, name, deep) && e.inert(n.Else, name, deep)
	case *syntax.Assign:
		switch n.Op.Kind {
		case syntax.TEqual:
			return !n.ByRef && e.inert(n.Value, name, deep)
		case syntax.TCoalesceEqual, syntax.TPlusEqual:
			return false
		}
		return true
	case *syntax.FuncCall:
		fname, builtin, result, documented := e.callee(n)
		kind := e.kinds(result, documented)
		if !builtin || kind&(kindKnown|kindObject|kindOpaque) != kindKnown {
			return false
		}
		return !deep || kind&kindArray == 0 || e.given(n, fname, name)
	}
	return false
}

// given is sealed for a call read without its values: each argument is
// written as an expression that yields no object and holds none.
func (e *Env) given(call *syntax.FuncCall, fname, name string) bool {
	switch fname {
	case "func_get_args", "func_get_arg", "get_defined_vars", "debug_backtrace", "compact", "get_object_vars", "get_class_vars":
		return false
	}
	params, _, _ := e.declaration(call)
	var few [8]int
	for i, at := range parameters(call.Args, params, few[:0]) {
		arg, ok := call.Args.Args[i].(*syntax.Arg)
		if !ok || at < 0 || arg.Unpack || invoked(params[at]) && !e.harmless(arg.Value) {
			return false
		}
		if kind := e.kinds(params[at].Type, params[at].DocType); !params[at].ByRef && !e.inert(arg.Value, name, true) && (kind&kindKnown == 0 || kind&(kindObject|kindOpaque|kindArray) != 0) {
			return false
		}
	}
	return true
}

// byValue reports an argument bound to a parameter that is known not to be
// a reference: the callee cannot write the variable given.
func (e *Env) byValue(arg *syntax.Arg) bool {
	list, ok := arg.Parent().(*syntax.ArgList)
	if !ok || arg.ByRef || arg.Unpack {
		return false
	}
	params, _, known := e.declaration(list.Parent())
	var few [8]int
	at := parameters(list, params, few[:0])[slices.Index(list.Args, syntax.Expr(arg))]
	return known && at >= 0 && !params[at].ByRef
}

// destructured reports an array item that is a slot of a destructuring: the
// target of an assignment or of a loop.
func destructured(item *syntax.ArrayItem) bool {
	for {
		array := item.Parent()
		if _, list := array.(*syntax.List); list {
			return true
		}
		switch p := array.Parent().(type) {
		case *syntax.ArrayItem:
			item = p
			continue
		case *syntax.Assign:
			return p.Var == array
		case *syntax.Foreach:
			return p.Value == array || p.Key == array
		}
		return false
	}
}

// imports reports a body that can write a variable it does not name: through
// a name chosen at run time, `extract`, `parse_str` filling the scope, or
// code it evaluates or includes.
func (e *Env) imports(body syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Variable:
			found = found || n.NameExpr != nil
		case *syntax.Eval, *syntax.Include:
			found = true
		case *syntax.FuncCall:
			name, builtin := e.callName(n)
			found = found || builtin && (name == "extract" || name == "parse_str" && n.Args != nil && len(n.Args.Args) == 1)
		}
		return !found
	})
	return found
}

// fixed reports a constant that holds no object: `true`, `false` or `null`,
// one PHP itself defines, or any constant before PHP 8.1, since when the
// project can define one as an object.
func (e *Env) fixed(c *syntax.ConstFetch) bool {
	if constantName(c) != "" || !e.types.PHP.AtLeast(phpversion.PHP81) || c.Name == nil {
		return true
	}
	fqn, fallback := e.types.Names.Const(c.Name.Value, c.Span().Start)
	known := e.types.Index.Constant(fqn, e.types.PHP)
	if known == nil && fallback != "" {
		// A constant defined under a computed name may be the namespaced
		// one, which PHP then reads instead of its own.
		if e.computedDefines() {
			return false
		}
		known = e.types.Index.Constant(fallback, e.types.PHP)
	}
	return known != nil && known.Builtin
}
