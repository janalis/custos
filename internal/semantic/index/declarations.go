package index

import (
	"strings"

	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/names"
)

func (x *extractor) function(n *syntax.Function) {
	d := x.doc(n)
	saved, savedClass := x.methodTpl, x.classTpl
	x.methodTpl, x.classTpl = nil, nil
	if d != nil {
		for _, t := range d.Templates() {
			if x.methodTpl == nil {
				x.methodTpl = map[string]bool{}
			}
			x.methodTpl[t] = true
		}
	}
	x.withTemplates(d, func() { x.functionBody(n, d) })
	x.methodTpl, x.classTpl = saved, savedClass
}

func (x *extractor) functionBody(n *syntax.Function, d *phpdoc.Doc) {
	at := n.Span().Start
	ns := x.r.Namespace(at)
	fqn := n.Name.Value
	if ns != "" {
		fqn = ns + `\` + fqn
	}
	fn := &Function{
		FQN: fqn, Params: x.params(n.Params, d, at), ByRef: n.ByRef,
		File: x.f.Path, Span: n.Span(), Avail: x.avail(n.Attrs, d), CSPRNG: callsCSPRNG(n.Body),
	}
	fn.Return, fn.RetVer = x.returnType(n.ReturnType, n.Attrs, at)
	if d != nil {
		fn.DocReturn = voidDoc(x.docTypeStr(d.EffectiveReturnType(), at), n.Body)
		fn.CondReturn = x.condReturn(d, at)
		fn.Asserts = x.assertions(d, n.Params, false, at)
		fn.Tpl = x.funcTemplates(d, n.Params, fn.Asserts, at)
		fn.Deprecated = d.Has("deprecated")
	}
	x.out.Functions = append(x.out.Functions, fn)
}

// declarationBuiltin identifies the global declaration builtins, allowing
// namespace fallback only when this file does not declare its preferred target.
func (x *extractor) declarationBuiltin(call *syntax.FuncCall, declared map[string]bool) string {
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return ""
	}
	fqn, fallback := x.r.Function(name.Value, name.Span().Start)
	if fallback != "" {
		if declared[key(fqn)] {
			return ""
		}
		fqn = fallback
	}
	switch strings.ToLower(fqn) {
	case "define":
		return "define"
	case "class_alias":
		return "class_alias"
	}
	return ""
}

// declarationArgs binds the two required and one optional builtin parameters.
// Unpacking and malformed calls cannot safely declare statically known symbols.
func (x *extractor) declarationArgs(call *syntax.FuncCall, params [3]string) ([3]*syntax.Arg, bool) {
	var bound [3]*syntax.Arg
	if call.Args == nil {
		return bound, false
	}
	named := false
	for i, expr := range call.Args.Args {
		a, ok := expr.(*syntax.Arg)
		if !ok || a.Unpack || a.ByRef || a.Value == nil {
			return bound, false
		}
		pos := i
		if a.Name != nil {
			if x.f.Version != 0 && x.f.Version < phpversion.PHP80 {
				return bound, false
			}
			named = true
			pos = -1
			for j, name := range params {
				if a.Name.Value == name {
					pos = j
					break
				}
			}
		} else if named {
			return bound, false
		}
		if pos < 0 || pos >= len(bound) || bound[pos] != nil {
			return bound, false
		}
		bound[pos] = a
	}
	return bound, bound[0] != nil && bound[1] != nil
}

// define() calls with a literal name declare global constants.
func (x *extractor) define(call *syntax.FuncCall) {
	args, ok := x.declarationArgs(call, [3]string{"constant_name", "value", "case_insensitive"})
	if !ok {
		return
	}
	lit, ok := syntax.UnwrapParens(args[0].Value).(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitString {
		return
	}
	x.out.Constants = append(x.out.Constants, &Constant{FQN: strings.TrimPrefix(unquote(lit.Raw), `\`), Value: x.text(args[1].Value), File: x.f.Path, Span: call.Span(), DeclarationFallback: x.declarationFallback(call)})
}

// classAlias records `class_alias(Original::class, 'Alias')` (class
// constants or string literals for both names).
func (x *extractor) classAlias(call *syntax.FuncCall) {
	args, ok := x.declarationArgs(call, [3]string{"class", "alias", "autoload"})
	if !ok {
		return
	}
	var fqns [2]string
	for i := range fqns {
		switch v := syntax.UnwrapParens(args[i].Value).(type) {
		case *syntax.ClassConstFetch:
			id, ok := v.Name.(*syntax.Identifier)
			nm, ok2 := v.Class.(*syntax.Name)
			if !ok || !ok2 || !strings.EqualFold(id.Value, "class") || names.IsSpecialClass(nm.Value) {
				return
			}
			fqns[i] = x.r.Class(nm.Value, nm.Span().Start)
		case *syntax.Literal:
			if v.LitKind != syntax.LitString {
				return
			}
			fqns[i] = strings.TrimPrefix(unquote(v.Raw), `\`)
		default:
			return
		}
		if fqns[i] == "" {
			return
		}
	}
	if fallback := x.declarationFallback(call); fallback != "" {
		if x.out.ClassAliasFallbacks == nil {
			x.out.ClassAliasFallbacks = map[int]string{}
		}
		x.out.ClassAliasFallbacks[len(x.out.ClassAliases)] = fallback
	}
	x.out.ClassAliases = append(x.out.ClassAliases, [2]string{fqns[1], fqns[0]})
}

// declarationFallback retains the preferred namespace candidate so a project
// function discovered later can suppress and subsequently restore the symbol.
func (x *extractor) declarationFallback(call *syntax.FuncCall) string {
	name := call.Name.(*syntax.Name) // declarationBuiltin accepted a named call
	fqn, fallback := x.r.Function(name.Value, name.Span().Start)
	if fallback != "" {
		return key(fqn)
	}
	return ""
}
