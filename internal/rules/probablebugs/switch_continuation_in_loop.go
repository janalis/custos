package probablebugs

import (
	"strconv"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// switchContinuationInLoop reports a bare `continue` inside a `switch` that
// sits in a loop: it only leaves the switch.
type switchContinuationInLoop struct{}

func init() { register(switchContinuationInLoop{}) }

const switchContinuationInLoopMsg = "Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop."

func (switchContinuationInLoop) ID() string { return "SwitchContinuationInLoop" }

func (switchContinuationInLoop) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KContinue}
}

func (switchContinuationInLoop) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.Continue)
	if c.Num != nil || c.Span().Len() == 0 { // D1
		return
	}
	switches := 0
	for p := c.Parent(); p != nil; p = p.Parent() { // D2
		if util.IsFuncLike(p) {
			return
		}
		switch p.(type) {
		case *syntax.Switch:
			switches++
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if switches == 0 {
				return
			}
			span := c.Span()
			// Divergence: count nested switches so the fix reaches the loop.
			repl := "continue " + strconv.Itoa(switches+1) + ";"
			ctx.Report(span, switchContinuationInLoopMsg, analysis.Fix{
				Title: "Use '" + repl[:len(repl)-1] + "'",
				Edits: func() []analysis.TextEdit {
					return []analysis.TextEdit{{Span: span, NewText: repl}}
				},
			})
			return
		}
	}
}
