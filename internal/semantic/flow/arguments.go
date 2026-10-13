package flow

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
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

func (e *Env) declaration(at syntax.Node) ([]index.Param, string, bool) {
	switch n := at.(type) {
	case *syntax.FuncCall:
		if f := e.types.ResolveFunction(n); f != nil {
			return f.Params, strings.ToLower(f.FQN), true
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
		untest(a.Value, f)
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
			next := unsettled(f.vars[variable.Name], Value{})
			if known {
				next = Value{}
				name, builtin, _ := e.callee(at)
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
			}
			if old := f.vars[variable.Name]; target != syntax.UnwrapParens(a.Value) {
				// Only the element can be bound: the callee may change or
				// create it, with what it writes, but the variable stays the
				// array it was, with the request data its other elements hold.
				next = unsettled(old, next)
				next.Complete = next.Complete || old.Complete
				next.NonFalse, next.NonNull = old.NonFalse, old.NonNull
				next.elements = old.elements.without(boundElement(a.Value))
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
	case *syntax.MethodCall, *syntax.StaticCall:
		sealed = false
		// A method that is not visible from here is not the one that runs.
		if m := e.method(call); m != nil && e.visible(call, m) {
			name, sealed = strings.ToLower(m.Class+"::"+m.Name), m.Final || m.Visibility == index.Private
			// The class of the object the method is looked up on at run
			// time is the receiver's, the enclosing one for `$this` and
			// `static`, or one that extends it.
			within := syntax.EnclosingClass(call)
			receiver := e.types.ClassFQN(within)
			if static, ok := n.(*syntax.StaticCall); ok {
				class, _ := static.Class.(*syntax.Name)
				sealed = sealed || !strings.EqualFold(class.Value, "static")
			} else if typed := e.types.TypeOf(n.(*syntax.MethodCall).Var).Classes(); len(typed) == 1 {
				receiver, within = typed[0], nil
			}
			// A trait method runs in classes that are not known here, and
			// a variable is the object inferred for it only while the flow
			// follows it: created here or received, and bound to nothing.
			known := within == nil || within.ClassKind != syntax.KindTrait
			if method, ok := n.(*syntax.MethodCall); ok && plainVariable(method.Var) != "" && plainVariable(method.Var) != "this" {
				_, shared := f.aliased[plainVariable(method.Var)]
				known = !shared && f.vars[plainVariable(method.Var)].Identity != 0
			}
			if !sealed && known {
				sealed = !e.replaced(receiver, m)
			}
		}
	}
	summary, ok := e.contract(name)
	written, described := summary.ByRef[positions[at]]
	if !ok || !described {
		return unsettled(old, Value{})
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
			if tainted(in) {
				v = merge(v, Value{Complete: true, Sources: in.Sources, Safe: in.Safe})
			}
		}
	}
	v.Safe |= written.Safe
	next := Value{}
	if tainted(v) {
		next = Value{Complete: true, Sources: v.Sources, Safe: v.Safe}
	}
	if !summary.Fixed || !sealed {
		return unsettled(old, next)
	}
	// An array of numbers the callee left is one for the caller.
	next.scalars = written.scalars
	return next
}

// replaced reports that a class of the project extending the receiver's
// declares, inherits or uses another method of that name than the one
// resolved. A class extends the receiver's through any of its declarations
// and through the names class_alias() gives it; one declared more than once
// may be either declaration. The classes the project does not contain
// cannot be seen: a method meant to be replaced outside it is not known to
// be. The answer is kept per receiver class and method.
func (e *Env) replaced(receiver string, m *index.Method) bool {
	at := [2]string{strings.ToLower(receiver), strings.ToLower(m.Class + "::" + m.Name)}
	if known, ok := e.overrides[at]; ok {
		return known
	}
	if e.overrides == nil {
		e.overrides = map[[2]string]bool{}
	}
	e.overrides[at] = e.overridden(receiver, m)
	return e.overrides[at]
}

func (e *Env) overridden(receiver string, m *index.Method) bool {
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
				folded := strings.ToLower(strings.TrimPrefix(sub, `\`))
				if seen[folded] {
					continue
				}
				seen[folded] = true
				// The summaries read here depend on what each subclass declares.
				e.contract(folded + "::" + strings.ToLower(m.Name))
				other := ix.FindMethod(sub, m.Name, e.types.PHP)
				if ix.ClassCount(sub) != 1 || other == nil || !strings.EqualFold(other.Class, m.Class) {
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
	next := unsettled(old, inputs)
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
// required parameter whose declared type does not admit it, in every
// version. The proof holds where that error cannot be caught: outside any
// `try` with a `catch`.
func (e *Env) accepted(call syntax.Node, list *syntax.ArgList, builtin bool, f *frame) {
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
	default:
		return
	}
	if builtin && !e.types.PHP.AtLeast(phpversion.PHP80) {
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
		} else if !builtin && !params[at].Optional && rejectsNull(params[at].Type) {
			exclude(arg.Value, f, false, true)
		}
	}
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
