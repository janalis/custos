package controlflow

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// invertedIfElseConstructs reports if/else pairs whose condition is negated.
type invertedIfElseConstructs struct{}

func init() { register(invertedIfElseConstructs{}) }

const invertedIfElseMsg = "Negated condition with an else branch; swap the branches and drop the negation."

func (invertedIfElseConstructs) ID() string { return "InvertedIfElseConstructs" }

func (invertedIfElseConstructs) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KElse} }

func (invertedIfElseConstructs) Check(ctx *analysis.Context, n syntax.Node) {
	els := n.(*syntax.Else)
	if els.Span().Len() == 0 {
		return
	}
	elseBody, ok := bracedBlock(ctx, els.Body) // D1
	if !ok {
		return
	}
	owner, ok := els.Parent().(*syntax.If)
	if !ok || owner.Alt {
		return
	}
	// D2
	cond, body := owner.Cond, owner.Body
	if k := len(owner.ElseIfs); k > 0 {
		cond, body = owner.ElseIfs[k-1].Cond, owner.ElseIfs[k-1].Body
	}
	partnerBody, ok := bracedBlock(ctx, body) // D3
	if !ok || cond == nil {
		return
	}
	var newCond string // D4
	switch c := cond.(type) {
	case *syntax.Unary:
		if c.Op.Kind != syntax.TExclaim {
			return
		}
		x := util.UnwrapParens(c.Expr)
		if _, isEmpty := x.(*syntax.Empty); isEmpty || x == nil {
			return
		}
		newCond = ctx.Text(x)
	case *syntax.Binary:
		if c.Op.Kind != syntax.TIsIdentical {
			return
		}
		var other syntax.Expr
		if v, ok := util.BoolConst(c.Left); ok && !v {
			other = c.Right
		} else if v, ok := util.BoolConst(c.Right); ok && !v {
			other = c.Left
		} else {
			return
		}
		x := util.UnwrapParens(other)
		atoms := ctx.TypeOf(x).Atoms()
		if len(atoms) != 1 {
			return
		}
		if atoms[0] == "bool" { // F1: `false !== X` is just X for a bool
			newCond = ctx.Text(x)
		} else {
			newCond = ctx.Text(c.Left) + " !== " + ctx.Text(c.Right)
		}
	default:
		return
	}
	kw, ok := util.FindToken(ctx.File, els.Span(), syntax.TElse)
	if !ok {
		return
	}
	lp, lok := invertedSignificantBefore(ctx.File, cond.Span().Start, syntax.TLParen)
	rp, rok := util.NextSignificant(ctx.File, cond.Span().End)
	var fixes []analysis.Fix
	if lok && rok && rp.Kind == syntax.TRParen {
		fixes = append(fixes, analysis.Fix{
			Title: "Swap the branches",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{
					{Span: syntax.Span{Start: lp.Start, End: rp.End}, NewText: "(" + newCond + ")"},
					{Span: partnerBody.Span(), NewText: ctx.Text(elseBody)},
					{Span: elseBody.Span(), NewText: ctx.Text(partnerBody)},
				}
			},
		})
	}
	ctx.Report(syntax.Span{Start: kw.Start, End: kw.End}, invertedIfElseMsg, fixes...)
}

// invertedSignificantBefore returns the last non-trivia token ending at or before off
// when it has kind k.
func invertedSignificantBefore(f *syntax.File, off uint32, k syntax.TokenKind) (syntax.Token, bool) {
	for i := util.TokenIndex(f, off) - 1; i >= 0; i-- {
		t := f.Tokens[i]
		if t.Kind.IsTrivia() {
			continue
		}
		return t, t.Kind == k
	}
	return syntax.Token{}, false
}
