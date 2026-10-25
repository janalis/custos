package forgottendebugoutput

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// forgottenDebugOutput reports calls of dumping/tracing helpers that were
// probably left over from debugging.
type forgottenDebugOutput struct{}

func (forgottenDebugOutput) ID() string { return "ForgottenDebugOutput" }
func (forgottenDebugOutput) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall}
}

// Semantic marks the rule as needing the project index.
func (forgottenDebugOutput) Semantic() {}

var debugDefaultEntries = []string{
	`\Codeception\Util\Debug::pause`, `\Codeception\Util\Debug::debug`,
	`\Doctrine\Common\Util\Debug::dump`, `\Doctrine\Common\Util\Debug::export`,
	`\Symfony\Component\Debug\Debug::enable`, `\Symfony\Component\Debug\ErrorHandler::register`,
	`\Symfony\Component\Debug\ExceptionHandler::register`, `\Symfony\Component\Debug\DebugClassLoader::enable`,
	`\Zend\Debug\Debug::dump`, `\Zend\Di\Display\Console::export`,
	`\TYPO3\CMS\Core\Utility\DebugUtility::debug`, `\Illuminate\Support\Debug\Dumper::dump`,
	"dd", "dump", "trap", "debug_print_backtrace", "debug_zval_dump", "error_log", "phpinfo", "print_r",
	"var_export", "var_dump", "dpm", "dsm", "dvm", "kpr", "dpq",
	"xdebug_break", "xdebug_call_class", "xdebug_call_file", "xdebug_call_function", "xdebug_call_line",
	"xdebug_code_coverage_started", "xdebug_debug_zval", "xdebug_debug_zval_stdout", "xdebug_dump_superglobals",
	"xdebug_enable", "xdebug_get_code_coverage", "xdebug_get_collected_errors", "xdebug_get_declared_vars",
	"xdebug_get_function_stack", "xdebug_get_headers", "xdebug_get_monitored_functions",
	"xdebug_get_profiler_filename", "xdebug_get_stack_depth", "xdebug_get_tracefile_name", "xdebug_is_enabled",
	"xdebug_memory_usage", "xdebug_peak_memory_usage", "xdebug_print_function_stack",
	"xdebug_start_code_coverage", "xdebug_start_error_collection", "xdebug_start_function_monitor",
	"xdebug_start_trace", "xdebug_stop_code_coverage", "xdebug_stop_error_collection",
	"xdebug_stop_function_monitor", "xdebug_stop_trace", "xdebug_time_index", "xdebug_var_dump",
}

// debugArgAllowance: calls with exactly this many arguments are intentional.
var debugArgAllowance = map[string]int{"phpinfo": 1, "print_r": 2, "var_export": 2}

const debugOutputMsg = "Debug output call; remove it if it was left over from debugging."

// debugEntries returns the effective, trimmed debug entry list.
func debugEntries(ctx *analysis.Context) []string {
	var out []string
	if !ctx.Bool("migratedIntoUserSpace") {
		out = append(out, debugDefaultEntries...)
	}
	conf := ctx.List("configuration")
	if conf == nil && ctx.String("configuration") != "" { // comma-separated text form
		conf = strings.Split(ctx.String("configuration"), ",")
	}
	for _, e := range conf {
		out = append(out, strings.TrimSpace(e))
	}
	for _, c := range ctx.List("@calls") { // registerCustomDebugMethod("<entry>")
		if !strings.HasPrefix(c, "registerCustomDebugMethod(") {
			continue
		}
		i, j := strings.IndexByte(c, '"'), strings.LastIndexByte(c, '"')
		if i < 0 || j <= i {
			continue
		}
		raw := c[i+1 : j]
		raw = strings.NewReplacer(`\\`, `\`, `\"`, `"`).Replace(raw)
		out = append(out, strings.TrimSpace(raw))
	}
	return out
}

func (forgottenDebugOutput) Check(ctx *analysis.Context, n syntax.Node) {
	entries := ctx.Memo("entries", func() any { return debugEntries(ctx) }).([]string)
	if debugCallableSyntax(n) { // E5: f(...) only creates a Closure
		return
	}
	switch x := n.(type) {
	case *syntax.FuncCall:
		name := astquery.CallLastName(x)
		if name == "" || !debugHasEntry(entries, name) { // D1
			return
		}
		if want, ok := debugArgAllowance[strings.ToLower(name)]; ok && astquery.ArgCount(x) == want { // D3
			return
		}
		if debugOutputBuffered(ctx.File, x) || debugCaptured(ctx.File, x) || debugInWrapper(ctx, x, entries) { // D4
			return
		}
		ctx.ReportNode(x, debugOutputMsg)
	case *syntax.MethodCall:
		debugCheckMethod(ctx, x, x.Name, x.Var, entries)
	case *syntax.StaticCall:
		debugCheckMethod(ctx, x, x.Name, x.Class, entries)
	}
}

// debugCallableSyntax reports a first-class callable `f(...)`.
func debugCallableSyntax(n syntax.Node) bool {
	var args *syntax.ArgList
	switch x := n.(type) {
	case *syntax.FuncCall:
		args = x.Args
	case *syntax.MethodCall:
		args = x.Args
	case *syntax.StaticCall:
		args = x.Args
	}
	if args == nil || len(args.Args) != 1 {
		return false
	}
	_, ok := args.Args[0].(*syntax.VariadicPlaceholder)
	return ok
}

// debugHasEntry compares case-insensitively: PHP function, class and method
// names are case-insensitive.
func debugHasEntry(entries []string, s string) bool {
	for _, e := range entries {
		if strings.EqualFold(e, s) {
			return true
		}
	}
	return false
}

func debugCheckMethod(ctx *analysis.Context, call syntax.Node, name, recv syntax.Expr, entries []string) {
	id, ok := name.(*syntax.Identifier)
	if !ok {
		return
	}
	var classes []string // D5: entries whose method part matches
	for _, e := range entries {
		cls, meth, ok := strings.Cut(e, "::")
		if ok && strings.EqualFold(meth, id.Value) {
			classes = append(classes, cls)
		}
	}
	if len(classes) == 0 {
		return
	}
	declaring := map[string]bool{} // D6
	for _, c := range semanticquery.ReferencedClasses(ctx, recv) {
		if m := ctx.Index().FindMethod(c, id.Value, ctx.PHP); m != nil {
			declaring[strings.ToLower(`\`+strings.TrimPrefix(m.Class, `\`))] = true
		}
	}
	for _, cls := range classes {
		if declaring[strings.ToLower(cls)] {
			if !debugInWrapper(ctx, call, entries) { // D7
				ctx.ReportNode(call, debugOutputMsg)
			}
			return
		}
	}
}

// debugOutputBuffered implements E2: `ob_start(); <call>;`.
func debugOutputBuffered(f *syntax.File, call *syntax.FuncCall) bool {
	p := call.Parent()
	if u, ok := p.(*syntax.Unary); ok && u.Op.Kind == syntax.TAt {
		p = u.Parent()
	}
	st, ok := p.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	prev, ok := astquery.PrevStmt(f, st)
	if !ok {
		return false
	}
	ps, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	pc, ok := ps.Expr.(*syntax.FuncCall)
	return ok && strings.EqualFold(astquery.CallLastName(pc), "ob_start")
}

// debugBufferReads lists the calls that end or read the output buffer.
var debugBufferReads = map[string]bool{
	"ob_get_clean": true, "ob_get_contents": true, "ob_get_flush": true,
	"ob_end_clean": true, "ob_end_flush": true,
}

// debugCaptured implements E2a: a statement holding the call sits between an
// `ob_start()` statement and a later one that reads or ends the buffer, in
// the same statement list: the output is captured, not printed.
func debugCaptured(f *syntax.File, call *syntax.FuncCall) bool {
	for n := call.Parent(); n != nil && !syntax.IsFuncLike(n); n = n.Parent() {
		st, isStmt := n.(syntax.Stmt)
		if !isStmt {
			continue
		}
		list, i, ok := astquery.StmtList(f, st)
		if !ok {
			continue
		}
		if debugBufferOpened(list[:i]) && slices.ContainsFunc(list[i+1:], debugReadsBuffer) {
			return true
		}
	}
	return false
}

// debugBufferOpened reports an `ob_start()` statement in stmts that no later
// one of them reads or ends.
func debugBufferOpened(stmts []syntax.Stmt) bool {
	for j := len(stmts) - 1; j >= 0; j-- {
		if debugReadsBuffer(stmts[j]) {
			return false
		}
		if es, ok := stmts[j].(*syntax.ExprStmt); ok {
			if c, ok := es.Expr.(*syntax.FuncCall); ok && strings.EqualFold(astquery.CallLastName(c), "ob_start") {
				return true
			}
		}
	}
	return false
}

// debugReadsBuffer reports a call in s, outside nested function-likes, that
// reads or ends the output buffer.
func debugReadsBuffer(s syntax.Stmt) bool {
	found := false
	syntax.Inspect(s, func(n syntax.Node) bool {
		if found || syntax.IsFuncLike(n) {
			return false
		}
		if c, ok := n.(*syntax.FuncCall); ok && debugBufferReads[strings.ToLower(astquery.CallLastName(c))] {
			found = true
		}
		return !found
	})
	return found
}

// debugInWrapper implements E3: the nearest enclosing function-like is a
// named function listed as a function entry, or a method whose
// `\Class::method` identity is listed as a method entry.
func debugInWrapper(ctx *analysis.Context, n syntax.Node, entries []string) bool {
	switch fn := syntax.EnclosingFuncLike(n).(type) {
	case *syntax.Function:
		return fn.Name != nil && debugHasEntry(entries, fn.Name.Value)
	case *syntax.Method:
		if debugHasEntry(entries, fn.Name.Value) { // a `dump()` method dumping
			return true
		}
		fqn := ctx.Types().ClassFQN(syntax.EnclosingClass(fn))
		return fqn != "" && debugHasEntry(entries, `\`+fqn+"::"+fn.Name.Value)
	}
	return false
}
