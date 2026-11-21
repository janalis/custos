// Package unavailablefunction implements the native UnavailableFunction
// inspection.
package unavailablefunction

import (
	"slices"

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

// alsoPECL lists the extensions PHP bundles that PECL also builds for PHP
// versions before the one bundling them, with the same functions: a project
// requiring one of them may run that build.
var alsoPECL = map[string]bool{"sodium": true, "zip": true}

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
	// at every version: only builtins introduced later are reported. This
	// cheap test comes first, as almost every call fails it.
	// Only an extension shipping with PHP dates its functions by PHP
	// release, and none can bring back the functions PHP removed from it.
	f := ix.Function(global, ctx.PHP)
	if f == nil || !f.Builtin || !f.Bundled || f.Avail.In(ctx.PHP) {
		return
	}
	// The namespaced function, when it exists, is the one called.
	if fallback != "" && ix.Function(fqn, 0) != nil {
		return
	}
	world := ix.World()
	if !world.Complete || (alsoPECL[f.Ext] && world.Extensions[f.Ext]) || semanticquery.ExistenceChecked(ctx, call, global) {
		return
	}
	target := "; the project targets PHP " + ctx.PHP.String() + "."
	if f.Avail.From > ctx.PHP {
		ctx.ReportNode(name, global+"() needs PHP "+f.Avail.From.String()+target)
		return
	}
	ctx.ReportNode(name, global+"() was removed in PHP "+after(f.Avail.To).String()+target)
}

// releases are the PHP versions custos supports, in order.
var releases = []phpversion.Version{
	phpversion.PHP53, phpversion.PHP54, phpversion.PHP55, phpversion.PHP56,
	phpversion.PHP70, phpversion.PHP71, phpversion.PHP72, phpversion.PHP73, phpversion.PHP74,
	phpversion.PHP80, phpversion.PHP81, phpversion.PHP82, phpversion.PHP83, phpversion.PHP84, phpversion.PHP85,
}

// after returns the release following v, the last version a declaration
// is available in (stubs write 7.99 for "until 8.0"): the one removing it.
// A declaration removed after the target is available there, so v is below
// the latest release.
func after(v phpversion.Version) phpversion.Version {
	i, _ := slices.BinarySearch(releases, v+1)
	return releases[i]
}
