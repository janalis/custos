// Package unavailablefunction implements the native UnavailableFunction
// inspection.
package unavailablefunction

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "UnavailableFunction" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (rule) Semantic() {}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP.AtLeast(phpversion.Max) {
		return
	}
	call := n.(*syntax.FuncCall)
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return
	}
	ix := ctx.Index()
	fqn, fallback := ctx.Names().Function(name.Value, name.Span().Start)
	global := fqn
	if fallback != "" {
		// The namespaced function, when it exists, is the one called.
		if ix.Function(fqn, 0) != nil {
			return
		}
		global = fallback
	}
	if !ix.World().Complete {
		return
	}
	// A declaration of the project or a package (a polyfill) is available
	// at every version: only builtins introduced later are reported.
	f := ix.Function(global, ctx.PHP)
	if f == nil || !f.Builtin || ctx.PHP.AtLeast(f.Avail.From) {
		return
	}
	if semanticquery.ExistenceChecked(ctx, call, global) {
		return
	}
	ctx.ReportNode(name, global+"() needs PHP "+f.Avail.From.String()+"; the project targets PHP "+ctx.PHP.String()+".")
}
