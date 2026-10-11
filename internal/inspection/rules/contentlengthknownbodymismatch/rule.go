// Package contentlengthknownbodymismatch implements the native ContentLengthKnownBodyMismatch inspection.
package contentlengthknownbodymismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match Content-Length to the emitted bytes."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ContentLengthKnownBodyMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, value, _, known := semanticquery.ExpansionHeader(ctx, c)
	if !known || key != "content-length" {
		return
	}
	size, known := semanticquery.ExpansionDecimalLength(value)
	if !known {
		return
	}
	response := semanticquery.ExpansionResponseAt(ctx, c)
	if !response.Final {
		return
	}
	body, known := response.BodyLength, response.BodyKnown
	if !known {
		body, known = emptyExitBody(ctx, c)
	}
	if known && body != size {
		ctx.ReportNode(c, message)
	}
}

// emptyExitBody proves the body of a path closed by exit() or die() with an
// empty argument list, which the shared response scan leaves unproven. It
// accepts only output of known length and plain literal headers that cannot
// change the status or the framing; anything else stays unproven.
func emptyExitBody(ctx *analysis.Context, c *syntax.FuncCall) (int64, bool) {
	// A final header is always a direct statement of the scanned list.
	n := c.Parent()
	for {
		if _, ok := n.(*syntax.ExprStmt); ok {
			break
		}
		n = n.Parent()
	}
	list := ctx.File.Stmts
	switch parent := n.Parent().(type) {
	case *syntax.Block:
		list = parent.Stmts
	case *syntax.Namespace:
		list = parent.Stmts
	}
	total := int64(0)
	for _, statement := range list {
		switch st := statement.(type) {
		case *syntax.Echo:
			for _, e := range st.Exprs {
				size, ok := byteLength(ctx, e, 0)
				if !ok {
					return 0, false
				}
				total += size
			}
		case *syntax.ExprStmt:
			switch value := syntax.UnwrapParens(st.Expr).(type) {
			case *syntax.Exit:
				return total, value.Args != nil && len(value.Args.Args) == 0
			case *syntax.Print:
				size, ok := byteLength(ctx, value.Expr, 0)
				if !ok {
					return 0, false
				}
				total += size
			case *syntax.FuncCall:
				key, _, _, ok := semanticquery.ExpansionHeader(ctx, value)
				if !ok || key == "location" || key == "status" || key == "transfer-encoding" ||
					semanticquery.CallArgument(value.Args, 2, "response_code") != nil {
					return 0, false
				}
			default:
				return 0, false
			}
		default:
			return 0, false
		}
	}
	return 0, false
}

// byteLength mirrors the shared scan's bounded length proof for literal
// strings, concatenations and str_repeat.
func byteLength(ctx *analysis.Context, e syntax.Expr, depth int) (int64, bool) {
	if depth > 8 {
		return 0, false
	}
	if text, ok := semanticquery.NativeString(ctx, e); ok {
		return int64(len(text)), true
	}
	switch value := semanticquery.NativeValue(ctx, e).(type) {
	case *syntax.Binary:
		if value.Op.Kind != syntax.TDot {
			return 0, false
		}
		a, ak := byteLength(ctx, value.Left, depth+1)
		b, bk := byteLength(ctx, value.Right, depth+1)
		return a + b, ak && bk && a+b <= 1<<30
	case *syntax.FuncCall:
		if !semanticquery.NativeBuiltin(ctx, value, "str_repeat") {
			return 0, false
		}
		times, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(value.Args, 1, "times"))
		if !known || times < 0 || times > 1<<30 {
			return 0, false
		}
		size, known := byteLength(ctx, semanticquery.CallArgument(value.Args, 0, "string"), depth+1)
		if !known || (size != 0 && times > (1<<30)/size) {
			return 0, false
		}
		return size * times, true
	}
	return 0, false
}
