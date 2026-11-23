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

// peclUntil maps the extensions PHP bundles that PECL also builds for older
// PHP versions to the version bundling them: below it, a project requiring
// one of them may run that build, which declares the same functions.
var peclUntil = map[string]phpversion.Version{"sodium": phpversion.PHP72, "zend-opcache": phpversion.PHP55}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	// Without a configured target, the fallback version says nothing about
	// where the code runs.
	if ctx.PHPAssumed() {
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
	// at every version, so only builtins are reported, and only those of an
	// extension shipping with PHP: it dates its functions by PHP release,
	// and no other build brings back what PHP removed from it. This cheap
	// test comes first, as almost every call fails it.
	f := ix.Function(global, ctx.PHP)
	if f == nil || !f.Builtin || !f.Bundled || f.Avail.In(ctx.PHP) {
		return
	}
	// The namespaced function, when it exists, is the one called.
	if fallback != "" && ix.Function(fqn, 0) != nil {
		return
	}
	world := ix.World()
	if !world.Complete || (ctx.PHP.Below(peclUntil[f.Ext]) && world.Extensions[f.Ext]) || semanticquery.ExistenceChecked(ctx, call, global) {
		return
	}
	target := "; the project targets PHP " + ctx.PHP.String() + "."
	if f.Avail.From > ctx.PHP {
		ctx.ReportNode(name, global+"() needs PHP "+f.Avail.From.String()+target)
		return
	}
	// Stubs write 7.99 for "until 8.0". The declaration ends below the
	// target, so a later release exists.
	removed, _ := phpversion.Next(f.Avail.To)
	ctx.ReportNode(name, global+"() was removed in PHP "+removed.String()+target)
}

// PartialCalls: a first-class callable or partial application refers to
// its callee as a call does.
func (rule) PartialCalls() {}
