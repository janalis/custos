package probablebugs

import (
	"sort"
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/meta"
	"custos/internal/syntax"
	"custos/internal/types"
)

// traitsPropertiesConflicts reports properties declared both by a class (or
// its parent) and by one of the traits the class uses.
type traitsPropertiesConflicts struct{}

func init() { register(traitsPropertiesConflicts{}) }

// Semantic marks the rule as needing the project symbol index.
func (traitsPropertiesConflicts) Semantic() {}

func (traitsPropertiesConflicts) ID() string { return "TraitsPropertiesConflicts" }

func (traitsPropertiesConflicts) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClassLike}
}

func (traitsPropertiesConflicts) Check(ctx *analysis.Context, n syntax.Node) {
	cl := n.(*syntax.ClassLike)
	if cl.Name == nil || cl.ClassKind == syntax.KindInterface {
		return
	}
	var traits []*index.Class
	refs := map[string]*syntax.Name{} // lower trait FQN -> first reference
	resolve := func(nm *syntax.Name) *index.Class {
		c := ctx.Index().Class(ctx.Names().Class(nm.Value, nm.Span().Start), ctx.PHP)
		if c == nil || c.Kind != syntax.KindTrait {
			return nil
		}
		return c
	}
	addRef := func(nm *syntax.Name, c *index.Class) {
		k := strings.ToLower(strings.TrimPrefix(c.FQN, `\`))
		if _, ok := refs[k]; !ok {
			refs[k] = nm
		}
	}
	for _, m := range cl.Members {
		tu, ok := m.(*syntax.TraitUse)
		if !ok {
			continue
		}
		for _, nm := range tu.Traits {
			if c := resolve(nm); c != nil {
				traits = append(traits, c)
				addRef(nm, c)
			}
		}
		for _, ad := range tu.Adaptations {
			names := append([]*syntax.Name{}, ad.Insteadof...)
			if ad.Trait != nil {
				names = append([]*syntax.Name{ad.Trait}, names...)
			}
			for _, nm := range names {
				if c := resolve(nm); c != nil {
					addRef(nm, c)
				}
			}
		}
	}
	if len(traits) == 0 {
		return
	}
	className := cl.Name.Value
	msg := func(t *index.Class, prop string) string {
		return className + " and trait " + util.LastNamePart(t.FQN) + " both declare property $" + prop + "."
	}
	first := func(name string) (*index.Class, *index.Property) {
		for _, t := range traits {
			if p := tpcFind(ctx, t, name, map[string]bool{}); p != nil {
				return t, p
			}
		}
		return nil, nil
	}

	// A. own properties: compatible duplicates are weak, incompatible ones
	// (a fatal error in PHP) are errors; undecidable pairs stay silent.
	reportOwn := func(v *syntax.Variable, own *index.Property, attributed bool) {
		t, tp := first(v.Name)
		if t == nil {
			return
		}
		switch tpcCompare(own, tp) {
		case tpcSame:
			if attributed { // re-declared to attach attributes (mapping metadata)
				return
			}
			ctx.ReportSeverity(v.Span(), meta.SeverityInfo, msg(t, v.Name))
		case tpcDiffers:
			ctx.ReportSeverity(v.Span(), meta.SeverityError, msg(t, v.Name))
		}
	}
	classReadonly := cl.Modifiers.Has(syntax.TReadonly)
	for _, m := range cl.Members {
		switch m := m.(type) {
		case *syntax.Property:
			if m.Modifiers.Has(syntax.TAbstract) || util.DocHasAnnotation(ctx.File, m) {
				continue
			}
			typ := tpcTypeString(ctx, m.Type)
			for _, it := range m.Props {
				if it.Var == nil || it.Var.Name == "" {
					continue
				}
				own := &index.Property{Name: it.Var.Name, Visibility: tpcVisibility(m.Modifiers),
					Static: m.Modifiers.Has(syntax.TStatic), Readonly: classReadonly || m.Modifiers.Has(syntax.TReadonly),
					Type: typ, HasDefault: it.Default != nil}
				if it.Default != nil {
					own.Default = ctx.Text(it.Default)
				}
				reportOwn(it.Var, own, len(m.Attrs) > 0)
			}
		case *syntax.Method:
			if !strings.EqualFold(m.Name.Value, "__construct") {
				continue
			}
			for _, p := range m.Params {
				if len(p.Modifiers) == 0 || p.Var == nil || p.Var.Name == "" {
					continue
				}
				// A promoted property never has a default value of its own.
				reportOwn(p.Var, &index.Property{Name: p.Var.Name, Visibility: tpcVisibility(p.Modifiers),
					Readonly: classReadonly || p.Modifiers.Has(syntax.TReadonly), Type: tpcTypeString(ctx, p.Type), Promoted: true}, len(p.Attrs) > 0)
			}
		}
	}

	// B. parent properties
	if len(cl.Extends) == 0 || cl.ClassKind != syntax.KindClass {
		return
	}
	parent := ctx.Index().Class(ctx.Names().Class(cl.Extends[0].Value, cl.Extends[0].Span().Start), ctx.PHP)
	if parent == nil {
		return
	}
	var props []*index.Property
	for _, p := range parent.Props {
		if !p.Magic && p.Visibility != index.Private {
			props = append(props, p)
		}
	}
	sort.Slice(props, func(i, j int) bool { return props[i].Span.Start < props[j].Span.Start })
	for _, q := range props {
		t, tp := first(q.Name)
		if t == nil {
			continue
		}
		// every trait in traits was registered in refs when collected
		ref := refs[strings.ToLower(strings.TrimPrefix(t.FQN, `\`))]
		sev := meta.SeverityError
		if tpcSameDefault(tpcHasDefault(q), q.Default, tp) && tpcCompare(q, tp) != tpcDiffers {
			sev = meta.SeverityInfo
		}
		ctx.ReportSeverity(ref.Span(), sev, msg(t, q.Name))
	}
}

// tpcFind looks a real (non-magic) property up in trait t and the traits it
// uses.
func tpcFind(ctx *analysis.Context, t *index.Class, name string, seen map[string]bool) *index.Property {
	k := strings.ToLower(t.FQN)
	if seen[k] {
		return nil
	}
	seen[k] = true
	if p, ok := t.Props[name]; ok && !p.Magic {
		return p
	}
	for _, sub := range t.Traits {
		if c := ctx.Index().Class(sub, ctx.PHP); c != nil {
			if p := tpcFind(ctx, c, name, seen); p != nil {
				return p
			}
		}
	}
	return nil
}

func tpcSameDefault(has bool, def string, tp *index.Property) bool {
	if has != tpcHasDefault(tp) {
		return false
	}
	return !has || tpcTokensEqual(def, tp.Default)
}

// tpcTokensEqual compares two expression texts token by token, ignoring
// whitespace and comments.
func tpcTokensEqual(a, b string) bool {
	if a == b {
		return true
	}
	sig := func(s string) []string {
		src := []byte("<?php " + s + ";")
		toks, _ := syntax.Lex(src, syntax.LexOptions{})
		var out []string
		for _, t := range toks {
			if !t.Kind.IsTrivia() && t.Kind != syntax.TOpenTag && t.Kind != syntax.TEOF {
				out = append(out, string(src[t.Start:t.End]))
			}
		}
		return out
	}
	x, y := sig(a), sig(b)
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// tpcHasDefault reports whether p declares a default value; a promoted
// property never does (the parameter default belongs to the constructor).
func tpcHasDefault(p *index.Property) bool { return p.HasDefault && !p.Promoted }

type tpcCompat int

const (
	tpcUnsure  tpcCompat = iota // cannot be decided statically
	tpcSame                     // PHP accepts the duplicate
	tpcDiffers                  // PHP rejects the duplicate (fatal error)
)

// tpcCompare applies PHP's trait property compatibility criteria: same
// visibility, static and readonly flags, same declared type and identical
// default value.
func tpcCompare(a, b *index.Property) tpcCompat {
	if a.Visibility != b.Visibility || a.Static != b.Static || a.Readonly != b.Readonly {
		return tpcDiffers
	}
	res := tpcSame
	switch {
	case (a.Type == "") != (b.Type == ""):
		return tpcDiffers
	case tpcSelfLike(a.Type) || tpcSelfLike(b.Type):
		res = tpcUnsure // self/static/parent depend on the composing class
	case !strings.EqualFold(a.Type, b.Type):
		return tpcDiffers
	}
	if tpcSameDefault(tpcHasDefault(a), a.Default, b) {
		return res
	}
	ka, okA := tpcDefaultValue(a)
	kb, okB := tpcDefaultValue(b)
	if !okA || !okB {
		return tpcUnsure
	}
	if ka != kb {
		return tpcDiffers
	}
	return res
}

// tpcSelfLike reports a type string mentioning self, static or parent.
func tpcSelfLike(t string) bool {
	for _, a := range strings.Split(strings.ToLower(t), "|") {
		switch strings.TrimLeft(a, `\?`) {
		case "self", "static", "parent":
			return true
		}
	}
	return false
}

// tpcDefaultValue canonicalises a property's initial value when it is a
// scalar literal (or absent); ok is false for anything else (constants,
// arrays, expressions), whose value is not compared.
func tpcDefaultValue(p *index.Property) (string, bool) {
	if !tpcHasDefault(p) {
		if p.Type == "" {
			return "null", true // untyped properties default to null
		}
		return "uninitialized", true
	}
	src := []byte("<?php " + p.Default + ";")
	toks, _ := syntax.Lex(src, syntax.LexOptions{})
	var sig []syntax.Token
	for _, t := range toks {
		if !t.Kind.IsTrivia() && t.Kind != syntax.TOpenTag && t.Kind != syntax.TEOF {
			sig = append(sig, t)
		}
	}
	text := func(t syntax.Token) string { return string(src[t.Start:t.End]) }
	if n := len(sig); n > 0 && text(sig[n-1]) == ";" {
		sig = sig[:n-1]
	}
	neg := false
	if len(sig) == 2 && text(sig[0]) == "-" {
		neg, sig = true, sig[1:]
	}
	if len(sig) != 1 {
		return "", false
	}
	t := sig[0]
	switch t.Kind {
	case syntax.TLNumber:
		v, ok := util.ParseIntLiteral(text(t))
		if !ok {
			return "", false
		}
		if neg {
			v = -v
		}
		return "i:" + strconv.FormatInt(v, 10), true
	case syntax.TDNumber:
		v, err := strconv.ParseFloat(strings.ReplaceAll(text(t), "_", ""), 64)
		if err != nil {
			return "", false
		}
		if neg {
			v = -v
		}
		return "f:" + strconv.FormatFloat(v, 'g', -1, 64), true
	case syntax.TConstantEncapsedString:
		raw := text(t)
		if neg || len(raw) < 2 || strings.ContainsAny(raw, `\$`) {
			return "", false
		}
		return "s:" + raw[1:len(raw)-1], true
	}
	if neg {
		return "", false
	}
	switch w := strings.ToLower(text(t)); w {
	case "true", "false", "null":
		return w, true
	}
	return "", false
}

func tpcVisibility(m syntax.Modifiers) index.Visibility {
	switch {
	case m.Has(syntax.TPrivate):
		return index.Private
	case m.Has(syntax.TProtected):
		return index.Protected
	}
	return index.Public
}

// tpcTypeString renders a declared type the way the symbol index stores it.
func tpcTypeString(ctx *analysis.Context, n syntax.Expr) string {
	if n == nil {
		return ""
	}
	at := n.Span().Start
	// String() of a type without atoms is "", like the index's rendering.
	return types.FromNode(n, func(w string) string { return ctx.Names().Class(w, at) }).String()
}
