// Package opensslverifytruthyresult implements the native OpenSslVerifyTruthyResult inspection.
package opensslverifytruthyresult

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Accept only verification result 1."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OpenSslVerifyTruthyResult" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "openssl_verify") {
		return
	}
	var condition syntax.Expr = call
	negated := false
	if u, ok := call.Parent().(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		condition = u
		negated = true
	}
	fixNode := condition
	for p, ok := condition.Parent().(*syntax.Paren); ok; p, ok = condition.Parent().(*syntax.Paren) {
		condition = p
	}
	used := false
	switch parent := condition.Parent().(type) {
	case *syntax.If:
		used = parent.Cond == condition
	case *syntax.While:
		used = parent.Cond == condition
	}
	if !used {
		return
	}
	text := ctx.Text(call) + " === 1"
	if negated {
		text = ctx.Text(call) + " !== 1"
	}
	fix := diagnostic.Fix{Title: "Require successful signature verification", Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: fixNode.Span(), NewText: text}} }}
	if strings.Contains(ctx.Text(fixNode), "/*") || strings.Contains(ctx.Text(fixNode), "//") || strings.Contains(ctx.Text(fixNode), "#") {
		ctx.ReportNode(call, message)
		return
	}
	ctx.ReportNode(call, message, fix)
}
