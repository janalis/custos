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
	target string // as written, without leading backslash
	fqn    string
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
			imp := ccImport{target: target, fqn: target}
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
			switch {
			case strings.EqualFold(u.fqn, cfqn): // L4
				if u.alias != "" {
					list = append(list, u.alias)
					break
				}
				r := ix.Class(u.target, ctx.PHP)
				if r == nil {
					break
				}
				rfqn := `\` + strings.TrimPrefix(r.FQN, `\`)
				precise := strings.HasSuffix(rfqn, u.target)
				if !precise || util.LastNamePart(rfqn) != t {
					list = append(list, rfqn)
				} else {
					list = append(list, t) // correct plain import (spec L4)
				}
			case u.alias != "" && strings.EqualFold(u.alias, t): // L5
				list = append(list, u.alias)
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
