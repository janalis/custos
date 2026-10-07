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
	scope := syntax.EnclosingFuncLike(a)
	for { // D3c: closures and arrow functions take the class scope of their method
		switch scope.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			scope = syntax.EnclosingFuncLike(scope)
			continue
		}
		break
	}
	meth, ok := scope.(*syntax.Method)
	if !ok { // D3b
		ctx.Report(span, msg)
		return
	}
	owner, _ := meth.Parent().(*syntax.ClassLike) // methods only live in class-likes
	ownerFQN := ctx.Types().ClassFQN(owner)       // "" for an anonymous class
	var target string
	switch strings.ToLower(cls.Value) {
	case "self", "static":
		target = ownerFQN
	case "parent":
		target = ctx.Names().ParentFQN(owner)
	default:
		target = ctx.Names().Class(cls.Value, cls.Span().Start)
	}
	if target == "" { // `static`/`self` in an anonymous class, `parent` without one
		return
	}
	p := ctx.Index().FindProperty(target, prop.Name, ctx.PHP) // D3a
	if p == nil {
		return
	}
	// The method's own class, or for an anonymous class the traits it uses
	// (trait-imported members count as own, see spec Divergences).
	start := []string{ownerFQN}
	if ownerFQN == "" {
		start = start[:0]
		for _, m := range owner.Members {
			if tu, ok := m.(*syntax.TraitUse); ok {
				for _, t := range tu.Traits {
					start = append(start, ctx.Names().Class(t.Value, t.Span().Start))
				}
			}
		}
	}
	if dwspOwns(ctx, start, p.Class) {
		return
	}
	ctx.Report(span, msg)
}

func dwspSameClass(a, b string) bool {
	return strings.EqualFold(strings.TrimPrefix(a, `\`), strings.TrimPrefix(b, `\`))
}

// dwspOwns reports whether decl is one of the classes in queue or a trait
// they use, directly or through other traits.
func dwspOwns(ctx *analysis.Context, queue []string, decl string) bool {
	ix := ctx.Index()
	seen := map[string]bool{}
	for len(queue) > 0 {
		k := strings.ToLower(strings.TrimPrefix(queue[0], `\`))
		queue = queue[1:]
		if seen[k] {
			continue
		}
		seen[k] = true
		if dwspSameClass(k, decl) {
			return true
		}
		if c := ix.Class(k, ctx.PHP); c != nil {
			queue = append(queue, c.Traits...)
		}
	}
	return false
}
