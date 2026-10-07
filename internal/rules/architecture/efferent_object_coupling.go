package architecture

import (
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/names"
	"custos/internal/syntax"
)

// efferentObjectCoupling reports class-likes referring to many distinct
// classes.
type efferentObjectCoupling struct{}

func init() { register(efferentObjectCoupling{}) }

func (efferentObjectCoupling) ID() string { return "EfferentObjectCoupling" }

func (efferentObjectCoupling) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClassLike}
}

func (efferentObjectCoupling) Check(ctx *analysis.Context, n syntax.Node) {
	cl := n.(*syntax.ClassLike)
	if cl.Name == nil { // E2
		return
	}
	seen := map[string]struct{}{}
	visit := func(x syntax.Node) bool {
		if name, ok := x.(*syntax.Name); ok {
			if isClassReference(name) {
				// class names are case-insensitive (D3)
				seen[strings.ToLower(ctx.Names().Class(name.Value, name.Span().Start))] = struct{}{}
			}
		}
		return true
	}
	ownAttrs := map[syntax.Node]bool{}
	for _, g := range cl.Attrs {
		ownAttrs[g] = true
	}
	syntax.Children(cl, func(c syntax.Node) {
		if !ownAttrs[c] {
			syntax.Inspect(c, visit)
		}
	})
	if limit := ctx.Int("optionCouplingLimit"); len(seen) >= limit { // D4
		ctx.ReportNode(cl.Name, "Depends on "+strconv.Itoa(len(seen))+" distinct classes; consider splitting it up.")
	}
}

// isClassReference reports whether a name node is used as a class name
// (D2), excluding builtin type keywords and self/static/parent (E3).
func isClassReference(n *syntax.Name) bool {
	if names.IsSpecialClass(n.Value) {
		return false
	}
	isType := false
	switch p := n.Parent().(type) {
	case *syntax.ClassLike, *syntax.TraitUse, *syntax.Catch, *syntax.Attribute:
		if a, ok := p.(*syntax.Attribute); ok && a.Name != n {
			return false
		}
		if c, ok := p.(*syntax.ClassLike); ok && c.EnumType == syntax.Expr(n) {
			return false
		}
		return true
	case *syntax.New:
		return p.Class == syntax.Expr(n)
	case *syntax.StaticCall:
		return p.Class == syntax.Expr(n)
	case *syntax.ClassConstFetch:
		return p.Class == syntax.Expr(n)
	case *syntax.StaticPropertyFetch:
		return p.Class == syntax.Expr(n)
	case *syntax.Instanceof:
		return p.Class == syntax.Expr(n)
	case *syntax.Param:
		isType = p.Type == syntax.Expr(n)
	case *syntax.Property:
		isType = p.Type == syntax.Expr(n)
	case *syntax.ClassConst:
		isType = p.Type == syntax.Expr(n)
	case *syntax.Method:
		isType = p.ReturnType == syntax.Expr(n)
	case *syntax.Function:
		isType = p.ReturnType == syntax.Expr(n)
	case *syntax.Closure:
		isType = p.ReturnType == syntax.Expr(n)
	case *syntax.ArrowFunction:
		isType = p.ReturnType == syntax.Expr(n)
	case *syntax.NullableType, *syntax.UnionType, *syntax.IntersectionType:
		isType = true
	}
	return isType && !names.IsBuiltinType(n.Value)
}
