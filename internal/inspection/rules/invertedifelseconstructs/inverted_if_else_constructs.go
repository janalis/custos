package invertedifelseconstructs

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// invertedIfElseConstructs reports if/else pairs whose condition is negated.
type invertedIfElseConstructs struct{}

const invertedIfElseMsg = "Negated condition with an else branch; swap the branches and drop the negation."

func (invertedIfElseConstructs) ID() string               { return "InvertedIfElseConstructs" }
func (invertedIfElseConstructs) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KElse} }
func (invertedIfElseConstructs) Check(ctx *analysis.Context, n syntax.Node) {
	els := n.(*syntax.Else)
	elseBody, ok := astquery.BracedBlock(ctx.Src, els.Body) // D1
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
	partnerBody, ok := astquery.BracedBlock(ctx.Src, body) // D3
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
		if v, ok := astquery.BoolConst(c.Left); ok && !v {
			other = c.Right
		} else if v, ok := astquery.BoolConst(c.Right); ok && !v {
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
	rp, rok := astquery.NextSignificant(ctx.File, cond.Span().End)
	var fixes []diagnostic.Fix
	if lok && rok && rp.Kind == syntax.TRParen {
		fixes = append(fixes, diagnostic.Fix{
			Title: "Swap the branches",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{
					{Span: syntax.Span{Start: lp.Start, End: rp.End}, NewText: "(" + newCond + ")"},
					{Span: partnerBody.Span(), NewText: ctx.Text(elseBody)},
					{Span: elseBody.Span(), NewText: ctx.Text(partnerBody)},
				}
			},
		})
	}
	ctx.Report(astquery.KeywordSpan(ctx.File, els), invertedIfElseMsg, fixes...)
}

// invertedSignificantBefore returns the last non-trivia token ending at or
// before off (a condition always follows the `if`/`elseif` keyword) and
// whether it has kind k.
func invertedSignificantBefore(f *syntax.File, off uint32, k syntax.TokenKind) (syntax.Token, bool) {
	i := astquery.TokenIndex(f, off) - 1
	for f.Tokens[i].Kind.IsTrivia() {
		i--
	}
	return f.Tokens[i], f.Tokens[i].Kind == k
}
