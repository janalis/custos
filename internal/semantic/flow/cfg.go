package flow

import (
	"sort"

	"custos/internal/php/syntax"
)

type (
	block struct {
		node        syntax.Node
		nodes       []syntax.Node
		next        []int
		exceptional int
		shared      bool
	}
	graph struct {
		blocks   []block
		complete bool
	}
	exits struct{ ret, thrown, brk, cont int }
)

func (g *graph) add(n syntax.Node, next ...int) int {
	if len(g.blocks) >= MaxBlocks {
		g.complete = false
		return 0
	}
	g.blocks = append(g.blocks, block{node: n, next: next, exceptional: -1})
	return len(g.blocks) - 1
}

func (g *graph) atom(n syntax.Node, next int, x exits) int {
	if next != 0 && !g.blocks[next].shared && len(g.blocks[next].nodes) > 0 && g.blocks[next].exceptional == x.thrown {
		g.blocks[next].nodes = append(g.blocks[next].nodes, n)
		return next
	}
	i := g.add(nil, next)
	if n != nil {
		g.blocks[i].exceptional = x.thrown
		g.blocks[i].nodes = []syntax.Node{n}
	}
	return i
}

func (g *graph) list(stmts []syntax.Stmt, next int, x exits) int {
	for i := len(stmts) - 1; i >= 0; i-- {
		next = g.stmt(stmts[i], next, x)
	}
	return next
}

func (g *graph) stmt(s syntax.Stmt, next int, x exits) int {
	if s == nil {
		return next
	}
	switch n := s.(type) {
	case *syntax.Block:
		return g.list(n.Stmts, next, x)
	case *syntax.Namespace:
		return g.list(n.Stmts, next, x)
	case *syntax.Function, *syntax.ClassLike:
		return next
	case *syntax.If:
		g.blocks[next].shared = true
		other := next
		if n.Else != nil {
			other = g.stmt(n.Else.Body, next, x)
		}
		for i := len(n.ElseIfs) - 1; i >= 0; i-- {
			el := n.ElseIfs[i]
			other = g.add(el.Cond, g.stmt(el.Body, next, x), other)
		}
		return g.add(n.Cond, g.stmt(n.Body, next, x), other)
	case *syntax.While:
		g.blocks[next].shared = true
		test := g.add(n.Cond)
		lx := x
		lx.brk, lx.cont = next, test
		g.blocks[test].next = []int{g.stmt(n.Body, test, lx), next}
		return test
	case *syntax.DoWhile:
		g.blocks[next].shared = true
		test := g.add(n.Cond)
		lx := x
		lx.brk, lx.cont = next, test
		body := g.stmt(n.Body, test, lx)
		g.blocks[test].next = []int{body, next}
		return body
	case *syntax.For:
		g.blocks[next].shared = true
		test := g.add(nil)
		loop := test
		for i := len(n.Loop) - 1; i >= 0; i-- {
			loop = g.atom(n.Loop[i], loop, x)
		}
		lx := x
		lx.brk, lx.cont = next, loop
		body := g.stmt(n.Body, loop, lx)
		cond := g.add(nil, body, next)
		for i := len(n.Cond) - 1; i >= 0; i-- {
			cond = g.atom(n.Cond[i], cond, x)
		}
		g.blocks[test].next = []int{cond}
		start := test
		for i := len(n.Init) - 1; i >= 0; i-- {
			start = g.atom(n.Init[i], start, x)
		}
		return start
	case *syntax.Foreach:
		g.blocks[next].shared = true
		test := g.add(n)
		lx := x
		lx.brk, lx.cont = next, test
		g.blocks[test].next = []int{g.stmt(n.Body, test, lx), next}
		return test
	case *syntax.Return:
		return g.atom(n, x.ret, x)
	case *syntax.ExprStmt:
		switch syntax.UnwrapParens(n.Expr).(type) {
		case *syntax.Throw:
			return g.atom(n, x.thrown, x)
		case *syntax.Exit:
			return g.atom(n, x.ret, x)
		}
	case *syntax.Break:
		if n.Num != nil {
			g.complete = false
		}
		return x.brk
	case *syntax.Continue:
		if n.Num != nil {
			g.complete = false
		}
		return x.cont
	case *syntax.Try:
		g.blocks[next].shared = true
		inside := x
		normal := next
		if n.Finally != nil {
			normal = g.stmt(n.Finally.Body, next, x)
			inside.ret = g.stmt(n.Finally.Body, x.ret, x)
			inside.thrown = g.stmt(n.Finally.Body, x.thrown, x)
			inside.brk = g.stmt(n.Finally.Body, x.brk, x)
			inside.cont = g.stmt(n.Finally.Body, x.cont, x)
		}
		if len(n.Catches) > 0 {
			choices := []int{inside.thrown}
			for _, c := range n.Catches {
				choices = append(choices, g.stmt(c.Body, normal, inside))
			}
			inside.thrown = g.add(nil, choices...)
		}
		return g.stmt(n.Body, normal, inside)
	case *syntax.Goto, *syntax.Switch:
		g.complete = false
	}
	return g.atom(s, next, x)
}

func (e *Env) scope(scope syntax.Node) *scopeResult {
	if r, ok := e.scopes[scope]; ok {
		return r
	}
	r := &scopeResult{values: map[syntax.Expr]Value{}, before: map[syntax.Node]map[uint32]State{}, callIndex: map[syntax.Node]int{}, complete: true}
	e.scopes[scope] = r
	g := graph{complete: true}
	exit := g.add(nil)
	x := exits{exit, exit, exit, exit}
	entry := exit
	if scope == nil {
		entry = g.list(e.file.Stmts, exit, x)
	} else {
		body := syntax.VariableScopeBody(scope)
		if st, ok := body.(syntax.Stmt); ok {
			entry = g.stmt(st, exit, x)
		} else if body != nil {
			entry = g.atom(body, exit, x)
		}
	}
	initial := frame{vars: map[string]Value{}, states: map[uint32]State{}}
	for i, p := range syntax.VariableScopeParams(scope) {
		if p.Var != nil {
			initial.vars[p.Var.Name] = Value{Complete: true, Identity: p.Var.Span().Start + 1, Sources: []Source{{Kind: "parameter", Parameter: i}}}
		}
	}
	inputs := map[int]frame{entry: initial}
	queue := []int{entry}
	queued := map[int]bool{entry: true}
	steps := 0
	for len(queue) > 0 && steps < MaxTransfers && r.operations <= MaxTransfers {
		i := queue[0]
		queue = queue[1:]
		delete(queued, i)
		steps++
		in := inputs[i]
		if !reserve(r, len(in.vars)+len(in.states)) {
			break
		}
		out := clone(in)
		b := g.blocks[i]
		var exceptional frame
		if b.exceptional > 0 {
			exceptional = clone(in)
		}
		if b.node != nil {
			e.transfer(b.node, &out, r)
		}
		for j := len(b.nodes) - 1; j >= 0; j-- {
			e.transfer(b.nodes[j], &out, r)
			if !r.complete || r.operations > MaxTransfers {
				break
			}
			if b.exceptional > 0 {
				exceptional = join(exceptional, out)
			}
		}
		propagate := func(target int, v frame) {
			old, exists := inputs[target]
			updated := v
			if exists {
				updated = join(old, v)
				if same(old, updated) {
					return
				}
			}
			if !reserve(r, len(updated.vars)+len(updated.states)) {
				return
			}
			inputs[target] = clone(updated)
			if !queued[target] {
				queue = append(queue, target)
				queued[target] = true
			}
		}
		for j, target := range b.next {
			branch := out
			if condition, ok := b.node.(syntax.Expr); ok && len(b.next) == 2 {
				branch = clone(out)
				e.refine(condition, &branch, j == 0)
			}
			propagate(target, branch)
		}
		if b.exceptional > 0 {
			propagate(b.exceptional, exceptional)
		}
	}
	r.complete = r.complete && g.complete && len(queue) == 0
	sort.Slice(r.calls, func(i, j int) bool { return r.calls[i].Node.Span().Start < r.calls[j].Node.Span().Start })
	return r
}
