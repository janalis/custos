// Package dynamiccorsoriginwithoutvary implements the native DynamicCorsOriginWithoutVary inspection.
package dynamiccorsoriginwithoutvary

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Include Vary: Origin for cacheable origin-dependent responses."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DynamicCorsOriginWithoutVary" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "header") {
		return
	}
	b, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(call.Args, 0, "header")).(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TDot {
		return
	}
	prefix, known := semanticquery.NativeString(ctx, b.Left)
	field, requestValue := request(ctx, b.Right)
	if !known || !requestValue || field != "HTTP_ORIGIN" || strings.ToLower(strings.TrimSpace(prefix)) != "access-control-allow-origin:" {
		return
	}
	stmt, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	list, index, ok := astquery.StmtList(ctx.File, stmt)
	if !ok {
		return
	}
	cached := false
	for _, s := range list[index+1:] {
		st, ok := s.(*syntax.ExprStmt)
		if !ok {
			return
		}
		c, ok := st.Expr.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, c, "header") {
			return
		}
		value, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "header"))
		if !known {
			return
		}
		key, v, found := strings.Cut(value, ":")
		if !found {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		v = strings.ToLower(strings.TrimSpace(v))
		switch key {
		case "access-control-allow-origin":
			return
		case "vary":
			for _, part := range strings.Split(v, ",") {
				if p := strings.TrimSpace(part); p == "origin" || p == "*" {
					return
				}
			}
		case "cache-control":
			cached = cacheable(v)
		}
	}
	if cached {
		ctx.ReportNode(call, message)
	}
}

func request(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	x, ok := semanticquery.NativeValue(ctx, e).(*syntax.ArrayDimFetch)
	if !ok {
		return "", false
	}
	v, ok := x.Var.(*syntax.Variable)
	if !ok {
		return "", false
	}
	switch v.Name {
	case "_SERVER":
	default:
		return "", false
	}
	if !semanticquery.NativeRequestUnwritten(ctx, x) {
		return "", false
	}
	key, known := semanticquery.NativeString(ctx, x.Dim)
	return key, known
}

func cacheable(value string) bool {
	shared := false
	for _, item := range strings.Split(value, ",") {
		key, v, hasValue := strings.Cut(strings.TrimSpace(item), "=")
		switch key {
		case "private", "no-store", "no-cache":
			return false
		case "public":
			shared = true
		case "max-age", "s-maxage":
			if hasValue {
				digits := strings.Trim(v, "\"")
				n, err := strconv.ParseInt(digits, 10, 64)
				if err == nil && n > 0 {
					shared = true
				}
			}
		}
	}
	return shared
}
