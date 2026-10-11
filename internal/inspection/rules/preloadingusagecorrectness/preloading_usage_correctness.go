package preloadingusagecorrectness

import (
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
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
	fn := semanticquery.QualifiedBuiltin(ctx, "opcache_compile_file", span.Start) // a namespaced function would capture a bare call
	ctx.Report(span, "Use opcache_compile_file() in a preload script instead of "+keyword+".", diagnostic.Fix{
		Title: "Use opcache_compile_file()",
		Edits: func() []diagnostic.TextEdit { return preloadEdits(ctx, inc, fn) },
	})
}

// preloadEdits rewrites the inclusion into a call of fn (F1). Without
// comments the whole expression is replaced and the parentheses around the
// argument are dropped; with a comment between the keyword and the argument,
// or inside those parentheses, only the keyword is replaced so that the
// comment stays where it was written.
func preloadEdits(ctx *analysis.Context, inc *syntax.Include, fn string) []diagnostic.TextEdit {
	span, src := inc.Span(), ctx.Src
	arg := syntax.UnwrapParens(inc.Expr).Span()
	if !astquery.HasComment(ctx.File, syntax.Span{Start: span.Start, End: arg.Start}) &&
		!astquery.HasComment(ctx.File, syntax.Span{Start: arg.End, End: span.End}) {
		return []diagnostic.TextEdit{{Span: span, NewText: fn + "(" + string(src[arg.Start:arg.End]) + ")"}}
	}
	kw := syntax.Span{Start: span.Start, End: inc.Keyword.Span.End}
	for strings.IndexByte(" \t\r\n", src[kw.End]) >= 0 {
		kw.End++
	}
	if _, ok := inc.Expr.(*syntax.Paren); ok { // its parentheses become the call's
		return []diagnostic.TextEdit{{Span: kw, NewText: fn}}
	}
	return []diagnostic.TextEdit{
		{Span: kw, NewText: fn + "("},
		{Span: syntax.Span{Start: span.End, End: span.End}, NewText: ")"},
	}
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

// preloadRunsCodeAfter reports an inclusion followed, in its statement list
// (anywhere later in its namespace block, function or file for variable reads),
// by code that may need it to have run (E4): `new`, a method or static call,
// a call to a function that is not a PHP builtin (EspoCRM's preload.php
// includes bootstrap.php, which registers the autoloader, then runs
// `(new Application())->run(Preload::class)`), or a read of a variable or
// constant that the script itself has not defined by then. Declarations are
// skipped.
func preloadRunsCodeAfter(ctx *analysis.Context, inc *syntax.Include) bool {
	stmt := inc.Parent().(*syntax.ExprStmt)
	list, i, ok := astquery.StmtList(ctx.File, stmt)
	// Variables outlive the block holding the inclusion: reads are looked
	// for up to the end of the enclosing namespace block, function or file.
	after := syntax.Span{Start: stmt.Span().End, End: stmt.Span().End}
	p := stmt.Parent()
	for ; p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		after.End = max(after.End, p.Span().End)
		if _, ns := p.(*syntax.Namespace); ns {
			break
		}
	}
	if p == nil {
		after.End = uint32(len(ctx.Src))
	}
	for _, off := range preloadFreeReads(ctx, syntax.EnclosingVariableScope(inc)) {
		if off >= after.Start && off < after.End {
			return true
		}
	}
	if !ok {
		return false
	}
	found := false
	for _, s := range list[i+1:] {
		syntax.Inspect(s, func(n syntax.Node) bool {
			if found { // a later sibling must not clear the verdict
				return false
			}
			switch x := n.(type) {
			case *syntax.Function, *syntax.ClassLike:
				return false
			case *syntax.New, *syntax.MethodCall, *syntax.StaticCall:
				found = true
			case *syntax.ConstFetch:
				fqn, fallback := ctx.Names().Const(x.Name.Value, x.Span().Start)
				found = !preloadKnownConst(ctx, fqn, fallback, x.Span().Start)
			case *syntax.FuncCall:
				name := ctx.GlobalFunctionName(x)
				if c, ok := preloadConstName(x); ok && (name == "defined" || name == "constant") {
					found = !preloadKnownConst(ctx, c, "", x.Span().Start)
					break
				}
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

// preloadFreeReads returns the offsets, in the given variable scope (nil for
// the top-level code), of the reads of variables the script has not assigned
// yet at that point: whatever they hold comes from an included file.
// Parameters, `$this` and the variables PHP itself provides are never free.
func preloadFreeReads(ctx *analysis.Context, scope syntax.Node) []uint32 {
	key := "free-reads"
	if scope != nil {
		key += strconv.Itoa(int(scope.Span().Start))
	}
	return ctx.Memo(key, func() any {
		given := map[string]bool{
			"this": true, "argv": true, "argc": true, "GLOBALS": true, "_SERVER": true, "_ENV": true,
			"_GET": true, "_POST": true, "_FILES": true, "_COOKIE": true, "_SESSION": true, "_REQUEST": true,
		}
		for _, p := range syntax.VariableScopeParams(scope) {
			given[p.Var.Name] = true
		}
		var free []uint32
		for name, accesses := range flowquery.VarAccessesByName(ctx.File, scope) {
			if given[name] {
				continue
			}
			// A write takes effect once its statement's right-hand side
			// has been evaluated: `$paths = array_merge($paths, $more)`
			// still reads the included file's $paths.
			set := ^uint32(0)
			for _, a := range accesses {
				if !a.Write {
					continue
				}
				at := a.Var.Span().Start
				if as, ok := a.By.(*syntax.Assign); ok {
					at = as.Span().End
				} else if a.Compound {
					at = a.Var.Span().End
				}
				set = min(set, at)
			}
			for _, a := range accesses {
				if at := a.Var.Span().Start; (!a.Write || a.Compound) && at < set {
					free = append(free, at)
				}
			}
		}
		return free
	}).([]uint32)
}

// preloadConstName returns the constant named by the string literal a
// `defined()` / `constant()` call receives first.
func preloadConstName(call *syntax.FuncCall) (string, bool) {
	if call.Args == nil || len(call.Args.Args) == 0 {
		return "", false
	}
	arg, ok := call.Args.Args[0].(*syntax.Arg)
	if !ok {
		return "", false
	}
	lit, ok := syntax.UnwrapParens(arg.Value).(*syntax.Literal)
	if !ok {
		return "", false
	}
	name, ok := astquery.StringLiteralValue(lit.Raw)
	return strings.TrimPrefix(name, `\`), ok
}

// preloadKnownConst reports whether a constant read at offset at exists
// without the help of an included file: `true`/`false`/`null`, a PHP builtin,
// or a constant the script declared (`const`, `define('…')`) before the read.
// fallback is the global name an unqualified reference falls back to.
func preloadKnownConst(ctx *analysis.Context, fqn, fallback string, at uint32) bool {
	declared := ctx.Memo("consts", func() any {
		m := map[string]uint32{}
		add := func(name string, end uint32) {
			if _, ok := m[name]; !ok {
				m[name] = end
			}
		}
		collect := func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.ConstStmt:
				ns := ctx.Names().Namespace(x.Span().Start)
				for _, c := range x.Consts {
					add(strings.TrimPrefix(ns+`\`+c.Name.Value, `\`), x.Span().End)
				}
			case *syntax.FuncCall:
				if name, ok := preloadConstName(x); ok && ctx.GlobalFunctionName(x) == "define" {
					add(name, x.Span().End)
				}
			}
			return true
		}
		for _, st := range ctx.File.Stmts {
			syntax.Inspect(st, collect)
		}
		return m
	}).(map[string]uint32)
	for _, name := range []string{fqn, fallback} {
		switch strings.ToLower(name) {
		case "":
			continue
		case "true", "false", "null":
			return true
		}
		if end, ok := declared[name]; ok && end <= at {
			return true
		}
		if stubs.Index().Constant(name, ctx.PHP) != nil {
			return true
		}
	}
	return false
}
