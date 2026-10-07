package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// disallowWritingIntoStaticProperties flags assignments to static
// properties (by default only from outside the declaring class).
type disallowWritingIntoStaticProperties struct{}

func init() { register(disallowWritingIntoStaticProperties{}) }

func (disallowWritingIntoStaticProperties) ID() string { return "DisallowWritingIntoStaticProperties" }

func (disallowWritingIntoStaticProperties) Semantic() {}

func (disallowWritingIntoStaticProperties) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KAssign}
}

func (disallowWritingIntoStaticProperties) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Assign)
	fetch, ok := a.Var.(*syntax.StaticPropertyFetch) // D1
	if !ok || a.Value == nil || a.Span().Len() == 0 {
		return
	}
	prop, ok := fetch.Name.(*syntax.Variable)
	if !ok || prop.NameExpr != nil || prop.Name == "" { // E4
		return
	}
	span := syntax.Span{Start: a.Span().Start, End: a.Value.Span().End}
	if !ctx.Bool("ALLOW_WRITE_FROM_SOURCE_CLASS") { // D2
		ctx.Report(span, "Avoid modifying static properties.")
		return
	}
	cls, ok := fetch.Class.(*syntax.Name)            // D3
	if !ok || strings.EqualFold(cls.Value, "self") { // any letter case
		return
	}
	const msg = "Modify this static property only from the class that declares it."
	meth, ok := syntax.EnclosingFuncLike(a).(*syntax.Method)
	if !ok { // D3b
		ctx.Report(span, msg)
		return
	}
	owner, ok := meth.Parent().(*syntax.ClassLike)
	if !ok {
		return
	}
	env := ctx.Types()
	ownerFQN := env.ClassFQN(owner)
	var target string
	switch strings.ToLower(cls.Value) {
	case "self", "static":
		target = ownerFQN
	case "parent":
		if owner.ClassKind != syntax.KindInterface && len(owner.Extends) > 0 {
			target = ctx.Names().Class(owner.Extends[0].Value, owner.Span().Start)
		}
	default:
		target = ctx.Names().Class(cls.Value, cls.Span().Start)
	}
	if target == "" || ownerFQN == "" {
		return
	}
	p := ctx.Index().FindProperty(target, prop.Name, ctx.PHP) // D3a
	if p == nil {
		return
	}
	if dwspSameClass(p.Class, ownerFQN) || dwspUsesTrait(ctx, ownerFQN, p.Class) {
		return
	}
	ctx.Report(span, msg)
}

func dwspSameClass(a, b string) bool {
	return strings.EqualFold(strings.TrimPrefix(a, `\`), strings.TrimPrefix(b, `\`))
}

// dwspUsesTrait reports whether class (directly or through its traits) uses
// trait (see spec Divergences: trait-imported members count as own).
func dwspUsesTrait(ctx *analysis.Context, class, trait string) bool {
	ix := ctx.Index()
	seen := map[string]bool{}
	queue := []string{class}
	for len(queue) > 0 {
		k := strings.ToLower(strings.TrimPrefix(queue[0], `\`))
		queue = queue[1:]
		if seen[k] {
			continue
		}
		seen[k] = true
		c := ix.Class(k, ctx.PHP)
		if c == nil {
			continue
		}
		for _, t := range c.Traits {
			if dwspSameClass(t, trait) {
				return true
			}
			queue = append(queue, t)
		}
	}
	return false
}
