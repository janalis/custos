package flow

import (
	"maps"
	"sort"
	"strings"
	"sync"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

func (e *Env) scope(scope syntax.Node) *scopeResult {
	if r, ok := e.scopes[scope]; ok {
		e.depend(r.reads)
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
		earlier := r
		r = e.follow(scope, distrust, round == staticRounds)
		addReads(r.reads, earlier.reads)
	}
	r.finished = true
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

type expressionSizesKey struct{}

// expressionSizes estimates the two map capacities separately. Lexical names,
// declaration headers and plain assignment targets are not evaluated values.
// Control wrappers lower to their tests and bodies rather than transfer nodes.
// These are allocation hints only: both maps grow for any additional entries.
func (e *Env) expressionSizes(scope syntax.Node) (int, int) {
	cache := e.file.Memo(expressionSizesKey{}, func() any { return &sync.Map{} }).(*sync.Map)
	if cached, ok := cache.Load(scope); ok {
		sizes := cached.([2]int)
		return sizes[0], sizes[1]
	}
	values, before := e.countExpressions(scope)
	cache.Store(scope, [2]int{values, before})
	return values, before
}

func (e *Env) countExpressions(scope syntax.Node) (values, before int) {
	count := func(x syntax.Node) bool {
		switch x.(type) {
		case *syntax.Block, *syntax.Namespace, *syntax.Function, *syntax.Method, *syntax.ClassLike,
			*syntax.If, *syntax.ElseIf, *syntax.Else, *syntax.While, *syntax.DoWhile, *syntax.For,
			*syntax.Try, *syntax.Finally, *syntax.Switch, *syntax.Case, *syntax.Label, *syntax.Goto, *syntax.Break, *syntax.Continue:
		case syntax.Expr:
			switch n := x.(type) {
			case *syntax.Name, *syntax.Identifier, *syntax.ArgList, *syntax.VariadicPlaceholder:
			case *syntax.Variable:
				assignment, target := n.Parent().(*syntax.Assign)
				if !target || assignment.Var != n || assignment.Op.Kind != syntax.TEqual || n.NameExpr != nil {
					values++
					before++
				}
			default:
				values++
				before++
			}

		case syntax.Stmt:
			before++
		}
		return !syntax.IsVariableScope(x)
	}
	if scope == nil {
		syntax.InspectFile(e.file, count)
	} else {
		syntax.Inspect(syntax.VariableScopeBody(scope), count)
	}
	return values, before
}

func (e *Env) follow(scope syntax.Node, distrust map[string]bool, distrustAll bool) *scopeResult {
	return e.followMode(scope, distrust, distrustAll, nil)
}

func (e *Env) followMode(scope syntax.Node, distrust map[string]bool, distrustAll bool, target syntax.Expr) *scopeResult {
	return e.followSeeded(scope, distrust, distrustAll, target, nil, nil, nil)
}

func (e *Env) followSeeded(scope syntax.Node, distrust map[string]bool, distrustAll bool, target syntax.Expr, seeds map[string]Value, states map[uint32]State, resources map[uint32]bool) *scopeResult {
	values, before := e.expressionSizes(scope)
	r := &scopeResult{sharing: e.sharingArena(), resourceInputs: resources, originTypes: e.types, values: newValueStore(values), before: make(map[syntax.NodeIdentity]map[uint32]State, before), callIndex: map[syntax.Node]int{}, complete: true, distrust: distrust, distrustAll: distrustAll, modern: e.types.PHP.AtLeast(phpversion.PHP80)}
	e.scopes[scope] = r
	r.reads = e.begin()
	defer e.end(r.reads)
	r.generator = e.generator(scope)
	// Block 0 ends the scope however it is left; a normal completion passes
	// through the block before it, which tells whether the scope can return.
	g := graph{complete: true, env: e}
	exit := g.add(nil)
	done := g.add(nil, exit)
	x := exits{ret: done, thrown: exit, brk: done, cont: done}
	fallthroughExit := done
	if body, ok := syntax.VariableScopeBody(scope).(*syntax.Block); ok && body != nil {
		terminalReturn := false
		if len(body.Stmts) > 0 {
			_, terminalReturn = body.Stmts[len(body.Stmts)-1].(*syntax.Return)
		}
		if !terminalReturn {
			fallthroughExit = g.add(nil, done)
		}
	}
	entry := fallthroughExit
	if scope == nil {
		entry = g.list(e.file.Stmts, fallthroughExit, x)
	} else {
		body := syntax.VariableScopeBody(scope)
		if st, ok := body.(syntax.Stmt); ok {
			g.body = body
			entry = g.stmt(st, fallthroughExit, x)
		} else if body != nil {
			entry = g.atom(body, fallthroughExit, x)
		}
	}
	// All label bodies remain built, including those entered by goto inside a
	// constant-false branch. Only propagation from the condition is pruned.
	hasExceptional := false
	for i := range g.blocks {
		b := &g.blocks[i]
		hasExceptional = hasExceptional || b.exceptional > 0
		if condition, ok := b.node.(syntax.Expr); ok && len(g.successors(i)) == 2 {
			if truth, known := constantTruth(condition); known {
				b.choice = 2
				if truth {
					b.choice = 1
				}
			}
		}
	}
	if hasExceptional {
		g.exceptionProofs(scope)
	}
	// A jump to a label that is not there goes nowhere known.
	g.complete = g.complete && len(g.jumps) == 0
	initial := frame{owner: r, definite: scope != nil}
	initial.resetVars()
	initial.states = maps.Clone(states)
	switch captured := scope.(type) {
	case *syntax.Closure:
		// A captured variable holds a value without a binding.
		initial.definite = len(captured.Uses) == 0
	case *syntax.ArrowFunction:
		initial.definite = false
	}
	for i, p := range syntax.VariableScopeParams(scope) {
		if p.Var != nil {
			initial.setVar(p.Var.Name, Value{Complete: true, Identity: p.Var.Span().Start + 1, Sources: []Source{{Kind: "parameter", Parameter: i}}, plain: e.typed(p, false), deep: e.typed(p, true)})
		}
	}
	if seeds != nil {
		initial.definite = true
		for name, value := range seeds {
			initial.setVar(name, value)
		}
	}
	e.bindings(scope, r, &initial)
	if target != nil {
		g.entryPrefix(entry, initial, r, target)
		return r
	}
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
	once := g.once(sweep, exit, done)
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
				v = clone(&v)
			}
			*input = v
		} else {
			before := len(input.vars) + len(input.states)
			changed, work := input.absorbAt(v, g.blocks[target].head)
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
		out := in
		if !once[i] {
			out = clone(&in)
		}
		b := g.blocks[i]
		throwing := b.exceptional > 0 && !g.proofs[i].holds(in)
		var exceptional frame
		if throwing {
			exceptional = clone(&in)
		}
		if b.node != nil {
			e.transfer(b.node, &out, r)
			if throwing {
				// What the condition wrote before it threw.
				_, work := exceptional.absorb(out)
				r.copies += work
			}
		}
		if held := out.vars.value(b.pin); b.pin != "" && tainted(held) {
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
			if throwing {
				_, work := exceptional.absorb(out)
				r.copies += work
			}
		}
		next := g.successors(i)
		condition, conditional := b.node.(syntax.Expr)
		conditional = conditional && len(next) == 2
		loop, iterating := b.node.(*syntax.Foreach)
		for j, target := range next {
			if b.choice != 0 && int(b.choice)-1 != j {
				continue
			}
			// The last successor takes the working frame itself.
			last := b.choice != 0 || j == len(next)-1
			branch := out
			if (conditional || iterating) && !last {
				r.copies += len(out.vars) + len(out.states)
				branch = clone(&out)
			}
			switch {
			case conditional:
				e.refine(condition, &branch, j == 0)
			case iterating && j == 0:
				// The body starts with an element in the targets.
				e.each(loop, &branch, r)
			}
			propagate(int(target), branch, conditional || iterating || last)
		}
		if throwing {
			propagate(b.exceptional, exceptional, true)
		}
	}
	// What each variable may hold when the scope ends is what a holder of a
	// reference to it has written.
	r.exit = inputs[exit].vars
	r.exitStates = inputs[done].states
	leaving(&inputs[exit], r)
	r.leaves = inputs[done].vars == nil
	r.fallsThrough = fallthroughExit != done && inputs[fallthroughExit].vars != nil
	if !g.complete {
		reason := "unsupported-control-flow"
		if g.full {
			reason = "flow-budget"
		}
		span := e.file.Span()
		if scope != nil {
			span = scope.Span()
		}
		r.gap(reason, span)
	}
	r.complete = r.complete && g.complete && queued == 0
	if !g.complete || queued != 0 || r.spent || len(g.labels) > 0 {
		r.cut = 0
	}
	sort.Slice(r.calls, func(i, j int) bool { return r.calls[i].Node.Span().Start < r.calls[j].Node.Span().Start })
	// Evaluation order can differ from source order for nested calls.
	// Publish indices into the same ordered history that queries observe.
	for i, call := range r.calls {
		r.callIndex[call.Node] = i
	}
	return r
}
