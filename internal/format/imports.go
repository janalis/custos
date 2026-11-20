package format

import (
	"sort"
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
)

// imports applies the import settings to every statement list holding
// `use` imports: the file and each namespace.
func (e *engine) imports() {
	e.importList(e.f.Stmts, 0)
	for _, st := range e.f.Stmts {
		if ns, ok := st.(*syntax.Namespace); ok {
			lvl := 0
			if ns.Braced {
				lvl = 1
			}
			e.importList(ns.Stmts, lvl)
		}
	}
}

// importList handles the imports among list, statements at level lvl.
func (e *engine) importList(list []syntax.Stmt, lvl int) {
	var run []*syntax.Use
	flush := func() {
		if len(run) > 0 {
			e.importRun(run, lvl)
		}
		run = nil
	}
	for _, st := range list {
		u, ok := st.(*syntax.Use)
		if !ok {
			flush()
			continue
		}
		if k := e.unitAt(u.Span().Start); k >= 0 && e.deleted[k] {
			continue // removed as unused
		}
		if u.Prefix != nil {
			flush() // a group import stays as written and ends the run
			continue
		}
		if len(run) > 0 && e.commentBetween(run[len(run)-1].Span().End, u.Span().Start) {
			flush()
		}
		run = append(run, u)
	}
	flush()
}

// commentBetween reports whether a comment lies in [from, to).
func (e *engine) commentBetween(from, to uint32) bool {
	for i := e.index(from); i < len(e.toks) && e.toks[i].Start < to; i++ {
		if k := e.toks[i].Kind; k == syntax.TComment || k == syntax.TDocComment {
			return true
		}
	}
	return false
}

// importEntry is one import of a run, rendered on its own.
type importEntry struct {
	kind syntax.UseKind
	name string // as sorted: without leading backslash
	text string // the statement
}

// importRun rewrites a run of consecutive import statements when the
// settings ask to split, strip, sort or group them.
func (e *engine) importRun(run []*syntax.Use, lvl int) {
	defer e.newGroup()()
	s := e.s
	if !s.NoLeadingImportSlash && !s.SingleImportPerStatement && s.ImportOrder == "none" && s.ImportGroups == "none" {
		return
	}
	for _, u := range run {
		sp := u.Span()
		if e.commentBetween(sp.Start, sp.End) || e.inRegions(e.ignored, sp.Start) {
			return // annotated imports stay as written
		}
	}
	var entries []importEntry
	rule := ""
	for _, u := range run {
		kind := useKind(u)
		if len(u.Items) > 1 && !s.SingleImportPerStatement {
			names := make([]string, len(u.Items))
			for i, it := range u.Items {
				names[i] = e.importName(it, &rule)
			}
			entries = append(entries, importEntry{kind: kind, name: strings.TrimPrefix(u.Items[0].Name.Value, `\`), text: useWord(kind) + strings.Join(names, ", ") + ";"})
			continue
		}
		if len(u.Items) > 1 {
			rule = "single_import_per_statement"
		}
		for _, it := range u.Items {
			entries = append(entries, importEntry{kind: kind, name: strings.TrimPrefix(it.Name.Value, `\`), text: useWord(kind) + e.importName(it, &rule) + ";"})
		}
	}
	sorted := append([]importEntry(nil), entries...)
	order := groupOrder(s.ImportGroups)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if order != nil && order[a.kind] != order[b.kind] {
			return order[a.kind] < order[b.kind]
		}
		switch s.ImportOrder {
		case "alpha":
			return importLess(a.name, b.name)
		case "length":
			if len(a.name) != len(b.name) {
				return len(a.name) < len(b.name)
			}
			return importLess(a.name, b.name)
		}
		return false
	})
	for i := range sorted {
		if sorted[i].text != entries[i].text {
			if s.ImportOrder != "none" {
				rule = "import_order"
			} else {
				rule = "import_groups"
			}
			break
		}
	}
	if rule == "" {
		return
	}
	ind := e.indentString(lvl)
	var b strings.Builder
	for i, en := range sorted {
		if i > 0 {
			b.WriteString(e.eol)
			if s.BlankLineBetweenImportGroups && en.kind != sorted[i-1].kind {
				b.WriteString(e.eol)
			}
			b.WriteString(ind)
		}
		b.WriteString(en.text)
	}
	first := e.unitAt(run[0].Span().Start)
	last := e.unitBefore(run[len(run)-1].Span().End)
	e.merged = true
	e.rewrite(first, b.String(), rule)
	for k := first + 1; k <= last; k++ {
		e.remove(k, rule)
		e.gaps[k].drop = true
	}
	if last+1 < len(e.gaps) {
		e.gaps[last+1].drop = false
	}
	e.gaps[first+1].drop = true
}

// importName renders one imported name with its alias.
func (e *engine) importName(it *syntax.UseItem, rule *string) string {
	name := it.Name.Value
	if strings.HasPrefix(name, `\`) && e.s.NoLeadingImportSlash {
		name = name[1:]
		if *rule == "" {
			*rule = "no_leading_import_slash"
		}
	}
	if it.Alias != nil {
		name += " as " + it.Alias.Value
	}
	return name
}

func useWord(k syntax.UseKind) string {
	switch k {
	case syntax.UseFunction:
		return "use function "
	case syntax.UseConst:
		return "use const "
	}
	return "use "
}

func groupOrder(v string) map[syntax.UseKind]int {
	switch v {
	case "class_function_const":
		return map[syntax.UseKind]int{syntax.UseNormal: 0, syntax.UseFunction: 1, syntax.UseConst: 2}
	case "const_class_function":
		return map[syntax.UseKind]int{syntax.UseConst: 0, syntax.UseNormal: 1, syntax.UseFunction: 2}
	}
	return nil
}

// importLess orders names case-insensitively, namespace separators first.
func importLess(a, b string) bool {
	la, lb := strings.ToLower(a), strings.ToLower(b)
	la = strings.ReplaceAll(la, `\`, " ")
	lb = strings.ReplaceAll(lb, `\`, " ")
	if la != lb {
		return la < lb
	}
	return a < b
}
