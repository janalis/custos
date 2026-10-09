package unnecessaryissetarguments

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// unnecessaryIssetArguments reports isset() arguments implied by a deeper
// array access in the same isset().
type unnecessaryIssetArguments struct{}

func (unnecessaryIssetArguments) ID() string               { return "UnnecessaryIssetArguments" }
func (unnecessaryIssetArguments) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIsset} }

func (unnecessaryIssetArguments) Check(ctx *analysis.Context, n syntax.Node) {
	is := n.(*syntax.Isset)
	args := is.Vars
	if len(args) < 2 { // D1
		return
	}
	reported := make([]bool, len(args))
	for ci, c := range args { // D2
		access, ok := c.(*syntax.ArrayDimFetch)
		if !ok || reported[ci] {
			continue
		}
		for base := access.Var; base != nil; {
			for mi, m := range args {
				if mi == ci || reported[mi] || !astquery.EquivalentFoldNames(ctx.File, base, m) {
					continue
				}
				reported[mi] = true
			}
			next, ok := base.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			base = next.Var
		}
	}
	// F1 when a kept argument follows, F2 otherwise: this keeps the edits of
	// all redundant arguments of one isset() disjoint.
	keptAfter := false
	trailing := make([]bool, len(args))
	for i := len(args) - 1; i >= 0; i-- {
		trailing[i] = !keptAfter
		if !reported[i] {
			keptAfter = true
		}
	}
	for i, m := range args {
		if reported[i] {
			reportIssetArg(ctx, m, trailing[i])
		}
	}
}

func reportIssetArg(ctx *analysis.Context, m syntax.Expr, last bool) {
	f := ctx.File
	span := m.Span()
	ctx.Report(span, "Redundant isset() argument: a deeper array access already covers it.", diagnostic.Fix{
		Title: "Remove the argument",
		Edits: func() []diagnostic.TextEdit {
			// isset() arguments are always comma-separated (the parser
			// stops at a missing comma), and a trailing redundant argument
			// is never the first one (the covering access is kept).
			del := span
			if !last { // F1
				comma, _ := astquery.NextSignificant(f, span.End)
				del.End = comma.End
				del = astquery.WithTrailingWhitespace(f, del)
			} else { // F2
				i := astquery.TokenIndex(f, span.Start) - 1
				for f.Tokens[i].Kind.IsTrivia() {
					i--
				}
				del.Start = f.Tokens[i].Start
			}
			return []diagnostic.TextEdit{{Span: del}}
		},
	})
}
