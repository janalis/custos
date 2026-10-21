// Package hashfilefailureusedasdigest implements HashFileFailureUsedAsDigest.
package hashfilefailureusedasdigest

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Reject file hashing failure before using the digest."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HashFileFailureUsedAsDigest" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	for _, input := range inputs(ctx, n) {
		if guarded(ctx, input) {
			return
		}
	}
	check(ctx, n, message)
}

// inputs lists the operands the shared check treats as a consumed digest.
func inputs(ctx *analysis.Context, n syntax.Node) []syntax.Expr {
	var out []syntax.Expr
	switch c := n.(type) {
	case *syntax.FuncCall:
		if semanticquery.NativeBuiltin(ctx, c, "file_put_contents") {
			out = append(out, semanticquery.CallArgument(c.Args, 1, "data"))
		}
	case *syntax.Binary:
		if c.Op.Kind != syntax.TIsIdentical && c.Op.Kind != syntax.TIsNotIdentical {
			return nil
		}
		if _, ok := semanticquery.NativeString(ctx, c.Right); ok {
			out = append(out, c.Left)
		}
		if _, ok := semanticquery.NativeString(ctx, c.Left); ok {
			out = append(out, c.Right)
		}
	}
	return out
}

// guarded proves that false cannot reach input: the flow facts exclude it, or
// the conditions guarding the read narrow it to a digest.
func guarded(ctx *analysis.Context, input syntax.Expr) bool {
	return semanticquery.NativeFailureExcluded(ctx, input, types.Of("string", "false"), "false")
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	var inputs []syntax.Expr
	switch c := n.(type) {
	case *syntax.FuncCall:
		if !semanticquery.NativeBuiltin(ctx, c, "file_put_contents") {
			return
		}
		inputs = []syntax.Expr{semanticquery.CallArgument(c.Args, 1, "data")}
	case *syntax.Binary:
		if c.Op.Kind != syntax.TIsIdentical && c.Op.Kind != syntax.TIsNotIdentical {
			return
		}
		if _, ok := semanticquery.NativeString(ctx, c.Right); ok {
			inputs = append(inputs, c.Left)
		}
		if _, ok := semanticquery.NativeString(ctx, c.Left); ok {
			inputs = append(inputs, c.Right)
		}
	}
	for _, e := range inputs {
		if semanticquery.OriginCall(ctx, e, "hash_file") != nil && !semanticquery.NativeFailureExcluded(ctx, e, types.Of("string", "false"), "false") {
			ctx.ReportNode(n, message)
			return
		}
	}
}
