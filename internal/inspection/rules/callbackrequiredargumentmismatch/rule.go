// Package callbackrequiredargumentmismatch implements the native CallbackRequiredArgumentMismatch inspection.
package callbackrequiredargumentmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the callback parameters to the supplied arrays."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CallbackRequiredArgumentMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_map") {
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) < 2 {
		return
	}
	params, body := semanticquery.NativeCallback(ctx, args[0])
	if body == nil || !settledDeclaration(ctx.File, body) {
		return
	}
	required := 0
	for _, param := range params {
		if param.Default == nil && !param.Variadic {
			required++
		}
	}
	if required > len(args)-1 {
		ctx.ReportNode(call, message)
	}
}

// settledDeclaration reports whether a callback body is the one definition
// the call can reach. Closures and arrow functions always are; a named
// function must be a direct statement of the file or of a namespace and the
// only function of that name in the file, since a conditional or repeated
// declaration leaves the signature bound at run time unknown.
func settledDeclaration(f *syntax.File, body syntax.Node) bool {
	decl, ok := body.Parent().(*syntax.Function)
	if !ok {
		return true
	}
	if parent := decl.Parent(); parent != nil {
		_, namespace := parent.(*syntax.Namespace)
		_, file := parent.(*syntax.File)
		if !namespace && !file {
			return false
		}
	}
	count := 0
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if other, ok := n.(*syntax.Function); ok && strings.EqualFold(other.Name.Value, decl.Name.Value) {
			count++
		}
		return true
	})
	return count == 1
}
