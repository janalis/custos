package semanticquery

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

// LocalStability selects the variants of StableLocalValue a rule's spec
// states; the zero value is the default proof.
type LocalStability struct {
	// SharedOpaque treats a local shared with other code (top-level code, a
	// `global` or `static` declaration, a by-reference parameter, a
	// by-reference capture by the closure that is the scope) as changed only
	// when the scope also holds code able to run application logic: a call
	// that is not a resolved builtin function, a method or static call,
	// `new`, include/eval, or a closure or arrow function expression.
	SharedOpaque bool
	// CopiesOnly follows plain copies (`$copy = $source`) instead of every
	// local an assigned value reads: the proof is about the value's origin,
	// which the operands of a computation do not change.
	CopiesOnly bool
	// TrustOwnCaptures keeps the proof when the closure that is the scope
	// captures the local by reference: only its caller can rebind it,
	// between two calls.
	TrustOwnCaptures bool
}

// StableLocalValue reports whether the value flow resolution gives e can be
// trusted where e is read. Flow resolution follows plain assignments only, so
// the proof fails when a local e reads, or a local whose value was assigned
// into one of them (transitively), can change another way anywhere in its
// scope:
//
//   - `++`/`--` on it or on one of its elements or properties;
//   - a `catch` clause or a foreach key or value binding it;
//   - a reference to it: a by-reference assignment on either side, a
//     by-reference foreach over it, a by-reference array item, a
//     by-reference capture by a nested closure, or an argument a resolved
//     callee may bind by reference (other than e itself, which the call
//     reads first);
//   - a `global` or `static` declaration of it, or a by-reference capture by
//     the closure that is the scope; in top-level code, where every local is
//     a global, a `global` declaration naming it anywhere in the file;
//   - a write naming its target at run time: a variable variable written,
//     extract(), single-argument parse_str() or mb_parse_str(),
//     include/require or eval;
//   - a `$GLOBALS` access in the scope or, for top-level code, in the file,
//     unless it names another entry by a known string.
//
// Nested function, method and class bodies are separate scopes and are not
// scanned. When e is read in an arrow function, the scopes it captures from
// are checked too. Statement order is ignored.
func StableLocalValue(ctx *analysis.Context, e syntax.Expr, opts LocalStability) bool {
	if e == nil {
		return true
	}
	var start []string
	syntax.Inspect(e, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil {
			start = append(start, v.Name)
		}
		return true
	})
	if len(start) == 0 {
		return true
	}
	for scope := syntax.EnclosingVariableScope(e); ; scope = syntax.EnclosingVariableScope(scope) {
		if !memoScopeFacts(ctx, scope, opts).stable(ctx, e, start, opts) {
			return false
		}
		if _, arrow := scope.(*syntax.ArrowFunction); !arrow {
			return true
		}
	}
}

// localScopeFacts summarizes the bindings of one variable scope that flow
// resolution does not follow.
type localScopeFacts struct {
	// wild: a write may target any local of the scope.
	wild bool
	// opaque: the scope runs code that can run application logic.
	opaque bool
	// unstable: locals bound outside plain assignments.
	unstable map[string]bool
	// allShared, shared: locals shared with other code (see LocalStability);
	// every local of top-level code is.
	allShared bool
	shared    map[string]bool
	// feeds: assigned local -> locals read by the assigned values.
	feeds map[string][]string
	// passed: local -> argument lists passing it (or one of its elements),
	// which may bind it by reference once the callee is resolved.
	passed map[string][]*syntax.ArgList
}

// scopeFactsKey memoizes the facts of a scope under given options: every
// rule and variable asking about the scope shares one walk of it.
type scopeFactsKey struct {
	scope syntax.Node
	opts  LocalStability
}

func memoScopeFacts(ctx *analysis.Context, scope syntax.Node, opts LocalStability) *localScopeFacts {
	return ctx.File.Memo(scopeFactsKey{scope, opts}, func() any { return scopeFactsOf(ctx, scope, opts) }).(*localScopeFacts)
}

func scopeFactsOf(ctx *analysis.Context, scope syntax.Node, opts LocalStability) *localScopeFacts {
	f := &localScopeFacts{
		allShared: scope == nil,
		unstable:  map[string]bool{},
		shared:    map[string]bool{},
		feeds:     map[string][]string{},
		passed:    map[string][]*syntax.ArgList{},
	}
	into := func(names map[string]bool) func(syntax.Expr) {
		return func(e syntax.Expr) {
			if e == nil {
				return
			}
			syntax.Inspect(e, func(n syntax.Node) bool {
				if v, ok := n.(*syntax.Variable); ok {
					names[v.Name] = true
				}
				return true
			})
		}
	}
	mark := into(f.unstable)
	// Shared storage ends the proof at once unless the rule weighs it
	// against the code of the scope.
	share := mark
	if opts.SharedOpaque {
		share = into(f.shared)
	}
	captures := func(c *syntax.Closure, bind func(syntax.Expr)) {
		for _, use := range c.Uses {
			if use.ByRef && use.Var != nil {
				bind(use.Var)
			}
		}
	}
	for _, p := range syntax.FuncLikeParams(scope) {
		if p.ByRef && p.Var != nil {
			f.shared[p.Var.Name] = true
		}
	}
	// globalsEntry records a `$GLOBALS` access: an entry named by a known
	// string, or any entry.
	globalsEntry := func(x *syntax.Variable) {
		if x.Name != "GLOBALS" {
			return
		}
		if fetch, ok := x.Parent().(*syntax.ArrayDimFetch); ok && fetch.Var == syntax.Expr(x) && fetch.Dim != nil {
			if entry, known := NativeString(ctx, fetch.Dim); known {
				f.unstable[entry] = true
				return
			}
		}
		f.wild = true
	}
	var visit func(n syntax.Node) bool
	visit = func(n syntax.Node) bool {
		if n == scope {
			if c, ok := n.(*syntax.Closure); ok && !opts.TrustOwnCaptures {
				captures(c, share)
			}
			return true
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.PropertyHook:
			return false
		case *syntax.ClassLike:
			// The constructor arguments of `new class (...)` run here.
			if x.Args != nil {
				syntax.Inspect(x.Args, visit)
			}
			return false
		case *syntax.Closure:
			// May be invoked later, including by a builtin taking a callback.
			captures(x, mark)
			f.opaque = true
			return false
		case *syntax.ArrowFunction:
			f.opaque = true
			return false
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
			f.opaque = true
		case *syntax.Include, *syntax.Eval:
			f.opaque, f.wild = true, true
		case *syntax.FuncCall:
			if NativeBuiltinName(ctx, x) == "" {
				f.opaque = true
			}
			// The two-argument parse_str() binds its result by reference.
			single := x.Args != nil && len(x.Args.Args) == 1
			f.wild = f.wild || importsLocals(ctx.GlobalFunctionName(x), single)
		case *syntax.IncDec:
			f.wild = writeRoots(x.Var, func(v *syntax.Variable) { f.unstable[v.Name] = true }) || f.wild
		case *syntax.Assign:
			var targets []string
			f.wild = writeRoots(x.Var, func(v *syntax.Variable) { targets = append(targets, v.Name) }) || f.wild
			if x.ByRef {
				mark(x.Var)
				mark(x.Value)
			}
			feed := func(read *syntax.Variable) {
				for _, target := range targets {
					f.feeds[target] = append(f.feeds[target], read.Name)
				}
			}
			if !opts.CopiesOnly {
				syntax.Inspect(x.Value, func(r syntax.Node) bool {
					if read, ok := r.(*syntax.Variable); ok && read.NameExpr == nil {
						feed(read)
					}
					return true
				})
			} else if read, ok := syntax.UnwrapParens(x.Value).(*syntax.Variable); ok && read.NameExpr == nil {
				feed(read)
			}
		case *syntax.Foreach:
			mark(x.Key)
			mark(x.Value)
			if x.ByRef {
				writeRoots(x.Expr, func(v *syntax.Variable) { f.unstable[v.Name] = true })
			}
		case *syntax.ArrayItem:
			if x.ByRef {
				mark(x.Value)
			}
		case *syntax.Catch:
			if x.Var != nil {
				f.unstable[x.Var.Name] = true
			}
		case *syntax.Global:
			for _, g := range x.Vars {
				share(g)
			}
		case *syntax.StaticVar:
			share(x.Var)
		case *syntax.ArgList:
			for _, a := range x.Args {
				arg, ok := a.(*syntax.Arg)
				if !ok {
					continue
				}
				if v := argumentRoot(arg.Value); v != nil {
					if arg.ByRef {
						f.unstable[v.Name] = true
					}
					f.passed[v.Name] = append(f.passed[v.Name], x)
				}
			}
		case *syntax.Variable:
			globalsEntry(x)
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
		return f
	}
	syntax.InspectFile(ctx.File, visit)
	// Top-level locals are globals: any function of the file can reach them.
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Global:
			for _, g := range x.Vars {
				share(g)
			}
		case *syntax.Variable:
			globalsEntry(x)
		}
		return true
	})
	return f
}

// importsLocals reports builtins that define locals named at run time;
// single tells a call with one argument.
func importsLocals(name string, single bool) bool {
	switch name {
	case "extract":
		return true
	case "parse_str", "mb_parse_str":
		return single
	}
	return false
}

// writeRoots calls bind for every local a write to e binds or alters (the
// variable itself, the base of an element or property, each destructuring
// slot) and reports whether a target is named at run time.
func writeRoots(e syntax.Expr, bind func(*syntax.Variable)) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Variable:
		if x.NameExpr != nil {
			return true
		}
		bind(x)
	case *syntax.ArrayDimFetch:
		return writeRoots(x.Var, bind)
	case *syntax.PropertyFetch:
		return writeRoots(x.Var, bind)
	case *syntax.Array:
		return itemRoots(x.Items, bind)
	case *syntax.List:
		return itemRoots(x.Items, bind)
	}
	return false
}

func itemRoots(items []*syntax.ArrayItem, bind func(*syntax.Variable)) bool {
	dynamic := false
	for _, item := range items {
		if item != nil && writeRoots(item.Value, bind) {
			dynamic = true
		}
	}
	return dynamic
}

// argumentRoot returns the local an argument passes, itself or through one
// of its elements.
func argumentRoot(e syntax.Expr) *syntax.Variable {
	for {
		switch x := syntax.UnwrapParens(e).(type) {
		case *syntax.ArrayDimFetch:
			e = x.Var
			continue
		case *syntax.Variable:
			if x.NameExpr == nil {
				return x
			}
		}
		return nil
	}
}

// stable follows the names and the locals assigned into them.
func (f *localScopeFacts) stable(ctx *analysis.Context, e syntax.Expr, start []string, opts LocalStability) bool {
	if f.wild {
		return false
	}
	seen := map[string]bool{}
	for queue := start; len(queue) > 0; queue = queue[1:] {
		name := queue[0]
		if seen[name] {
			continue
		}
		seen[name] = true
		if f.unstable[name] || opts.SharedOpaque && f.opaque && (f.allShared || f.shared[name]) || f.boundByReference(ctx, e, name) {
			return false
		}
		queue = append(queue, f.feeds[name]...)
	}
	return true
}

// boundByReference reports whether a callee may bind an argument passing
// name by reference, other than e itself: a resolved callee taking it by
// reference, or one that does not resolve, whose parameters are unknown.
func (f *localScopeFacts) boundByReference(ctx *analysis.Context, e syntax.Expr, name string) bool {
	for _, list := range f.passed[name] {
		callee, ok := ResolveCallee(ctx, list.Parent())
		if !ok {
			for _, node := range list.Args {
				if arg, ok := node.(*syntax.Arg); ok && !arg.Unpack {
					if v := argumentRoot(arg.Value); v != nil && v.Name == name && syntax.UnwrapParens(arg.Value) != syntax.UnwrapParens(e) {
						return true
					}
				}
			}
			continue
		}
		for _, ref := range ReferenceArguments(list, callee.Params, true) {
			if v := argumentRoot(ref.Argument.Value); v != nil && v.Name == name && syntax.UnwrapParens(ref.Argument.Value) != syntax.UnwrapParens(e) {
				return true
			}
		}
	}
	return false
}
