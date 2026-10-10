// Package contentencodingbodymismatch implements the native ContentEncodingBodyMismatch inspection.
package contentencodingbodymismatch

import (
	"custos/internal/inspection/analysis"
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
	if name == "gzencode" {
		if e := semanticquery.CallArgument(c.Args, 2, "encoding"); e != nil {
			v, ok := semanticquery.NativeInt(ctx, e)
			if !ok {
				return
			}
			if v == 15 {
				coding = "deflate"
			} else if v != 31 {
				return
			}
		}
	}
	if name == "gzcompress" {
		if e := semanticquery.CallArgument(c.Args, 2, "encoding"); e != nil {
			v, ok := semanticquery.NativeInt(ctx, e)
			if !ok {
				return
			}
			switch v {
			case 31:
				coding = "gzip"
			case -15:
				coding = "raw"
			case 15:
			default:
				return
			}
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
