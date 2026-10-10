// Package credentialedcorsuseswildcardorigin implements the native CredentialedCorsUsesWildcardOrigin inspection.
package credentialedcorsuseswildcardorigin

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use an explicit origin for credentialed CORS."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CredentialedCorsUsesWildcardOrigin" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "header") {
		return
	}
	headers, last, known := finalHeaders(ctx, call)
	if known && last == call && headers["access-control-allow-origin"] == "*" && headers["access-control-allow-credentials"] == "true" {
		ctx.ReportNode(call, message)
	}
}

func finalHeaders(ctx *analysis.Context, call *syntax.FuncCall) (map[string]string, syntax.Node, bool) {
	stmt, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return nil, nil, false
	}
	list, index, ok := astquery.StmtList(ctx.File, stmt)
	if !ok {
		return nil, nil, false
	}
	start := index
	for start > 0 {
		if !headerStmt(ctx, list[start-1]) {
			break
		}
		start--
	}
	headers := map[string]string{}
	var last syntax.Node
	for _, s := range list[start:] {
		st, ok := s.(*syntax.ExprStmt)
		if !ok {
			return nil, nil, false
		}
		c, ok := st.Expr.(*syntax.FuncCall)
		if !ok {
			return nil, nil, false
		}
		name := semanticquery.NativeBuiltinName(ctx, c)
		if name != "header" && name != "header_remove" {
			return nil, nil, false
		}
		if name == "header_remove" {
			e := semanticquery.CallArgument(c.Args, 0, "name")
			if e == nil {
				clear(headers)
				continue
			}
			value, known := semanticquery.NativeString(ctx, e)
			if !known {
				return nil, nil, false
			}
			delete(headers, strings.ToLower(value))
			continue
		}
		replace := semanticquery.CallArgument(c.Args, 1, "replace")
		if replace != nil {
			b, known := semanticquery.NativeTruth(ctx, replace)
			if !known || !b {
				return nil, nil, false
			}
		}
		value, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "header"))
		if !known {
			return nil, nil, false
		}
		key, v, found := strings.Cut(value, ":")
		if !found {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		headers[key] = strings.ToLower(strings.TrimSpace(v))
		if key == "access-control-allow-origin" || key == "access-control-allow-credentials" {
			last = c
		}
	}
	return headers, last, true
}

func headerStmt(ctx *analysis.Context, s syntax.Stmt) bool {
	st, ok := s.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	c, ok := st.Expr.(*syntax.FuncCall)
	if !ok {
		return false
	}
	name := semanticquery.NativeBuiltinName(ctx, c)
	return name == "header" || name == "header_remove"
}
