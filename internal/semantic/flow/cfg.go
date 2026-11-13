package flow

import (
	"custos/internal/php/syntax"
)

// Compact successor indices can represent every permitted block ID.
const _ uint16 = MaxBlocks - 1

type (
	block struct {
		node        syntax.Node
		nodes       []syntax.Node
		more        int
		exceptional int
		next        [2]uint16
		nextCount   int8
		shared      bool
		head        bool
		choice      int8
		// pin names a variable that, on entry to the block, was just found
		// loosely equal to a string literal no request text can equal
		// without being that literal (see allowed), or the request table
		// element found so (see requestElement).
		pin string
		// differs holds a switch subject that, on entry to the block, was
		// just found loosely different from false or null, or loosely equal
		// to true: it is neither false nor null.
		differs syntax.Expr
	}
	graph struct {
		blocks   []block
		fanouts  [][]uint16
		complete bool
		proofs   map[int]*exceptionProof
		// full marks a graph that reached MaxBlocks.
		full bool
		env  *Env
		// body is what the graph is built from, nil for the statements of
		// the file. labels gives the block each label starts, and jumps the
		// `goto` statements written before their label, which gets them when
		// it is built. written holds the label statements of the body, once
		// a jump out of a try statement needs to know where one is.
		body    syntax.Node
		labels  map[string]int
		jumps   map[string][]pendingJump
		written map[string]*syntax.Label
		// first is where the first label of the body is written, plus one.
		first uint32
	}
	exits struct {
		ret, thrown, brk, cont int
		// levels lists the break and continue targets of the enclosing
		// loops and switches, innermost last, for `break 2`.
		levels [][2]int
		// finals lists the finally blocks of the enclosing try statements,
		// innermost last, which a `goto` out of them runs on its way.
		finals []final
	}
	// final is the finally block of a try statement and the exits around
	// that statement, where the block continues.
	final struct {
		try  *syntax.Try
		body syntax.Stmt
		out  exits
	}
	// pendingJump is a `goto` built before its label, with the finally
	// blocks it runs on its way there.
	pendingJump struct {
		block  int
		finals []final
	}
)

func (g *graph) add(n syntax.Node, next ...int) int {
	if len(g.blocks) >= MaxBlocks {
		g.complete, g.full = false, true
		return 0
	}
	g.blocks = append(g.blocks, block{node: n, exceptional: -1})
	at := len(g.blocks) - 1
	g.link(at, next...)
	return at
}

// link stores the usual one or two successors within the block. Exception
// dispatch can have any number of successors and uses a separate list.
func (g *graph) link(at int, next ...int) {
	b := &g.blocks[at]
	if len(next) <= len(b.next) {
		b.nextCount = int8(len(next))
		for i, target := range next {
			b.next[i] = uint16(target)
		}
		return
	}
	b.nextCount = -1
	b.more = len(g.fanouts)
	targets := make([]uint16, len(next))
	for i, target := range next {
		targets[i] = uint16(target)
	}
	g.fanouts = append(g.fanouts, targets)
}

// Successors are read only after construction; the block buffer can no longer
// grow or change the inline entries while a caller reads them.
func (g *graph) successors(at int) []uint16 {
	b := &g.blocks[at]
	if b.nextCount < 0 {
		return g.fanouts[b.more]
	}
	return b.next[:b.nextCount]
}

// test adds the block of a condition, or of what a loop iterates: its
// evaluation can throw into the enclosing handler like a statement.
func (g *graph) test(n syntax.Node, x exits, next ...int) int {
	i := g.add(n, next...)
	if i != 0 {
		g.blocks[i].exceptional = x.thrown
	}
	return i
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
	// A graph past its size limit is not followed: nothing more is built.
	if s == nil || g.full {
		return next
	}
	switch n := s.(type) {
	case *syntax.Block:
		return g.list(n.Stmts, next, x)
	case *syntax.Namespace:
		return g.list(n.Stmts, next, x)
	case *syntax.Function, *syntax.ClassLike:
		return next
	case *syntax.Declare:
		// The statements of a declare block run in place.
		if n.Body != nil {
			return g.stmt(n.Body, next, x)
		}
	case *syntax.If:
		g.blocks[next].shared = true
		other := next
		if n.Else != nil {
			other = g.stmt(n.Else.Body, next, x)
		}
		for i := len(n.ElseIfs) - 1; i >= 0; i-- {
			el := n.ElseIfs[i]
			other = g.test(el.Cond, x, g.stmt(el.Body, next, x), other)
		}
		return g.test(n.Cond, x, g.stmt(n.Body, next, x), other)
	case *syntax.While:
		g.blocks[next].shared = true
		test := g.test(n.Cond, x)
		g.blocks[test].head = true
		lx := x.inside(next, test)
		g.link(test, repeats(n.Cond, g.stmt(n.Body, test, lx), next)...)
		return test
	case *syntax.DoWhile:
		g.blocks[next].shared = true
		test := g.test(n.Cond, x)
		g.blocks[test].head = true
		lx := x.inside(next, test)
		body := g.stmt(n.Body, test, lx)
		g.link(test, repeats(n.Cond, body, next)...)
		// The body runs again: what is written before the loop is not part
		// of its first block.
		g.blocks[body].shared = true
		g.blocks[body].head = true
		return body
	case *syntax.For:
		g.blocks[next].shared = true
		test := g.add(nil)
		g.blocks[test].head = true
		g.blocks[test].shared = true
		loop := test
		for i := len(n.Loop) - 1; i >= 0; i-- {
			loop = g.atom(n.Loop[i], loop, x)
		}
		// `continue` runs the step, not what the body ends with.
		g.blocks[loop].shared = true
		lx := x.inside(next, loop)
		body := g.stmt(n.Body, loop, lx)
		// The last expression of the condition decides, like a while test:
		// it refines the body on its true outcome and the exit on its false
		// one. No condition is a true one: only a `break` leaves the loop.
		var deciding syntax.Node
		earlier := n.Cond
		if len(earlier) > 0 {
			deciding, earlier = earlier[len(earlier)-1], earlier[:len(earlier)-1]
		}
		cond := g.test(deciding, x, repeats(deciding, body, next)...)
		for i := len(earlier) - 1; i >= 0; i-- {
			cond = g.atom(earlier[i], cond, x)
		}
		g.link(test, cond)
		start := test
		for i := len(n.Init) - 1; i >= 0; i-- {
			start = g.atom(n.Init[i], start, x)
		}
		return start
	case *syntax.Foreach:
		g.blocks[next].shared = true
		test := g.test(n, x)
		g.blocks[test].head = true
		lx := x.inside(next, test)
		g.link(test, g.stmt(n.Body, test, lx), next)
		return test
	case *syntax.Return:
		return g.atom(n, x.ret, x)
	case *syntax.ExprStmt:
		switch expr := syntax.UnwrapParens(n.Expr).(type) {
		case *syntax.Throw:
			return g.atom(n, x.thrown, x)
		case *syntax.Exit:
			return g.atom(n, 0, x)
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
			if g.env.never(expr) {
				return g.atom(n, x.thrown, x)
			}
		case *syntax.Binary:
			// `test or exit;` and `test and throw …;` continue on one
			// outcome of the test only.
			and := expr.Op.Kind == syntax.TBooleanAnd || expr.Op.Kind == syntax.TAnd
			or := expr.Op.Kind == syntax.TBooleanOr || expr.Op.Kind == syntax.TOr
			if leave, leaves := g.leave(expr.Right, x); leaves && and {
				return g.test(expr.Left, x, leave, next)
			} else if leaves && or {
				return g.test(expr.Left, x, next, leave)
			}
		case *syntax.Ternary:
			if leave, leaves := g.leave(expr.Else, x); leaves && expr.Then == nil {
				return g.test(expr.Cond, x, next, leave)
			}
		}
	case *syntax.Break:
		return g.leaveLevels(n.Num, x, 0)
	case *syntax.Continue:
		return g.leaveLevels(n.Num, x, 1)
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
			inside.finals = append(inside.finals[:len(inside.finals):len(inside.finals)], final{try: n, body: n.Finally.Body, out: x})
			// A jump over several levels runs the finally block on its way
			// to each of them: the copies are built when one is written.
			inside.levels = nil
			if len(x.levels) > 0 && jumpsLevels(n) {
				inside.levels = make([][2]int, len(x.levels))
				inside.levels[len(x.levels)-1] = [2]int{inside.brk, inside.cont}
				for i, outer := range x.levels[:len(x.levels)-1] {
					inside.levels[i] = [2]int{g.stmt(n.Finally.Body, outer[0], x), g.stmt(n.Finally.Body, outer[1], x)}
				}
			}
		}
		if len(n.Catches) > 0 {
			choices := []int{inside.thrown}
			for _, c := range n.Catches {
				body := g.stmt(c.Body, normal, inside)
				if c.Var != nil {
					// The clause rebinds its variable before the body runs.
					body = g.atom(c, body, inside)
				}
				choices = append(choices, body)
			}
			inside.thrown = g.add(nil, choices...)
		}
		return g.stmt(n.Body, normal, inside)
	case *syntax.Switch:
		return g.choose(n, next, x)
	case *syntax.Label:
		// Nothing written before the label runs after a jump to it.
		at := g.add(nil, next)
		g.blocks[at].shared = true
		if g.labels == nil {
			g.labels = map[string]int{}
		}
		g.labels[n.Name.Value] = at
		g.first = n.Span().Start + 1
		for _, jump := range g.jumps[n.Name.Value] {
			g.link(jump.block, g.through(jump.finals, at))
		}
		delete(g.jumps, n.Name.Value)
		return at
	case *syntax.Goto:
		leaves := g.leaves(n.Label.Value, x.finals)
		if at, built := g.labels[n.Label.Value]; built {
			return g.add(nil, g.through(leaves, at))
		}
		jump := g.add(nil)
		if g.jumps == nil {
			g.jumps = map[string][]pendingJump{}
		}
		g.jumps[n.Label.Value] = append(g.jumps[n.Label.Value], pendingJump{block: jump, finals: leaves})
		return jump
	}
	return g.atom(s, next, x)
}

// once marks the blocks a scope visits once at most: no edge leads from a
// block at or below one to a block at or above it, so once the block is
// visited, as the highest waiting, nothing visited after it reaches it
// again. Its input is not kept for a later arrival to be absorbed in: the
// visit works on it in place. The exit blocks keep theirs, read when the
// scope is done, and a sweep over labels visits blocks out of that order.
func (g *graph) once(sweep bool, keep ...int) []bool {
	if sweep {
		return make([]bool, len(g.blocks))
	}
	spans := make([]int, len(g.blocks)+1)
	back := func(from, to int) {
		if to >= from {
			spans[from]++
			spans[to+1]--
		}
	}
	for j, b := range g.blocks {
		for _, t := range g.successors(j) {
			back(j, int(t))
		}
		if b.exceptional >= 0 {
			back(j, b.exceptional)
		}
	}
	once := make([]bool, len(g.blocks))
	depth := 0
	for i := range once {
		depth += spans[i]
		once[i] = depth == 0
	}
	for _, i := range keep {
		once[i] = false
	}
	return once
}
