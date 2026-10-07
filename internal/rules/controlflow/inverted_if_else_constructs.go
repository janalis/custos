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
	elseBody, ok := bracedBlock(ctx, els.Body) // D1
	if !ok || len(elseBody.Stmts) == 0 {
		// An empty else is dead weight: swapping would leave an empty if
		// body; dropping the else is the better change (custos diverges).
		return
	}
	owner := els.Parent().(*syntax.If) // alternative syntax fails D3
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
		x := syntax.UnwrapParens(c.Expr)
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
		x := syntax.UnwrapParens(other)
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
	ctx.Report(keywordSpan(ctx, els), invertedIfElseMsg, fixes...)
}

// invertedSignificantBefore returns the last non-trivia token ending at or
// before off (a condition always follows the `if`/`elseif` keyword) and
// whether it has kind k.
func invertedSignificantBefore(f *syntax.File, off uint32, k syntax.TokenKind) (syntax.Token, bool) {
	i := util.TokenIndex(f, off) - 1
	for f.Tokens[i].Kind.IsTrivia() {
		i--
	}
	return f.Tokens[i], f.Tokens[i].Kind == k
}
