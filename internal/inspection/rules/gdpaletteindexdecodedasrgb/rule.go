// Package gdpaletteindexdecodedasrgb implements the native GdPaletteIndexDecodedAsRgb inspection.
package gdpaletteindexdecodedasrgb

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Look up palette components instead of shifting the index."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdPaletteIndexDecodedAsRgb" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TAmpersand {
		return
	}
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		mask, mk := semanticquery.NativeContractInt(ctx, pair[1])
		if !mk || mask != 255 {
			continue
		}
		s, ok := syntax.UnwrapParens(pair[0]).(*syntax.Binary)
		if !ok || s.Op.Kind != syntax.TSr {
			continue
		}
		shift, sk := semanticquery.NativeContractInt(ctx, s.Right)
		if !sk || (shift != 8 && shift != 16) {
			continue
		}
		// Both the index and its image must be the one value every path
		// agrees on: a packed colour or a truecolor image on another path
		// makes the shift valid there.
		c, ok := localOnly(ctx, s.Left).(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, c, "imagecolorallocate") {
			continue
		}
		e := semanticquery.CallArgument(c.Args, 0, "image")
		if !semanticquery.ExpansionDUnaliased(ctx, e, n) {
			return
		}
		creation, ok := localOnly(ctx, e).(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, creation, "imagecreate") {
			continue
		}
		if converted(ctx, creation, c) {
			continue
		}
		ctx.ReportNode(n, message)
		return
	}
}

// converted reports whether any call between the creation and the allocation
// may turn the palette image into a truecolor one. The call need not dominate
// the allocation: a conversion on one path already removes the palette proof.
func converted(ctx *analysis.Context, creation, alloc *syntax.FuncCall) bool {
	found := false
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= alloc.Span().Start {
			return false
		}
		c, ok := n.(*syntax.FuncCall)
		if ok && c.Span().Start > creation.Span().End && semanticquery.NativeBuiltin(ctx, c, "imagepalettetotruecolor") {
			found = semanticquery.NativeLocalValue(ctx, semanticquery.CallArgument(c.Args, 0, "image")) == creation
		}
		return true
	}
	if scope := syntax.EnclosingVariableScope(alloc); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}

// localOnly is NativeLocalValue for an inspection that needs the one
// value a variable holds wherever the read runs: nil when the producers of
// several paths reach the read, or one producer and a path that left the
// variable unknown. NativeLocalValue answers such a read with a call result
// no guard covers, which is what a failure check asks, not what the variable
// is known to hold.
func localOnly(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	if reaching, single := ctx.Flow().Reaching(e); !single && len(reaching) > 0 {
		return nil
	}
	return semanticquery.NativeLocalValue(ctx, e)
}
