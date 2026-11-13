package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// ReturnContract is the return type a function, method, closure or arrow
// function declares itself, never one inherited from a parent declaration.
type ReturnContract struct {
	// Native is the declared return type: `self` and `parent` name the
	// enclosing class, and so does `static` (a subset of it, which only
	// rejections may rely on). Unknown when none is declared, and for
	// relative types in a trait or outside a class.
	Native types.Type
	// Doc is the effective `@return` type (PHPStan and Psalm dialects
	// first), names resolved as Native's; DocText is its text. Unknown
	// when there is no tag, when it names a template of an enclosing
	// declaration, or when it is conditional.
	Doc     types.Type
	DocText string
}

// ReturnContract returns what function-like fn declares it returns.
func (e *Env) ReturnContract(fn syntax.Node) ReturnContract {
	var ret syntax.Expr
	switch f := fn.(type) {
	case *syntax.Function:
		ret = f.ReturnType
	case *syntax.Method:
		ret = f.ReturnType
	case *syntax.Closure:
		ret = f.ReturnType
	case *syntax.ArrowFunction:
		ret = f.ReturnType
	default:
		return ReturnContract{}
	}
	at := fn.Span().Start
	out := ReturnContract{Native: e.bindAtClass(types.FromNode(ret, e.resolver(at)), fn)}
	d := e.DocOf(fn)
	if d == nil {
		return out
	}
	text := d.EffectiveReturnType()
	if text == "" || strings.HasPrefix(text, "(") && strings.Contains(text, " is ") {
		return out
	}
	ds, templated := e.scopeDocs(fn), false
	doc := e.docType(text, func(w string) string {
		if ds != nil && ds.tpl[w] {
			templated = true
			return ""
		}
		if ds != nil {
			if def, ok := ds.aliases[w]; ok {
				return "=" + def
			}
		}
		return e.className(w, at)
	})
	if !templated {
		out.Doc, out.DocText = e.bindAtClass(doc, fn), text
	}
	return out
}
