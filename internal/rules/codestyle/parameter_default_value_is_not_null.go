package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/index"
	"custos/internal/syntax"
)

// parameterDefaultValueIsNotNull flags optional parameters whose default is
// a non-null sentinel although `null` could be used.
type parameterDefaultValueIsNotNull struct{}

func init() { register(parameterDefaultValueIsNotNull{}) }

func (parameterDefaultValueIsNotNull) ID() string { return "ParameterDefaultValueIsNotNull" }

func (parameterDefaultValueIsNotNull) Semantic() {}

func (parameterDefaultValueIsNotNull) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure, syntax.KArrowFunction}
}

func (parameterDefaultValueIsNotNull) Check(ctx *analysis.Context, n syntax.Node) {
	params := syntax.FuncLikeParams(n)
	if len(params) == 0 {
		return
	}
	var candidates []*syntax.Param
	for _, p := range params {
		if p.Default == nil || p.Default.Span().Len() == 0 || syntax.IsNullConst(syntax.UnwrapParens(p.Default)) { // D1
			continue
		}
		if p.Type != nil && !pdvTypeHasNull(ctx, p.Type) { // D2 / E1
			continue
		}
		candidates = append(candidates, p)
	}
	if len(candidates) == 0 {
		return
	}
	if m, ok := n.(*syntax.Method); ok && pdvOverridesParent(ctx, m) { // E2 / E3
		return
	}
	for _, p := range candidates {
		ctx.Report(syntax.Span{Start: p.Span().Start, End: p.Default.Span().End}, "Prefer null as the default value for this parameter.")
	}
}

// pdvTypeHasNull reports whether a declared type includes null.
func pdvTypeHasNull(ctx *analysis.Context, t syntax.Expr) bool {
	switch t := t.(type) {
	case *syntax.NullableType:
		return true
	case *syntax.UnionType:
		for _, x := range t.Types {
			if pdvTypeHasNull(ctx, x) {
				return true
			}
		}
		return false
	case *syntax.Paren:
		return false
	}
	return strings.EqualFold(strings.TrimSpace(ctx.Text(t)), "null")
}

// pdvOverridesParent reports whether m's class extends a resolvable
// parent class whose hierarchy (excluding interfaces) has a non-private
// method of the same name.
func pdvOverridesParent(ctx *analysis.Context, m *syntax.Method) bool {
	cls, ok := m.Parent().(*syntax.ClassLike)
	if !ok || cls.ClassKind == syntax.KindInterface || len(cls.Extends) == 0 || m.Name == nil {
		return false
	}
	parent := ctx.Names().Class(cls.Extends[0].Value, cls.Span().Start)
	ix := ctx.Index()
	lname := strings.ToLower(m.Name.Value)
	for _, c := range ix.Ancestors(parent, ctx.PHP) {
		if c.Kind == syntax.KindInterface {
			continue
		}
		if pm, ok := c.Methods[lname]; ok && pm.Visibility != index.Private {
			return true
		}
	}
	return false
}
