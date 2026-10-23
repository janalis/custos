package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// callbackSlot is the parameter position of the callback each builtin with a
// fixed callback contract takes.
var callbackSlot = map[string]int{
	"array_map": 0, "array_filter": 1, "array_walk": 1, "array_reduce": 1,
	"usort": 1, "uasort": 1, "uksort": 1, "preg_replace_callback": 1,
}

// contextParamType types parameter i of closure or arrow function fn when fn
// is written directly as the callback of a builtin that decides what it
// passes: the elements (or keys) of the array argument for array_map(),
// array_filter() (by its mode), array_walk(), array_reduce() and the usort()
// family; otherwise what the called function or method documents its
// callback receives (see documentedCallbackParam). Unknown otherwise.
func (e *Env) contextParamType(fn syntax.Node, i int) types.Type {
	arg, ok := fn.Parent().(*syntax.Arg)
	if !ok {
		return types.Unknown
	}
	if t := e.builtinCallbackParam(fn, arg, i); !t.IsUnknown() {
		return t
	}
	return e.documentedCallbackParam(fn, arg.Parent().(*syntax.ArgList), i)
}

// builtinCallbackParam is contextParamType for the builtins with a fixed
// callback contract (see callbackSlot).
func (e *Env) builtinCallbackParam(fn syntax.Node, arg *syntax.Arg, i int) types.Type {
	call, ok := arg.Parent().Parent().(*syntax.FuncCall)
	if !ok {
		return types.Unknown
	}
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return types.Unknown
	}
	fqn, fb := e.Names.Function(name.Value, name.Span().Start)
	if fb != "" && e.Index.Function(fqn, e.PHP) == nil {
		if e.annotating {
			e.deps = append(e.deps, fqn)
		}
		fqn = fb
	}
	builtin := strings.ToLower(fqn)
	slot, ok := callbackSlot[builtin]
	function := e.Index.Function(fqn, e.PHP)
	if !ok || function == nil || !function.Builtin || !builtinBindingsKnown(call.Args, function.Params, e.PHP, false) {
		return types.Unknown
	}
	at := func(j int) syntax.Expr {
		x, state := bindArgument(call.Args, function.Params, j, e.PHP)
		if state != bindingBound {
			return nil
		}
		return x
	}
	if at(slot) != fn {
		return types.Unknown
	}
	switch builtin {
	case "array_map":
		// array_map(cb, $a, $b, …): parameter i receives the elements of
		// the (i+1)th array, padded with null when there are several.
		inputs := len(call.Args.Args) - 1
		if inputs == 1 {
			if i == 0 {
				return e.foreachElem(at(1))
			}
			return types.Unknown
		}
		if i >= inputs || namedArgs(call.Args) {
			return types.Unknown
		}
		if el := e.foreachElem(call.Args.Args[i+1].(*syntax.Arg).Value); !el.IsUnknown() {
			return types.Union(el, types.Null)
		}
	case "array_filter":
		arr := at(0)
		mode := ""
		if m := at(2); m != nil {
			c, ok := syntax.UnwrapParens(m).(*syntax.ConstFetch)
			if !ok {
				return types.Unknown
			}
			mode = strings.ToUpper(strings.TrimPrefix(c.Name.Value, `\`))
		}
		switch {
		case i == 0 && mode == "":
			return e.foreachElem(arr)
		case i == 0 && mode == "ARRAY_FILTER_USE_KEY":
			return e.foreachKey(arr)
		case mode == "ARRAY_FILTER_USE_BOTH":
			if i == 0 {
				return e.foreachElem(arr)
			}
			if i == 1 {
				return e.foreachKey(arr)
			}
		}
	case "array_walk":
		switch i {
		case 0:
			return e.foreachElem(at(0))
		case 1:
			return e.foreachKey(at(0))
		}
	case "array_reduce":
		if i == 1 {
			return e.foreachElem(at(0))
		}
	case "usort", "uasort":
		if i < 2 {
			return e.foreachElem(at(0))
		}
	case "uksort":
		if i < 2 {
			return e.foreachKey(at(0))
		}
	case "preg_replace_callback":
		// The matches, as preg_match() stores them by its flags.
		flags := "0"
		if f := at(5); f != nil {
			flags = flagsValue(f)
		}
		if el := matchesElement(flags, false); i == 0 && el != "" {
			return types.FromDoc(el+"[]", nil)
		}
	}
	return types.Unknown
}

// namedArgs reports a list with a named argument.
func namedArgs(args *syntax.ArgList) bool {
	for _, x := range args.Args {
		if x.(*syntax.Arg).Name != nil {
			return true
		}
	}
	return false
}

// documentedCallbackParam types parameter i of closure or arrow function
// fn, an argument of argument list al, from the callable signature the
// called function or method documents for the parameter fn binds
// (`@param Closure(T, int): bool $p`): the receiver binds the templates of
// the method's class, the other arguments its own templates. Unknown when
// the callee or the signature is unknown, or callees disagree.
func (e *Env) documentedCallbackParam(fn syntax.Node, al *syntax.ArgList, i int) types.Type {
	switch call := al.Parent().(type) {
	case *syntax.FuncCall:
		if f := e.ResolveFunction(call); f != nil {
			return e.signatureParam(f.Params, f.Tpl, nil, nil, al, fn, i)
		}
	case *syntax.MethodCall:
		id, ok := call.Name.(*syntax.Identifier)
		if !ok {
			return types.Unknown
		}
		recv := e.chainReceiver(call.Var, call.NullSafe)
		var t types.Type
		for n, cls := range e.memberClasses(recv, func(c string) bool { return e.Index.FindMethod(c, id.Value, e.PHP) != nil }) {
			c := strings.TrimPrefix(cls, `\`)
			ct := e.methodCallbackParam(c, id.Value, c, recv.TypeArgs(cls), al, fn, i)
			if ct.IsUnknown() || n > 0 && ct.DocString() != t.DocString() {
				return types.Unknown
			}
			t = ct
		}
		return t
	case *syntax.StaticCall:
		id, ok := call.Name.(*syntax.Identifier)
		if !ok {
			return types.Unknown
		}
		if cls := e.classRef(call.Class); cls != "" {
			return e.methodCallbackParam(cls, id.Value, cls, nil, al, fn, i)
		}
	}
	return types.Unknown
}

// methodCallbackParam is documentedCallbackParam for method name of class
// cls, called on origin (with generic arguments targs).
func (e *Env) methodCallbackParam(cls, name, origin string, targs []types.Type, al *syntax.ArgList, fn syntax.Node, i int) types.Type {
	m := e.Index.FindMethod(cls, name, e.PHP)
	if m == nil {
		return types.Unknown
	}
	var classB tplBindings
	c := e.Index.Class(m.Class, e.PHP)
	if c != nil && len(c.Templates) > 0 {
		classB = e.genBindings(origin, targs)[strings.ToLower(strings.TrimPrefix(m.Class, `\`))]
	}
	t := e.signatureParam(m.Params, m.Tpl, classB, c, al, fn, i)
	return e.bindMember(t, m.Class, m.TypeClass, origin)
}

// signatureParam is the type the documented callable signature of the
// parameter of params that callback fn binds gives its parameter i, with
// the templates of ft bound by the other arguments of al and those of
// class cls by classB (see substTemplateText). Native typing never asks:
// it keeps a parameter's declaration.
func (e *Env) signatureParam(params []index.Param, ft *index.FuncTemplates, classB tplBindings, cls *index.Class, al *syntax.ArgList, fn syntax.Node, i int) types.Type {
	for j, p := range params {
		if x, state := e.boundArgument(al, params, j); state != bindingBound || x != fn {
			continue
		}
		text := e.resolvedDoc(p.DocType)
		switch {
		case ft != nil && j < len(ft.Params) && ft.Params[j] != "":
			text = ft.Params[j]
		case p.Gen != "":
			text = p.Gen
		}
		t, _ := substTemplateText(text, e.bindCallSkipping(ft, params, al, j), classB, cls)
		return types.CallableParam(t, i)
	}
	return types.Unknown
}
