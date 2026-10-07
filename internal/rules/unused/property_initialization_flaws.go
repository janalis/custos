package unused

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// propertyInitializationFlaws reports redundant property defaults and
// constructor assignments writing the default value.
type propertyInitializationFlaws struct{}

func init() { register(propertyInitializationFlaws{}) }

func (propertyInitializationFlaws) ID() string { return "PropertyInitializationFlaws" }

func (propertyInitializationFlaws) Semantic() {}

func (propertyInitializationFlaws) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KProperty, syntax.KMethod}
}

const (
	pifMsgNull     = "Explicit null default is redundant; remove it."
	pifMsgSame     = "Default repeats the inherited value; drop the re-declaration."
	pifMsgReplaced = "Default is always replaced by the constructor; remove it."
	pifMsgWrites   = "Assignment writes the property's default value; remove it."
)

func (propertyInitializationFlaws) Check(ctx *analysis.Context, n syntax.Node) {
	switch n := n.(type) {
	case *syntax.Property:
		if ctx.Bool("REPORT_DEFAULTS_FLAWS") {
			pifDefaults(ctx, n)
		}
	case *syntax.Method:
		if ctx.Bool("REPORT_INIT_FLAWS") {
			pifConstructor(ctx, n)
		}
	}
}

// pifNullableTyped implements E1.
func pifNullableTyped(ctx *analysis.Context, typ syntax.Expr) bool {
	if typ == nil || ctx.PHP < phpver.PHP74 {
		return false
	}
	nullish := func(e syntax.Expr) bool {
		nm, ok := e.(*syntax.Name)
		return ok && (strings.EqualFold(nm.Value, "null") || strings.EqualFold(nm.Value, "mixed"))
	}
	switch t := typ.(type) {
	case *syntax.NullableType:
		return true
	case *syntax.UnionType:
		for _, x := range t.Types {
			if nullish(x) {
				return true
			}
		}
		return false
	}
	return nullish(typ)
}

// pifRemoveDefault deletes from after the property name through the end of
// the default value (F1).
func pifRemoveDefault(item *syntax.PropertyItem) analysis.Fix {
	return analysis.Fix{
		Title: "Remove the default value",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: syntax.Span{Start: item.Var.Span().End, End: item.Default.Span().End}}}
		},
	}
}

// pifDefaults is Check 1 (D1, D2).
func pifDefaults(ctx *analysis.Context, prop *syntax.Property) {
	cl := prop.Parent().(*syntax.ClassLike) // properties only parse as class members
	parent := ""
	if cl.ClassKind == syntax.KindClass {
		if pc := ctx.Index().Class(ctx.Names().ParentFQN(cl), ctx.PHP); pc != nil {
			parent = pc.FQN
		}
	}
	for _, item := range prop.Props {
		d := item.Default
		if d == nil || d.Span().Len() == 0 {
			continue
		}
		if syntax.IsNullConst(d) { // D1
			if !pifNullableTyped(ctx, prop.Type) {
				ctx.Report(d.Span(), pifMsgNull, pifRemoveDefault(item))
			}
			continue
		}
		// D2; a static re-declaration gives the subclass its own storage
		// (custos: not reported even with the same default).
		if parent == "" || prop.Modifiers.Has(syntax.TStatic) {
			continue
		}
		o := util.PropertyInChain(ctx.Index(), parent, item.Var.Name, ctx.PHP)
		if o == nil || !o.HasDefault || o.Visibility == index.Private {
			continue
		}
		if pifSameAsInherited(ctx, d, o) {
			ctx.Report(d.Span(), pifMsgSame)
		}
	}
}

// pifSameAsInherited compares a default with the inherited one, including
// the class-reference guard.
func pifSameAsInherited(ctx *analysis.Context, d syntax.Expr, o *index.Property) bool {
	od := pifPropertyDefault(ctx, o)
	if od == nil {
		// Declared in another file: text comparison, no class references.
		if pifClassRefs(ctx, d) != nil {
			return false
		}
		return spmSquash(ctx.Text(d)) == spmSquash(o.Default)
	}
	if !util.EquivalentFoldNames(ctx.File, d, od) {
		return false
	}
	theirs := pifClassRefs(ctx, od)
	if len(theirs) == 0 {
		return true
	}
	for k := range pifClassRefs(ctx, d) {
		if !theirs[k] {
			return false
		}
	}
	return true
}

// pifPropertyDefault finds the default expression node of an indexed
// property declared in this file.
func pifPropertyDefault(ctx *analysis.Context, p *index.Property) syntax.Expr {
	cl := util.ClassDecl(ctx.File, ctx.Index().Class(p.Class, ctx.PHP))
	if cl == nil {
		return nil
	}
	var out syntax.Expr
	for _, m := range cl.Members {
		if pr, ok := m.(*syntax.Property); ok {
			for _, it := range pr.Props {
				if it.Span() == p.Span {
					out = it.Default
				}
			}
		}
	}
	return out
}

// pifClassRefs collects the resolved class references inside e ("?" for
// unresolvable ones); nil when there are none. Property defaults are
// constant expressions, so `X::NAME` / `X::class` are the only class
// references they can hold. self/parent resolve against the class declaring
// the default (custos: a child's self::X may name its own constant).
func pifClassRefs(ctx *analysis.Context, e syntax.Expr) map[string]bool {
	var out map[string]bool
	syntax.Inspect(e, func(x syntax.Node) bool {
		cf, ok := x.(*syntax.ClassConstFetch)
		if !ok {
			return true
		}
		nm, ok := cf.Class.(*syntax.Name)
		if !ok {
			return true
		}
		if out == nil {
			out = map[string]bool{}
		}
		key := "?"
		fqn := ctx.Names().Class(nm.Value, nm.Span().Start)
		switch strings.ToLower(fqn) {
		case "self", "static":
			fqn = ctx.Names().DeclFQN(syntax.EnclosingClass(nm))
		case "parent":
			fqn = ctx.Names().ParentFQN(syntax.EnclosingClass(nm))
		}
		if c := ctx.Index().Class(fqn, ctx.PHP); c != nil {
			key = strings.ToLower(strings.TrimPrefix(c.FQN, `\`))
		}
		out[key] = true
		return true
	})
	return out
}

type pifCandidate struct {
	item *syntax.PropertyItem
	prop *syntax.Property
}

// pifConstructor is Check 2 (D3–D7).
func pifConstructor(ctx *analysis.Context, m *syntax.Method) {
	if m.Name == nil || !strings.EqualFold(m.Name.Value, "__construct") || m.Body == nil || len(m.Body.Stmts) == 0 {
		return
	}
	cl, ok := m.Parent().(*syntax.ClassLike)
	if !ok || cl.ClassKind != syntax.KindClass {
		return
	}
	cands := map[string]pifCandidate{} // D3
	for _, mem := range cl.Members {
		pr, ok := mem.(*syntax.Property)
		if !ok || !pr.Modifiers.Has(syntax.TPrivate) || pr.Modifiers.Has(syntax.TStatic) {
			continue
		}
		for _, it := range pr.Props {
			if it.Var.Name != "" {
				cands[it.Var.Name] = pifCandidate{item: it, prop: pr}
			}
		}
	}
	if len(cands) == 0 {
		return
	}
	reported := map[string]bool{}
	earlyReturn := false              // a `return` before this statement: later writes are conditional
	for _, st := range m.Body.Stmts { // D4
		mayReturn := pifHasReturn(st)
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			earlyReturn = earlyReturn || mayReturn
			continue
		}
		as, ok := es.Expr.(*syntax.Assign)
		if !ok || as.Op.Kind != syntax.TEqual {
			continue
		}
		pf, ok := as.Var.(*syntax.PropertyFetch)
		if !ok || pf.NullSafe {
			continue
		}
		if v, ok := pf.Var.(*syntax.Variable); !ok || v.Name != "this" {
			continue
		}
		id, ok := pf.Name.(*syntax.Identifier)
		if !ok {
			continue
		}
		c, ok := cands[id.Value]
		if !ok {
			continue
		}
		d := c.item.Default
		if d != nil && syntax.IsNullConst(d) {
			d = nil
		}
		if (d == nil && syntax.IsNullConst(as.Value)) || (d != nil && util.EquivalentFoldNames(ctx.File, as.Value, d)) { // D5
			if !pifNullableTyped(ctx, c.prop.Type) {
				ctx.Report(es.Span(), pifMsgWrites)
			}
			continue
		}
		if d == nil { // D6
			continue
		}
		reuses := false // D7
		syntax.Inspect(as.Value, func(x syntax.Node) bool {
			if p, ok := x.(*syntax.PropertyFetch); ok && util.EquivalentFoldNames(ctx.File, p, pf) {
				reuses = true
			}
			return !reuses
		})
		// A typed property without default is uninitialised: objects made
		// without the constructor (unserialize, reflection, ORM hydration)
		// would then throw on access, so its default is kept (custos).
		typed := c.prop.Type != nil
		if reuses || earlyReturn || typed || !ctx.Bool("REPORT_DEFAULTS_FLAWS") || reported[id.Value] {
			continue
		}
		reported[id.Value] = true
		ctx.Report(d.Span(), pifMsgReplaced, pifRemoveDefault(c.item))
	}
}

// pifHasReturn reports whether st contains a return statement of the
// enclosing function (nested functions and closures excluded).
func pifHasReturn(st syntax.Node) bool {
	found := false
	syntax.Inspect(st, func(x syntax.Node) bool {
		if found {
			return false
		}
		if _, ok := x.(*syntax.Return); ok {
			found = true
			return false
		}
		return x == st || !syntax.IsFuncLike(x)
	})
	return found
}
