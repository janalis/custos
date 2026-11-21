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
	// Without a configured target, the fallback version says nothing about
	// where the code runs.
	if ctx.PHP.AtLeast(phpversion.Max) || ctx.PHPAssumed() {
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
		global = fallback
	}
	// A declaration of the project or a package (a polyfill) is available
	// at every version: only builtins introduced later are reported. This
	// cheap test comes first, as almost every call fails it.
	f := ix.Function(global, ctx.PHP)
	if f == nil || !f.Builtin || ctx.PHP.AtLeast(f.Avail.From) {
		return
	}
	// The namespaced function, when it exists, is the one called.
	if fallback != "" && ix.Function(fqn, 0) != nil {
		return
	}
	if !ix.World().Complete || semanticquery.ExistenceChecked(ctx, call, global) {
		return
	}
	ctx.ReportNode(name, global+"() needs PHP "+f.Avail.From.String()+"; the project targets PHP "+ctx.PHP.String()+".")
}
