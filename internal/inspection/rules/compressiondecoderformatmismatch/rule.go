// Package compressiondecoderformatmismatch implements the native CompressionDecoderFormatMismatch inspection.
package compressiondecoderformatmismatch

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Match the decoder to the compression format."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CompressionDecoderFormatMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	decoder := semanticquery.NativeBuiltinName(ctx, c)
	switch decoder {
	case "gzdecode", "gzuncompress", "gzinflate":
	default:
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "data")
	encoder, ok := semanticquery.NativeValue(ctx, arg).(*syntax.FuncCall)
	if !ok {
		return
	}
	if capturedByReference(ctx, c, arg) || !defaultFormat(encoder) {
		return
	}

	match := map[string]string{"gzencode": "gzdecode", "gzcompress": "gzuncompress", "gzdeflate": "gzinflate"}[semanticquery.NativeBuiltinName(ctx, encoder)]
	if match != "" && match != decoder {
		var fixes []diagnostic.Fix
		if fn := ctx.Index().Function(match, ctx.PHP); fn != nil && fn.Builtin && (match != "gzdecode" || ctx.PHP >= phpversion.PHP54) {
			fixes = append(fixes, astquery.ReplaceFix(c.Name.Span(), semanticquery.QualifiedBuiltinFor(ctx, match, c)))
		}
		ctx.ReportNode(c, message, fixes...)
	}
}

// defaultFormat reports whether the encoder call provably omits the encoding
// argument: at most the data and level slots are bound, and nothing is unpacked.
func defaultFormat(encoder *syntax.FuncCall) bool {
	for _, node := range encoder.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || (arg.Name != nil && arg.Name.Value == "encoding") {
			return false
		}
	}
	return semanticquery.CallArgument(encoder.Args, 2, "encoding") == nil
}

// capturedByReference reports whether the decoded local, or a local it was
// copied from, is bound by reference to a closure of its scope; invoking that
// closure can rewrite the bytes.
func capturedByReference(ctx *analysis.Context, at syntax.Node, arg syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(arg).(*syntax.Variable)
	if !ok {
		return false
	}
	scope := syntax.EnclosingFuncLike(at)
	captured := map[string]bool{}
	copies := map[string][]string{}
	visit := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Closure:
			for _, use := range x.Uses {
				if use.ByRef && use.Var != nil {
					captured[use.Var.Name] = true
				}
			}
		case *syntax.Assign:
			to, toOK := syntax.UnwrapParens(x.Var).(*syntax.Variable)
			from, fromOK := syntax.UnwrapParens(x.Value).(*syntax.Variable)
			if toOK && fromOK {
				copies[to.Name] = append(copies[to.Name], from.Name)
			}
		}
		return n == scope || !syntax.IsFuncLike(n)
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	seen := map[string]bool{}
	for queue := []string{v.Name}; len(queue) > 0; queue = queue[1:] {
		name := queue[0]
		if captured[name] {
			return true
		}
		if !seen[name] {
			seen[name] = true
			queue = append(queue, copies[name]...)
		}
	}
	return false
}
