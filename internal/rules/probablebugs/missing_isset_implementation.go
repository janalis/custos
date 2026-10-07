package probablebugs

import (
	"strings"

	"custos/internal/analysis"
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
		if report == "" {
			report = t
		}
	}
	if report != "" {
		ctx.ReportNode(pf, report+" has no __isset(); this isset/empty check is always false.")
	}
}

// misAllowsDynamic reports whether a class of cls' hierarchy carries
// #[\AllowDynamicProperties] (recorded by the index for every file).
func misAllowsDynamic(ctx *analysis.Context, cls string) bool {
	for _, c := range ctx.Index().Ancestors(cls, ctx.PHP) {
		if c.HasAttr("AllowDynamicProperties") {
			return true
		}
	}
	return false
}
