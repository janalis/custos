// Package validatedintegerzerorejected implements the native ValidatedIntegerZeroRejected inspection.
package validatedintegerzerorejected

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare integer validation failure strictly with false."

// arrayResultFlags is FILTER_REQUIRE_ARRAY | FILTER_FORCE_ARRAY: with either
// flag the call yields an array, so negation tests emptiness, not zero.
const arrayResultFlags = 16777216 | 67108864

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ValidatedIntegerZeroRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "filter_var") {
		return
	}
	filter, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 1, "filter"))
	if !known || filter != 257 {
		return
	}
	options := semanticquery.CallArgument(call.Args, 2, "options")
	if !integerAllowsZero(ctx, options) || !failsWithScalarFalse(ctx, options) {
		return
	}
	p, _ := astquery.ParentSkipParens(call)
	u, ok := p.(*syntax.Unary)
	if !ok || u.Op.Kind != syntax.TExclaim {
		return
	}
	ctx.ReportNode(call, message, diagnostic.Fix{Title: "Distinguish invalid integers from zero", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: u.Op.Span, NewText: "("}, {Span: syntax.Span{Start: u.Span().End, End: u.Span().End}, NewText: " === false)"}}
	}})
}

// failsWithScalarFalse reports whether a rejected value still yields false:
// array-shaping flags change the result type and a configured default replaces
// the failure value. The options are already proven static by the caller.
func failsWithScalarFalse(ctx *analysis.Context, options syntax.Expr) bool {
	if options == nil {
		return true
	}
	if flags, known := semanticquery.NativeInt(ctx, options); known {
		return flags&arrayResultFlags == 0
	}
	entries, _ := semanticquery.NativeArrayEntries(ctx, options)
	if shaped, _ := semanticquery.NativeFlag(ctx, entries["s:flags"], arrayResultFlags); shaped {
		return false
	}
	if inner := entries["s:options"]; inner != nil {
		values, _ := semanticquery.NativeArrayEntries(ctx, inner)
		if _, replaced := values["s:default"]; replaced {
			return false
		}
	}
	return true
}

// integerAllowsZero proves the validator options neither exclude zero
// nor change the failure sentinel to null.
func integerAllowsZero(ctx *analysis.Context, options syntax.Expr) bool {
	if options == nil {
		return true
	}
	if n, known := semanticquery.NativeInt(ctx, options); known {
		return n&134217728 == 0 // FILTER_NULL_ON_FAILURE
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, options)
	if !known {
		return false
	}
	if flags := entries["s:flags"]; flags != nil {
		present, known := semanticquery.NativeFlag(ctx, flags, 134217728)
		if !known || present {
			return false
		}
	}
	if bounds := entries["s:options"]; bounds != nil {
		values, known := semanticquery.NativeArrayEntries(ctx, bounds)
		if !known {
			return false
		}
		for _, key := range []string{"min_range", "max_range"} {
			if value := values["s:"+key]; value != nil {
				n, known := semanticquery.NativeInt(ctx, value)
				if !known || (key == "min_range" && n > 0) || (key == "max_range" && n < 0) {
					return false
				}
			}
		}
	}
	return true
}
