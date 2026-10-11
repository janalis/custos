package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
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

// nativeLocalWrites lists the writes of one variable scope in source order.
// known is false for a scope with more writes than the budget, which then
// proves nothing.
type nativeLocalWrites struct {
	known  bool
	writes []nativeLocalWrite
}

const nativeLocalWriteBudget = 1024

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
	write := nativeLocalWrite{at: n.Span().Start}
	switch x := n.(type) {
	case *syntax.Assign:
		if x.ByRef {
			// The source is shared once the reference exists.
			w.target(x.Value, nativeLocalWrite{at: x.Span().End})
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
		}
	case *syntax.Catch:
		if x.Var != nil {
			w.target(x.Var, write)
		}
	case *syntax.Global:
		for _, v := range x.Vars {
			w.target(v, write)
		}
	case *syntax.StaticStmt:
		for _, v := range x.Vars {
			w.target(v.Var, write)
		}
	case *syntax.Unset:
		for _, v := range x.Vars {
			w.target(v, write)
		}
	case *syntax.ArrayItem:
		if x.ByRef {
			w.target(x.Value, nativeLocalWrite{at: x.Span().End})
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
		w.arguments(ctx, x, x.Args)
	case *syntax.MethodCall:
		w.arguments(ctx, x, x.Args)
	case *syntax.StaticCall:
		w.arguments(ctx, x, x.Args)
	case *syntax.New:
		w.arguments(ctx, x, x.Args)
	}
}

// arguments records the variables a call may replace through a reference
// parameter once it returns. A resolved callee decides by its parameter; an
// unresolved function is assumed to take references, as the local-value scan
// always did, an unresolved method or constructor is not.
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
		if arg.Name == nil {
			position++
		}
		if nativeRootVariable(arg.Value) == nil {
			continue
		}
		if !asked {
			asked = true
			params, resolved, assume, builtin = nativeCalleeParameters(ctx, call)
		}
		written := arg.ByRef || assume
		if resolved {
			written = arg.ByRef || arg.Unpack || !nativeByValueParameter(ctx, params, arg, at)
		}
		if !written {
			continue
		}
		w.target(arg.Value, nativeLocalWrite{at: call.Span().End, builtin: builtin && resolved})
	}
}

// nativeCalleeParameters resolves the parameters a call binds its arguments
// to, whether an unresolved callee must be assumed to take references, and
// whether PHP itself declares the function.
func nativeCalleeParameters(ctx *analysis.Context, call syntax.Node) (params []index.Param, resolved, assume, builtin bool) {
	switch x := call.(type) {
	case *syntax.New:
		if name, ok := x.Class.(*syntax.Name); ok {
			if m := ctx.Index().FindMethod(ctx.Names().Class(name.Value, name.Span().Start), "__construct", ctx.PHP); m != nil {
				return m.Params, true, false, false
			}
		}
		return nil, false, false, false
	case *syntax.FuncCall:
		callee, ok := ResolveCallee(ctx, x)
		return callee.Params, ok, true, NativeBuiltinName(ctx, x) != ""
	}
	callee, ok := ResolveCallee(ctx, call)
	return callee.Params, ok, false, false
}

// nativeByValueParameter proves the parameter an argument binds to is taken
// by value. Parameters unavailable at the target version are not counted, and
// a surplus argument of a non-variadic callee binds to nothing.
func nativeByValueParameter(ctx *analysis.Context, params []index.Param, arg *syntax.Arg, position int) bool {
	var last *index.Param
	seen := 0
	for i := range params {
		p := &params[i]
		if !p.Avail.In(ctx.PHP) {
			continue
		}
		last = p
		if arg.Name != nil {
			if p.Name == arg.Name.Value {
				return !p.ByRef
			}
			continue
		}
		if seen == position {
			return !p.ByRef
		}
		seen++
	}
	return arg.Name == nil && (last == nil || !last.Variadic || !last.ByRef)
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
	return false
}
