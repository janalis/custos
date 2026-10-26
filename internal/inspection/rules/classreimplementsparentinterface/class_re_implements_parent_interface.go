package classreimplementsparentinterface

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// classReImplementsParentInterface reports `implements` entries already
// provided by the parent class chain.
type classReImplementsParentInterface struct{}

func (classReImplementsParentInterface) ID() string { return "ClassReImplementsParentInterface" }
func (classReImplementsParentInterface) Semantic()  {}
func (classReImplementsParentInterface) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClassLike}
}

func (classReImplementsParentInterface) Check(ctx *analysis.Context, n syntax.Node) {
	cl := n.(*syntax.ClassLike)
	if cl.ClassKind != syntax.KindClass || len(cl.Implements) == 0 || len(cl.Extends) == 0 { // D1
		return
	}
	ix := ctx.Index()
	parent := ix.Class(ctx.Names().ParentFQN(cl), ctx.PHP) // D3
	if parent == nil {
		return
	}
	inherited := map[string]bool{} // D4
	for i, a := range ix.Ancestors(parent.FQN, ctx.PHP) {
		if i == 0 {
			continue // the parent itself
		}
		inherited[criKey(a.FQN)] = true
	}
	f := ctx.File
	redundant := make([]bool, len(cl.Implements))
	msgs := make([]string, len(cl.Implements))
	for i, entry := range cl.Implements {
		if entry.Span().Len() == 0 {
			continue
		}
		iface := ix.Class(ctx.Names().Class(entry.Value, entry.Span().Start), ctx.PHP) // D2
		if iface == nil || iface.Kind != syntax.KindInterface || ctx.PHP != 0 && !iface.Avail.In(ctx.PHP) {
			continue
		}
		k := criKey(iface.FQN)
		if !inherited[k] { // D5
			continue
		}
		redundant[i] = true
		msgs[i] = "'\\" + strings.TrimPrefix(iface.FQN, `\`) + "' is already implemented by '\\" + strings.TrimPrefix(parent.FQN, `\`) + "'; remove it here."
	}
	for i := range cl.Implements {
		if !redundant[i] {
			continue
		}
		edit := criRemoveEdit(f, cl.Extends[0], cl.Implements, redundant, i)
		if astquery.HasComment(f, edit.Span) { // F4: report only
			ctx.ReportNode(cl.Implements[i], msgs[i])
			continue
		}
		entry := cl.Implements[i]
		ctx.ReportNode(entry, msgs[i], diagnostic.Fix{
			Title: "Remove the redundant interface",
			Edits: func() []diagnostic.TextEdit {
				if imp, ok := criImportEdit(f, cl, entry); ok { // F5
					return []diagnostic.TextEdit{imp, edit}
				}
				return []diagnostic.TextEdit{edit}
			},
		})
	}
}

func criKey(fqn string) string { return strings.ToLower(strings.TrimPrefix(fqn, `\`)) }

// criRemoveEdit deletes entry idx of an implements list (F1–F3). The
// separator removed with an entry is chosen so that the fixes of all
// redundant entries of one list never overlap and compose when applied
// together: an entry followed by a kept entry takes its following comma,
// otherwise its preceding one; when every entry is redundant, the first
// entry's fix drops the whole clause.
func criRemoveEdit(f *syntax.File, ext *syntax.Name, list []*syntax.Name, redundant []bool, idx int) diagnostic.TextEdit {
	entry := list[idx].Span()
	all := true
	for _, r := range redundant {
		all = all && r
	}
	if len(list) == 1 || (all && idx == 0) { // F1: drop the whole clause
		end := list[len(list)-1].Span().End
		text := " "
		if int(end) < len(f.Src) && astquery.IsSpace(f.Src[end]) {
			text = ""
		}
		return diagnostic.TextEdit{Span: syntax.Span{Start: ext.Span().End, End: end}, NewText: text}
	}
	keptAfter := false
	for _, r := range redundant[idx+1:] {
		keptAfter = keptAfter || !r
	}
	// The parser only continues a name list after a comma, so consecutive
	// entries are always separated by one.
	if keptAfter { // F2: entry plus the following comma
		comma, _ := astquery.NextSignificant(f, entry.End)
		span := syntax.Span{Start: entry.Start, End: comma.End}
		// An entry on a line of its own takes the line with it.
		indent := astquery.IndentBefore(f.Src, entry.Start)
		end := int(comma.End)
		for end < len(f.Src) && (f.Src[end] == ' ' || f.Src[end] == '\t' || f.Src[end] == '\r') {
			end++
		}
		if start := entry.Start - uint32(len(indent)); f.Src[start-1] == '\n' && end < len(f.Src) && f.Src[end] == '\n' {
			span = syntax.Span{Start: start, End: uint32(end) + 1}
		}
		return diagnostic.TextEdit{Span: span}
	}
	// F3: the preceding comma plus the entry.
	comma, _ := astquery.FindToken(f, syntax.Span{Start: list[idx-1].Span().End, End: entry.Start}, syntax.TComma)
	return diagnostic.TextEdit{Span: syntax.Span{Start: comma.Start, End: entry.End}}
}

// criImportEdit deletes the `use` import of an unqualified entry when that
// import names only it and the alias appears nowhere else in the file
// (F5): in the class's namespace, the import, the entry and nothing else
// mention it.
func criImportEdit(f *syntax.File, cl *syntax.ClassLike, entry *syntax.Name) (diagnostic.TextEdit, bool) {
	alias := entry.Value
	if strings.ContainsRune(alias, '\\') {
		return diagnostic.TextEdit{}, false
	}
	stmts, scope := f.Stmts, f.Src
	if ns, ok := cl.Parent().(*syntax.Namespace); ok { // an import only reaches its namespace
		stmts, scope = ns.Stmts, f.Src[ns.Span().Start:ns.Span().End]
	}
	var use *syntax.Use
	for _, st := range stmts {
		u, ok := st.(*syntax.Use)
		if !ok || u.Type != syntax.UseNormal || u.Prefix != nil || len(u.Items) != 1 {
			continue
		}
		it := u.Items[0]
		name := it.Name.Value[strings.LastIndexByte(it.Name.Value, '\\')+1:]
		if it.Alias != nil {
			name = it.Alias.Value
		}
		if strings.EqualFold(name, alias) {
			use = u
		}
	}
	if use == nil || criMentions(scope, alias) != 2 {
		return diagnostic.TextEdit{}, false
	}
	span := use.Span()
	start := span.Start - uint32(len(astquery.IndentBefore(f.Src, span.Start)))
	end := int(span.End)
	for end < len(f.Src) && (f.Src[end] == ' ' || f.Src[end] == '\t' || f.Src[end] == '\r') {
		end++
	}
	if f.Src[start-1] == '\n' && end < len(f.Src) && f.Src[end] == '\n' { // the import fills its line
		span = syntax.Span{Start: start, End: uint32(end) + 1}
		if f.Src[start-2] == '\n' && end+1 < len(f.Src) && f.Src[end+1] == '\n' { // between blank lines: keep one
			span.End++
		}
	}
	return diagnostic.TextEdit{Span: span}, true
}

// criMentions counts the case-insensitive whole-word occurrences of name in
// src, comments and strings included.
func criMentions(src []byte, name string) int {
	low, key := strings.ToLower(string(src)), strings.ToLower(name)
	n := 0
	for i := 0; ; {
		j := strings.Index(low[i:], key)
		if j < 0 {
			return n
		}
		at, end := i+j, i+j+len(key)
		if (at == 0 || !astquery.NameByte(low[at-1])) && (end == len(low) || !astquery.NameByte(low[end])) {
			n++
		}
		i = end
	}
}
