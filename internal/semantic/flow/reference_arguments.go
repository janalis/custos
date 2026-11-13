package flow

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// ReferenceResult returns what a fixed project callable leaves in an argument.
// Sources include inputs retained on any path; Expr is present only for a
// single reaching expression from the local reference channel.
// Unpacked, variadic and unresolved bindings cannot establish this proof.
// Multiple reference outputs require distinct, unaliased local argument cells.
// A callee that may modify request globals can also write the actual argument
// through another route; its parameter-only summary cannot prove this result.
func (e *Env) ReferenceResult(call syntax.Node, target *syntax.Arg) (Value, bool) {
	list := arguments(call)
	if list == nil || target == nil {
		return Value{}, false
	}
	if method, ok := call.(*syntax.MethodCall); ok {
		native := e.types.Native().TypeOf(method.Var)
		if method.NullSafe || len(native.Classes()) != 1 || len(native.Atoms()) != 1 {
			return Value{}, false
		}
	}
	if static, ok := call.(*syntax.StaticCall); ok {
		name, named := static.Class.(*syntax.Name)
		if !named || strings.EqualFold(name.Value, "static") {
			return Value{}, false
		}
	}
	params, name, known := e.declaration(call)
	if _, function := call.(*syntax.FuncCall); !function && (e.sinkSummaryName(call) != name || e.reached(call) == followLoose) {
		return Value{}, false
	}
	summary, described := e.contract(name)
	if !known || !described || !summary.Fixed || !summary.Spares {
		return Value{}, false
	}
	r := e.scope(syntax.EnclosingVariableScope(call))
	if !r.complete {
		return Value{}, false
	}
	if len(list.Args) > 256 || len(params) > 64 {
		r.gap("argument-budget", call.Span())
		return Value{}, false
	}
	var few [8]int
	positions := parameters(list, params, few[:0])
	at, references := -1, 0
	for i, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		p := positions[i]
		if !ok || arg.Unpack || arg.ByRef || p < 0 || params[p].Variadic {
			return Value{}, false
		}
		if _, evaluated := r.values.get(arg.Value); !evaluated {
			return Value{}, false
		}
		if params[p].ByRef {
			references++
		}
		if arg == target {
			at = i
		}
	}
	if at < 0 || references != 1 && !r.distinctReferences[call] || !params[positions[at]].ByRef {
		return Value{}, false
	}
	if _, described := summary.ByRef[positions[at]]; !described {
		return Value{}, false
	}
	v := e.left(call, at, list, positions, r.values.at(target.Value), nil, r)
	if body := e.localBody(name); body != nil {
		p := syntax.VariableScopeParams(body)[positions[at]]
		v.Expr = e.scope(body).exit.value(received(p.Var.Name)).Expr
	}
	return v, true
}

// invalidateReferences discards values a callee may replace through reference
// parameters. Unresolved callees cannot establish a by-value parameter contract.
func (e *Env) invalidateReferences(at syntax.Node, list *syntax.ArgList, f *frame, r *scopeResult) {
	if list == nil {
		return
	}
	params, _, known := e.declaration(at)
	var few [8]int
	positions := parameters(list, params, few[:0])
	referenceCount := 0
	for _, param := range params {
		if param.ByRef {
			referenceCount++
		}
	}
	if known && referenceCount > 1 {
		distinct := distinctReferenceLocals(list, params, positions, f, r)
		if old, seen := r.distinctReferences[at]; seen {
			distinct = distinct && old
		}
		if r.distinctReferences == nil {
			r.distinctReferences = map[syntax.Node]bool{}
		}
		r.distinctReferences[at] = distinct
	}
	for i, x := range list.Args {
		a, ok := x.(*syntax.Arg)
		if !ok {
			continue
		}
		p := positions[i]
		if known && p >= 0 && !params[p].ByRef {
			continue
		}
		untest(a.Value, f, false)
		target := syntax.UnwrapParens(a.Value)
		for {
			dim, ok := target.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			target = syntax.UnwrapParens(dim.Var)
		}
		if variable, ok := target.(*syntax.Variable); ok && variable.NameExpr == nil {
			// A declared reference parameter replaces the value. An unresolved
			// callee may not write at all: request data stays possible.
			// The write reaches the names aliased with the argument too.
			next := unsettled(f.vars.value(variable.Name), Value{}, r)
			if known {
				next = Value{}
				name, builtin, _, _ := e.callee(at)
				switch {
				case !builtin:
					next = e.left(at, i, list, positions, f.vars.value(variable.Name), f, r)
				case name == "settype" && scalarType(list):
				case name == "mysqli_stmt_bind_result" || name == "bind_result" || name == "bindcolumn":
					// The variable will receive what the database returns.
				case name == "mysqli_stmt_bind_param" || name == "bind_param" || name == "bindparam":
					// Binding a statement parameter reads the variable when
					// the statement runs; it never writes it.
					continue
				default:
					next = e.updated(name, i, list, positions, params, f.vars.value(variable.Name), r)
				}
				if builtin && p >= 0 {
					// A PHP function leaves what it declares for the parameter:
					// an array it filled, or the one it was given, reordered.
					kind := e.kinds(params[p].Type, params[p].DocType)
					next.plain = kind&(kindKnown|kindObject|kindOpaque) == kindKnown
					next.deep = next.plain && (kind&kindArray == 0 || e.sealed(at, name, list, r) && (params[p].Optional || params[p].Variadic || f.vars.value(variable.Name).deep))
				}
			}
			if old := f.vars.value(variable.Name); target != syntax.UnwrapParens(a.Value) {
				// Only the element can be bound: the callee may change or
				// create it, with what it writes, but the variable stays the
				// array it was, with the request data its other elements hold.
				next = unsettled(old, next, r)
				next.Complete = next.Complete || old.Complete
				next.NonFalse, next.NonNull, next.plain, next.deep = old.NonFalse, old.NonNull, old.plain, old.deep && next.deep
				next.elements = old.elements.without(boundElement(a.Value))
				next.Expr, next.origins = nil, origins(Value{}, old, r)
			}
			f.write(r, variable.Name, next)
		}
	}
}

// left is what a function or method of the project leaves in the reference
// argument at index at, from its summary: the request data it writes there
// and the values of the arguments that may end up in it. Without a summary
// the callee may not write at all. A body that may not be the one called (a
// conditional declaration, a method a subclass can replace, a generator)
// adds to what the argument held instead of replacing it.
func (e *Env) left(call syntax.Node, at int, list *syntax.ArgList, positions []int, old Value, f *frame, r *scopeResult) Value {
	// `new C` runs the constructor of C itself; a method may be replaced.
	// The summary is the one of the body that runs when the callee is exactly
	// known: a final or private method, a method of a final class, a static
	// call that names its class, as `self::` and `parent::` do too, or a
	// method no class of the project extending the receiver's replaces
	// (`static::` and a method called on an object bind late).
	name, sealed := e.summaryName(call), true
	switch n := call.(type) {
	case *syntax.MethodCall:
		followed := e.reached(call)
		if f != nil {
			followed = e.follows(n.Var, f)
		}
		name, sealed = e.dispatched(call, followed, name)
	case *syntax.StaticCall:
		name, sealed = e.dispatched(call, followTyped, name)
	}
	summary, ok := e.contract(name)
	written, described := summary.ByRef[positions[at]]
	if !ok || !described {
		return unsettled(old, Value{}, r)
	}
	v := Value{Complete: true, Safe: written.Safe}
	for _, source := range written.Sources {
		if source.Parameter < 0 {
			v.Sources, _ = union(v.Sources, []Source{source})
			continue
		}
		for i, position := range positions {
			arg, isArg := list.Args[i].(*syntax.Arg)
			if position != source.Parameter || !isArg {
				continue
			}
			in := r.values.at(arg.Value)
			if i == at {
				in = old
			}
			in = projectSource(in, source)
			if len(in.Sources) > 0 {
				v = merge(v, Value{Complete: true, Sources: in.Sources, Safe: in.Safe})
			}
		}
	}
	v.Safe |= written.Safe
	next := Value{}
	if len(v.Sources) > 0 {
		next = Value{Complete: true, Sources: v.Sources, Safe: v.Safe}
	}
	if !summary.Fixed || !sealed {
		return unsettled(old, next, r)
	}
	// An array of numbers the callee left is one for the caller.
	next.scalars = written.scalars
	if body := e.localBody(name); body != nil {
		params := syntax.VariableScopeParams(body)
		at := positions[at]
		if at >= 0 && at < len(params) && params[at].Var != nil {
			result := e.scope(body)
			written := result.exit.value(received(params[at].Var.Name))
			next.origins = result.producers(written)
		}
	}
	return next
}

// boundElement returns the key of the element of a variable an argument
// reaches when every offset on the way is constant, "" and false otherwise.
func boundElement(argument syntax.Expr) (string, bool) {
	key, known := "", true
	for x := syntax.UnwrapParens(argument); known; {
		fetch, ok := x.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		key, known = arrayKey(fetch.Dim)
		x = syntax.UnwrapParens(fetch.Var)
	}
	return key, known
}

// updated is what a PHP function leaves in the reference argument at index
// at: the request data of its other arguments may reach it, and a required
// reference parameter is worked on in place, so it also keeps what it held.
// A parameter that receives a number or a boolean holds no text, a callback
// is not data, and what a command prints is not what the request sent.
// Nothing else is known about the result.
func (e *Env) updated(name string, at int, list *syntax.ArgList, positions []int, params []index.Param, old Value, r *scopeResult) Value {
	p := positions[at]
	if p >= 0 && counter(params[p]) || name == "exec" {
		return Value{}
	}
	if name == "sscanf" && len(list.Args) > 1 {
		if format, ok := list.Args[1].(*syntax.Arg); ok {
			if text, known := literalText(format.Value); known && numericFormat(text) {
				return Value{}
			}
		}
	}
	inputs := Value{Complete: true}
	for i, x := range list.Args {
		other, ok := x.(*syntax.Arg)
		if !ok || i == at || positions[i] >= 0 && strings.Contains(params[positions[i]].Type, "callable") {
			continue
		}
		if in := r.values.at(other.Value); tainted(in) {
			inputs = merge(inputs, Value{Complete: true, Sources: in.Sources, Safe: in.Safe})
		}
	}
	if p < 0 || params[p].Optional || params[p].Variadic {
		old = Value{}
	}
	next := unsettled(old, inputs, r)
	// Reordering or shortening an array of numbers leaves one.
	next.scalars = old.scalars && !tainted(inputs) && !strings.HasPrefix(name, "array_walk")
	return next
}

// counter reports a reference parameter PHP fills with a number or a
// boolean: one declared so, or one of the untyped result parameters its
// functions name alike (a replacement count, an exit status, an error code).
func counter(p index.Param) bool {
	switch p.Name {
	case "count", "percent", "result_code", "error_code", "still_running", "would_block", "line":
		return true
	}
	return scalarResult(p.Type)
}

// numericFormat reports a scanf format whose every conversion reads a number.
func numericFormat(format string) bool {
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		i++
		for i < len(format) && (format[i] == '*' || format[i] >= '0' && format[i] <= '9') {
			i++
		}
		if i == len(format) || !strings.ContainsRune("%dDiufeEgxXon", rune(format[i])) {
			return false
		}
	}
	return true
}

// scalarType reports a settype() call whose literal target type is a number,
// a boolean or null: the converted variable holds no text.
func scalarType(list *syntax.ArgList) bool {
	if len(list.Args) != 2 {
		return false
	}
	if target, ok := list.Args[1].(*syntax.Arg); ok {
		name, _ := literalText(target.Value)
		switch strings.ToLower(name) {
		case "int", "integer", "float", "double", "bool", "boolean", "null":
			return true
		}
	}
	return false
}

// Only independent local cells may reuse parameter-only reference summaries.
// Alias groups, captured storage, globals, static locals and element lvalues
// can overlap through paths not represented by their written variable names.
func distinctReferenceLocals(list *syntax.ArgList, params []index.Param, positions []int, f *frame, r *scopeResult) bool {
	if r.global || !f.definite {
		return false
	}
	names := map[string]bool{}
	count := 0
	for i, node := range list.Args {
		p := positions[i]
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || arg.ByRef || p < 0 {
			return false
		}
		if !params[p].ByRef {
			continue
		}
		name := plainVariable(arg.Value)
		value := f.vars.value(name)
		_, shared := f.aliased[name]
		_, static := r.statics[name]
		if !value.Complete || value.replaced || value.Invalidated != 0 || name == "" || external(name) || name == "GLOBALS" || names[name] || shared || static || r.globalNames[name] {
			return false
		}
		names[name] = true
		count++
	}
	return count > 1
}
