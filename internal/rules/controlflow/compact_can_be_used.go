package controlflow

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// compactCanBeUsed reports array literals mapping each string key to the
// variable of the same name, which compact() writes more briefly.
type compactCanBeUsed struct{}

func init() { register(compactCanBeUsed{}) }

func (compactCanBeUsed) ID() string { return "CompactCanBeUsed" }

func (compactCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArray} }

func (compactCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	arr := n.(*syntax.Array)
	if len(arr.Items) < 2 || isDestructuringTarget(arr) { // D4, D2
		return
	}
	keys := make([]string, 0, len(arr.Items))
	for _, it := range arr.Items { // D3
		if it == nil || it.ByRef || it.Unpack || it.Key == nil {
			return
		}
		lit, ok := it.Key.(*syntax.Literal)
		if !ok || lit.LitKind != syntax.LitString || len(lit.Raw) < 2 || (lit.Raw[0] != '\'' && lit.Raw[0] != '"') {
			return
		}
		key, ok := util.StringLiteralValue(lit.Raw) // escapes decoded
		if !ok {
			return
		}
		v, ok := it.Value.(*syntax.Variable)
		if !ok || v.NameExpr != nil || v.Name != key {
			return
		}
		keys = append(keys, lit.Raw)
	}
	// A namespaced compact() would capture a bare call.
	repl := util.QualifiedBuiltin(ctx, "compact", arr.Span().Start) + "(" + strings.Join(keys, ", ") + ")"
	s := arr.Span()
	first := syntax.Span{Start: s.Start, End: s.Start + 1}
	if !arr.Short {
		first.End = s.Start + uint32(len("array"))
	}
	ctx.Report(first, "Replace with '"+repl+"'.", analysis.Fix{
		Title: "Use compact()",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: s, NewText: repl}}
		},
	})
}

// isDestructuringTarget reports whether arr (or an array/list it is nested
// in) is written to: the left side of an assignment or a foreach key/value.
// Upstream only checks the direct assignment target (see spec Divergences).
func isDestructuringTarget(arr *syntax.Array) bool {
	var cur syntax.Node = arr
	for {
		item, ok := cur.Parent().(*syntax.ArrayItem)
		if !ok || item.Value != cur {
			break
		}
		switch outer := item.Parent().(type) {
		case *syntax.Array, *syntax.List:
			cur = outer
			continue
		}
		break
	}
	switch p := cur.Parent().(type) {
	case *syntax.Assign:
		return p.Var == cur
	case *syntax.Foreach:
		return p.Value == cur || p.Key == cur
	case *syntax.List:
		return true
	}
	return false
}
