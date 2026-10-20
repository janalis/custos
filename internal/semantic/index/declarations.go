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
		File: x.f.Path, Span: n.Span(), Avail: x.avail(n.Attrs, d), CSPRNG: callsCSPRNG(n.Body), Deprecated: x.deprecated(n.Attrs, d), NoDiscard: x.noDiscard(n.Attrs),
	}
	fn.Return, fn.RetVer = x.returnType(n.ReturnType, n.Attrs, at)
	if d != nil {
		fn.DocReturn = voidDoc(x.docTypeStr(d.EffectiveReturnType(), at), n.Body)
		fn.CondReturn = x.condReturn(d, at)
		fn.Asserts = x.assertions(d, n.Params, false, at)
		fn.Tpl = x.funcTemplates(d, n.Params, fn.Asserts, at)
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
	name, known := x.declarationName(args[0].Value)
	if !known {
		return
	}
	// Unlike a class name, define's literal name keeps a leading separator.
	x.out.Constants = append(x.out.Constants, &Constant{FQN: name, Verbatim: strings.HasPrefix(name, `\`), Value: x.text(args[1].Value), ValueType: constArrayDoc(args[1].Value, nil), File: x.f.Path, Span: call.Span(), DeclarationFallback: x.declarationFallback(call)})
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
		value, known := x.declarationName(args[i].Value)
		if !known {
			return
		}
		fqns[i] = strings.TrimPrefix(value, `\`)
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

// declarationName evaluates a bounded string expression without invoking PHP.
// Dynamic constants, calls and late-bound class names establish no name.
func (x *extractor) declarationName(expression syntax.Expr) (string, bool) {
	if literal, ok := expression.(*syntax.Literal); ok {
		return syntax.LiteralStringValue(literal)
	}
	var out strings.Builder
	remaining := 256
	var appendValue func(syntax.Expr) bool
	appendValue = func(expression syntax.Expr) bool {
		remaining--
		if remaining < 0 {
			return false
		}
		switch value := expression.(type) {
		case *syntax.Paren:
			return appendValue(value.Expr)
		case *syntax.Binary:
			return value.Op.Kind == syntax.TDot && appendValue(value.Left) && appendValue(value.Right)
		case *syntax.Literal:
			text, known := syntax.LiteralStringValue(value)
			if !known {
				return false
			}
			out.WriteString(text)
		case *syntax.MagicConst:
			if value.Token.Kind != syntax.TNsC {
				return false
			}
			out.WriteString(x.r.Namespace(value.Span().Start))
		case *syntax.ClassConstFetch:
			member, plain := value.Name.(*syntax.Identifier)
			class, named := value.Class.(*syntax.Name)
			if !plain || !named || !strings.EqualFold(member.Value, "class") || names.IsSpecialClass(class.Value) || x.f.Version != 0 && x.f.Version < phpversion.PHP55 {
				return false
			}
			out.WriteString(x.r.Class(class.Value, class.Span().Start))
		default:
			return false
		}
		return true
	}
	if !appendValue(expression) {
		return "", false
	}
	return out.String(), true
}
