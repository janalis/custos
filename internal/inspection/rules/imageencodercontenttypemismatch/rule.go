// Package imageencodercontenttypemismatch implements the native ImageEncoderContentTypeMismatch inspection.
package imageencodercontenttypemismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the response content type to the image encoder."

// encoderTypes maps each response-capable encoder to its canonical MIME type.
var encoderTypes = map[string]string{"imagepng": "image/png", "imagejpeg": "image/jpeg", "imagegif": "image/gif", "imagewebp": "image/webp", "imageavif": "image/avif", "imagebmp": "image/bmp"}

// canonicalTypes is the closed set of header values the proof understands.
var canonicalTypes = map[string]bool{"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true, "image/avif": true, "image/bmp": true}

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ImageEncoderContentTypeMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	mime, known := encoderTypes[name]
	if !known || c.Args == nil {
		return
	}
	// An unpacked argument may carry the output path.
	for _, node := range c.Args.Args {
		if arg, ok := node.(*syntax.Arg); !ok || arg.Unpack {
			return
		}
	}
	output := semanticquery.CallArgument(c.Args, 1, "file")
	if output != nil {
		c, ok := output.(*syntax.ConstFetch)
		if !ok || c.Name.Value != "null" {
			return
		}
	}
	statement, ok := c.Parent().(*syntax.ExprStmt)
	if !ok || statement.Parent() != nil {
		return
	}
	list, index, known := astquery.StmtList(ctx.File, statement)
	if !known || index != len(list)-1 {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	count := 0
	for _, priorStatement := range prior {
		buffered := false
		syntax.Inspect(priorStatement, func(n syntax.Node) bool {
			count++
			if count > 4096 {
				buffered = true
				return false
			}
			if call, ok := n.(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, call, "ob_start") && semanticquery.CallArgument(call.Args, 0, "callback") != nil {
				buffered = true
			}
			return !buffered
		})
		if buffered {
			return
		}
	}
	// Only straight-line known statements preserve header state.
	for i := len(prior) - 1; i >= 0 && len(prior)-i <= 64; i-- {
		st, ok := prior[i].(*syntax.ExprStmt)
		if !ok {
			return
		}
		call, ok := st.Expr.(*syntax.FuncCall)
		if !ok {
			return
		}
		if !semanticquery.NativeBuiltin(ctx, call, "header") {
			return
		}
		header, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "header"))
		if !ok {
			return
		}
		lower := strings.ToLower(strings.TrimSpace(header))
		if strings.HasPrefix(lower, "content-type:") {
			actual := strings.TrimSpace(strings.SplitN(strings.TrimPrefix(lower, "content-type:"), ";", 2)[0])
			// Aliases and unlisted image types are not proven mismatches.
			if canonicalTypes[actual] && actual != mime {
				ctx.ReportNode(c, message)
			}
			return
		}
	}
}
