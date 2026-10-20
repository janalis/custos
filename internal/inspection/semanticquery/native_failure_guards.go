package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/types"
)

// NativeLengthProven proves that input is exactly length bytes long where it
// is read: the flow facts say so, or, in a scope without goto labels before
// it, the strlen() tests guarding it leave only that length.
func NativeLengthProven(ctx *analysis.Context, input syntax.Expr, length int64) bool {
	if input == nil {
		return false
	}
	if n, known := ctx.Flow().ExactLength(input); known && n == length {
		return true
	}
	return !NativeLabelled(ctx, input) && ctx.Types().Native().NarrowedFrom(input, types.String).LengthsOnly(length, length)
}

// NativeFailureExcluded proves that the failure value sentinel (`null`,
// `false`) cannot reach input: the flow facts exclude it, or, in a scope
// without goto labels before it, the conditions guarding input narrow it out
// (seed is the type its producer returns, for a member read).
func NativeFailureExcluded(ctx *analysis.Context, input syntax.Expr, seed types.Type, sentinel string) bool {
	if input == nil {
		return false
	}
	return ctx.Flow().Excludes(input, sentinel) || NativeNarrowedExcludes(ctx, input, seed, sentinel)
}

// NativeFailureExcludedSingle is NativeFailureExcluded where the narrowed
// type counts only when a single producer reaches input (see
// NativeFailureFlowOnly): with several, only the flow facts do.
func NativeFailureExcludedSingle(ctx *analysis.Context, input syntax.Expr, seed types.Type, sentinel string) bool {
	return ctx.Flow().Excludes(input, sentinel) || !NativeFailureFlowOnly(ctx, input) && NativeNarrowedExcludes(ctx, input, seed, sentinel)
}

// NativeIntegerAllowsZero proves the validator options neither exclude zero
// nor change the failure sentinel to null.
func NativeIntegerAllowsZero(ctx *analysis.Context, options syntax.Expr) bool {
	if options == nil {
		return true
	}
	if n, known := NativeInt(ctx, options); known {
		return n&134217728 == 0 // FILTER_NULL_ON_FAILURE
	}
	entries, known := NativeArrayEntries(ctx, options)
	if !known {
		return false
	}
	if flags := entries["s:flags"]; flags != nil {
		present, known := NativeFlag(ctx, flags, 134217728)
		if !known || present {
			return false
		}
	}
	if bounds := entries["s:options"]; bounds != nil {
		values, known := NativeArrayEntries(ctx, bounds)
		if !known {
			return false
		}
		for _, key := range []string{"min_range", "max_range"} {
			if value := values["s:"+key]; value != nil {
				n, known := NativeInt(ctx, value)
				if !known || (key == "min_range" && n > 0) || (key == "max_range" && n < 0) {
					return false
				}
			}
		}
	}
	return true
}

// NativeCurlDefault proves a local handle retains an option's disabled default.
// Any escape, ambiguous option, bulk update or write callback invalidates proof.
func NativeCurlDefault(ctx *analysis.Context, at syntax.Node, handle syntax.Expr, option string) bool {
	value := ctx.Flow().Value(handle)
	id := value.Identity
	if id == 0 || value.Invalidated == ^uint32(0) || !NativeHandleUnescaped(ctx, handle, at) {
		return false
	}
	origin := NativeValue(ctx, handle)
	init, initialized := origin.(*syntax.FuncCall)
	if !initialized || !NativeBuiltin(ctx, init, "curl_init") {
		return false
	}
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		if call.Node.Span().Start >= at.Span().Start {
			continue
		}
		uses := false
		for _, arg := range call.Arguments {
			if arg.Identity == id {
				uses = true
			}
		}
		if !uses {
			continue
		}
		fc, ok := call.Node.(*syntax.FuncCall)
		if !ok {
			return false
		}
		switch NativeBuiltinName(ctx, fc) {
		case "curl_setopt":
			opt := CallArgument(fc.Args, 1, "option")
			constant, ok := syntax.UnwrapParens(opt).(*syntax.ConstFetch)
			if !ok || constant.Name == nil {
				return false
			}
			name := GlobalConstName(ctx, constant)
			declaration := ctx.Index().Constant(name, ctx.PHP)
			if declaration == nil || !declaration.Builtin || !declaration.Avail.In(ctx.PHP) {
				return false
			}
			if name == "CURLOPT_WRITEFUNCTION" {
				return false
			}
			if name == option {
				truth, known := NativeTruth(ctx, CallArgument(fc.Args, 2, "value"))
				if !known || truth {
					return false
				}
			}
		case "curl_init":
		default:
			return false
		}
	}
	return true
}

// NativeLabelled reports that a `goto` label is written before a node in its
// variable scope. The node can then be reached by a jump, without the
// conditions written before it or around it having been evaluated: a guard
// proves nothing by where it is written, and only the facts the flow
// analysis established on the paths that reach the node count.
func NativeLabelled(ctx *analysis.Context, at syntax.Node) bool {
	return ctx.Flow().Labelled(at)
}

// nativeStaticReplaced reports what a static local held, read twice, with a
// call that may have entered the function again and assigned it in between.
// The flow stamps a binding with the last such call, and a copy keeps the
// stamp; paths that disagree leave a mark that says nothing of the order of
// the reads, which then proves nothing.
func nativeStaticReplaced(ctx *analysis.Context, a, b syntax.Expr, av, bv flow.Value) bool {
	if av.Invalidated == bv.Invalidated && av.Invalidated != ^uint32(0) {
		return false
	}
	return nativeReentryStamp(ctx, a, av) || nativeReentryStamp(ctx, b, bv)
}

// nativeReentryStamp tells the stamp of a call that may have assigned a
// static local from the one a value with an identity gets when it is handed
// to unknown code, which leaves its variable bound to it: that one only
// counts for a variable declared `static`.
func nativeReentryStamp(ctx *analysis.Context, e syntax.Expr, v flow.Value) bool {
	return v.Invalidated != 0 && (v.Identity == 0 || NativeStaticLocal(ctx, e))
}

// NativeStaticLocal reports a read of a variable declared `static` in its
// scope. A name computed at run time may denote one. What a test or a
// narrowed type established of such a variable holds until the next call
// that can enter the function again, not until its next written assignment.
func NativeStaticLocal(ctx *analysis.Context, e syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		return false
	}
	scope := syntax.EnclosingVariableScope(e)
	writes := nativeContractScope(ctx, scope).localWrites(ctx, scope)
	return variable.NameExpr != nil && len(writes.statics) > 0 || writes.static(variable.Name)
}
