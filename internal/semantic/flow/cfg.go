package flow

import (
	"sort"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

type (
	block struct {
		node        syntax.Node
		nodes       []syntax.Node
		next        []int
		exceptional int
		shared      bool
		// pin names a variable that, on entry to the block, was just found
		// loosely equal to a string literal no request text can equal
		// without being that literal (see allowed).
		pin string
	}
	graph struct {
		blocks   []block
		complete bool
		env      *Env
	}
	exits struct {
		ret, thrown, brk, cont int
		// levels lists the break and continue targets of the enclosing
		// loops and switches, innermost last, for `break 2`.
		levels [][2]int
	}
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
		lx := x.inside(next, test)
		g.blocks[test].next = []int{g.stmt(n.Body, test, lx), next}
		return test
	case *syntax.DoWhile:
		g.blocks[next].shared = true
		test := g.add(n.Cond)
		lx := x.inside(next, test)
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
		lx := x.inside(next, loop)
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
		lx := x.inside(next, test)
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
		return g.leaveLevels(n.Num, x, 0)
	case *syntax.Continue:
		return g.leaveLevels(n.Num, x, 1)
	case *syntax.Try:
		g.blocks[next].shared = true
		inside := x
		normal := next
		if n.Finally != nil {
			// A jump over several levels would run the finally block on
			// its way: it is not followed.
			inside.levels = nil
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
	case *syntax.Switch:
		return g.choose(n, next, x)
	case *syntax.Goto:
		g.complete = false
	}
	return g.atom(s, next, x)
}

// inside returns the exits of a loop or switch body.
func (x exits) inside(brk, cont int) exits {
	x.brk, x.cont = brk, cont
	x.levels = append(x.levels[:len(x.levels):len(x.levels)], [2]int{brk, cont})
	return x
}

// leaveLevels returns the target of `break` (which 0) or `continue` (which 1),
// over the number of levels a literal operand names.
func (g *graph) leaveLevels(operand syntax.Expr, x exits, which int) int {
	levels := int64(0)
	if literal, ok := syntax.UnwrapParens(operand).(*syntax.Literal); ok {
		levels, _ = strconv.ParseInt(literal.Raw, 10, 64)
	}
	switch {
	case operand == nil || levels == 1:
		return [2]int{x.brk, x.cont}[which]
	case levels < 1 || levels > int64(len(x.levels)):
		g.complete = false
		return [2]int{x.brk, x.cont}[which]
	}
	return x.levels[len(x.levels)-int(levels)][which]
}

// choose builds a switch: the subject runs once, the case expressions are
// compared with it in order until one matches, and a matching case runs into
// the bodies below it until a break. `continue` leaves the switch like
// `break`. Without a matching case the default body runs, wherever it is
// written, or nothing does. A comparison is loose: its outcome establishes no
// fact about the subject.
func (g *graph) choose(n *syntax.Switch, next int, x exits) int {
	g.blocks[next].shared = true
	inner := x.inside(next, next)
	bodies := make([]int, len(n.Cases))
	below, unmatched := next, next
	for i := len(n.Cases) - 1; i >= 0; i-- {
		below = g.list(n.Cases[i].Stmts, below, inner)
		g.blocks[below].shared = true
		bodies[i] = below
		if n.Cases[i].Cond == nil {
			unmatched = below
		}
	}
	// `switch (true)` tests each case expression for truth, like an
	// if/elseif chain: a match establishes the condition, a miss its
	// negation for the cases below and for the default.
	conditions := false
	if subject, ok := syntax.UnwrapParens(n.Cond).(*syntax.ConstFetch); ok && subject.Name != nil {
		conditions = strings.EqualFold(strings.TrimPrefix(subject.Name.Value, `\`), "true")
	}
	// A plain variable compared with constant labels only is the value the
	// labels were compared with: no label expression can have replaced it.
	pinned := plainVariable(n.Cond)
	for _, c := range n.Cases {
		switch syntax.UnwrapParens(c.Cond).(type) {
		case nil, *syntax.Literal, *syntax.ConstFetch, *syntax.ClassConstFetch:
		default:
			pinned = ""
		}
	}
	test := unmatched
	for i := len(n.Cases) - 1; i >= 0; i-- {
		label := n.Cases[i].Cond
		switch {
		case label == nil:
		case conditions:
			test = g.add(label, bodies[i], test)
			g.blocks[test].shared = true
		default:
			matched := bodies[i]
			if pinned != "" && allowed(label) {
				matched = g.add(nil, bodies[i])
				g.blocks[matched].pin = pinned
				g.blocks[matched].shared = true
			}
			test = g.add(nil, matched, test)
			g.blocks[test].exceptional = x.thrown
			g.blocks[test].nodes = []syntax.Node{label}
			g.blocks[test].shared = true
		}
	}
	return g.atom(n.Cond, test, x)
}

// allowed reports a case label that request text can only match by being it:
// a string literal that is not empty and does not start, after any leading
// whitespace, like a number. PHP compares a string with a non-numeric string
// byte for byte in every version, and an array never equals a string. Any
// other label lets text through a loose comparison: an integer, a float or a
// numeric string equals numeric strings with surrounding whitespace, signs
// and exponents (and, before PHP 8, any text starting with the number, or
// any non-numeric text for 0); `”`, `null`, `false` and `true` equal many
// strings. The only other values that equal an allowed label are the boolean
// true and, before PHP 8, the integer 0: neither holds text.
func allowed(label syntax.Expr) bool {
	literal, ok := syntax.UnwrapParens(label).(*syntax.Literal)
	if !ok || literal.LitKind != syntax.LitString {
		return false
	}
	text, known := literalText(literal)
	text = strings.TrimLeft(text, " \t\n\r\v\f")
	return known && text != "" && !strings.ContainsRune("0123456789+-.", rune(text[0]))
}

// leave builds the block of an expression that never completes: a throw, an
// exit or a call that does not return.
func (g *graph) leave(x syntax.Expr, out exits) (int, bool) {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Exit:
		return g.atom(n, 0, out), true
	case *syntax.Throw:
		// A throw is an expression since PHP 8.0 only.
		if g.env.types.PHP.AtLeast(phpversion.PHP80) {
			return g.atom(n, out.thrown, out), true
		}
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
	// A null-safe call may not happen, and `f(...)` only makes a closure.
	var list *syntax.ArgList
	switch n := call.(type) {
	case *syntax.FuncCall:
		list = n.Args
	case *syntax.MethodCall:
		if list = n.Args; conditional(n) {
			return false
		}
	case *syntax.StaticCall:
		if list = n.Args; conditional(n.Class) {
			return false
		}
	}
	if list != nil && len(list.Args) == 1 {
		if _, callable := list.Args[0].(*syntax.VariadicPlaceholder); callable {
			return false
		}
	}
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
		if within := syntax.EnclosingClass(call); within != nil && within.ClassKind == syntax.KindTrait && !m.Abstract && strings.EqualFold(m.Class, e.types.ClassFQN(within)) {
			// A class using the trait can replace the method, whatever its
			// modifiers, with one that returns: only an abstract method
			// binds the class to the declared type.
			declared, name = "", ""
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
		if this, direct := n.Var.(*syntax.Variable); ok && direct && this.Name == "this" {
			// Inside a trait `$this` has no class of its own yet.
			return e.types.Index.FindMethod(e.types.ClassFQN(syntax.EnclosingClass(n)), id.Value, e.types.PHP)
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
	x := exits{ret: done, thrown: exit, brk: done, cont: done}
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
		if held := out.vars[b.pin]; b.pin != "" && tainted(held) {
			// The request data was just compared equal to a fixed text.
			out.write(r, b.pin, Value{Complete: true})
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
