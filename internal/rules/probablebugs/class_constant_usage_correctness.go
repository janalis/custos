package probablebugs

import (
	"custos/internal/analysis/util"
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// classConstantUsageCorrectness reports `X::class` where X is written with
// a letter case that differs from the declaration.
type classConstantUsageCorrectness struct{}

func init() { register(classConstantUsageCorrectness{}) }

func (classConstantUsageCorrectness) ID() string { return "ClassConstantUsageCorrectness" }

func (classConstantUsageCorrectness) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClassConstFetch}
}

// Semantic marks the rule as needing the project index.
func (classConstantUsageCorrectness) Semantic() {}

type ccImport struct {
	target string // as written, without leading backslash (also the FQN)
	alias  string // explicit alias or ""
}

// ccScope returns the nearest enclosing namespace (nil when none) and the
// class imports that are direct statements of it (or of the file).
func ccScope(f *syntax.File, n syntax.Node) (*syntax.Namespace, []ccImport) {
	var ns *syntax.Namespace
	for p := n.Parent(); p != nil; p = p.Parent() {
		if x, ok := p.(*syntax.Namespace); ok {
			ns = x
			break
		}
	}
	stmts := f.Stmts
	if ns != nil {
		stmts = ns.Stmts
	}
	var out []ccImport
	for _, st := range stmts {
		u, ok := st.(*syntax.Use)
		if !ok {
			continue
		}
		prefix := ""
		if u.Prefix != nil {
			prefix = strings.Trim(u.Prefix.Value, `\`) + `\`
		}
		for _, it := range u.Items {
			kind := u.Type
			if u.Prefix != nil {
				kind = it.Type
			}
			if kind != syntax.UseNormal || it.Name == nil {
				continue
			}
			target := prefix + strings.TrimPrefix(it.Name.Value, `\`)
			imp := ccImport{target: target}
			if it.Alias != nil {
				imp.alias = it.Alias.Value
			}
			out = append(out, imp)
		}
	}
	return ns, out
}

func (classConstantUsageCorrectness) Check(ctx *analysis.Context, n syntax.Node) {
	fetch := n.(*syntax.ClassConstFetch)
	id, ok := fetch.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "class") { // D1 (case-insensitive keyword)
		return
	}
	x, ok := fetch.Class.(*syntax.Name) // D2
	if !ok {
		return
	}
	t := x.Value
	switch strings.ToLower(t) { // D3 (case-insensitive, see spec divergences)
	case "self", "static", "parent":
		return
	}
	ix := ctx.Index()
	c := ix.Class(ctx.Names().Class(t, x.Span().Start), ctx.PHP) // D4
	if c == nil {
		return
	}
	cfqn := strings.TrimPrefix(c.FQN, `\`)
	var list []string // D5
	ns, uses := ccScope(ctx.File, fetch)
	switch {
	case strings.HasPrefix(t, `\`): // L1
		if k := ix.Class(t, ctx.PHP); k != nil {
			list = append(list, `\`+strings.TrimPrefix(k.FQN, `\`))
		}
	case strings.Contains(t, `\`):
		if ns == nil {
			break
		}
		if ns.Name != nil { // L2
			nsFQN := strings.ToLower(strings.Trim(ns.Name.Value, `\`))
			lc := strings.ToLower(cfqn)
			if (strings.HasPrefix(lc, nsFQN) || strings.HasSuffix(lc, `\`+strings.ToLower(t))) && len(cfqn) >= len(t) {
				list = append(list, cfqn[len(cfqn)-len(t):])
			}
		}
		for _, u := range uses { // L3
			if u.alias != "" && strings.HasPrefix(strings.ToLower(t), strings.ToLower(u.alias)) {
				list = append(list, strings.TrimPrefix(strings.ReplaceAll(`\`+cfqn, u.target, u.alias), `\`))
			}
		}
	default:
		for _, u := range uses {
			// Only an import that resolves the written name (alias, or last
			// segment when unaliased) can say how T should be spelled; the
			// others (`use A\Bag as B;` with `Bag` resolved through the
			// namespace) are unrelated to this access (custos).
			eff := u.alias
			if eff == "" {
				eff = util.LastNamePart(u.target)
			}
			if !strings.EqualFold(eff, t) {
				continue
			}
			// custos: `::class` on an imported name yields the import's
			// target as written in the `use` statement, whatever the case
			// of the short name or alias at the access: only a wrong-case
			// import gives the wrong string.
			switch {
			case strings.EqualFold(u.target, cfqn): // L4
				if u.target == cfqn {
					list = append(list, t)
				} else {
					list = append(list, `\`+cfqn)
				}
			case u.alias != "": // L5
				list = append(list, t)
			}
		}
	}
	if len(list) == 0 {
		return
	}
	for _, s := range list {
		if s == t {
			return
		}
	}
	ctx.ReportNode(x, "Letter case of the class name differs from its declaration; ::class will return the wrong string.")
}
