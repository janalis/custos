package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// nativeLocalWrite is one place where a local variable may receive a value.
type nativeLocalWrite struct {
	// name is the written variable; empty for a write that can reach any.
	name string
	// at is the offset from which a later read sees the write.
	at uint32
	// plain marks `$v = value`, the only form the local-value scan reads.
	plain bool
	// element marks a write into the variable's array or object, which
	// leaves the variable bound to the same container.
	element bool
	// builtin marks a reference parameter of a resolved PHP function, the
	// way an API updates a handle it was given.
	builtin bool
}

type nativeLocalRelease struct {
	at    uint32
	owner syntax.Span
}

// nativeLocalWrites lists the writes of one variable scope in source order.
// known is false for a scope with more writes than the budget, which then
// proves nothing.
type nativeLocalWrites struct {
	known  bool
	writes []nativeLocalWrite
	// shared names are bound, from the recorded offset on, to storage another
	// name, a closure or the globals table also reaches; calls are the
	// offsets after which a call that can run project code has returned.
	shared      map[string][]uint32
	released    map[string][]nativeLocalRelease
	jumps       []uint32
	exceptional bool
	labelled    bool
	calls       []uint32
	// statics names the locals declared `static`: an invocation the function
	// makes of itself, through whatever it calls, assigns them too.
	statics map[string]bool
}

// share records a name whose storage is reachable without naming it.
// Multiple bindings matter when unset releases an earlier reference.
func (w *nativeLocalWrites) share(x syntax.Expr, at uint32) {
	if !w.known {
		return
	}
	if v := nativeRootVariable(x); v != nil && v.NameExpr == nil {
		if w.shared == nil {
			w.shared = map[string][]uint32{}
		}
		w.shared[v.Name] = append(w.shared[v.Name], at)
	}
}

// call records a call that can run project code.
func (w *nativeLocalWrites) call(n syntax.Node) {
	if len(w.calls) == nativeLocalWriteBudget {
		w.known = false
		return
	}
	w.calls = append(w.calls, n.Span().End)
}

const nativeLocalWriteBudget = 1024

// Bound the combined nested work of each reference-lifetime query.
const nativeLocalReferenceWork = 4096

// index walks the scope once. A nested closure contributes the names it
// binds by reference: its body may write them whenever it runs.
func (w *nativeLocalWrites) index(ctx *analysis.Context, scope syntax.Node) {
	w.known = true
	visit := func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			if closure, ok := n.(*syntax.Closure); ok {
				for _, use := range closure.Uses {
					if use.ByRef {
						w.target(use.Var, nativeLocalWrite{at: n.Span().Start})
						w.share(use.Var, n.Span().Start)
					}
				}
			}
			return false
		}
		w.record(ctx, n)
		return w.known
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
}

func (w *nativeLocalWrites) add(name string, write nativeLocalWrite) {
	if len(w.writes) == nativeLocalWriteBudget {
		w.known = false
		return
	}
	write.name = name
	w.writes = append(w.writes, write)
}

// target records the variables a write target replaces. A property or static
// property target changes no local binding.
func (w *nativeLocalWrites) target(x syntax.Expr, write nativeLocalWrite) {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		if n.NameExpr != nil {
			w.add("", nativeLocalWrite{at: write.at, builtin: write.builtin})
		} else {
			w.add(n.Name, write)
		}
	case *syntax.ArrayDimFetch:
		// An entry of the globals table is the variable of that name, or
		// any variable when the key is not a literal.
		if table, ok := n.Var.(*syntax.Variable); ok && table.Name == "GLOBALS" && table.NameExpr == nil {
			name, _ := astquery.QuotedStringValue(syntax.UnwrapParens(n.Dim))
			w.add(name, nativeLocalWrite{at: write.at})
			return
		}
		write.plain, write.element = false, true
		w.target(n.Var, write)
	default:
		write.plain = false
		for _, item := range nativeTargetItems(x) {
			if item != nil {
				w.target(item.Value, write)
			}
		}
	}
}

// nativeTargetItems returns the slots of a `[...]` or `list(...)` target.
func nativeTargetItems(x syntax.Expr) []*syntax.ArrayItem {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Array:
		return n.Items
	case *syntax.List:
		return n.Items
	}
	return nil
}

func (w *nativeLocalWrites) record(ctx *analysis.Context, n syntax.Node) {
	start := n.Span().Start
	write := nativeLocalWrite{at: start}
	switch x := n.(type) {
	case *syntax.Assign:
		if x.ByRef {
			// The source is shared once the reference exists.
			w.target(x.Value, nativeLocalWrite{at: x.Span().End})
			w.share(x.Value, start)
			w.share(x.Var, start)
		}
		write.plain = !x.ByRef && x.Op.Kind == syntax.TEqual

		w.target(x.Var, write)
	case *syntax.IncDec:
		w.target(x.Var, write)
	case *syntax.Foreach:
		w.target(x.Key, write)
		w.target(x.Value, write)
		if x.ByRef || nativeReferenceTarget(x.Value) {
			w.target(x.Expr, nativeLocalWrite{at: x.Expr.Span().End})
			w.share(x.Expr, start)
			w.share(x.Value, start)
		}
	case *syntax.Catch:
		if x.Var != nil {
			w.target(x.Var, write)
		}
	case *syntax.Global:
		for _, v := range x.Vars {
			w.target(v, write)
			w.share(v, start)
		}
	case *syntax.StaticStmt:
		for _, v := range x.Vars {
			w.target(v.Var, write)
			if w.statics == nil {
				w.statics = map[string]bool{}
			}
			w.statics[v.Var.Name] = true
		}
	case *syntax.Unset:
		for _, v := range x.Vars {
			w.target(v, write)
			if _, list := syntax.StmtListOf(x.Parent()); list && w.known {
				if variable, plain := syntax.UnwrapParens(v).(*syntax.Variable); plain && variable.NameExpr == nil {
					if w.released == nil {
						w.released = map[string][]nativeLocalRelease{}
					}
					w.released[variable.Name] = append(w.released[variable.Name], nativeLocalRelease{at: x.Span().End, owner: x.Parent().Span()})
				}
			}
		}
	case *syntax.Break, *syntax.Continue, *syntax.Throw:
		if len(w.jumps) == nativeLocalWriteBudget {
			w.known = false
			return
		}
		w.jumps = append(w.jumps, start)
	case *syntax.Try:
		w.exceptional = true
	case *syntax.Label:
		w.labelled = true
	case *syntax.ArrayItem:
		if x.ByRef {
			w.target(x.Value, nativeLocalWrite{at: x.Span().End})
			w.share(x.Value, start)
		}
	case *syntax.Include, *syntax.Eval:
		w.add("", write)
	case *syntax.FuncCall:
		if name, ok := x.Name.(*syntax.Name); ok && x.Args != nil {
			switch strings.ToLower(name.Value[strings.LastIndexByte(name.Value, '\\')+1:]) {
			case "extract":
				if NativeBuiltin(ctx, x, "extract") {
					w.add("", write)
				}
			case "parse_str":
				if len(x.Args.Args) < 2 && NativeBuiltin(ctx, x, "parse_str") {
					w.add("", write)
				}
			}
		}
		if NativeBuiltinName(ctx, x) == "" || nativeTakesCallback(ctx, x) {
			w.call(x)
		}
		w.arguments(ctx, x, x.Args)
	case *syntax.MethodCall:
		w.call(x)
		w.arguments(ctx, x, x.Args)
	case *syntax.StaticCall:
		w.call(x)
		w.arguments(ctx, x, x.Args)
	case *syntax.New:
		w.call(x)
		w.arguments(ctx, x, x.Args)
	}
}

// arguments records the variables a call may replace through a reference
// parameter once it returns. A resolved callee decides by its parameter; an
// unresolved callee may take a reference to a variable or element. Temporary
// arrays and values passed inside them are not rebindable argument slots.
func (w *nativeLocalWrites) arguments(ctx *analysis.Context, call syntax.Node, list *syntax.ArgList) {
	if list == nil {
		return
	}
	var params []index.Param
	resolved, assume, builtin, asked := false, false, false, false
	position := 0
	for _, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok {
			continue
		}
		at := position
		position = nextArgumentPosition(arg, position)
		if nativeRootVariable(arg.Value) == nil {
			continue
		}
		if !asked {
			asked = true
			params, resolved, assume, builtin = nativeCalleeParameters(ctx, call)
		}
		written := arg.ByRef || assume
		if resolved {
			byRef, known := referenceBinding(arg, at, params)
			written = arg.ByRef || !known || byRef
		}
		if !written {
			continue
		}
		w.target(arg.Value, nativeLocalWrite{at: call.Span().End, builtin: builtin && resolved})
	}
}

// nativeTakesCallback reports a resolved function with a callable parameter:
// it may run project code holding a reference.
func nativeTakesCallback(ctx *analysis.Context, call *syntax.FuncCall) bool {
	callee, _ := ResolveCallee(ctx, call)
	for _, p := range callee.Params {
		if strings.Contains(p.Type, "callable") {
			return true
		}
	}
	return false
}

// nativeCalleeParameters resolves the parameters a call binds its arguments
// to, whether an unresolved callee must be assumed to take references, and
// whether PHP itself declares the function.
func nativeCalleeParameters(ctx *analysis.Context, call syntax.Node) (params []index.Param, resolved, assume, builtin bool) {
	if x, ok := call.(*syntax.FuncCall); ok {
		callee, ok := ResolveCallee(ctx, x)
		return callee.Params, ok, true, NativeBuiltinName(ctx, x) != ""
	}
	callee, ok := ResolveCallee(ctx, call)
	return callee.Params, ok, !ok, false
}

// nativeRootVariable returns the variable an lvalue chain is stored in.
func nativeRootVariable(x syntax.Expr) *syntax.Variable {
	for {
		switch n := syntax.UnwrapParens(x).(type) {
		case *syntax.ArrayDimFetch:
			x = n.Var
		case *syntax.Variable:
			return n
		default:
			return nil
		}
	}
}

// nativeReferenceTarget reports a destructuring target with a reference slot.
func nativeReferenceTarget(x syntax.Expr) bool {
	for _, item := range nativeTargetItems(x) {
		if item != nil && (item.ByRef || nativeReferenceTarget(item.Value)) {
			return true
		}
	}
	return false
}

// rewritten reports whether name may hold another value at the
// read position than the one a plain assignment ending at from gave it: a
// write the local-value scan does not read lies after it and before the read.
// A reference argument of the assigned call itself is written first. A write into
// an element only matters for a value whose content is tracked.
func (w *nativeLocalWrites) rewritten(name string, from, read uint32, content bool) bool {
	if !w.known {
		return true
	}
	for _, write := range w.writes {
		if (write.name == name || write.name == "") && !write.plain && (content || !write.element) && write.at > from && write.at < read {
			return true
		}
	}
	return w.reached(name, from, read)
}

// reached reports that storage shared under name may have been written
// between two offsets without naming it: through another shared name, or by a
// call that can run code holding a reference.
func (w *nativeLocalWrites) reached(name string, after, before uint32) bool {
	if !w.known {
		return true
	}
	if len(w.shared[name]) == 0 {
		return false
	}
	budget := nativeLocalReferenceWork
	for _, write := range w.writes {
		budget--
		if budget < 0 {
			return true
		}
		if write.name != name && !write.builtin && write.at > after && write.at < before && w.sharedAt(name, write.at, &budget) && w.sharedAt(write.name, write.at, &budget) {
			return true
		}
		if budget < 0 {
			return true
		}
	}
	for _, at := range w.calls {
		budget--
		if budget < 0 {
			return true
		}
		if at > after && at < before && w.sharedAt(name, at, &budget) {
			return true
		}
	}
	return false
}

// sharedAt keeps releases local to the statement list where unset executes.
// A later reference binds the name again. Labels prevent this lexical proof:
// a jump could land past the unset without having released the binding.
func (w *nativeLocalWrites) sharedAt(name string, at uint32, budget *int) bool {
	for _, start := range w.shared[name] {
		(*budget)--
		if *budget < 0 {
			return true
		}
		if start >= at {
			continue
		}
		released := false
		if !w.labelled {
			for _, release := range w.released[name] {
				(*budget)--
				if *budget < 0 {
					return true
				}
				if release.at >= at || release.at <= start {
					continue
				}
				if release.owner.Contains(syntax.Span{Start: at, End: at}) {
					released = true
					break
				}
				if w.exceptional || !release.owner.Contains(syntax.Span{Start: start, End: start}) {
					continue
				}
				escaped := false
				for _, jump := range w.jumps {
					(*budget)--
					if *budget < 0 {
						return true
					}
					if jump > start && jump < release.at {
						escaped = true
						break
					}
				}
				if !escaped {
					released = true
					break
				}
			}
		}
		if !released {
			return true
		}
	}
	return false
}

// rebound reports a write strictly between two offsets after which
// name may denote another value. A PHP function updating the handle it was
// given, and a write into the value's elements, keep it the same value.
func (w *nativeLocalWrites) rebound(name string, after, before uint32) bool {
	if !w.known {
		return true
	}
	for _, write := range w.writes {
		if (write.name == name || write.name == "") && !write.builtin && !write.element && write.at > after && write.at < before {
			return true
		}
	}
	return w.reached(name, after, before)
}

// static reports a local declared `static` in the scope, or any local of a
// scope with more writes than the budget.
func (w *nativeLocalWrites) static(name string) bool {
	return !w.known || w.statics[name]
}
