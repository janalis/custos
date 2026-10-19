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
	semanticquery.CheckHashFileFailureUsedAsDigest(ctx, n, message)
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
	if input == nil {
		return false // a missing argument
	}
	if ctx.Flow().Excludes(input, "false") {
		return true
	}
	return !semanticquery.NativeLabelled(ctx, input) && semanticquery.NativeNarrowedExcludes(ctx, input, types.Of("string", "false"), "false")
}
