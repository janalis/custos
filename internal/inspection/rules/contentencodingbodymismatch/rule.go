// Package contentencodingbodymismatch implements the native ContentEncodingBodyMismatch inspection.
package contentencodingbodymismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Encode the body using the declared content coding."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ContentEncodingBodyMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	coding := ""
	switch name {
	case "gzencode":
		coding = "gzip"
	case "gzcompress":
		coding = "deflate"
	case "gzdeflate":
		coding = "raw"
	default:
		return
	}
	// All three functions take the same optional framing selector, which
	// overrides the function's default framing.
	if e, state := semanticquery.CallArgumentBinding(c.Args, 2, "encoding"); state != semanticquery.ArgumentMissing {
		v, ok := framing(ctx, e)
		if !ok {
			return
		}
		switch v {
		case 31:
			coding = "gzip"
		case 15:
			coding = "deflate"
		case -15:
			coding = "raw"
		default:
			return
		}
	}
	var statement syntax.Stmt
	emitted := false
	for p := c.Parent(); p != nil; p = p.Parent() {
		if paren, ok := p.(*syntax.Paren); ok {
			_ = paren
			continue
		}
		switch p.(type) {
		case *syntax.Echo:
			statement = p.(syntax.Stmt)
		case *syntax.ExprStmt:
			if !emitted {
				return
			}
			statement = p.(syntax.Stmt)
		case *syntax.Print:
			emitted = true
			continue
		default:
			return
		}
		break
	}
	prior := flowquery.NativePriorStatements(ctx.File, statement)
	if len(prior) > 4096 {
		return
	}
	var declared *syntax.FuncCall
	value := ""
	for i := len(prior) - 1; i >= 0; i-- {
		st, ok := prior[i].(*syntax.ExprStmt)
		if !ok {
			return
		}
		h, ok := st.Expr.(*syntax.FuncCall)
		if !ok {
			return
		}
		key, text, _, known := semanticquery.ExpansionHeader(ctx, h)
		if !known {
			return
		}
		if key == "content-encoding" && declared == nil {
			declared = h
			value = text
		}
	}
	if declared == nil {
		return
	}
	response := semanticquery.ExpansionResponseAt(ctx, declared)
	if response.Final && response.Coding == value && (value == "gzip" || value == "deflate" || value == "identity") && coding != value {
		ctx.ReportNode(c, message)
	}
}

// framing evaluates an encoding argument. Negative builtin constants such as
// ZLIB_ENCODING_RAW are resolved here because the shared integer query only
// understands unsigned constant values.
func framing(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if v, ok := semanticquery.NativeInt(ctx, e); ok {
		return v, true
	}
	if k, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch); ok {
		if resolved := semanticquery.ResolveConstant(ctx.Types(), k); resolved != nil {
			v, ok := astquery.ParseIntLiteral(strings.TrimPrefix(resolved.Value, "-"))
			return -v, ok
		}
	}
	return 0, false
}
