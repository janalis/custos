package architecture

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// classReImplementsParentInterface reports `implements` entries already
// provided by the parent class chain.
type classReImplementsParentInterface struct{}

func init() { register(classReImplementsParentInterface{}) }

func (classReImplementsParentInterface) ID() string { return "ClassReImplementsParentInterface" }

func (classReImplementsParentInterface) Semantic() {}

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
		inherited[strings.ToLower(strings.TrimPrefix(a.FQN, `\`))] = true
	}
	f := ctx.File
	redundant := make([]bool, len(cl.Implements))
	msgs := make([]string, len(cl.Implements))
	for i, entry := range cl.Implements {
		if entry.Span().Len() == 0 {
			continue
		}
		iface := ix.Class(ctx.Names().Class(entry.Value, entry.Span().Start), ctx.PHP) // D2
		if iface == nil || iface.Kind != syntax.KindInterface {
			continue
		}
		if !inherited[strings.ToLower(strings.TrimPrefix(iface.FQN, `\`))] { // D5
			continue
		}
		redundant[i] = true
		msgs[i] = "'\\" + strings.TrimPrefix(iface.FQN, `\`) + "' is already implemented by '\\" + strings.TrimPrefix(parent.FQN, `\`) + "'; remove it here."
	}
	for i, entry := range cl.Implements {
		if !redundant[i] {
			continue
		}
		idx, ext := i, cl.Extends[0]
		ctx.ReportNode(entry, msgs[i], analysis.Fix{
			Title: "Remove the redundant interface",
			Edits: func() []analysis.TextEdit { return criRemoveEdits(f, ext, cl.Implements, redundant, idx) },
		})
	}
}

// criRemoveEdits deletes entry idx of an implements list (F1–F3). The
// separator removed with an entry is chosen so that the fixes of all
// redundant entries of one list never overlap and compose when applied
// together: an entry followed by a kept entry takes its following comma,
// otherwise its preceding one; when every entry is redundant, the first
// entry's fix drops the whole clause.
func criRemoveEdits(f *syntax.File, ext *syntax.Name, list []*syntax.Name, redundant []bool, idx int) []analysis.TextEdit {
	entry := list[idx].Span()
	all := true
	for _, r := range redundant {
		all = all && r
	}
	if len(list) == 1 || (all && idx == 0) { // F1: drop the whole clause
		end := list[len(list)-1].Span().End
		text := " "
		if int(end) < len(f.Src) && util.IsSpace(f.Src[end]) {
			text = ""
		}
		return []analysis.TextEdit{{Span: syntax.Span{Start: ext.Span().End, End: end}, NewText: text}}
	}
	keptAfter := false
	for _, r := range redundant[idx+1:] {
		keptAfter = keptAfter || !r
	}
	// The parser only continues a name list after a comma, so consecutive
	// entries are always separated by one.
	if keptAfter { // F2: entry plus the following comma
		comma, _ := util.NextSignificant(f, entry.End)
		return []analysis.TextEdit{{Span: syntax.Span{Start: entry.Start, End: comma.End}}}
	}
	// F3: the preceding comma plus the entry.
	comma, _ := util.FindToken(f, syntax.Span{Start: list[idx-1].Span().End, End: entry.Start}, syntax.TComma)
	return []analysis.TextEdit{{Span: syntax.Span{Start: comma.Start, End: entry.End}}}
}
