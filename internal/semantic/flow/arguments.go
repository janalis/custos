package flow

import (
	"strings"

	"custos/internal/php/syntax"
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
		fqn := e.types.Names.Class(class.Value, class.Span().Start)
		if m := e.types.Index.FindMethod(fqn, id.Value, e.types.PHP); m != nil && !m.Magic {
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
				case !builtin || name == "settype" && scalarType(list):
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
			f.write(r, variable.Name, next)
		}
	}
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

// accepted records what a PHP function proves by returning: since PHP 8 it
// throws a TypeError for false or null given to a parameter that only takes
// arrays, objects or callables (booleans and null still convert to scalar
// parameters). The proof holds where that error cannot be caught: outside
// any `try` with a `catch`.
func (e *Env) accepted(call syntax.Node, list *syntax.ArgList, f *frame) {
	if _, function := call.(*syntax.FuncCall); !function {
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
		if compound, nullable := compoundOnly(params[at].Type); compound {
			exclude(arg.Value, f, true, !nullable)
		}
	}
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
