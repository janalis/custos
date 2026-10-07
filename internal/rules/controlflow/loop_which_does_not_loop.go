package controlflow

import (
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// loopWhichDoesNotLoop reports braced loops whose body always leaves on the
// first pass (or is empty).
type loopWhichDoesNotLoop struct{}

func init() { register(loopWhichDoesNotLoop{}) }

const loopNoLoopMsg = "Loop body exits on the first iteration; the loop never repeats."

func (loopWhichDoesNotLoop) ID() string { return "LoopWhichDoesNotLoop" }

func (loopWhichDoesNotLoop) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach, syntax.KFor, syntax.KWhile, syntax.KDoWhile}
}

func (loopWhichDoesNotLoop) Check(ctx *analysis.Context, n syntax.Node) {
	if strings.HasSuffix(ctx.File.Path, ".blade.php") {
		return
	}
	var body syntax.Stmt
	switch l := n.(type) {
	case *syntax.Foreach:
		body = l.Body
	case *syntax.For:
		body = l.Body
	case *syntax.While:
		body = l.Body
	case *syntax.DoWhile:
		body = l.Body
	}
	blk, ok := body.(*syntax.Block) // D1
	if !ok || blk.Span().Len() == 0 {
		return
	}
	var last syntax.Stmt // D2
	for i := len(blk.Stmts) - 1; i >= 0; i-- {
		if blk.Stmts[i].Span().Len() > 0 {
			last = blk.Stmts[i]
			break
		}
	}
	if last == nil {
		// An empty while/do/for body still loops: the condition (or step)
		// does the work (`while (@ob_end_flush()) {}`).
		if _, ok := n.(*syntax.Foreach); !ok {
			return
		}
	}
	isThrow := false
	if last != nil {
		switch x := last.(type) {
		case *syntax.Break, *syntax.Return:
		case *syntax.ExprStmt:
			if _, ok := x.Expr.(*syntax.Throw); !ok {
				return
			}
			isThrow = true
		default:
			return
		}
	}
	if continuesLoop(blk, n) { // D3
		return
	}
	if fe, ok := n.(*syntax.Foreach); ok && last != nil && !isThrow { // D4
		for _, a := range ctx.TypeOf(fe.Expr).Atoms() {
			switch strings.ToLower(a) {
			case `\generator`, `\traversable`, `\iterator`, "iterable":
				return
			}
			if strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "]") && ctx.Index().IsSubtype(a, `\Traversable`, ctx.PHP) {
				return
			}
		}
	}
	ctx.Report(keywordSpan(ctx, n), loopNoLoopMsg)
}

func loopNoLoopIsLoop(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.Foreach, *syntax.For, *syntax.While, *syntax.DoWhile:
		return true
	}
	return false
}

// continuesLoop reports whether a `continue` inside body targets loop.
func continuesLoop(body syntax.Node, loop syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if found {
			return false
		}
		if syntax.IsFuncLike(x) {
			return false
		}
		c, ok := x.(*syntax.Continue)
		if !ok {
			return true
		}
		level := 1
		if c.Num != nil {
			if lit, ok := c.Num.(*syntax.Literal); ok {
				if v, err := strconv.Atoi(lit.Raw); err == nil {
					level = v
				}
			}
		}
		k := 0
		// The walk never crosses a function boundary: Inspect does not enter
		// function-likes, so the loop is reached first.
		for p := c.Parent(); p != nil; p = p.Parent() {
			_, isSwitch := p.(*syntax.Switch)
			if isSwitch || loopNoLoopIsLoop(p) {
				// PHP counts a switch as a loop structure for continue.
				k++
				if p == loop {
					if level == k {
						found = true
					}
					break
				}
			}
		}
		return true
	})
	return found
}
