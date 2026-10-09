package switchcontinuationinloop

import (
	"strconv"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// switchContinuationInLoop reports a bare `continue` inside a `switch` that
// sits in a loop: it only leaves the switch.
type switchContinuationInLoop struct{}

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
	var inner syntax.Node                           // innermost switch
	for p := c.Parent(); p != nil; p = p.Parent() { // D2
		if syntax.IsFuncLike(p) {
			return
		}
		switch p.(type) {
		case *syntax.Switch:
			switches++
			if inner == nil {
				inner = p
			}
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if switches == 0 {
				return
			}
			span := c.Span()
			if switchFollowed(ctx, inner, p) {
				// custos: `continue N` would also skip the statements
				// after the switch (`++$i;` keeping a position in step).
				ctx.Report(span, switchContinuationInLoopMsg)
				return
			}
			// Divergence: count nested switches so the fix reaches the loop.
			repl := "continue " + strconv.Itoa(switches+1) + ";"
			ctx.Report(span, switchContinuationInLoopMsg, diagnostic.Fix{
				Title: "Use '" + repl[:len(repl)-1] + "'",
				Edits: func() []diagnostic.TextEdit {
					return []diagnostic.TextEdit{{Span: span, NewText: repl}}
				},
			})
			return
		}
	}
}

// switchFollowed reports whether a statement follows the switch sw, or any
// statement enclosing it, inside loop: `continue` leaves the switch and runs
// it, `continue 2` skips it.
func switchFollowed(ctx *analysis.Context, sw, loop syntax.Node) bool {
	for n := sw; n != loop; n = n.Parent() {
		if st, ok := n.(syntax.Stmt); ok {
			if list, i, ok := astquery.StmtList(ctx.File, st); ok && i < len(list)-1 {
				return true
			}
		}
	}
	return false
}
