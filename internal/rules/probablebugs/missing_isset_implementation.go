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
	types := ctx.TypeOf(pf.Var).Classes()
	for _, t := range types { // D2
		if ix.FindProperty(strings.TrimPrefix(t, `\`), id.Value, ctx.PHP) != nil {
			return
		}
	}
	for _, t := range types { // D4
		if issetExempt(t) {
			continue
		}
		cls := strings.TrimPrefix(t, `\`)
		c := ix.Class(cls, ctx.PHP)
		if c == nil || issetExempt(c.FQN) {
			continue
		}
		if ix.FindMethod(cls, "__isset", ctx.PHP) == nil {
			ctx.ReportNode(pf, t+" has no __isset(); this isset/empty check is always false.")
			return
		}
	}
}
