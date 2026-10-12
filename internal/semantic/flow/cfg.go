package flow

import (
	"sort"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
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
		env      *Env
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
		// The body runs again: what is written before the loop is not part
		// of its first block.
		g.blocks[body].shared = true
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
		// The last expression of the condition decides, like a while test:
		// it refines the body on its true outcome and the exit on its false
		// one. No condition is a true one.
		var deciding syntax.Node
		earlier := n.Cond
		if len(earlier) > 0 {
			deciding, earlier = earlier[len(earlier)-1], earlier[:len(earlier)-1]
		}
		cond := g.add(deciding, body, next)
		for i := len(earlier) - 1; i >= 0; i-- {
			cond = g.atom(earlier[i], cond, x)
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
				return g.add(expr.Left, leave, next)
			} else if leaves && or {
				return g.add(expr.Left, next, leave)
			}
		case *syntax.Ternary:
			if leave, leaves := g.leave(expr.Else, x); leaves && expr.Then == nil {
				return g.add(expr.Cond, next, leave)
			}
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
	case *syntax.Goto, *syntax.Switch:
		g.complete = false
	}
	return g.atom(s, next, x)
}

// leave builds the block of an expression that never completes: a throw, an
// exit or a call that does not return.
func (g *graph) leave(x syntax.Expr, out exits) (int, bool) {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Exit:
		return g.atom(n, 0, out), true
	case *syntax.Throw:
		return g.atom(n, out.thrown, out), true
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		if g.env.never(n) {
			return g.atom(n, out.thrown, out), true
		}
	}
	return 0, false
}

// never reports a call that does not return: its callee is declared `never`,
// or is a function, or a method no subclass can replace, every path of which
// ends in a throw, an exit or such a call.
func (e *Env) never(call syntax.Node) bool {
	declared, name := "", ""
	if fn, ok := call.(*syntax.FuncCall); ok {
		if f := e.types.ResolveFunction(fn); f != nil {
			declared, name = f.Return, f.FQN
		}
	} else if m := e.method(call); m != nil {
		declared = m.Return
		if m.Final || m.Visibility == index.Private {
			name = m.Class + "::" + m.Name
		}
	}
	if strings.EqualFold(strings.TrimPrefix(declared, `\`), "never") {
		return true
	}
	summary, ok := e.contract(strings.ToLower(name))
	return ok && summary.Never
}

// method resolves the declaration a method call names through `$this` or a
// typed receiver, or a static call through a class name, `self` or `static`.
func (e *Env) method(call syntax.Node) *index.Method {
	switch n := call.(type) {
	case *syntax.MethodCall:
		id, ok := n.Name.(*syntax.Identifier)
		if classes := e.types.TypeOf(n.Var).Classes(); ok && len(classes) == 1 {
			return e.types.Index.FindMethod(classes[0], id.Value, e.types.PHP)
		}
	case *syntax.StaticCall:
		id, ok := n.Name.(*syntax.Identifier)
		class, named := n.Class.(*syntax.Name)
		if !ok || !named {
			return nil
		}
		fqn := e.types.Names.Class(class.Value, class.Span().Start)
		switch strings.ToLower(class.Value) {
		case "self", "static":
			fqn = e.types.ClassFQN(syntax.EnclosingClass(n))
		}
		return e.types.Index.FindMethod(fqn, id.Value, e.types.PHP)
	}
	return nil
}

func (e *Env) scope(scope syntax.Node) *scopeResult {
	if r, ok := e.scopes[scope]; ok {
		return r
	}
	r := &scopeResult{values: map[syntax.Expr]Value{}, before: map[syntax.Node]map[uint32]State{}, callIndex: map[syntax.Node]int{}, complete: true}
	e.scopes[scope] = r
	// Block 0 ends the scope however it is left; a normal completion passes
	// through the block before it, which tells whether the scope can return.
	g := graph{complete: true, env: e}
	exit := g.add(nil)
	done := g.add(nil, exit)
	x := exits{done, exit, done, done}
	entry := done
	if scope == nil {
		entry = g.list(e.file.Stmts, done, x)
	} else {
		body := syntax.VariableScopeBody(scope)
		if st, ok := body.(syntax.Stmt); ok {
			entry = g.stmt(st, done, x)
		} else if body != nil {
			entry = g.atom(body, done, x)
		}
	}
	initial := frame{vars: map[string]Value{}, states: map[uint32]State{}}
	e.bindings(scope, r, &initial)
	for i, p := range syntax.VariableScopeParams(scope) {
		if p.Var != nil {
			initial.vars[p.Var.Name] = Value{Complete: true, Identity: p.Var.Span().Start + 1, Sources: []Source{{Kind: "parameter", Parameter: i}}}
		}
	}
	// A block is visited when no block before it in the source waits: blocks
	// are numbered from the end of the scope, a loop body above its test, so
	// a loop settles before what follows it runs.
	inputs := make([]frame, len(g.blocks))
	waiting := make([]bool, len(g.blocks))
	inputs[entry], waiting[entry] = initial, true
	top, queued, steps := entry, 1, 0
	// A block input is retained until the scope is done: only its growth is
	// charged to the memory budget. owned marks a frame nothing else uses.
	propagate := func(target int, v frame, owned bool) {
		input := &inputs[target]
		if input.vars == nil {
			if !retain(r, len(v.vars)+len(v.states)) {
				return
			}
			if !owned {
				r.copies += len(v.vars) + len(v.states)
				v = clone(v)
			}
			*input = v
		} else {
			before := len(input.vars) + len(input.states)
			changed, work := input.absorb(v)
			r.copies += work
			if !changed || !retain(r, max(0, len(input.vars)+len(input.states)-before)) {
				return
			}
		}
		if !waiting[target] {
			waiting[target] = true
			queued++
			top = max(top, target)
		}
	}
	for queued > 0 && steps < MaxTransfers && !r.spent && r.operations <= MaxTransfers && r.copies <= MaxCopied {
		for !waiting[top] {
			top--
		}
		i := top
		waiting[i] = false
		queued--
		steps++
		in := inputs[i]
		r.copies += len(in.vars) + len(in.states)
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
				_, work := exceptional.absorb(out)
				r.copies += work
			}
		}
		condition, conditional := b.node.(syntax.Expr)
		conditional = conditional && len(b.next) == 2
		for j, target := range b.next {
			// The last successor takes the working frame itself.
			last := j == len(b.next)-1
			branch := out
			if conditional {
				if !last {
					r.copies += len(out.vars) + len(out.states)
					branch = clone(out)
				}
				e.refine(condition, &branch, j == 0)
			}
			propagate(target, branch, conditional || last)
		}
		if b.exceptional > 0 {
			propagate(b.exceptional, exceptional, true)
		}
	}
	// What each variable may hold when the scope ends is what a holder of a
	// reference to it has written.
	r.exit = inputs[exit].vars
	r.leaves = inputs[done].vars == nil
	r.complete = r.complete && g.complete && queued == 0
	sort.Slice(r.calls, func(i, j int) bool { return r.calls[i].Node.Span().Start < r.calls[j].Node.Span().Start })
	return r
}
