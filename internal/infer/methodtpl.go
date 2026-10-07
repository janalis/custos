package infer

import (
	"strings"

	"custos/internal/index"
	"custos/internal/syntax"
	"custos/internal/types"
)

// Method-level templates: a function or method declaring `@template T`
// whose documented return type uses T (index.FuncTemplates) returns the
// type T takes from the call's arguments:
//
//	@param class-string<T> $class  ← `Foo::class` (or a class-string<Foo> value)
//	@param T $x                    ← the argument's type (`?T`: without null)
//	@param array<T> $xs / T[] / list<T> / iterable<T>  ← the element type
//	@param Box<T> $b               ← the argument's own Box<…> arguments
//
// A template bound by several parameters takes the union. When a template
// of the return type stays unbound, the call is typed as before (templates
// read as mixed).

// tplArg returns the argument expression bound to parameter i of params
// (positional, or named after it), nil when absent, unpacked or variadic.
func tplArg(args *syntax.ArgList, params []index.Param, i int) syntax.Expr {
	if args == nil || i >= len(params) || params[i].Variadic {
		return nil
	}
	for j, x := range args.Args {
		a, ok := x.(*syntax.Arg)
		if !ok || a.Unpack {
			return nil
		}
		if a.Name == nil {
			if j == i {
				return a.Value
			}
			continue
		}
		if strings.EqualFold(a.Name.Value, strings.TrimPrefix(params[i].Name, "$")) {
			return a.Value
		}
	}
	return nil
}

// tplReturn types a call to a function or method with method-level
// templates ft, given its parameters and declared return type; classB and
// cls bind the class templates of the declaring class (nil when unknown).
// ok is false when the call must be typed as before.
func (e *Env) tplReturn(ft *index.FuncTemplates, params []index.Param, args *syntax.ArgList, declared string,
	classB tplBindings, cls *index.Class) (types.Type, bool) {
	if ft == nil || ft.Return == "" {
		return types.Unknown, false
	}
	mb := e.bindCall(ft, params, args)
	if len(mb) == 0 {
		return types.Unknown, false
	}
	t, ok := substFuncTemplates(ft.Return, mb, classB, cls)
	if !ok {
		return types.Unknown, false
	}
	return e.withDeclared(t, declared)
}

// bindCall binds the method-level templates of ft from the call's
// arguments (see bindPattern); templates bound to unknown or mixed types
// are left out.
func (e *Env) bindCall(ft *index.FuncTemplates, params []index.Param, args *syntax.ArgList) tplBindings {
	if ft == nil || args == nil || len(ft.Params) == 0 {
		return nil
	}
	found := map[string][]types.Type{}
	for i, pat := range ft.Params {
		if pat == "" {
			continue
		}
		arg := tplArg(args, params, i)
		if arg == nil {
			continue
		}
		e.bindPattern(types.FromDoc(pat, nil), e.TypeOf(arg), arg, found)
	}
	mb := tplBindings{}
	for name, ts := range found {
		if u := types.Union(ts...); !u.IsUnknown() && !u.Has("mixed") {
			mb[name] = u
		}
	}
	return mb
}

// substFuncTemplates parses doc type text whose `\~~T` atoms are method
// templates bound by mb and `\~T` atoms class templates of cls bound by
// classB (else their bound, else mixed). ok is false when a method
// template is unbound or the result unknown.
func substFuncTemplates(text string, mb, classB tplBindings, cls *index.Class) (types.Type, bool) {
	allBound := true
	res := func(w string) string {
		name := strings.TrimPrefix(w, `\`)
		switch {
		case strings.HasPrefix(name, "~~"):
			if bt, ok := mb[name[2:]]; ok {
				if ds := bt.DocString(); len(ds) <= types.MaxDocTypeLen {
					return "=" + ds
				}
			}
			allBound = false
			return ""
		case strings.HasPrefix(name, "~"):
			name = name[1:]
			if bt, ok := classB[name]; ok && !bt.IsUnknown() && !bt.Has("mixed") {
				if ds := bt.DocString(); len(ds) <= types.MaxDocTypeLen {
					return "=" + ds
				}
			}
			if cls != nil {
				for _, tp := range cls.Templates {
					if tp.Name == name && tp.Bound != "" {
						return "=" + tp.Bound
					}
				}
			}
			return ""
		}
		return name
	}
	t := types.FromDoc(text, res)
	if !allBound || t.IsUnknown() {
		return types.Unknown, false
	}
	return t, true
}

// bindPattern matches the argument type at (of expression arg) against the
// parameter's documented type pat, recording template bindings in found.
func (e *Env) bindPattern(pat, at types.Type, arg syntax.Expr, found map[string][]types.Type) {
	if pat.IsUnknown() || at.IsUnknown() {
		return
	}
	tpl := func(a string) (string, bool) {
		if strings.HasPrefix(a, `\~~`) && !strings.HasSuffix(a, "[]") {
			return a[3:], true
		}
		return "", false
	}
	// The pattern's other members (`?T`: null) are not part of T.
	var direct string
	var others []string
	for _, a := range pat.Atoms() {
		if name, ok := tpl(a); ok {
			if direct != "" {
				return // `T|U`: ambiguous
			}
			direct = name
			continue
		}
		others = append(others, a)
	}
	if direct != "" {
		if len(others) > 0 && len(pat.Atoms()) > 1 {
			for _, o := range others {
				if strings.Contains(o, "~~") || pat.TypeArgs(o) != nil {
					return // `T|class-string<T>`, `T|T[]`: not handled
				}
			}
		}
		if rest := at.Without(others...); len(rest.Atoms()) > 0 {
			found[direct] = append(found[direct], rest)
		}
		return
	}
	if len(pat.Atoms()) != 1 {
		return
	}
	a := pat.Atoms()[0]
	switch {
	case a == "string":
		// class-string<T>
		if args := pat.TypeArgs(a); len(args) == 1 {
			if cs := args[0].Classes(); len(cs) == 1 && len(args[0].Atoms()) == 1 {
				if name, ok := tpl(cs[0]); ok {
					if c := types.ClassStringOf(at); !c.IsUnknown() {
						found[name] = append(found[name], c)
					}
				}
			}
		}
	case strings.HasPrefix(a, `\~~`) && strings.HasSuffix(a, "[]") && !strings.HasSuffix(a, "[][]"):
		// T[], array<T>, list<T>, array<K, T>
		if el := e.elemOf(at, arg); !el.IsUnknown() && !el.Has("mixed") {
			found[strings.TrimSuffix(a[3:], "[]")] = append(found[strings.TrimSuffix(a[3:], "[]")], el)
		}
	case a == "iterable":
		args := pat.TypeArgs(a)
		if len(args) == 0 {
			return
		}
		k, v := e.iterTypes(at)
		if v.IsUnknown() && at.IsArrayLike() {
			v = e.elemOf(at, arg)
		}
		e.bindSingle(args[len(args)-1], v, found)
		if len(args) == 2 {
			e.bindSingle(args[0], k, found)
		}
	case strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]"):
		// Box<T>: positional arguments of the same class.
		pargs := pat.TypeArgs(a)
		aargs := at.TypeArgs(a)
		if len(pargs) == 0 || len(pargs) != len(aargs) || len(at.Classes()) != 1 {
			return
		}
		for i := range pargs {
			e.bindSingle(pargs[i], aargs[i], found)
		}
	}
}

// bindSingle binds pattern p when it is a bare template `T`.
func (e *Env) bindSingle(p, t types.Type, found map[string][]types.Type) {
	if len(p.Atoms()) != 1 || t.IsUnknown() || t.Has("mixed") {
		return
	}
	a := p.Atoms()[0]
	if strings.HasPrefix(a, `\~~`) && !strings.HasSuffix(a, "[]") {
		found[a[3:]] = append(found[a[3:]], t)
	}
}

// elemOf is the element type of array type at (of expression arg): its
// `T[]` members, or the values of a sealed shape.
func (e *Env) elemOf(at types.Type, arg syntax.Expr) types.Type {
	if !at.IsArrayLike() {
		return types.Unknown
	}
	if el := iterElem(at); !el.IsUnknown() {
		return el
	}
	return e.shapeElem(at, arg)
}

// withDeclared returns the bound return type t when it agrees with the
// declared return type (unknown, mixed, a supertype of every member, or an
// intersection refining it: `@return T&Stub` on `: Stub`); else ok is false.
func (e *Env) withDeclared(t types.Type, declared string) (types.Type, bool) {
	d := types.FromDoc(declared, nil)
	if d.IsUnknown() || d.Has("mixed") || strictSuperset(t, d) {
		return t, true
	}
	for _, a := range t.Atoms() {
		switch {
		case d.Has(a):
		case (a == "true" || a == "false") && d.Has("bool"):
		case strings.HasSuffix(a, "[]") && (d.Has("array") || d.Has("iterable")):
		case strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]"):
			if d.Has("object") || e.subtypeOfAny(strings.TrimPrefix(a, `\`), d.Classes()) {
				continue
			}
			return types.Unknown, false
		default:
			return types.Unknown, false
		}
	}
	return t, true
}

func (e *Env) subtypeOfAny(cls string, parents []string) bool {
	for _, p := range parents {
		if e.Index.IsSubtype(cls, strings.TrimPrefix(p, `\`), e.PHP) {
			return true
		}
	}
	return false
}
