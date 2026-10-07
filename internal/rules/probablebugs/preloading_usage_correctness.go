package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// preloadingUsageCorrectness reports include/require statements in an
// opcache preload script.
type preloadingUsageCorrectness struct{}

func init() { register(preloadingUsageCorrectness{}) }

func (preloadingUsageCorrectness) ID() string { return "PreloadingUsageCorrectness" }

func (preloadingUsageCorrectness) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KInclude}
}

func (preloadingUsageCorrectness) Check(ctx *analysis.Context, n syntax.Node) {
	p := strings.ReplaceAll(ctx.File.Path, "\\", "/")
	if p[strings.LastIndexByte(p, '/')+1:] != "preload.php" { // D1
		return
	}
	inc := n.(*syntax.Include)
	if inc.Expr == nil || inc.Expr.Span().Len() == 0 { // D2
		return
	}
	if _, ok := inc.Parent().(*syntax.ExprStmt); !ok { // D3
		return
	}
	if preloadRunsScript(inc.Expr) { // E3
		return
	}
	keyword := strings.ToLower(ctx.SpanText(inc.Keyword.Span))
	span := inc.Span()
	arg := util.UnwrapParens(inc.Expr).Span()
	src := ctx.Src
	fn := util.QualifiedBuiltin(ctx, "opcache_compile_file", span.Start) // a namespaced function would capture a bare call
	ctx.Report(span, "Use opcache_compile_file() in a preload script instead of "+keyword+".", analysis.Fix{
		Title: "Use opcache_compile_file()",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: fn + "(" + string(src[arg.Start:arg.End]) + ")"}}
		},
	})
}

// preloadRunsScript reports an inclusion whose path literal names a file that
// must run rather than merely be compiled: a Composer autoloader
// (`vendor/autoload.php`) or another preload script (`….preload.php`, as
// Symfony generates). opcache_compile_file() would not execute it.
func preloadRunsScript(arg syntax.Expr) bool {
	found := false
	syntax.Inspect(arg, func(n syntax.Node) bool {
		raw := ""
		switch x := n.(type) {
		case *syntax.Literal:
			if x.LitKind == syntax.LitString {
				raw = x.Raw
			}
		case *syntax.StringPart:
			raw = x.Raw
		}
		if low := strings.ToLower(raw); strings.Contains(low, "autoload") || strings.Contains(low, "preload") {
			found = true
		}
		return !found
	})
	return found
}
