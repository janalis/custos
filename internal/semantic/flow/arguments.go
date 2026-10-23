package flow

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
)

// BoundArguments returns values in declaration parameter order. Unknown
// targets, variadic/unpacked arguments and duplicate bindings are incomplete.
func (e *Env) BoundArguments(at syntax.Node) ([]Value, bool) {
	r := e.scope(syntax.EnclosingVariableScope(at))
	for _, c := range r.calls {
		if c.Node == at {
			return e.bind(at, c.Arguments)
		}
	}
	return nil, false
}

// ReferenceResult returns what a fixed project function leaves in an argument.
// Sources include inputs retained on any path; Expr is present only for a
// single reaching expression from the local reference channel.
// Unpacked, variadic and unresolved bindings cannot establish this proof.
// Multiple reference arguments may alias, so their outputs are not projected.
// A callee that may modify request globals can also write the actual argument
// through another route; its parameter-only summary cannot prove this result.
func (e *Env) ReferenceResult(call *syntax.FuncCall, target *syntax.Arg) (Value, bool) {
	if call == nil || call.Args == nil || target == nil {
		return Value{}, false
	}
	params, name, known := e.declaration(call)
	summary, described := e.contract(name)
	if !known || !described || !summary.Fixed || !summary.Spares {
		return Value{}, false
	}
	r := e.scope(syntax.EnclosingVariableScope(call))
	if !r.complete {
		return Value{}, false
	}
	if len(call.Args.Args) > 256 || len(params) > 64 {
		r.gap("argument-budget", call.Span())
		return Value{}, false
	}
	var few [8]int
	positions := parameters(call.Args, params, few[:0])
	at, references := -1, 0
	for i, node := range call.Args.Args {
		arg, ok := node.(*syntax.Arg)
		p := positions[i]
		if !ok || arg.Unpack || arg.ByRef || p < 0 || params[p].Variadic {
			return Value{}, false
		}
		if _, evaluated := r.values[arg.Value]; !evaluated {
			return Value{}, false
		}
		if params[p].ByRef {
			references++
		}
		if arg == target {
			at = i
		}
	}
	if at < 0 || references != 1 || !params[positions[at]].ByRef {
		return Value{}, false
	}
	if _, described := summary.ByRef[positions[at]]; !described {
		return Value{}, false
	}
	v := e.left(call, at, call.Args, positions, r.values[target.Value], nil, r)
	if body := e.localBody(name); body != nil {
		p := syntax.VariableScopeParams(body)[positions[at]]
		v.Expr = e.scope(body).exit[received(p.Var.Name)].Expr
	}
	return v, true
}

func (e *Env) declaration(at syntax.Node) ([]index.Param, string, bool) {
	switch n := at.(type) {
	case *syntax.FuncCall:
		if f := e.types.ResolveFunction(n); f != nil {
			return f.Params, names.Fold(f.FQN), true
		}
	case *syntax.MethodCall:
		id, ok := n.Name.(*syntax.Identifier)
		if !ok {
			return nil, "", false
		}
		classes := e.types.TypeOf(n.Var).Classes()
		if len(classes) != 1 {
			return nil, "", false
		}
		if m := e.types.Index.FindMethod(classes[0], id.Value, e.types.PHP); m != nil && !m.Magic {
			return m.Params, strings.ToLower(m.Class + "::" + m.Name), true
		}
	case *syntax.StaticCall:
		id, ok := n.Name.(*syntax.Identifier)
		class, named := n.Class.(*syntax.Name)
		if !ok || !named {
			return nil, "", false
		}
		if m := e.types.Index.FindMethod(e.named(class), id.Value, e.types.PHP); m != nil && !m.Magic {
			return m.Params, strings.ToLower(m.Class + "::" + m.Name), true
		}
	case *syntax.New:
		if class, ok := n.Class.(*syntax.Name); ok {
			fqn := e.types.Names.Class(class.Value, class.Span().Start)
			if m := e.types.Index.FindMethod(fqn, "__construct", e.types.PHP); m != nil {
				return m.Params, strings.ToLower(m.Class + "::__construct"), true
			}
		}
	}
	return nil, "", false
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
			next := unsettled(f.vars[variable.Name], Value{}, r)
			if known {
				next = Value{}
				name, builtin, _, _ := e.callee(at)
				switch {
				case !builtin:
					next = e.left(at, i, list, positions, f.vars[variable.Name], f, r)
				case name == "settype" && scalarType(list):
				case name == "mysqli_stmt_bind_result" || name == "bind_result" || name == "bindcolumn":
					// The variable will receive what the database returns.
				case name == "mysqli_stmt_bind_param" || name == "bind_param" || name == "bindparam":
					// Binding a statement parameter reads the variable when
					// the statement runs; it never writes it.
					continue
				default:
					next = e.updated(name, i, list, positions, params, f.vars[variable.Name], r)
				}
				if builtin && p >= 0 {
					// A PHP function leaves what it declares for the parameter:
					// an array it filled, or the one it was given, reordered.
					kind := e.kinds(params[p].Type, params[p].DocType)
					next.plain = kind&(kindKnown|kindObject|kindOpaque) == kindKnown
					next.deep = next.plain && (kind&kindArray == 0 || e.sealed(at, name, list, r) && (params[p].Optional || params[p].Variadic || f.vars[variable.Name].deep))
				}
			}
			if old := f.vars[variable.Name]; target != syntax.UnwrapParens(a.Value) {
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
		name, sealed = e.dispatched(call, e.follows(n.Var, f), name)
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
			in := r.values[arg.Value]
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
			written := result.exit[received(params[at].Var.Name)]
			next.origins = result.producers(written)
		}
	}
	return next
}

// dispatched names the summary of the method a call runs, the given one
// when the method is not resolved, and tells whether it is exactly known.
func (e *Env) dispatched(call syntax.Node, followed uint8, name string) (string, bool) {
	if m, sealed := e.dispatch(call, followed); m != nil {
		return strings.ToLower(m.Class + "::" + m.Name), sealed
	}
	return name, false
}

// dispatch resolves the method a method or static call names and tells
// whether that declaration is the one that runs: a final or private method,
// a method of a final class, a static call that names its class, as `self::`
// and `parent::` do too, a method called on an object made here by `new`, or
// a method no class of the project extending the receiver's replaces
// (`static::` and a method called on an object bind late). followed tells
// what is known of the receiver (see follows).
func (e *Env) dispatch(call syntax.Node, followed uint8) (*index.Method, bool) {
	m, receiver, sealed, open := e.resolve(call, followed)
	if m != nil && !sealed && open {
		sealed = !e.replaced(receiver, m)
	}
	return m, sealed
}

// resolve is dispatch without the classes of the project: sealed tells that
// the declaration is the one that runs whatever they declare, and open that
// it is unless one of those extending the receiver's class replaces it.
func (e *Env) resolve(call syntax.Node, followed uint8) (m *index.Method, receiver string, sealed, open bool) {
	// A method that is not visible from here is not the one that runs.
	m = e.method(call)
	if m == nil || !e.visible(call, m) {
		return nil, "", false, false
	}
	sealed = m.Final || m.Visibility == index.Private || followed == followExact
	// The class of the object the method is looked up on at run time is the
	// receiver's, the enclosing one for `$this` and `static`, or one that
	// extends it.
	within := syntax.EnclosingClass(call)
	receiver = e.types.ClassFQN(within)
	if static, ok := call.(*syntax.StaticCall); ok {
		class, _ := static.Class.(*syntax.Name)
		sealed = sealed || !strings.EqualFold(class.Value, "static")
	} else if typed := e.types.TypeOf(call.(*syntax.MethodCall).Var).Classes(); len(typed) == 1 {
		receiver, within = typed[0], nil
	}
	// A trait method runs in classes that are not known here.
	return m, receiver, sealed, followed != followLoose && (within == nil || within.ClassKind != syntax.KindTrait)
}

// What is known of the receiver of a method call.
const (
	// followExact: a variable that holds an object made by `new` of the
	// class inferred for it, on every path.
	followExact uint8 = iota + 1
	// followTyped: an object of the class inferred, or of one extending it.
	followTyped
	// followLoose: a variable that may hold something else.
	followLoose
)

// follows tells whether a receiver is the object inferred for it: a variable
// is only while the flow follows it, created here, received or produced by
// one expression every path agrees on, and bound to nothing. Any other
// receiver is what is declared or documented for it.
func (e *Env) follows(receiver syntax.Expr, f *frame) uint8 {
	name := plainVariable(receiver)
	if name == "" || name == "this" {
		return followTyped
	}
	held := f.vars[name]
	if _, shared := f.aliased[name]; shared || held.Identity == 0 && (held.Expr == nil || held.Invalidated != 0) {
		return followLoose
	}
	made, created := syntax.UnwrapParens(held.Expr).(*syntax.New)
	typed := e.types.TypeOf(receiver).Classes()
	if !created || held.Invalidated != 0 || len(typed) != 1 {
		return followTyped
	}
	if class, named := made.Class.(*syntax.Name); named && !strings.EqualFold(class.Value, "static") && strings.EqualFold(strings.TrimPrefix(e.named(class), `\`), strings.TrimPrefix(typed[0], `\`)) {
		return followExact
	}
	return followTyped
}

// Targets returns the declarations a method or static call may run: the one
// it resolves to when no other can (see dispatch), else that one and each
// method that replaces it in a class of the project extending the
// receiver's. Every class from the receiver's up to the one that declares
// the method, and every class extending it, must be declared once: of two
// declarations either may be the one in effect. Nothing is returned when
// the callee is not known that well. A method PHP itself declares only has
// to be the one no class of the project replaces.
func (e *Env) Targets(call syntax.Node) []*index.Method {
	return e.targets(call, e.reached(call))
}

// reached tells what the scope of a method call knows of its receiver (see
// follows): a variable is followed only in a scope analysed to its end.
func (e *Env) reached(call syntax.Node) uint8 {
	r := e.scope(syntax.EnclosingVariableScope(call))
	followed := followTyped
	if method, ok := call.(*syntax.MethodCall); ok && plainVariable(method.Var) != "" && plainVariable(method.Var) != "this" {
		if followed = r.receivers[call]; followed == 0 || !r.complete {
			followed = followLoose
		}
	}
	return followed
}

// targets is Targets for what is known of the receiver.
func (e *Env) targets(call syntax.Node, followed uint8) []*index.Method {
	m, receiver, sealed, open := e.resolve(call, followed)
	if m == nil || m.Magic || !sealed && !open {
		return nil
	}
	targets := []*index.Method{m}
	if !m.Builtin && !e.single(call, m) {
		if targets = e.copies(call, m); targets == nil {
			return nil
		}
	}
	if sealed {
		return targets
	}
	ix := e.types.Index
	if e.below(receiver, func(sub string) bool {
		other := ix.FindMethod(sub, m.Name, e.types.PHP)
		if ix.ClassCount(sub) != 1 || other == nil {
			return true
		}
		if !slices.Contains(targets, other) {
			targets = append(targets, other)
		}
		return false
	}) {
		return nil
	}
	return targets
}

// single reports that each class on the way from the one a call looks its
// method up on to the one that declares it is declared once: none of them
// may be another declaration, with another method.
func (e *Env) single(call syntax.Node, m *index.Method) bool {
	ix := e.types.Index
	from := e.lookedUp(call)
	for steps := 0; steps < MaxDepth*4 && ix.ClassCount(from) == 1; steps++ {
		class := ix.Class(from, e.types.PHP)
		if strings.EqualFold(strings.TrimPrefix(from, `\`), strings.TrimPrefix(m.Class, `\`)) || class.Parent == "" {
			// A method of a trait is declared outside the classes above.
			return ix.ClassCount(m.Class) == 1
		}
		from = class.Parent
	}
	return false
}

// lookedUp names the class a call looks its method up on.
func (e *Env) lookedUp(call syntax.Node) string {
	from := e.types.ClassFQN(syntax.EnclosingClass(call))
	switch n := call.(type) {
	case *syntax.New:
		from = e.named(n.Class.(*syntax.Name))
	case *syntax.StaticCall:
		from = e.named(n.Class.(*syntax.Name))
	case *syntax.MethodCall:
		if typed := e.types.TypeOf(n.Var).Classes(); len(typed) == 1 {
			from = typed[0]
		}
	}
	return from
}

// copies returns the method of each declaration of the class a call looks it
// up on, when the project declares that class more than once and every copy
// declares the method itself: any of them may be the one that runs.
func (e *Env) copies(call syntax.Node, m *index.Method) []*index.Method {
	from := e.lookedUp(call)
	if !strings.EqualFold(strings.TrimPrefix(from, `\`), strings.TrimPrefix(m.Class, `\`)) {
		return nil
	}
	// single found more than one declaration of the class.
	declarations := e.types.Index.ClassDecls(from, e.types.PHP)
	methods := make([]*index.Method, 0, len(declarations))
	for _, class := range declarations {
		own := class.Methods[strings.ToLower(m.Name)]
		if own == nil || own.Magic {
			return nil
		}
		methods = append(methods, own)
	}
	return methods
}

// Spares reports a written call or a construction that stores nothing in a
// request table. A PHP function or method does not, unless it runs code of
// the project: it is given a function to call, loads or evaluates code,
// resumes or invokes some (see runner, loads, invokes and builds), or is a
// method, not final, of a class the project extends. A call given an array
// that shares a variable (`[&$x]`) may store anything; `f(...)` calls
// nothing. A
// callable of the project does not when it is the declaration that runs, or
// each of those that may (see Targets), declared once, takes no argument by
// reference, and its body spares the tables too and calls nothing of the
// project (see Summary.Spares). A
// construction runs the constructor and, once the object is dropped, the
// destructor the class finds: none is none. Anything not known that well may
// store request text anywhere. Code that runs without a call being written
// (a conversion to text, a magic method, an autoloader) is not counted.
func (e *Env) Spares(call syntax.Node) bool {
	return e.spares(call, e.reached(call), false)
}

// spares is Spares for what is known of the receiver. within tells that the
// call is written in a body being summarized: a callable of the project is
// then not looked into, so that a summary depends on no other.
func (e *Env) spares(call syntax.Node, followed uint8, within bool) bool {
	ix := e.types.Index
	if list := arguments(call); list != nil {
		if len(list.Args) == 1 {
			if _, callable := list.Args[0].(*syntax.VariadicPlaceholder); callable {
				// `f(...)` only makes a closure: nothing is called.
				return true
			}
		}
		if shares(list) {
			return false
		}
	}
	switch n := call.(type) {
	case *syntax.FuncCall:
		f := e.types.ResolveFunction(n)
		if f == nil || !f.Avail.In(e.types.PHP) {
			return false
		}
		if f.Builtin {
			name := strings.ToLower(f.FQN)
			return name != "create_function" && !runner(name) && !loads(name, n.Args) && !e.callbackGiven(n.Args, f.Params)
		}
		if within {
			return e.nestedSpares(f)
		}
		summary, ok := e.contract(strings.ToLower(f.FQN))
		return ok && summary.Spares && !referenced(f.Params) && len(ix.FunctionDecls(f.FQN, e.types.PHP)) == 1
	case *syntax.MethodCall:
		return e.sparing(e.targets(call, followed), n.Args, within)
	case *syntax.StaticCall:
		return e.sparing(e.targets(call, followed), n.Args, within)
	case *syntax.New:
		class, named := n.Class.(*syntax.Name)
		if !named || strings.EqualFold(class.Value, "static") {
			return false
		}
		fqn := e.named(class)
		declared := ix.Class(fqn, e.types.PHP)
		if declared == nil || ix.ClassCount(fqn) != 1 {
			return false
		}
		short := fqn[strings.LastIndexByte(fqn, '\\')+1:]
		for _, name := range []string{"__construct", "__destruct", short} {
			m := ix.FindMethod(fqn, name, e.types.PHP)
			switch {
			case m == nil:
			case name == short:
				// Before PHP 8 a method named like its class constructs it.
				if !e.types.PHP.AtLeast(phpversion.PHP80) {
					return false
				}
			case !m.Builtin && !e.single(call, m), !e.sparing([]*index.Method{m}, n.Args, within):
				return false
			}
		}
		return true
	}
	return false
}

// sparing reports that none of the declarations a call may run stores
// anything in a request table, and that one of them at least has a body.
func (e *Env) sparing(targets []*index.Method, list *syntax.ArgList, within bool) bool {
	bodies := 0
	for _, m := range targets {
		switch class := strings.ToLower(strings.TrimPrefix(m.Class, `\`)); {
		case referenced(m.Params):
			return false
		case m.Abstract:
			continue
		case !m.Builtin && within:
			return false
		case !m.Builtin:
			if summary, ok := e.contract(strings.ToLower(m.Class + "::" + m.Name)); !ok || !summary.Spares {
				return false
			}
		case class == "closure" || class == "generator" || class == "fiber" || strings.HasPrefix(class, "reflection") || e.callbackGiven(list, m.Params):
			return false
		case strings.EqualFold(m.Name, "__construct"):
			// The object made is of that very class.
			if builds(class, list) {
				return false
			}
		case invokes(class+"::"+strings.ToLower(m.Name), list) || !m.Final && e.extended(m.Class):
			return false
		}
		bodies++
	}
	return bodies > 0
}

// shares reports arguments that hold an array written with an element by
// reference (`[&$_GET]`): the callee then holds what the caller's variable
// holds, and writes it without a parameter by reference.
func shares(list *syntax.ArgList) bool {
	shared := false
	syntax.Inspect(list, func(n syntax.Node) bool {
		item, ok := n.(*syntax.ArrayItem)
		shared = shared || ok && item.ByRef
		return !shared
	})
	return shared
}

// arguments returns the argument list of a call or a construction, nil when
// none is written.
func arguments(call syntax.Node) *syntax.ArgList {
	switch n := call.(type) {
	case *syntax.FuncCall:
		return n.Args
	case *syntax.MethodCall:
		return n.Args
	case *syntax.StaticCall:
		return n.Args
	case *syntax.New:
		return n.Args
	}
	return nil
}

// extended reports a class of PHP that a class of the project extends: a
// method PHP declares may then call one the project replaces, as the
// insertion into a heap calls its comparison. The answer is kept per class.
func (e *Env) extended(class string) bool {
	at := strings.ToLower(strings.TrimPrefix(class, `\`))
	if known, ok := e.extendedBy[at]; ok {
		return known
	}
	if e.extendedBy == nil {
		e.extendedBy = map[string]bool{}
	}
	ix := e.types.Index
	e.extendedBy[at] = e.below(class, func(sub string) bool {
		declared := ix.Class(sub, e.types.PHP)
		return declared != nil && !strings.HasPrefix(declared.File, "stubs/")
	})
	return e.extendedBy[at]
}

// referenced reports a parameter taken by reference: the callee writes to
// what the caller hands it.
func referenced(params []index.Param) bool {
	return slices.ContainsFunc(params, func(p index.Param) bool { return p.ByRef })
}

// callbackGiven reports an argument that may be a function for PHP to call:
// one bound to a parameter that takes one, unless it names a PHP function
// that only computes a number, a boolean or text, or one whose parameter is
// not known.
func (e *Env) callbackGiven(list *syntax.ArgList, params []index.Param) bool {
	if list == nil {
		return false
	}
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		if !ok || at < 0 || arg.Unpack || invoked(params[at]) && !e.harmless(arg.Value) {
			return true
		}
	}
	return false
}

// sparesTables reports a callable whose body is the one every call runs and
// stores nothing in a request table: it names no request table, no
// `$GLOBALS` and no variable by a name computed at run time, evaluates and
// includes no code, and all it calls or constructs, in the default values
// of its parameters and in the functions it declares inline too, is PHP's own and
// spares them (see Spares), or a class of the project without constructor
// and destructor. A callable of the project it calls is not looked into: one
// summary then depends on no other, and none has to be computed again for
// this. The receivers of the calls in the inline functions are not followed.
func (e *Env) sparesTables(scope, body syntax.Node) bool {
	if body == nil || !unconditional(scope) {
		return false
	}
	spared := true
	look := func(n syntax.Node) bool {
		if !spared {
			// What is written after a refill does not undo it.
			return false
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.ClassLike:
			// Declared here, run where it is called or constructed.
			return false
		case *syntax.Variable:
			// A name computed at run time may be the one of a request table.
			spared = x.NameExpr == nil && !external(x.Name) && x.Name != "GLOBALS"
		case *syntax.Eval, *syntax.Include:
			spared = false
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
			followed := followLoose
			if syntax.EnclosingVariableScope(n) == scope {
				followed = e.reached(n)
			}
			spared = e.spares(n, followed, true)
		}
		return spared
	}
	// The value a parameter takes when none is given is computed by each
	// call that leaves it out.
	for _, p := range syntax.VariableScopeParams(scope) {
		if p.Default != nil {
			syntax.Inspect(p.Default, look)
		}
	}
	syntax.Inspect(body, look)
	return spared
}

// maxNested bounds how many calls deep nestedSpares looks.
const maxNested = 4

// nestedSpares looks into a function of the file that a body being
// summarized calls, so that a summary still depends on no other: the function
// must be declared once, unconditionally, take nothing by reference, and its
// own body must spare the tables, a few calls deep. A call that comes back
// to a body being looked into is taken to fill them.
func (e *Env) nestedSpares(f *index.Function) bool {
	declaration := e.localBody(strings.ToLower(strings.TrimPrefix(f.FQN, `\`)))
	if declaration == nil || referenced(f.Params) || len(e.types.Index.FunctionDecls(f.FQN, e.types.PHP)) != 1 || len(e.nested) >= maxNested || e.nested[declaration] {
		return false
	}
	if e.nested == nil {
		e.nested = map[syntax.Node]bool{}
	}
	e.nested[declaration] = true
	defer delete(e.nested, declaration)
	return e.sparesTables(declaration, syntax.VariableScopeBody(declaration))
}

// replaced reports that a class of the project extending the receiver's
// declares, inherits or uses another method of that name than the one
// resolved. A class extends the receiver's through any of its declarations
// and through the names class_alias() gives it; one declared more than once
// may be either declaration. The classes the project does not contain
// cannot be seen: a method meant to be replaced outside it is not known to
// be. The answer is kept per receiver class and method.
func (e *Env) replaced(receiver string, m *index.Method) bool {
	return e.overrides(receiver, m.Name, m.Class)
}

// overrides is replaced for a method name and the class that declares the
// method resolved: none when the receiver's class has no such method, which
// a class extending it then must not have either.
func (e *Env) overrides(receiver, method, declaring string) bool {
	at := [2]string{strings.ToLower(receiver), strings.ToLower(declaring + "::" + method)}
	if known, ok := e.overridden[at]; ok {
		return known
	}
	if e.overridden == nil {
		e.overridden = map[[2]string]bool{}
	}
	e.overridden[at] = e.override(receiver, method, declaring)
	return e.overridden[at]
}

func (e *Env) override(receiver, method, declaring string) bool {
	ix := e.types.Index
	return e.below(receiver, func(sub string) bool {
		// The summaries read here depend on what each subclass declares.
		e.contract(strings.ToLower(sub) + "::" + strings.ToLower(method))
		other := ix.FindMethod(sub, method, e.types.PHP)
		return ix.ClassCount(sub) != 1 || (other == nil) != (declaring == "") || other != nil && !strings.EqualFold(other.Class, declaring)
	})
}

// below visits each class of the project that extends the receiver's, through
// any of its declarations and through the names class_alias() gives it, until
// found reports one; it tells whether one was. A final class has none. The
// receiver may be written with a leading backslash; the index names the
// classes extending it without one.
func (e *Env) below(receiver string, found func(sub string) bool) bool {
	ix := e.types.Index
	if declared := ix.Class(receiver, e.types.PHP); declared != nil && declared.Final {
		return false
	}
	seen := map[string]bool{strings.ToLower(strings.TrimPrefix(receiver, `\`)): true}
	for queue := []string{receiver}; len(queue) > 0; queue = queue[1:] {
		// The names of the class: its own, its aliases and theirs.
		names := []string{queue[0]}
		for i := 0; i < len(names); i++ {
			for _, alias := range ix.AliasesOf(names[i]) {
				if !seen[strings.ToLower(alias)] {
					seen[strings.ToLower(alias)] = true
					names = append(names, alias)
				}
			}
		}
		for _, name := range names {
			for _, sub := range ix.ChildrenAll(name) {
				folded := strings.ToLower(sub)
				if seen[folded] {
					continue
				}
				seen[folded] = true
				if found(sub) {
					return true
				}
				queue = append(queue, sub)
			}
		}
	}
	return false
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

// parameters maps each argument of a call to the index of the parameter it
// binds, -1 when there is none. Arguments past the last parameter bind it
// when it is variadic. The result extends positions.
func parameters(list *syntax.ArgList, params []index.Param, positions []int) []int {
	position := 0
	for _, x := range list.Args {
		p := position
		if a, ok := x.(*syntax.Arg); ok && a.Name != nil {
			p = -1
			for j, param := range params {
				if param.Name == a.Name.Value {
					p = j
					break
				}
			}
		} else {
			position++
		}
		if last := len(params) - 1; p > last && last >= 0 && params[last].Variadic {
			p = last
		}
		if p >= len(params) {
			p = -1
		}
		positions = append(positions, p)
	}
	return positions
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
		if in := r.values[other.Value]; tainted(in) {
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

func (e *Env) summaryName(at syntax.Node) string { _, name, _ := e.declaration(at); return name }

func (e *Env) bind(at syntax.Node, values []Value) ([]Value, bool) {
	params, _, known := e.declaration(at)
	if !known {
		return nil, false
	}
	var list *syntax.ArgList
	switch n := at.(type) {
	case *syntax.FuncCall:
		list = n.Args
	case *syntax.MethodCall:
		list = n.Args
	case *syntax.StaticCall:
		list = n.Args
	}
	args := make([]Value, len(params))
	seen := make([]bool, len(params))
	position := 0
	if list == nil {
		return args, len(params) == 0
	}
	for i, x := range list.Args {
		a, ok := x.(*syntax.Arg)
		if !ok || a.Unpack || a.ByRef || i >= len(values) {
			return nil, false
		}
		p := position
		if a.Name != nil {
			p = -1
			for j, param := range params {
				if param.Name == a.Name.Value {
					p = j
					break
				}
			}
		} else {
			position++
		}
		if p < 0 || p >= len(params) || seen[p] || params[p].Variadic || params[p].ByRef {
			return nil, false
		}
		seen[p] = true
		args[p] = values[i]
	}
	for i, p := range params {
		if !seen[i] {
			if !p.Optional {
				return nil, false
			}
			args[i] = Value{Complete: true}
		}
	}
	return args, true
}

// accepted records what a call proves by returning. Since PHP 8 a PHP
// function throws a TypeError for false or null given to a parameter that
// only takes arrays, objects or callables (booleans and null still convert
// to scalar parameters). A function of the project, or a static method
// named by its class, `self` or `parent`, throws one for null given to a
// required parameter whose declared type does not admit it, as does a method
// called on an object when it is the one resolved (see dispatch), since PHP 7:
// before, the refusal is an error a handler can recover from, and a scalar
// type names a class. The proof holds where that error cannot be caught:
// outside any `try` with a `catch`.
func (e *Env) accepted(call syntax.Node, list *syntax.ArgList, builtin bool, followed uint8, f *frame) {
	switch n := call.(type) {
	case *syntax.FuncCall:
		// One of two declarations of a name may be the one in effect.
		if f := e.types.ResolveFunction(n); !builtin && (f == nil || len(e.types.Index.FunctionDecls(f.FQN, e.types.PHP)) != 1) {
			return
		}
	case *syntax.StaticCall:
		// A method bound late, or called on an object, may be one that
		// declares a wider type; one that is not visible from here leaves
		// the call to `__callStatic`.
		class, named := n.Class.(*syntax.Name)
		m := e.method(call)
		if builtin || !named || m == nil || strings.EqualFold(class.Value, "static") {
			return
		}
		if !e.visible(call, m) || e.types.Index.ClassCount(m.Class) != 1 {
			return
		}
	case *syntax.MethodCall:
		// A method called on an object is the one resolved when no class
		// can replace it; a null-safe call may not be made.
		if builtin || conditional(n) || !e.provable(call, list, f) {
			return
		}
		if m, exact := e.dispatch(call, followed); m == nil || !exact || !e.single(call, m) {
			return
		}
	}
	if builtin && !e.types.PHP.AtLeast(phpversion.PHP80) || !e.types.PHP.AtLeast(phpversion.PHP70) {
		return
	}
	params, _, _ := e.declaration(call)
	for p := call.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if guarded, ok := p.(*syntax.Try); ok && len(guarded.Catches) > 0 {
			return
		}
	}
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		if !ok || at < 0 || arg.Unpack || params[at].ByRef || params[at].Variadic {
			continue
		}
		// A later argument that assigns the variable replaces what this
		// parameter received.
		written := false
		for j, other := range list.Args {
			if rest, ok := other.(*syntax.Arg); ok && j != i {
				written = written || e.rebinds(rest.Value, arg.Value, f)
			}
		}
		if written {
			continue
		}
		if compound, nullable := compoundOnly(params[at].Type); builtin && compound {
			exclude(arg.Value, f, true, !nullable)
		} else if resource, nullable := resourceOnly(params[at]); builtin && resource {
			excludeValue(arg.Value, f, true, !nullable)
		} else if !builtin && !params[at].Optional && rejectsNull(params[at].Type) {
			exclude(arg.Value, f, false, true)
		}
	}
}

// provable reports a call with something to prove: an argument that is a
// variable not yet known to differ from null, bound to a required parameter
// that does not admit null. Only then is what runs worth resolving.
func (e *Env) provable(call syntax.Node, list *syntax.ArgList, f *frame) bool {
	params, _, _ := e.declaration(call)
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		if !ok || at < 0 || params[at].Optional || !rejectsNull(params[at].Type) {
			continue
		}
		if variable := guarded(arg.Value); variable != nil {
			if held, tracked := f.vars[variable.Name]; tracked && !held.NonNull {
				return true
			}
		}
	}
	return false
}

// rejectsNull reports a declared parameter type that does not admit null.
func rejectsNull(declared string) bool {
	if declared == "" || strings.HasPrefix(declared, "?") {
		return false
	}
	for declared != "" {
		var part string
		part, declared, _ = strings.Cut(declared, "|")
		switch strings.ToLower(strings.TrimPrefix(part, `\`)) {
		case "null", "mixed":
			return false
		}
	}
	return true
}

// resourceOnly reports a builtin parameter PHP checks as a resource (no
// declared type, documented `resource` in the stubs: `fread($stream)`),
// and whether it also takes null.
func resourceOnly(p index.Param) (bool, bool) {
	if p.Type != "" || p.DocType == "" {
		return false, false
	}
	resource, nullable := false, false
	for _, part := range strings.Split(p.DocType, "|") {
		switch strings.ToLower(strings.TrimPrefix(part, "?")) {
		case "resource":
			resource = true
		case "null":
			nullable = true
		default:
			return false, false
		}
		nullable = nullable || strings.HasPrefix(part, "?")
	}
	return resource, nullable
}

// compoundOnly reports a declared parameter type made only of arrays,
// iterables, callables, objects and classes, and whether it also takes null.
func compoundOnly(declared string) (bool, bool) {
	nullable := strings.HasPrefix(declared, "?")
	declared = strings.TrimPrefix(declared, "?")
	if declared == "" {
		return false, false
	}
	for declared != "" {
		var part string
		part, declared, _ = strings.Cut(declared, "|")
		switch strings.ToLower(strings.TrimPrefix(part, `\`)) {
		case "null":
			nullable = true
		case "array", "iterable", "callable", "object", "countable", "traversable", "iterator", "arrayaccess", "closure":
		default:
			return false, false
		}
	}
	return true, nullable
}

// localBody finds a unique declaration in this file. Producer syntax is kept
// in this environment only; exported summaries remain independent of the AST.
func (e *Env) localBody(name string) syntax.Node {
	if e.localBodies == nil {
		e.localBodies = map[string]syntax.Node{}
		seen := map[string]bool{}
		syntax.InspectFile(e.file, func(n syntax.Node) bool {
			key := ""
			switch n := n.(type) {
			case *syntax.Function:
				if n.Name != nil {
					key = n.Name.Value
					if ns := e.types.Names.Namespace(n.Span().Start); ns != "" {
						key = ns + `\` + key
					}
				}
			case *syntax.Method:
				if n.Name != nil {
					key = e.types.ClassFQN(syntax.EnclosingClass(n)) + "::" + n.Name.Value
				}
			}
			key = strings.ToLower(key)
			if key != "" {
				if seen[key] {
					e.localBodies[key] = nil
				} else {
					e.localBodies[key] = n
					seen[key] = true
				}
			}
			return true
		})
	}
	body := e.localBodies[name]
	if body != nil && unconditional(body) {
		return body
	}
	return nil
}
