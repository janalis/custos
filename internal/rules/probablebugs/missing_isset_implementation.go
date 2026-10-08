package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// missingIssetImplementation reports isset()/empty() on undeclared
// properties of classes without __isset().
type missingIssetImplementation struct{}

func init() { register(missingIssetImplementation{}) }

func (missingIssetImplementation) ID() string { return "MissingIssetImplementation" }

func (missingIssetImplementation) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KIsset, syntax.KEmpty}
}

// Semantic marks the rule as needing the project index.
func (missingIssetImplementation) Semantic() {}

func issetExempt(fqn string) bool {
	switch strings.ToLower(strings.TrimPrefix(fqn, `\`)) {
	case "simplexmlelement", "stdclass", "domdocument":
		return true
	}
	return false
}

func (missingIssetImplementation) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.Isset:
		for _, a := range x.Vars {
			checkMissingIsset(ctx, a)
		}
	case *syntax.Empty:
		checkMissingIsset(ctx, x.Expr)
	}
}

func checkMissingIsset(ctx *analysis.Context, a syntax.Expr) {
	pf, ok := a.(*syntax.PropertyFetch) // D1
	if !ok || pf.NullSafe {
		return
	}
	id, ok := pf.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	if ctx.Text(pf.Var) == "$this" { // D3
		return
	}
	ix := ctx.Index()
	typ := ctx.TypeOf(pf.Var)
	types := typ.Classes()
	for _, t := range types { // D2
		if ix.FindProperty(strings.TrimPrefix(t, `\`), id.Value, ctx.PHP) != nil {
			return
		}
	}
	// custos: the check is "always false" only when every possible value
	// is an object of a concrete class without __isset() and without
	// dynamic properties; a non-class member (object, mixed, array…), an
	// unresolvable class, an interface or abstract class (implementations
	// may declare __isset()) or an #[\AllowDynamicProperties] hierarchy
	// makes the outcome undecidable.
	for _, a := range typ.Atoms() {
		if a != "null" && !strings.HasPrefix(a, `\`) {
			return
		}
	}
	report := ""
	for _, t := range types { // D4
		if issetExempt(t) {
			return
		}
		cls := strings.TrimPrefix(t, `\`)
		c := ix.Class(cls, ctx.PHP)
		if c == nil || issetExempt(c.FQN) || c.Kind != syntax.KindClass || c.Abstract {
			return
		}
		if ix.FindMethod(cls, "__isset", ctx.PHP) != nil || misAllowsDynamic(ctx, cls) {
			return
		}
		if !c.Final && misSubclassMayHave(ctx, cls, id.Value) {
			return
		}
		// custos: a missing ancestor may declare the property or
		// __isset().
		if !util.HierarchyResolved(ix, cls, ctx.PHP) {
			return
		}
		// custos: without __set() a write creates a real (dynamic)
		// property, which isset() does see; skip it when the file writes
		// properties of that name or by a computed name (an importer's
		// `$entity->$field = $value`).
		if ix.FindMethod(cls, "__set", ctx.PHP) == nil && misDynamicWrite(ctx, id.Value) {
			return
		}
		if report == "" {
			report = t
		}
	}
	if report != "" {
		ctx.ReportNode(pf, report+" has no __isset(); this isset/empty check is always false.")
	}
}

// misAllowsDynamic reports whether a class of cls' hierarchy carries
// #[\AllowDynamicProperties] (recorded by the index for every file) or is
// one of the exempt builtins.
func misAllowsDynamic(ctx *analysis.Context, cls string) bool {
	for _, c := range ctx.Index().Ancestors(cls, ctx.PHP) {
		// custos: subclasses of stdClass (always dynamic) and of the
		// exempt builtins (SimpleXMLElement's native isset handler).
		if c.HasAttr("AllowDynamicProperties") || issetExempt(c.FQN) {
			return true
		}
	}
	return false
}

// misDynamicWrite reports whether the file assigns a property named name on
// a receiver other than $this, or assigns a property with a computed name.
func misDynamicWrite(ctx *analysis.Context, name string) bool {
	w := ctx.Memo("writes", func() any {
		names := map[string]bool{}
		for _, st := range ctx.File.Stmts {
			syntax.Inspect(st, func(n syntax.Node) bool {
				as, ok := n.(*syntax.Assign)
				if !ok {
					return true
				}
				if pf, ok := as.Var.(*syntax.PropertyFetch); ok {
					if id, ok := pf.Name.(*syntax.Identifier); !ok {
						names[""] = true // computed name
					} else if ctx.Text(pf.Var) != "$this" {
						names[id.Value] = true
					}
				}
				return true
			})
		}
		return names
	}).(map[string]bool)
	return w[""] || w[name]
}

// misSubclassMayHave reports whether a descendant of cls declares the
// property or __isset() or allows dynamic properties (custos): the value may be such a subclass instance
// (`isset($e->errorcode)` on `Exception`, `isset($node->tagName)` on
// `DOMNode`).
func misSubclassMayHave(ctx *analysis.Context, cls, prop string) bool {
	ix := ctx.Index()
	queue := []string{cls}
	seen := map[string]bool{strings.ToLower(cls): true}
	for len(queue) > 0 {
		k := queue[0]
		queue = queue[1:]
		for _, ch := range ix.ChildrenAll(k) {
			lk := strings.ToLower(strings.TrimPrefix(ch, `\`))
			if c := ix.Class(ch, ctx.PHP); c != nil && !seen[lk] {
				if c.Props[prop] != nil || c.Methods["__isset"] != nil || c.HasAttr("AllowDynamicProperties") {
					return true
				}
				seen[lk] = true
				queue = append(queue, ch)
			}
		}
	}
	return false
}
