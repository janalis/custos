package classoverridesfieldofsuperclass

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/types"
)

// classOverridesFieldOfSuperClass reports properties re-declaring a
// property of the parent chain.
type classOverridesFieldOfSuperClass struct{}

func (classOverridesFieldOfSuperClass) ID() string { return "ClassOverridesFieldOfSuperClass" }
func (classOverridesFieldOfSuperClass) Semantic()  {}
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
	if ctx.IsTestFile() || astquery.IsTestClassFQN(ctx.Names().DeclFQN(cl)) {
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
	parent := ix.Class(ctx.Names().ParentFQN(cl), ctx.PHP)
	if parent == nil {
		return
	}
	if ctor := semanticquery.MethodInChain(ix, parent.FQN, "__construct", ctx.PHP); ctor != nil && (ctor.Final || ctor.Visibility == index.Private) { // D3
		return
	}
	own := index.Public
	switch {
	case prop.Modifiers.Has(syntax.TPrivate):
		own = index.Private
	case prop.Modifiers.Has(syntax.TProtected):
		own = index.Protected
	}
	// Effective write visibility and finality of the own declaration:
	// `private(set)` is implicitly final.
	ownSet, ownFinal := own, prop.Modifiers.Has(syntax.TFinal)
	switch {
	case prop.Modifiers.Has(syntax.TPrivateSet):
		ownSet, ownFinal = index.Private, true
	case prop.Modifiers.Has(syntax.TProtectedSet):
		ownSet = index.Protected
	case prop.Modifiers.Has(syntax.TPublicSet):
		ownSet = index.Public
	}
	for _, item := range prop.Props {
		if item.Var.Name == "" { // recovered declaration without a variable
			continue
		}
		found := semanticquery.PropertyInChain(ix, parent.FQN, item.Var.Name, ctx.PHP) // D4
		if found == nil {
			continue
		}
		holder := `\` + strings.TrimPrefix(found.Class, `\`)
		if found.Visibility == index.Private { // D5
			if ctx.Bool("REPORT_PRIVATE_REDEFINITION") {
				ctx.ReportSeverity(item.Var.Span(), diagnostic.SeverityInfo, holder+" already has a private property with this name; consider a different name.")
			}
			continue
		}
		if own < found.Visibility { // D6: deliberately widened
			continue
		}
		// D6c: attributes, hooks, `final` and asymmetric visibility are
		// behaviour the inherited declaration does not have (or, for an
		// abstract or hooked one, that this declaration implements or
		// replaces).
		foundSet := found.Visibility
		if found.SetVisibility != nil {
			foundSet = *found.SetVisibility
		}
		asymmetric := ownSet != own || foundSet != found.Visibility
		if len(prop.Attrs) > 0 || len(prop.Hooks) > 0 || found.Hooked || ownFinal != found.Final || asymmetric && ownSet != foundSet {
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
		if item.Default != nil && !cofsSameContext(ctx, cl, item.Default, found) {
			continue
		}
		// custos: a re-declaration whose @var narrows the documented type
		// is the only way to refine it (native property types are
		// invariant); dropping it loses the type.
		if d := cofsOwnDocType(ctx, prop, item.Var.Name); d != "" && d != found.DocType && d != found.Type {
			continue
		}
		ctx.ReportSeverity(item.Var.Span(), diagnostic.SeverityInfo, "Property '"+item.Var.Name+"' is already declared in "+holder+"; drop this re-declaration.")
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

// cofsSameContext reports whether def, the own default whose text equals
// the default of found, also evaluates to the same value there: magic
// constants never do, `self::X` only when both classes reach the same
// constant declaration, and names that are not fully qualified only when
// both declarations share one namespace scope of this file.
func cofsSameContext(ctx *analysis.Context, cl *syntax.ClassLike, def syntax.Expr, found *index.Property) bool {
	ix, nr := ctx.Index(), ctx.Names()
	holder := ix.Class(found.Class, ctx.PHP)
	sameScope := holder != nil && holder.File == ctx.File.Path && nr.ScopeAt(holder.Span.Start) == nr.ScopeAt(def.Span().Start)
	same := true
	syntax.Inspect(def, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.MagicConst:
			same = false
		case *syntax.ConstFetch:
			switch strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`)) {
			case "null", "true", "false":
				return false
			}
		case *syntax.ClassConstFetch:
			if c, ok := n.Class.(*syntax.Name); ok && names.IsSpecialClass(c.Value) {
				id, _ := n.Name.(*syntax.Identifier)
				same = same && id != nil && strings.EqualFold(c.Value, "self") && !strings.EqualFold(id.Value, "class") &&
					cofsSameConst(ix.FindConst(nr.SymbolFQN(cl), id.Value, ctx.PHP), ix.FindConst(found.Class, id.Value, ctx.PHP))
				return false
			}
		case *syntax.Name:
			same = same && (n.NameKind == syntax.NameFullyQualified || sameScope)
		}
		return same
	})
	return same
}

// cofsSameConst reports whether two class-constant lookups reached the same
// declaration.
func cofsSameConst(a, b *index.ClassConst) bool {
	return a != nil && b != nil && strings.EqualFold(a.Class, b.Class)
}

// cofsOwnDocType is the canonical @var type documented on the property
// declaration for name ("" when there is none).
func cofsOwnDocType(ctx *analysis.Context, prop *syntax.Property, name string) string {
	c := index.DocComment(ctx.File, prop)
	at := prop.Span().Start
	t := types.FromDoc(phpdoc.Parse(c).VarType(name), func(w string) string { return ctx.Names().Class(w, at) })
	if t.IsUnknown() {
		return ""
	}
	return t.DocString()
}
