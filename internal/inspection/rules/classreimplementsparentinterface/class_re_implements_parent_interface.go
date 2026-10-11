package classreimplementsparentinterface

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
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
	var gated map[string]bool // E4, built for the first candidate only
	for i, entry := range cl.Implements {
		if entry.Span().Len() == 0 {
			continue
		}
		iface := ix.Class(ctx.Names().Class(entry.Value, entry.Span().Start), ctx.PHP) // D2
		if iface == nil || iface.Kind != syntax.KindInterface {
			continue
		}
		k := criKey(iface.FQN)
		if !inherited[k] { // D5
			continue
		}
		if gated == nil {
			gated = criInherited(ix, parent, ctx.PHP)
		}
		if !gated[k] {
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
		ctx.ReportNode(cl.Implements[i], msgs[i], diagnostic.Fix{
			Title: "Remove the redundant interface",
			Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{edit} },
		})
	}
}

func criKey(fqn string) string { return strings.ToLower(strings.TrimPrefix(fqn, `\`)) }

// criInherited is the inherited set of parent (D4) without the links a
// built-in class only has in a later PHP version than ver (E4): the stubs
// record one interface list per class, so `Throwable extends Stringable`
// shows at every version, and a project polyfill of the interface would
// make the entry look redundant where it is the only thing providing it. A
// link leaving a built-in class is followed only when the stubs declare
// its target at ver.
func criInherited(ix *index.Index, parent *index.Class, ver phpversion.Version) map[string]bool {
	builtin := stubs.Index()
	seen := map[string]bool{criKey(parent.FQN): true}
	queue := []*index.Class{parent}
	for len(queue) > 0 && len(seen) < index.MaxAncestors {
		c := queue[0]
		queue = queue[1:]
		stub := builtin.Class(c.FQN, ver) == c
		links := c.Interfaces
		if c.Parent != "" {
			links = append([]string{c.Parent}, links...)
		}
		for _, l := range links {
			if seen[criKey(l)] || (stub && len(builtin.ClassDecls(l, ver)) == 0) {
				continue
			}
			if next := ix.Class(l, ver); next != nil {
				seen[criKey(l)] = true
				queue = append(queue, next)
			}
		}
	}
	return seen
}

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
		return diagnostic.TextEdit{Span: syntax.Span{Start: entry.Start, End: comma.End}}
	}
	// F3: the preceding comma plus the entry.
	comma, _ := astquery.FindToken(f, syntax.Span{Start: list[idx-1].Span().End, End: entry.Start}, syntax.TComma)
	return diagnostic.TextEdit{Span: syntax.Span{Start: comma.Start, End: entry.End}}
}
