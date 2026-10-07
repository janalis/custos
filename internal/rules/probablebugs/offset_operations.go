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
	sp := access.Span()
	if sp.Len() == 0 || access.Var == nil || access.Var.Span().Len() == 0 {
		return
	}
	if last := ctx.Src[sp.End-1]; last != ']' && last != '}' {
		return
	}

	// D1, D2
	// The occurrence's own type comes first: it honours narrowing by the
	// guarding condition (is_array(), !== false, instanceof…). The single
	// discovered value is only a fallback when the occurrence is untyped.
	// An unknown discovery result (unstable variable) empties S: no
	// fallback to the occurrence type.
	vals, known := util.PossibleValuesKnown(ctx.File, access.Var)
	if !known {
		return
	}
	t := ctx.TypeOf(access.Var)
	if t.IsUnknown() && len(vals) == 1 {
		t = ctx.TypeOf(vals[0])
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
	// D4
	var s []string
	for _, a := range set {
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
		if ctx.Index().Class(cls, ctx.PHP) == nil {
			continue
		}
		for _, m := range []string{"offsetGet", "offsetSet", "__get", "__set"} {
			meth := ctx.Index().FindMethod(cls, m, ctx.PHP)
			if meth == nil {
				continue
			}
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
	}
	if !supported { // D6
		ctx.ReportNode(access, "'"+ctx.Text(access.Var)+"' does not support offset access (types: "+strings.Join(s, "|")+").")
		return
	}

	// D7
	if len(allowed) == 0 || access.Dim == nil || offsetHas(allowed, "mixed") {
		return
	}
	it, ok := offsetNormalize(ctx.TypeOf(access.Dim))
	if !ok {
		return
	}
	var rest []string
	for _, a := range it {
		if a == "mixed" || a == "null" || offsetHas(allowed, a) {
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
	if len(rest) > 0 {
		ctx.ReportNode(access.Dim, "Index of type "+strings.Join(rest, "|")+" does not fit the accepted "+strings.Join(allowed, "|")+".")
	}
}

// offsetSubtypeOfAllowed reports whether the class type a extends or
// implements a class type of the allowed set.
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
