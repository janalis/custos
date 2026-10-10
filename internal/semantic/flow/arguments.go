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
func (e *Env) invalidateReferences(at syntax.Node, list *syntax.ArgList, f *frame) {
	if list == nil {
		return
	}
	params, _, known := e.declaration(at)
	position := 0
	for _, x := range list.Args {
		a, ok := x.(*syntax.Arg)
		if !ok {
			continue
		}
		p := position
		if a.Name != nil {
			p = -1
			for i, param := range params {
				if param.Name == a.Name.Value {
					p = i
					break
				}
			}
		} else {
			position++
		}
		if known && p >= 0 && p < len(params) && !params[p].ByRef {
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
			f.vars[variable.Name] = Value{}
		}
	}
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
