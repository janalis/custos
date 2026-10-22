// Package fastdigestusedforpasswordstorage implements the native FastDigestUsedForPasswordStorage inspection.
package fastdigestusedforpasswordstorage

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

const message = "Store passwords with a password hashing API."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FastDigestUsedForPasswordStorage" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)

	if !semanticquery.CallsAnnotatedFunction(ctx, c, "@custos-credential-store") {
		return
	}
	digest, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "")).(*syntax.FuncCall)
	if !ok {
		return
	}
	hash := semanticquery.NativeBuiltinName(ctx, digest)
	if hash != "md5" && hash != "sha1" && hash != "hash" {
		return
	}
	position := 0
	if hash == "hash" {
		position = 1
	}
	if passwordParameter(ctx, c, semanticquery.CallArgument(digest.Args, position, "data")) {
		ctx.ReportNode(c, message)
	}
}

func passwordParameter(ctx *analysis.Context, at syntax.Node, input syntax.Expr) bool {
	scope := syntax.EnclosingVariableScope(at)
	if scope == nil {
		return false
	}
	comment := index.DocComment(ctx.File, scope)
	value := ctx.Flow().Value(input)
	if !value.Complete {
		return false
	}
	for _, source := range value.Sources {
		if source.Kind != "parameter" {
			continue
		}
		params := syntax.VariableScopeParams(scope)
		name := params[source.Parameter].Var.Name
		for _, line := range strings.Split(comment, "\n") {
			fields := strings.Fields(strings.TrimSpace(strings.Trim(line, "/*")))
			for i, field := range fields {
				if field == "@param" && i+2 < len(fields) && fields[i+1] == "password-string" && fields[i+2] == "$"+name {
					return true
				}
			}
		}
	}
	return false
}
