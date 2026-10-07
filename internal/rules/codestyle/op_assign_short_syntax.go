package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// opAssignShortSyntax suggests compound assignments for `$v = $v op x`.
type opAssignShortSyntax struct{}

func init() { register(opAssignShortSyntax{}) }

func (opAssignShortSyntax) ID() string { return "OpAssignShortSyntax" }

func (opAssignShortSyntax) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }

// opAssignEligible lists the binary operators with a compound form.
var opAssignEligible = map[syntax.TokenKind]string{
	syntax.TPlus: "+", syntax.TMinus: "-", syntax.TMul: "*", syntax.TDiv: "/", syntax.TMod: "%",
	syntax.TDot: ".", syntax.TAmpersand: "&", syntax.TBar: "|", syntax.TCaret: "^",
	syntax.TSl: "<<", syntax.TSr: ">>",
}

func (opAssignShortSyntax) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Assign)
	if a.Op.Kind != syntax.TEqual || a.ByRef || a.Var == nil || a.Value == nil || a.Span().Len() == 0 {
		return
	}
	switch a.Var.(type) {
	case *syntax.Array, *syntax.List:
		return
	}
	bin, ok := syntax.UnwrapParens(a.Value).(*syntax.Binary)
	if !ok {
		return
	}
	op, ok := opAssignEligible[bin.Op.Kind]
	if !ok {
		return
	}
	// D2: walk the left spine.
	var frags []syntax.Expr
	var base syntax.Expr = bin
	for {
		b, ok := base.(*syntax.Binary)
		if !ok || b.Op.Kind != bin.Op.Kind {
			break
		}
		frags = append(frags, b.Right)
		base = b.Left
	}
	if _, ok := base.(*syntax.Binary); ok { // D3: stopped on another operator
		return
	}
	if !util.EquivalentFoldNames(ctx.File, base, a.Var) {
		return
	}
	if len(frags) > 1 && op != "+" && op != "." && op != "*" { // D4
		return
	}
	for _, f := range frags { // D5
		if _, ok := f.(*syntax.Binary); ok || f == nil {
			return
		}
	}
	if dim, ok := a.Var.(*syntax.ArrayDimFetch); ok { // E1
		if t := ctx.TypeOf(dim.Var); !t.IsUnknown() && t.Has("string") {
			return
		}
	}
	var sb strings.Builder
	sb.WriteString(ctx.Text(base))
	sb.WriteString(" " + op + "= ")
	for i := len(frags) - 1; i >= 0; i-- {
		sb.WriteString(ctx.Text(frags[i]))
		if i > 0 {
			sb.WriteString(" " + op + " ")
		}
	}
	repl := sb.String()
	span := a.Span()
	ctx.Report(span, "Use the compound form '"+repl+"'.", analysis.Fix{
		Title: "Use the compound assignment",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}
