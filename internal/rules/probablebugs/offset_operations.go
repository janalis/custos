package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
	"custos/internal/types"
)

// offsetOperations reports offset access on containers that do not support
// it and index types the container does not accept.
type offsetOperations struct{}

func init() { register(offsetOperations{}) }

// Semantic marks the rule as needing the project symbol index.
func (offsetOperations) Semantic() {}

func (offsetOperations) ID() string { return "OffsetOperations" }

func (offsetOperations) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KArrayDimFetch}
}

// offsetNormalize normalises inferred atoms (D1); ok is false when t is
// unknown.
func offsetNormalize(t types.Type) ([]string, bool) {
	if t.IsUnknown() {
		return nil, false
	}
	var out []string
	add := func(a string) {
		for _, x := range out {
			if x == a {
				return
			}
		}
		out = append(out, a)
	}
	for _, a := range t.Atoms() {
		switch {
		case a == "true" || a == "false" || a == "bool":
			add("bool")
		case strings.HasSuffix(a, "[]"):
			add("array")
		case strings.EqualFold(a, `\Closure`):
			add("callable")
		case a == "iterable": // custos: array|\Traversable
			add("array")
			add(`\Traversable`)
		case a == "self" || a == "static":
		default:
			add(a)
		}
	}
	return out, true
}

func offsetHas(set []string, a string) bool {
	for _, x := range set {
		if x == a {
			return true
		}
	}
	return false
}

func offsetSetIs(set []string, want ...string) bool {
	if len(set) != len(want) {
		return false
	}
	for _, w := range want {
		if !offsetHas(set, w) {
			return false
		}
	}
	return true
}

func (offsetOperations) Check(ctx *analysis.Context, n syntax.Node) {
	access := n.(*syntax.ArrayDimFetch)
	sp := access.Span() // never empty: it includes the opening bracket
	if last := ctx.Src[sp.End-1]; last != ']' && last != '}' {
		return
	}

	// D1, D2
	// The occurrence's own type comes first: it honours narrowing by the
	// guarding condition (is_array(), !== false, instanceof…). The single
	// discovered value is only a fallback when the occurrence is untyped.
	// An unknown discovery result (unstable variable) empties S: no
	// fallback to the occurrence type.
	if _, known := util.PossibleValuesKnown(ctx.File, access.Var); !known {
		return
	}
	t := ctx.TypeOf(access.Var)
	if t.IsUnknown() {
		// The fallback needs every value: a parameter default, a foreach
		// variable or an unresolved source is only one of the values.
		if all, complete := util.PossibleValuesComplete(ctx.File, access.Var); complete && len(all) == 1 {
			t = ctx.TypeOf(all[0])
		}
	}
	set, ok := offsetNormalize(t)
	if !ok || len(set) == 0 {
		return
	}
	// D3
	if offsetHas(set, "mixed") || offsetSetIs(set, "string", "int") || offsetSetIs(set, "string", "bool") ||
		offsetSetIs(set, "array", "bool") {
		return
	}
	// D4. A boolean next to an array or string is a failure marker
	// (`array|false` from a lookup), handled like null (custos refinement).
	falsy := offsetHas(set, "array") || offsetHas(set, "string") || offsetHas(set, "callable")
	for _, a := range set {
		// custos: also next to a class (`simplexml_load_string()` is
		// SimpleXMLElement|false).
		falsy = falsy || strings.HasPrefix(a, `\`)
	}
	var s []string
	for _, a := range set {
		if a == "bool" && falsy {
			continue
		}
		switch a {
		case "callable":
			for _, x := range []string{"array", "string"} {
				if !offsetHas(s, x) {
					s = append(s, x)
				}
			}
		case "null", "void", "object":
		default:
			if !offsetHas(s, a) {
				s = append(s, a)
			}
		}
	}
	if len(s) == 0 {
		return
	}
	// custos: a PHPDoc union that admits arrays or strings next to other
	// scalars (`@return array|int|string|float|bool`) is loose
	// documentation unless the native types confirm the scalar members.
	if (offsetHas(s, "array") || offsetHas(s, "string")) && ctx.Types().Native().TypeOf(access.Var).IsUnknown() {
		for _, a := range s {
			if a != "array" && a != "string" && !strings.HasPrefix(a, `\`) {
				return
			}
		}
	}

	// D5
	supported := false
	var allowed []string
	allow := func(as ...string) {
		for _, a := range as {
			if !offsetHas(allowed, a) {
				allowed = append(allowed, a)
			}
		}
	}
	for _, a := range s {
		if a == "array" || a == "string" {
			supported = true
			allow("string", "int")
			continue
		}
		if !strings.HasPrefix(a, `\`) {
			supported = false
			break
		}
		cls := strings.TrimPrefix(a, `\`)
		c := ctx.Index().Class(cls, ctx.PHP)
		if c == nil {
			return // D1: an unresolvable class empties S
		}
		if offsetNativeAccess(ctx, cls) {
			supported = true
			allow("string", "int")
			continue
		}
		classSupports := false
		for _, m := range []string{"offsetGet", "offsetSet", "__get", "__set"} {
			meth := ctx.Index().FindMethod(cls, m, ctx.PHP)
			if meth == nil {
				continue
			}
			classSupports = true
			supported = true
			if strings.HasPrefix(m, "__") {
				allow("string", "int")
				continue
			}
			if len(meth.Params) > 0 {
				p := meth.Params[0]
				pt := types.FromDoc(p.Type, nil)
				if pt.IsUnknown() {
					pt = types.FromDoc(p.DocType, nil)
				}
				if norm, ok := offsetNormalize(pt); ok {
					for _, x := range norm {
						// unknown parts (e.g. stub template names) are dropped
						if !strings.HasPrefix(x, `\`) || ctx.Index().Class(strings.TrimPrefix(x, `\`), ctx.PHP) != nil {
							allow(x)
						}
					}
				}
			}
		}
		// custos: implementations of an interface or of a non-final
		// abstract class may be ArrayAccess (Laravel's `$app['config']`).
		if !classSupports && (c.Kind == syntax.KindInterface || c.Abstract && !c.Final) {
			return
		}
	}
	if !supported { // D6
		// custos: PHP returns null without an error for these reads when
		// the value is a scalar; with an array or string member the union
		// is a deliberate "array or not" lookup (`$counts[$id] ?? 0`).
		if quietOffsetRead(access) && (offsetHas(s, "array") || offsetHas(s, "string")) {
			return
		}
		ctx.ReportNode(access, "'"+ctx.Text(access.Var)+"' does not support offset access (types: "+strings.Join(s, "|")+").")
		return
	}

	// D7
	if len(allowed) == 0 || access.Dim == nil || offsetHas(allowed, "mixed") {
		return
	}
	it, ok := offsetNormalize(ctx.TypeOf(access.Dim))
	if !ok || offsetHas(it, "mixed") { // custos: mixed admits any key
		return
	}
	var rest []string
	for _, a := range it {
		// PHP casts bool keys to int (custos refinement).
		if a == "null" || a == "void" || offsetHas(allowed, a) || a == "bool" && offsetHas(allowed, "int") {
			continue
		}
		if offsetHas(allowed, "object") && strings.HasPrefix(a, `\`) {
			continue
		}
		if offsetSubtypeOfAllowed(ctx, a, allowed) {
			continue
		}
		rest = append(rest, a)
	}
	// custos: an index documented as a union admitting an accepted type
	// (`@return string[]|string` by input) is loose documentation unless
	// the native types confirm the other members.
	accepted := len(it) - len(rest)
	for _, a := range it {
		if a == "null" || a == "void" {
			accepted--
		}
	}
	if len(rest) > 0 && accepted > 0 && ctx.Types().Native().TypeOf(access.Dim).IsUnknown() {
		return
	}
	if len(rest) > 0 {
		ctx.ReportNode(access.Dim, "Index of type "+strings.Join(rest, "|")+" does not fit the accepted "+strings.Join(allowed, "|")+".")
	}
}

// offsetSubtypeOfAllowed reports whether the class type a extends or
// implements a class type of the allowed set.
// offsetNativeAccess reports whether cls is (or extends) a builtin class
// whose objects support `$o[…]` natively without declaring offsetGet() in
// the stubs (custos refinement).
func offsetNativeAccess(ctx *analysis.Context, cls string) bool {
	for _, b := range []string{"DOMNodeList", "DOMNamedNodeMap", "ResourceBundle", `Dom\NodeList`,
		`Dom\NamedNodeMap`, `Dom\HTMLCollection`, `FFI\CData`} {
		if strings.EqualFold(cls, b) || ctx.Index().IsSubtype(cls, b, ctx.PHP) {
			return true
		}
	}
	return false
}

func offsetSubtypeOfAllowed(ctx *analysis.Context, a string, allowed []string) bool {
	if !strings.HasPrefix(a, `\`) {
		return false
	}
	for _, x := range allowed {
		if strings.HasPrefix(x, `\`) && ctx.Index().IsSubtype(a, x, ctx.PHP) {
			return true
		}
	}
	return false
}

// quietOffsetRead reports whether access (possibly the base of a longer
// access chain) is read inside isset(), empty() or on the left of `??`.
func quietOffsetRead(access *syntax.ArrayDimFetch) bool {
	var n syntax.Node = access
	for {
		p, child := util.ParentSkipParens(n)
		switch x := p.(type) {
		case *syntax.ArrayDimFetch:
			if x.Var != child {
				return false
			}
			n = x
			continue
		case *syntax.PropertyFetch:
			if x.Var != child {
				return false
			}
			n = x
			continue
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			return x.Op.Kind == syntax.TCoalesce && x.Left == child
		}
		return false
	}
}
