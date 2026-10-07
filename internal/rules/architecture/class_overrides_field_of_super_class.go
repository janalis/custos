package architecture

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/meta"
	"custos/internal/phpdoc"
	"custos/internal/syntax"
)

// classOverridesFieldOfSuperClass reports properties re-declaring a
// property of the parent chain.
type classOverridesFieldOfSuperClass struct{}

func init() { register(classOverridesFieldOfSuperClass{}) }

func (classOverridesFieldOfSuperClass) ID() string { return "ClassOverridesFieldOfSuperClass" }

func (classOverridesFieldOfSuperClass) Semantic() {}

func (classOverridesFieldOfSuperClass) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KProperty}
}

func (classOverridesFieldOfSuperClass) Check(ctx *analysis.Context, n syntax.Node) {
	prop := n.(*syntax.Property)
	if prop.Modifiers.Has(syntax.TStatic) { // D1
		return
	}
	cl, ok := prop.Parent().(*syntax.ClassLike)
	if !ok || cl.ClassKind != syntax.KindClass || len(cl.Extends) == 0 {
		return
	}
	if util.IsTestPath(ctx.File.Path) || util.IsTestClassFQN(util.ClassDeclFQN(ctx.Names(), cl)) {
		return
	}
	if c := index.DocComment(ctx.File, prop); c != "" { // D2
		for _, t := range phpdoc.Parse(c).Tags {
			if strings.ToLower(t.Name) != t.Name {
				return
			}
		}
	}
	ix := ctx.Index()
	parent := ix.Class(util.ParentFQN(ctx.Names(), cl), ctx.PHP)
	if parent == nil {
		return
	}
	if ctor := util.MethodInChain(ix, parent.FQN, "__construct", ctx.PHP); ctor != nil && (ctor.Final || ctor.Visibility == index.Private) { // D3
		return
	}
	own := index.Public
	switch {
	case prop.Modifiers.Has(syntax.TPrivate):
		own = index.Private
	case prop.Modifiers.Has(syntax.TProtected):
		own = index.Protected
	}
	for _, item := range prop.Props {
		if item.Var == nil || item.Var.Name == "" || item.Var.Span().Len() == 0 {
			continue
		}
		found := util.PropertyInChain(ix, parent.FQN, item.Var.Name, ctx.PHP) // D4
		if found == nil {
			continue
		}
		holder := `\` + strings.TrimPrefix(found.Class, `\`)
		if found.Visibility == index.Private { // D5
			if ctx.Bool("REPORT_PRIVATE_REDEFINITION") {
				ctx.ReportSeverity(item.Var.Span(), meta.SeverityInfo, holder+" already has a private property with this name; consider a different name.")
			}
			continue
		}
		if own < found.Visibility { // D6: deliberately widened
			continue
		}
		// D6b: a re-declaration that changes the default value is how a
		// subclass overrides it; dropping it would change behaviour.
		ownDef := ""
		if item.Default != nil {
			ownDef = ctx.Text(item.Default)
		}
		if cofsDefault(ownDef, item.Default != nil, prop.Type != nil) != cofsDefault(found.Default, found.HasDefault, found.Type != "") {
			continue
		}
		ctx.ReportSeverity(item.Var.Span(), meta.SeverityInfo, "Property '"+item.Var.Name+"' is already declared in "+holder+"; drop this re-declaration.")
	}
}

// cofsDefault normalises a property default for comparison: whitespace
// removed, lower-cased `null`; an untyped property without default holds
// null, a typed one is uninitialised.
func cofsDefault(text string, has, typed bool) string {
	if !has {
		if typed {
			return "<uninitialised>"
		}
		return "null"
	}
	t := strings.Join(strings.Fields(text), "")
	if strings.EqualFold(strings.TrimPrefix(t, `\`), "null") {
		return "null"
	}
	return t
}
