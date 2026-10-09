package loopwhichdoesnotloop

import (
	"strconv"
	"strings"

	"custos/internal/inspection/astquery"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// loopWhichDoesNotLoop reports braced loops whose body always leaves on the
// first pass (or is empty).
type loopWhichDoesNotLoop struct{}

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
	if fe, ok := n.(*syntax.Foreach); ok && !isThrow { // D4
		t := ctx.TypeOf(fe.Expr)
		// custos: an empty body iterates for the iterator's side effects
		// (initialising a lazy collection) unless the subject is known not
		// to be an object.
		if last == nil && (t.IsUnknown() || t.Has("object") || t.Has("iterable") || t.Has("mixed") || len(t.Classes()) > 0) {
			return
		}
		// custos: `foreach ($this as …)` in a trait iterates the using
		// class (CakePHP's CollectionTrait::isEmpty()).
		if v, ok := syntax.UnwrapParens(fe.Expr).(*syntax.Variable); ok && v.Name == "this" {
			if cl := syntax.EnclosingClass(fe); cl != nil && cl.ClassKind == syntax.KindTrait {
				return
			}
		}
		for _, a := range t.Atoms() {
			switch strings.ToLower(a) {
			case `\generator`, `\traversable`, `\iterator`, "iterable":
				return
			}
			if strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "]") && ctx.Index().IsSubtype(a, `\Traversable`, ctx.PHP) {
				return
			}
		}
	}
	ctx.Report(astquery.KeywordSpan(ctx.File, n), loopNoLoopMsg)
}

func loopNoLoopIsLoop(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.Foreach, *syntax.For, *syntax.While, *syntax.DoWhile:
		return true
	}
	return false
}

// continuesLoop reports whether a `continue` inside body targets loop.
func continuesLoop(body, loop syntax.Node) bool {
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
