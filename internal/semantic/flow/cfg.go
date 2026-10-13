package flow

import (
	"sort"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
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
		complete bool
		// full marks a graph that reached MaxBlocks.
		full bool
		env  *Env
		// body is what the graph is built from, nil for the statements of
		// the file. labels gives the block each label starts, and jumps the
		// blocks of the `goto` statements written before their label, which
		// gets them when it is built. guarded tells whether the body holds a
		// try statement, once that was looked up.
		body    syntax.Node
		labels  map[string]int
		jumps   map[string][]int
		guarded int8
		// first is where the first label of the body is written, plus one.
		first uint32
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
		g.complete, g.full = false, true
		return 0
	}
	g.blocks = append(g.blocks, block{node: n, next: next, exceptional: -1})
	return len(g.blocks) - 1
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
		lx := x.inside(next, test)
		g.blocks[test].next = repeats(n.Cond, g.stmt(n.Body, test, lx), next)
		return test
	case *syntax.DoWhile:
		g.blocks[next].shared = true
		test := g.test(n.Cond, x)
		lx := x.inside(next, test)
		body := g.stmt(n.Body, test, lx)
		g.blocks[test].next = repeats(n.Cond, body, next)
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
		g.blocks[test].next = []int{cond}
		start := test
		for i := len(n.Init) - 1; i >= 0; i-- {
			start = g.atom(n.Init[i], start, x)
		}
		return start
	case *syntax.Foreach:
		g.blocks[next].shared = true
		test := g.test(n, x)
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
			g.blocks[jump].next = []int{at}
		}
		delete(g.jumps, n.Name.Value)
		return at
	case *syntax.Goto:
		// A jump out of a try statement runs its finally blocks on the
		// way: such a scope is not followed.
		if g.tries() {
			g.complete = false
			break
		}
		if at, built := g.labels[n.Label.Value]; built {
			return g.add(nil, at)
		}
		jump := g.add(nil)
		if g.jumps == nil {
			g.jumps = map[string][]int{}
		}
		g.jumps[n.Label.Value] = append(g.jumps[n.Label.Value], jump)
		return jump
	}
	return g.atom(s, next, x)
}

// tries reports whether the body the graph is built from holds a try
// statement, outside the functions declared in it.
func (g *graph) tries() bool {
	if g.guarded == 0 {
		g.guarded = -1
		look := func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Function, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
				return false
			case *syntax.Try:
				g.guarded = 1
			}
			return g.guarded < 0
		}
		if g.body == nil {
			syntax.InspectFile(g.env.file, look)
		} else {
			syntax.Inspect(g.body, look)
		}
	}
	return g.guarded > 0
}

// repeats returns the successors of a loop test: the body and, unless the
// condition is absent or the constant true, what follows the loop. A loop
// that always repeats is left through `break`, `return` or a throw only.
func repeats(cond syntax.Node, body, next int) []int {
	if cond == nil {
		return []int{body}
	}
	if test, ok := cond.(syntax.Expr); ok && constantName(test) == "true" {
		return []int{body}
	}
	return []int{body, next}
}

// jumpsLevels reports a `break` or `continue` written with a level inside a
// try statement, outside the functions declared in it.
func jumpsLevels(n *syntax.Try) bool {
	found := false
	syntax.Inspect(n, func(child syntax.Node) bool {
		switch jump := child.(type) {
		case *syntax.Function, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.Break:
			found = found || jump.Num != nil
		case *syntax.Continue:
			found = found || jump.Num != nil
		}
		return !found
	})
	return found
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
// written, or nothing does. A comparison is loose: the only facts its outcome
// establishes about the subject are those of pin and differs.
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
	// `switch (false)` runs the first case whose expression is not true.
	conditions := constantName(n.Cond)
	// A plain variable compared with constant labels only is the value the
	// labels were compared with: no label expression can have replaced it.
	// An element of a request table read under a constant key is compared
	// the same way; it has no binding, so the match is recorded for it.
	pinned := plainVariable(n.Cond)
	matching := requestElement(n.Cond)
	for _, c := range n.Cases {
		switch syntax.UnwrapParens(c.Cond).(type) {
		case nil, *syntax.Literal, *syntax.ConstFetch, *syntax.ClassConstFetch:
		default:
			pinned, matching = "", ""
		}
	}
	test := unmatched
	for i := len(n.Cases) - 1; i >= 0; i-- {
		label := n.Cases[i].Cond
		switch {
		case label == nil:
		case conditions == "true":
			test = g.test(label, x, bodies[i], test)
			g.blocks[test].shared = true
		case conditions == "false":
			test = g.test(label, x, test, bodies[i])
			g.blocks[test].shared = true
		default:
			matched, missed := bodies[i], test
			if pinned+matching != "" && allowed(label) {
				matched = g.add(nil, bodies[i])
				g.blocks[matched].pin = pinned + matching
				g.blocks[matched].shared = true
			}
			switch constantName(label) {
			case "false", "null":
				// Every falsy value equals both: the cases below and the
				// default see a subject that is neither.
				if pinned != "" {
					missed = g.add(nil, test)
					g.blocks[missed].differs = n.Cond
					g.blocks[missed].shared = true
				}
			case "true":
				if pinned != "" {
					matched = g.add(nil, bodies[i])
					g.blocks[matched].differs = n.Cond
					g.blocks[matched].shared = true
				}
			}
			test = g.add(nil, matched, missed)
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
// typed receiver, or a static call through a class name, `self`, `static` or
// `parent`.
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
		return e.types.Index.FindMethod(e.named(class), id.Value, e.types.PHP)
	}
	return nil
}

// named resolves the class a static call names: the enclosing class for
// `self` and `static`, the class it extends for `parent`. The method is then
// looked up through the classes above it.
func (e *Env) named(class *syntax.Name) string {
	switch strings.ToLower(class.Value) {
	case "self", "static":
		return e.types.ClassFQN(syntax.EnclosingClass(class))
	case "parent":
		return e.types.Names.ParentFQN(syntax.EnclosingClass(class))
	}
	return e.types.Names.Class(class.Value, class.Span().Start)
}

func (e *Env) scope(scope syntax.Node) *scopeResult {
	if r, ok := e.scopes[scope]; ok {
		return r
	}
	// What the functions of the file write to global variables is computed
	// before the scope has a result of its own: one that is such a function
	// is analysed there, and has no result to read while it is.
	if !e.globalsKnown {
		e.globals()
		return e.scope(scope)
	}
	// A static variable the function assigns is first assumed never to be
	// false where another invocation can read it: at its calls and where it
	// ends. Then a guard on it holds across a call that enters the function
	// again. When some path breaks the assumption for a variable the scope
	// is analysed again without it for that one, a bounded number of times,
	// then without it for any.
	r := e.follow(scope, nil, false)
	var distrust map[string]bool
	for round := 0; len(r.broken) > 0; round++ {
		if distrust == nil {
			distrust = map[string]bool{}
		}
		for name := range r.broken {
			distrust[name] = true
		}
		r = e.follow(scope, distrust, round == staticRounds)
	}
	return r
}

// firstLabel returns where the first `goto` label of a scope is written,
// outside the functions and classes declared in it, plus one; 0 without one.
func (e *Env) firstLabel(scope syntax.Node) uint32 {
	first := uint32(0)
	look := func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Function, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
			return n == scope
		case *syntax.Label:
			first = n.Span().Start + 1
		}
		return first == 0
	}
	if scope == nil {
		syntax.InspectFile(e.file, look)
	} else {
		syntax.Inspect(scope, look)
	}
	return first
}

// jumpShare is the part of the time budget a scope with labels gets.
const jumpShare = 16

// staticRounds bounds the analyses of a scope made to find which of its
// static variables may be left false.
const staticRounds = 2

// typed reports a parameter whose declared type admits no object: an array,
// what a variadic parameter collects, or since PHP 7 a number, a boolean or
// text. Before, a scalar type names a class. With deep the type admits no
// array either, which may hold objects.
func (e *Env) typed(p *syntax.Param, deep bool) bool {
	declared := types.FromNode(p.Type, nil)
	if e.types.PHP.AtLeast(phpversion.PHP70) && declared.OnlyOf("int", "float", "string", "bool", "true", "false", "null") {
		return !deep || !p.Variadic
	}
	return !deep && (p.Variadic || declared.OnlyOf("array") || e.types.PHP.AtLeast(phpversion.PHP70) && declared.OnlyOf("int", "float", "string", "bool", "true", "false", "null", "array"))
}

func (e *Env) follow(scope syntax.Node, distrust map[string]bool, distrustAll bool) *scopeResult {
	r := &scopeResult{values: map[syntax.Expr]Value{}, before: map[syntax.Node]map[uint32]State{}, callIndex: map[syntax.Node]int{}, complete: true, distrust: distrust, distrustAll: distrustAll, modern: e.types.PHP.AtLeast(phpversion.PHP80)}
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
			g.body = body
			entry = g.stmt(st, done, x)
		} else if body != nil {
			entry = g.atom(body, done, x)
		}
	}
	// A jump to a label that is not there goes nowhere known.
	g.complete = g.complete && len(g.jumps) == 0
	initial := frame{vars: map[string]Value{}, states: map[uint32]State{}, definite: scope != nil}
	switch captured := scope.(type) {
	case *syntax.Closure:
		// A captured variable holds a value without a binding.
		initial.definite = len(captured.Uses) == 0
	case *syntax.ArrowFunction:
		initial.definite = false
	}
	for i, p := range syntax.VariableScopeParams(scope) {
		if p.Var != nil {
			initial.vars[p.Var.Name] = Value{Complete: true, Identity: p.Var.Span().Start + 1, Sources: []Source{{Kind: "parameter", Parameter: i}}, plain: e.typed(p, false), deep: e.typed(p, true)}
		}
	}
	e.bindings(scope, r, &initial)
	// A block is visited when no block before it in the source waits: blocks
	// are numbered from the end of the scope, a loop body above its test, so
	// a loop settles before what follows it runs.
	inputs := make([]frame, len(g.blocks))
	waiting := make([]bool, len(g.blocks))
	inputs[entry], waiting[entry] = initial, true
	top, queued, steps := entry, 1, 0
	// A jump back to a label is no loop the numbering knows: a scope with
	// labels is swept from its start to its end, the blocks a jump back
	// reaches waiting for the next sweep, which settles it in far fewer
	// visits. A scope of jumps between many labels still takes more of them
	// than it is worth: it has a fraction of the time budget.
	sweep, again, budget := g.labels != nil, 0, MaxCopied
	if sweep {
		budget = MaxCopied / jumpShare
	}
	if r.labelled = g.first; g.full {
		// The statements before the one that filled the graph were not read.
		r.labelled = e.firstLabel(scope)
	}
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
			if sweep {
				again = max(again, target)
			} else {
				top = max(top, target)
			}
		}
	}
	// A broken assumption about a static variable ends the first analysis:
	// the scope is analysed again without it.
	for queued > 0 && steps < MaxTransfers && !r.spent && len(r.broken) == 0 && r.operations <= MaxTransfers && r.copies <= budget {
		for top >= 0 && !waiting[top] {
			top--
		}
		if top < 0 {
			// The sweep is over: the next one starts at the first block a
			// jump back reached.
			top, again = again, 0
			continue
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
			if b.exceptional > 0 {
				// What the condition wrote before it threw.
				_, work := exceptional.absorb(out)
				r.copies += work
			}
		}
		if held := out.vars[b.pin]; b.pin != "" && tainted(held) {
			// The request data was just compared equal to a fixed text.
			out.write(r, b.pin, Value{Complete: true, matched: true, plain: true, deep: true})
		} else if strings.Contains(b.pin, "\x00") {
			out.match(r, b.pin)
		}
		exclude(b.differs, &out, true, true)
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
		loop, iterating := b.node.(*syntax.Foreach)
		for j, target := range b.next {
			// The last successor takes the working frame itself.
			last := j == len(b.next)-1
			branch := out
			if (conditional || iterating) && !last {
				r.copies += len(out.vars) + len(out.states)
				branch = clone(out)
			}
			switch {
			case conditional:
				e.refine(condition, &branch, j == 0)
			case iterating && j == 0:
				// The body starts with an element in the targets.
				e.each(loop, &branch, r)
			}
			propagate(target, branch, conditional || iterating || last)
		}
		if b.exceptional > 0 {
			propagate(b.exceptional, exceptional, true)
		}
	}
	// What each variable may hold when the scope ends is what a holder of a
	// reference to it has written.
	r.exit = inputs[exit].vars
	leaving(&inputs[exit], r)
	r.leaves = inputs[done].vars == nil
	r.complete = r.complete && g.complete && queued == 0
	sort.Slice(r.calls, func(i, j int) bool { return r.calls[i].Node.Span().Start < r.calls[j].Node.Span().Start })
	return r
}
