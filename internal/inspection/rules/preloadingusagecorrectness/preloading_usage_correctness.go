package preloadingusagecorrectness

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/stubs"
)

// preloadingUsageCorrectness reports include/require statements in an
// opcache preload script.
type preloadingUsageCorrectness struct{}

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
	if preloadRunsScript(inc.Expr) || preloadRunsCodeAfter(ctx, inc) { // E3, E4
		return
	}
	keyword := strings.ToLower(ctx.SpanText(inc.Keyword.Span))
	span := inc.Span()
	arg := syntax.UnwrapParens(inc.Expr).Span()
	src := ctx.Src
	fn := semanticquery.QualifiedBuiltin(ctx, "opcache_compile_file", span.Start) // a namespaced function would capture a bare call
	ctx.Report(span, "Use opcache_compile_file() in a preload script instead of "+keyword+".", diagnostic.Fix{
		Title: "Use opcache_compile_file()",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: fn + "(" + string(src[arg.Start:arg.End]) + ")"}}
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

// preloadRunsCodeAfter reports an inclusion followed, in its statement list,
// by code that may need it to have run (E4): `new`, a method or static call,
// or a call to a function that is not a PHP builtin (EspoCRM's preload.php
// includes bootstrap.php, which registers the autoloader, then runs
// `(new Application())->run(Preload::class)`). Declarations are skipped.
func preloadRunsCodeAfter(ctx *analysis.Context, inc *syntax.Include) bool {
	list, i, ok := astquery.StmtList(ctx.File, inc.Parent().(*syntax.ExprStmt))
	if !ok {
		return false
	}
	found := false
	for _, s := range list[i+1:] {
		syntax.Inspect(s, func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.Function, *syntax.ClassLike:
				return false
			case *syntax.New, *syntax.MethodCall, *syntax.StaticCall:
				found = true
			case *syntax.FuncCall:
				name := ctx.GlobalFunctionName(x)
				found = name == "" || stubs.Index().Function(name, ctx.PHP) == nil
			}
			return !found
		})
		if found {
			return true
		}
	}
	return false
}
