package semanticquery

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/php/syntax"
)

// OutputParameterCall locates the most recent dominating out-parameter producer.
func OutputParameterCall(ctx *analysis.Context, at syntax.Node, value syntax.Expr, functions []string, position int, name string) *syntax.FuncCall {
	v, ok := syntax.UnwrapParens(value).(*syntax.Variable)
	if !ok {
		return nil
	}
	scope := syntax.EnclosingVariableScope(at)
	calls := ctx.Flow().Calls(scope)
	if len(calls) > 256 {
		return nil
	}
	var found *syntax.FuncCall
	for _, fact := range calls {
		c, ok := fact.Node.(*syntax.FuncCall)
		if !ok || !NativeDominates(c, at) {
			continue
		}
		out, ok := CallArgument(c.Args, position, name).(*syntax.Variable)
		if !ok || out.Name != v.Name {
			continue
		}
		builtin := NativeBuiltinName(ctx, c)
		for _, name := range functions {
			if builtin == name && (found == nil || found.Span().Start < c.Span().Start) {
				found = c
			}
		}
	}
	if found == nil {
		return nil
	}
	valid := true
	visited := 0
	visit := func(n syntax.Node) bool {
		visited++
		if visited > 256 {
			valid = false
			return false
		}
		if n.Span().Start >= at.Span().Start {
			return false
		}
		if n != scope && syntax.EnclosingVariableScope(n) != scope {
			return false
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		for _, stmt := range ctx.File.Stmts {
			if !valid || visited > 256 || stmt.Span().Start >= at.Span().Start {
				break
			}
			syntax.Inspect(stmt, visit)
		}
	}
	if valid && outputParameterRebound(ctx, scope, v.Name, found, at) {
		valid = false
	}
	for _, fact := range calls {
		if fact.Node.Span().Start <= found.Span().End || fact.Node.Span().Start >= at.Span().Start {
			continue
		}
		var args *syntax.ArgList
		switch call := fact.Node.(type) {
		case *syntax.FuncCall:
			switch NativeBuiltinName(ctx, call) {
			case "pcntl_wifexited", "pcntl_wifsignaled", "pcntl_wifstopped", "pcntl_wexitstatus", "pcntl_wtermsig", "pcntl_wstopsig":
				continue
			}
			args = call.Args
		case *syntax.MethodCall:
			args = call.Args
		case *syntax.StaticCall:
			args = call.Args
		}
		if args != nil {
			for _, node := range args.Args {
				if arg, ok := node.(*syntax.Arg); ok && SameStableExpr(ctx, arg.Value, value) {
					valid = false
				}
			}
		}
	}
	if !valid {
		return nil
	}
	return found
}

// outputParameterRebound reports a write of the output variable between found and
// at: any rebinding (assignment, destructuring, increment, foreach target,
// unset), an element write or a by-reference capture. A write later in a loop
// around at does not count: the first pass still reads found's output.
func outputParameterRebound(ctx *analysis.Context, scope syntax.Node, name string, found *syntax.FuncCall, at syntax.Node) bool {
	from, to := found.Span().End, at.Span().Start
	for _, access := range flowquery.VarAccesses(ctx.File, scope, name) {
		use, captured := access.By.(*syntax.ClosureUse)
		write := access.Write || access.ElemWrite || captured && use.ByRef
		if start := access.Var.Span().Start; write && start > from && start < to {
			return true
		}
	}
	return false
}

// SameStableExpr compares stable source expressions; useful for output argument identities.
func SameStableExpr(ctx *analysis.Context, a, b syntax.Expr) bool {
	return a != nil && b != nil && astquery.Equivalent(ctx.File, a, b)
}

// SameStableLocal proves a stable local variable even when its value comes from
// a parameter. Unknown calls, references and writes invalidate the relation.
func SameStableLocal(ctx *analysis.Context, a, b syntax.Expr) bool {
	if a == nil || b == nil {
		return false
	}
	facts := nativeContractScope(ctx, syntax.EnclosingVariableScope(a))
	events := facts.nodes
	if events == nil {
		return false
	}
	at := syntax.Node(b)
	if a.Span().Start > b.Span().Start {
		at = a
	}
	if !UnaliasedLocal(ctx, a, at) || !UnaliasedLocal(ctx, b, at) {
		return false
	}
	if NativeSameValue(ctx, a, b) {
		return true
	}
	if x, xk := NativeString(ctx, a); xk {
		if y, yk := NativeString(ctx, b); yk && x == y {
			return true
		}
	}
	if x, xk := NativeContractInt(ctx, a); xk {
		if y, yk := NativeContractInt(ctx, b); yk && x == y {
			return true
		}
	}
	av, aok := syntax.UnwrapParens(a).(*syntax.Variable)
	bv, bok := syntax.UnwrapParens(b).(*syntax.Variable)
	if !aok || !bok || av.Name != bv.Name || syntax.EnclosingVariableScope(a) != syntax.EnclosingVariableScope(b) {
		return false
	}
	start, end := a.Span().End, b.Span().Start
	if start > end {
		start, end = b.Span().End, a.Span().Start
	}
	// Only a write of the variable itself separates the two reads; using it
	// as an index or a property name of another target does not.
	if facts.localWrites(ctx, syntax.EnclosingVariableScope(a)).rebound(av.Name, start, end) {
		return false
	}
	for _, event := range events {
		if event.Span().Start <= start || event.Span().End >= end {
			continue
		}
		switch e := event.(type) {
		case *syntax.FuncCall:
			if NativeBuiltinName(ctx, e) == "" && astquery.MentionsVariable(e.Args, av.Name) {
				return false
			}
		case *syntax.MethodCall:
			if astquery.MentionsVariable(e.Args, av.Name) {
				return false
			}
		}
	}
	return true
}

// ScopeCallEvents returns the assignments, unsets and function and method
// calls of the scope of at, nested scopes excluded, cached per scope; none
// when there are more than 512.
func ScopeCallEvents(ctx *analysis.Context, at syntax.Node) []syntax.Node {
	return ContractEvents(ctx, syntax.EnclosingVariableScope(at))
}

// OriginCall returns the call to one of the builtins names that the local e
// holds, when e has no alias that could have changed it.
func OriginCall(ctx *analysis.Context, e syntax.Expr, names ...string) *syntax.FuncCall {
	if e == nil || !UnaliasedLocal(ctx, e, e) {
		return nil
	}
	c, ok := NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok {
		return nil
	}
	for _, name := range names {
		if NativeBuiltin(ctx, c, name) {
			return c
		}
	}
	return nil
}

// EarlierMethodCalls returns the calls, before at in its scope, of the
// methods names of class on the same stable local as receiver that run
// before at: they dominate it or at runs only when they succeeded.
func EarlierMethodCalls(ctx *analysis.Context, at, receiver syntax.Expr, class string, names ...string) []*syntax.MethodCall {
	var out []*syntax.MethodCall
	for _, event := range ScopeCallEvents(ctx, at) {
		c, ok := event.(*syntax.MethodCall)
		if !ok || c.Span().End >= at.Span().Start || (!NativeDominates(c, at) && !NativeCallSucceeded(ctx, c, at)) || !SameStableLocal(ctx, c.Var, receiver) {
			continue
		}
		for _, name := range names {
			if NativeMethod(ctx, c, class, name) {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// ObjectMethodHistory retains ordered same-object calls; ambiguous path effects
// discard the complete state history rather than inventing a state transition.
func ObjectMethodHistory(ctx *analysis.Context, at, obj syntax.Expr) []*syntax.MethodCall {
	calls := ctx.Flow().Calls(syntax.EnclosingVariableScope(at))
	if len(calls) > 256 {
		return nil
	}
	out := []*syntax.MethodCall{}
	for _, record := range calls {
		c, ok := record.Node.(*syntax.MethodCall)
		if !ok || c.Args.Partial() || c.Span().End >= at.Span().Start || !NativeSameValue(ctx, c.Var, obj) {
			continue
		}
		if !NativeDominates(c, at) && !NativeCallSucceeded(ctx, c, at) {
			return nil
		}
		out = append(out, c)
	}
	return out
}

// MembersUnwrittenBefore excludes local writes through XML property/array handles.
func MembersUnwrittenBefore(ctx *analysis.Context, obj syntax.Expr, at syntax.Node) bool {
	if !UnaliasedLocal(ctx, obj, at) {
		return false
	}
	good := true
	visited := 0
	scope := syntax.EnclosingVariableScope(at)
	visit := func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		visited++
		if visited > 512 {
			good = false
			return false
		}
		if n.Span().Start >= at.Span().Start {
			return false
		}
		var targets []syntax.Expr
		switch x := n.(type) {
		case *syntax.Assign:
			targets = []syntax.Expr{x.Var}
		case *syntax.IncDec:
			targets = []syntax.Expr{x.Var}
		case *syntax.Unset:
			targets = x.Vars
		}
		for _, target := range targets {
			var base syntax.Expr
			switch x := target.(type) {
			case *syntax.PropertyFetch:
				base = x.Var
			case *syntax.ArrayDimFetch:
				base = x.Var
			}
			for {
				switch x := base.(type) {
				case *syntax.PropertyFetch:
					base = x.Var
				case *syntax.ArrayDimFetch:
					base = x.Var
				default:
					goto rooted
				}
			}
		rooted:
			if base != nil && NativeSameValue(ctx, base, obj) {
				good = false
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return good
}

type escapeScopeKey struct{ scope syntax.Node }

type opaqueEscapeScopeKey struct{ scope syntax.Node }

type escapeCandidates struct {
	nodes []syntax.Node
	known bool
}

// escapeCandidatesOf caches syntax-only escape candidates once per lexical scope.
// Receiver identity and types remain context-dependent and are never cached.
func escapeCandidatesOf(ctx *analysis.Context, scope syntax.Node, opaque bool) escapeCandidates {
	var key any
	if opaque {
		key = opaqueEscapeScopeKey{scope}
	} else {
		key = escapeScopeKey{scope}
	}
	return ctx.File.Memo(key, func() any {
		result := escapeCandidates{known: true}
		visited, events := 0, 0
		visit := func(n syntax.Node) bool {
			if !result.known {
				return false
			}
			if n != scope && syntax.IsVariableScope(n) {
				if arrow, ok := n.(*syntax.ArrowFunction); ok && opaque {
					parameters := map[string]bool{}
					for _, param := range arrow.Params {
						if param.Var != nil {
							parameters[param.Var.Name] = true
						}
					}
					syntax.Inspect(arrow.Expr, func(child syntax.Node) bool {
						visited++
						if visited > 4096 {
							result.known = false
							result.nodes = nil
							return false
						}
						if variable, ok := child.(*syntax.Variable); ok && !parameters[variable.Name] {
							result.nodes = append(result.nodes, variable)
						}
						return true
					})
				}
				if closure, ok := n.(*syntax.Closure); ok {
					visited += len(closure.Uses)
					if visited > 4096 {
						result.known = false
						result.nodes = nil
						return false
					}
					for _, use := range closure.Uses {
						result.nodes = append(result.nodes, use)
					}
				}
				return false
			}
			visited++
			switch n.(type) {
			case *syntax.Assign, *syntax.Unset, *syntax.FuncCall, *syntax.MethodCall:
				events++
			}
			if visited > 4096 || events > 512 {
				result.known = false
				result.nodes = nil
				return false
			}
			switch value := n.(type) {
			case *syntax.Assign:
				if opaque || value.ByRef {
					result.nodes = append(result.nodes, n)
				}
			case *syntax.ArrayItem:
				if opaque {
					result.nodes = append(result.nodes, n)
				}
			}
			return true
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		return result
	}).(escapeCandidates)
}

// UnaliasedLocal rejects escaping aliases using bounded cached syntax.
func UnaliasedLocal(ctx *analysis.Context, obj syntax.Expr, at syntax.Node) bool {
	candidates := escapeCandidatesOf(ctx, syntax.EnclosingVariableScope(at), false)
	if !candidates.known {
		return false
	}
	for _, candidate := range candidates.nodes {
		if candidate.Span().Start >= at.Span().Start {
			break
		}
		switch x := candidate.(type) {
		case *syntax.Assign:
			if x.ByRef && (astquery.Equivalent(ctx.File, x.Value, obj) || NativeSameValue(ctx, x.Value, obj)) {
				return false
			}
		case *syntax.ClosureUse:
			if !astquery.Equivalent(ctx.File, x.Var, obj) {
				continue
			}
			if x.ByRef {
				return false
			}
			object := ctx.Types().Native().TypeOf(obj)
			if len(object.Classes()) > 0 || object.Has("resource") {
				return false
			}
		}
	}
	return true
}

// handleUnescaped rejects visible storage escapes of an opaque handle.
// Callers separately establish its allocation identity and call effects.
func handleUnescaped(ctx *analysis.Context, handle syntax.Expr, at syntax.Node) bool {
	return nativeHandleUnescaped(ctx, handle, at, false)
}

func nativeHandleUnescaped(ctx *analysis.Context, handle syntax.Expr, at syntax.Node, allowLocalCopies bool) bool {
	candidates := escapeCandidatesOf(ctx, syntax.EnclosingVariableScope(at), true)
	if !candidates.known {
		return false
	}
	// Keep possible local copies even across conditional rebinding. Reads inside
	// deferred arrow bodies need not carry the enclosing scope's flow identity.
	var aliases map[string]bool
	matches := func(expr syntax.Expr) bool {
		if variable, ok := syntax.UnwrapParens(expr).(*syntax.Variable); ok && aliases[variable.Name] {
			return true
		}
		return astquery.Equivalent(ctx.File, expr, handle) || NativeSameValue(ctx, expr, handle)
	}
	for _, node := range candidates.nodes {
		if node.Span().Start >= at.Span().Start {
			break
		}
		switch value := node.(type) {
		case *syntax.Variable:
			if matches(value) {
				return false
			}
		case *syntax.Assign:
			if local, plain := value.Var.(*syntax.Variable); allowLocalCopies && plain && local.NameExpr == nil && !value.ByRef {
				if matches(value.Value) && !astquery.Equivalent(ctx.File, local, syntax.UnwrapParens(handle)) {
					if aliases == nil {
						aliases = make(map[string]bool)
					}
					aliases[local.Name] = true
				}
				continue
			}
			if value.ByRef && astquery.Equivalent(ctx.File, value.Var, handle) {
				return false
			}
			if value.ByRef || !astquery.Equivalent(ctx.File, value.Var, handle) {
				if matches(value.Value) {
					return false
				}
			}
		case *syntax.ArrayItem:
			if matches(value.Value) {
				return false
			}
		case *syntax.ClosureUse:
			if matches(value.Var) {
				return false
			}
		}
	}
	return true
}

// LocalEvents returns bounded lexical effects, excluding nested scopes.
func LocalEvents(ctx *analysis.Context, at syntax.Node) []syntax.Node {
	return ContractEvents(ctx, syntax.EnclosingVariableScope(at))
}

// SameLocalObject compares retained local binding provenance when the flow
// model cannot assign allocation identities to extension-created objects.
func SameLocalObject(ctx *analysis.Context, a, b syntax.Expr) bool {
	x := ctx.Flow().Value(a)
	y := ctx.Flow().Value(b)
	if x.Complete && y.Complete && x.Identity != 0 && x.Identity == y.Identity && x.Invalidated == y.Invalidated {
		return true
	}
	left := NativeValue(ctx, a)
	right := NativeValue(ctx, b)
	if left == nil || left != right {
		return false
	}
	switch left.(type) {
	case *syntax.New, *syntax.MethodCall:
		return true
	}
	return false
}
