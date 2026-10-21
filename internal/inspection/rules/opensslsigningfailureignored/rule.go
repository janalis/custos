// Package opensslsigningfailureignored implements OpenSslSigningFailureIgnored.
package opensslsigningfailureignored

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check signing success before publishing the signature."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OpenSslSigningFailureIgnored" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if signatureChecked(ctx, n.(*syntax.FuncCall)) {
		return
	}
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "openssl_sign") {
		return
	}
	if _, ok := c.Parent().(*syntax.ExprStmt); !ok {
		return
	}
	signature := semanticquery.CallArgument(c.Args, 1, "signature")
	for _, event := range expansionDOutputs(ctx, c) {
		if !semanticquery.NativeDominates(c, event) {
			continue
		}
		var inputs []syntax.Expr
		switch use := event.(type) {
		case *syntax.FuncCall:
			switch semanticquery.NativeBuiltinName(ctx, use) {
			case "base64_encode", "bin2hex":
				inputs = []syntax.Expr{semanticquery.CallArgument(use.Args, 0, "string")}
			case "file_put_contents":
				inputs = []syntax.Expr{semanticquery.CallArgument(use.Args, 1, "data")}
			}
		case *syntax.Echo:
			inputs = use.Exprs
		}
		for _, input := range inputs {
			if semanticquery.SameStableLocal(ctx, signature, input) {
				ctx.ReportNode(c, message)
				return
			}
		}
	}
}

func expansionDOutputs(ctx *analysis.Context, at syntax.Node) []syntax.Node {
	scope := syntax.EnclosingVariableScope(at)
	return ctx.File.Memo(expansionDOutputKey{scope}, func() any {
		var out []syntax.Node
		visit := func(n syntax.Node) bool {
			if n != scope && syntax.IsVariableScope(n) {
				return false
			}
			switch n.(type) {
			case *syntax.Echo, *syntax.FuncCall:
				out = append(out, n)
			}
			return true
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		if len(out) > 512 {
			return []syntax.Node(nil)
		}
		return out
	}).([]syntax.Node)
}

type expansionDOutputKey struct{ scope syntax.Node }
