package semanticquery

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func nativeStateCall(ctx *analysis.Context, at syntax.Node, domain string) *syntax.FuncCall {
	state, known := ctx.Flow().GlobalStateBefore(at, domain)
	if !known {
		return nil
	}
	for _, c := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		if call, ok := c.Node.(*syntax.FuncCall); ok && call.Span() == state.Span {
			return call
		}
	}
	return nil
}

func nativeSessionMutation(n syntax.Node) bool {
	var expr syntax.Expr
	switch n := n.(type) {
	case *syntax.Assign:
		expr = n.Var
	case *syntax.IncDec:
		expr = n.Var
	default:
		return false
	}
	for {
		fetch, ok := expr.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		expr = fetch.Var
	}
	v, ok := expr.(*syntax.Variable)
	return ok && v.Name == "_SESSION"
}

func nativeAnnotatedCall(ctx *analysis.Context, c *syntax.FuncCall, annotation string) bool {
	function := ctx.Types().ResolveFunction(c)
	if function == nil || function.Builtin {
		return false
	}
	decl := FunctionDecl(ctx.File, function)
	if decl == nil {
		return false
	}
	for _, field := range strings.Fields(index.DocComment(ctx.File, decl)) {
		if strings.Trim(field, "/*") == annotation {
			return true
		}
	}
	return false
}

func nativeProtectedRedirect(ctx *analysis.Context, c *syntax.FuncCall) bool {
	if !nativeAnnotatedCall(ctx, c, "@custos-protected") {
		return false
	}
	scope := syntax.EnclosingVariableScope(c)
	// authorized collects the variables proven truthy on every path from an
	// enclosing statement list down to the protected call.
	authorized := map[string]bool{}
	var child syntax.Node = c
	for p := c.Parent(); p != nil && p != scope; child, p = p, p.Parent() {
		nativeRedirectArmGuards(p, child, authorized)
		if block, ok := p.(*syntax.Block); ok && nativeRedirectFallsThrough(ctx, scope, block.Stmts, c, authorized) {
			return true
		}
	}
	// Top-level statements have no enclosing block node.
	return scope == nil && nativeRedirectFallsThrough(ctx, scope, ctx.File.Stmts, c, authorized)
}

// nativeRedirectFallsThrough scans the statements preceding the protected
// call, nearest first, for a redirecting unauthorized branch whose predicate
// is not proven truthy by a guard closer to the call.
func nativeRedirectFallsThrough(ctx *analysis.Context, scope syntax.Node, stmts []syntax.Stmt, c *syntax.FuncCall, authorized map[string]bool) bool {
	for i := len(stmts) - 1; i >= 0; i-- {
		branch, ok := stmts[i].(*syntax.If)
		if !ok || branch.Span().End > c.Span().Start || branch.Else != nil || len(branch.ElseIfs) > 0 {
			continue
		}
		if syntax.Terminates(branch.Body) {
			// Passing an early-exit guard proves its condition false.
			nativeRedirectTruthyVars(branch.Cond, false, authorized)
			continue
		}
		unauthorized, ok := branch.Cond.(*syntax.Unary)
		if !ok || unauthorized.Op.Kind != syntax.TExclaim {
			continue
		}
		predicate, ok := unauthorized.Expr.(*syntax.Variable)
		if !ok || predicate.Name == "" || authorized[predicate.Name] {
			continue
		}
		if nativeRedirectBranch(ctx, branch.Body) && !nativeRedirectPredicateRebound(ctx, scope, predicate.Name, branch.Span().Start, c.Span().Start) {
			return true
		}
	}
	return false
}

// nativeRedirectBranch reports whether body emits a literal Location header
// in its own scope; nested function bodies run later, if at all.
func nativeRedirectBranch(ctx *analysis.Context, body syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(child syntax.Node) bool {
		if found || syntax.IsVariableScope(child) {
			return false
		}
		header, ok := child.(*syntax.FuncCall)
		if !ok || NativeBuiltinName(ctx, header) != "header" {
			return true
		}
		value, known := NativeString(ctx, CallArgument(header.Args, 0, "header"))
		if known && strings.HasPrefix(strings.ToLower(value), "location:") {
			found = true
		}
		return true
	})
	return found
}

// nativeRedirectPredicateRebound reports whether the authorization variable
// is written between the redirect branch and the protected call, which makes
// the relationship between the two unknown.
func nativeRedirectPredicateRebound(ctx *analysis.Context, scope syntax.Node, name string, from, to uint32) bool {
	for _, access := range flowquery.VarAccesses(ctx.File, scope, name) {
		if at := access.Var.Span().Start; access.Write && at >= from && at < to {
			return true
		}
	}
	return false
}

// nativeRedirectArmGuards records the variables that must be truthy for
// control to enter child, an arm of the conditional construct parent.
func nativeRedirectArmGuards(parent, child syntax.Node, out map[string]bool) {
	switch p := parent.(type) {
	case *syntax.If:
		switch arm := child.(type) {
		case *syntax.ElseIf:
			nativeRedirectTruthyVars(p.Cond, false, out)
			for _, earlier := range p.ElseIfs {
				if earlier == arm {
					break
				}
				nativeRedirectTruthyVars(earlier.Cond, false, out)
			}
		case *syntax.Else:
			nativeRedirectTruthyVars(p.Cond, false, out)
			for _, earlier := range p.ElseIfs {
				nativeRedirectTruthyVars(earlier.Cond, false, out)
			}
		default:
			if child == syntax.Node(p.Body) {
				nativeRedirectTruthyVars(p.Cond, true, out)
			}
		}
	case *syntax.ElseIf:
		if child == syntax.Node(p.Body) {
			nativeRedirectTruthyVars(p.Cond, true, out)
		}
	case *syntax.While:
		if child == syntax.Node(p.Body) {
			nativeRedirectTruthyVars(p.Cond, true, out)
		}
	case *syntax.Ternary:
		if p.Then != nil && child == syntax.Node(p.Then) {
			nativeRedirectTruthyVars(p.Cond, true, out)
		} else if child == syntax.Node(p.Else) {
			nativeRedirectTruthyVars(p.Cond, false, out)
		}
	case *syntax.Binary:
		if child != syntax.Node(p.Right) {
			return
		}
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			nativeRedirectTruthyVars(p.Left, true, out)
		case syntax.TBooleanOr, syntax.TOr:
			nativeRedirectTruthyVars(p.Left, false, out)
		}
	}
}

// nativeRedirectTruthyVars adds the plain variables that are certainly truthy
// when cond evaluates to outcome.
func nativeRedirectTruthyVars(cond syntax.Expr, outcome bool, out map[string]bool) {
	switch e := cond.(type) {
	case *syntax.Paren:
		nativeRedirectTruthyVars(e.Expr, outcome, out)
	case *syntax.Variable:
		if outcome && e.Name != "" {
			out[e.Name] = true
		}
	case *syntax.Unary:
		if e.Op.Kind == syntax.TExclaim {
			nativeRedirectTruthyVars(e.Expr, !outcome, out)
		}
	case *syntax.Binary:
		// A true conjunction and a false disjunction fix both operands.
		conjunction := e.Op.Kind == syntax.TBooleanAnd || e.Op.Kind == syntax.TAnd
		disjunction := e.Op.Kind == syntax.TBooleanOr || e.Op.Kind == syntax.TOr
		if (conjunction && outcome) || (disjunction && !outcome) {
			nativeRedirectTruthyVars(e.Left, outcome, out)
			nativeRedirectTruthyVars(e.Right, outcome, out)
		}
	}
}

// CheckContentLengthUsesCharacterCount checks the independently specified contract.
func CheckContentLengthUsesCharacterCount(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false
	var fixes []diagnostic.Fix

	if name != "header" {
		return
	}
	value := CallArgument(c.Args, 0, "header")
	binary, ok := value.(*syntax.Binary)
	if !ok || binary.Op.Kind != syntax.TDot {
		return
	}
	prefix, known := NativeString(ctx, binary.Left)
	if !known || !strings.EqualFold(strings.TrimSpace(prefix), "Content-Length:") {
		return
	}
	length, ok := binary.Right.(*syntax.FuncCall)
	if !ok || (NativeBuiltinName(ctx, length) != "mb_strlen" && NativeBuiltinName(ctx, length) != "grapheme_strlen") {
		return
	}
	body := CallArgument(length.Args, 0, "string")
	if NativeBuiltinName(ctx, length) == "mb_strlen" && nativeLengthCountsBytes(ctx, length) {
		return
	}
	if !nativeEmittedBody(ctx, c, body) {
		return
	}
	report = true
	if _, bound := ctx.Flow().BoundArguments(length); bound && body != nil && nativeLengthEncodingSafe(ctx, length) && nativeLengthFixSafe(length) && nativeCleanEdit(ctx, length.Span()) {
		fixes = append(fixes, nativeFix(length.Span(), nativeGlobalCallName(ctx, "strlen", c)+"("+ctx.Text(body)+")", "Count response bytes"))
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckHeadersAfterCommittedOutput checks the independently specified contract.
func CheckHeadersAfterCommittedOutput(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false
	var fixes []diagnostic.Fix

	if name != "header" && name != "setcookie" && name != "http_response_code" {
		return
	}
	if name == "http_response_code" && !headersStatusWrite(ctx, c) {
		return
	}
	commit, report := ctx.Flow().GlobalStateBefore(c, "committed-output")

	if report && !headersUnsentGuard(ctx, c, commit.Span) {
		ctx.ReportNode(c, message, fixes...)
	}
}

// headersStatusWrite reports whether http_response_code is called as a setter.
// Without a status argument, or with one known to be falsy, it only reads.
func headersStatusWrite(ctx *analysis.Context, c *syntax.FuncCall) bool {
	code := CallArgument(c.Args, 0, "response_code")
	if code == nil {
		return false
	}
	set, known := NativeTruth(ctx, code)
	return set || !known
}

// headersUnsentGuard proves that a headers_sent() test evaluated before the
// call excludes sent headers on its path. A commit recorded between the test
// and the call invalidates the test.
func headersUnsentGuard(ctx *analysis.Context, c *syntax.FuncCall, commit syntax.Span) bool {
	proves := func(cond syntax.Expr, truth bool) bool {
		stale := commit.Start >= cond.Span().End && commit.End <= c.Span().Start
		return !stale && headersUnsentWhen(ctx, cond, truth)
	}
	if nativeGuard(ctx, c, proves) {
		return true
	}
	for child, p := syntax.Node(c), c.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch x := p.(type) {
		case *syntax.Binary:
			and := x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd
			or := x.Op.Kind == syntax.TBooleanOr || x.Op.Kind == syntax.TOr
			if x.Right == child && (and || or) && proves(x.Left, and) {
				return true
			}
		case *syntax.Ternary:
			if (x.Then == child && proves(x.Cond, true)) || (x.Else == child && proves(x.Cond, false)) {
				return true
			}
		case *syntax.ElseIf:
			if syntax.Node(x.Body) == child && proves(x.Cond, true) {
				return true
			}
		}
	}
	return false
}

// headersUnsentWhen reports whether cond evaluating to truth implies that
// headers_sent() returned false.
func headersUnsentWhen(ctx *analysis.Context, cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.FuncCall:
		return !truth && NativeBuiltin(ctx, x, "headers_sent")
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && headersUnsentWhen(ctx, x.Expr, !truth)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
			// A true conjunction and a false disjunction fix both operands.
			if truth == (x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd) {
				return headersUnsentWhen(ctx, x.Left, truth) || headersUnsentWhen(ctx, x.Right, truth)
			}
		case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
			equal := x.Op.Kind == syntax.TIsIdentical || x.Op.Kind == syntax.TIsEqual
			for _, pair := range [][2]syntax.Expr{{x.Left, x.Right}, {x.Right, x.Left}} {
				if expected, ok := astquery.BoolConst(syntax.UnwrapParens(pair[1])); ok {
					return headersUnsentWhen(ctx, pair[0], truth == (expected == equal))
				}
			}
		}
	}
	return false
}

// CheckNoContentResponseWithBody checks the independently specified contract.
func CheckNoContentResponseWithBody(ctx *analysis.Context, n syntax.Node, message string) {
	echo, ok := n.(*syntax.Echo)
	if !ok {
		return
	}
	call := nativeStateCall(ctx, echo, "response")
	if call == nil {
		return
	}
	status, known := NativeInt(ctx, CallArgument(call.Args, 0, "response_code"))
	if !known || status != 204 || noContentStatusReplaced(ctx, call, echo) {
		return
	}
	// Output written into an open buffer may still be discarded.
	if _, buffering := ctx.Flow().GlobalStateBefore(echo, "output-buffer"); buffering {
		return
	}
	for _, expr := range echo.Exprs {
		if content, known := NativeString(ctx, expr); known && content != "" {
			ctx.ReportNode(echo, message)
			return
		}
	}
}

// noContentStatusReplaced reports whether a header() call that may replace the
// status set by the given http_response_code call can run before the echo: one
// placed between them, or later in a loop that repeats the echo.
func noContentStatusReplaced(ctx *analysis.Context, set *syntax.FuncCall, echo *syntax.Echo) bool {
	scope := syntax.EnclosingVariableScope(echo)
	limit := echo.Span().Start
	for p := echo.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			limit = p.Span().End
		}
	}
	for _, c := range ctx.Flow().Calls(scope) {
		header, ok := c.Node.(*syntax.FuncCall)
		if !ok || header.Span().Start < set.Span().End || header.Span().Start >= limit {
			continue
		}
		if NativeBuiltinName(ctx, header) == "header" && !noContentHeaderKeepsStatus(ctx, header) {
			return true
		}
	}
	return false
}

// noContentHeaderKeepsStatus proves that a header() call leaves a 204 status
// in place: no other explicit status argument, and a field line that is
// neither a status line nor a Location or Status field.
func noContentHeaderKeepsStatus(ctx *analysis.Context, c *syntax.FuncCall) bool {
	explicit, state := CallArgumentBinding(c.Args, 2, "response_code")
	if state == ArgumentUnknown {
		return false
	}
	if state == ArgumentBound {
		code, known := NativeInt(ctx, explicit)
		if !known {
			return false
		}
		if code != 0 {
			return code == 204
		}
	}
	line := syntax.UnwrapParens(CallArgument(c.Args, 0, "header"))
	text, complete := NativeString(ctx, line)
	if !complete {
		// A concatenation still exposes the field name in its leading literal.
		for {
			concat, ok := line.(*syntax.Binary)
			if !ok || concat.Op.Kind != syntax.TDot {
				break
			}
			line = syntax.UnwrapParens(concat.Left)
		}
		text, _ = NativeString(ctx, line)
	}
	text = strings.ToLower(text)
	if strings.HasPrefix(text, "http/") {
		fields := strings.Fields(text)
		return complete && len(fields) > 1 && fields[1] == "204"
	}
	name, _, named := strings.Cut(text, ":")
	if !named {
		return complete
	}
	name = strings.TrimSpace(name)
	return name != "location" && name != "status"
}

// CheckRedirectContinuesProtectedExecution checks the independently specified contract.
func CheckRedirectContinuesProtectedExecution(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}

	report := false
	var fixes []diagnostic.Fix

	report = nativeProtectedRedirect(ctx, c)

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckRepeatedCookieHeaderReplacement checks the independently specified contract.
func CheckRepeatedCookieHeaderReplacement(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false
	var fixes []diagnostic.Fix

	if name != "header" {
		return
	}
	value, known := NativeString(ctx, CallArgument(c.Args, 0, "header"))
	if !known || !strings.HasPrefix(strings.ToLower(value), "set-cookie:") {
		return
	}
	replace := CallArgument(c.Args, 1, "replace")
	if replace != nil {
		truth, known := NativeTruth(ctx, replace)
		if !known || !truth {
			return
		}
	}
	previous := nativeStateCall(ctx, c, "cookie-header")
	if previous == nil {
		return
	}
	old, _ := NativeString(ctx, CallArgument(previous.Args, 0, "header"))
	// Only two proven, different cookie names show a lost cookie: replacing
	// a cookie by a new value of itself is the intended use of the default.
	oldName, newName := cookieHeaderName(old), cookieHeaderName(value)
	if oldName == "" || newName == "" || oldName == newName || cookieHeaderRemovedBetween(ctx, previous, c) {
		return
	}
	report = true
	if _, bound := ctx.Flow().BoundArguments(c); bound && replace == nil && nativeCleanEdit(ctx, c.Args.Span()) {
		insert := syntax.Span{Start: c.Args.Span().End - 1, End: c.Args.Span().End - 1}
		addition := ", false"
		for _, node := range c.Args.Args {
			if arg, ok := node.(*syntax.Arg); ok && arg.Name != nil {
				addition = ", replace: false"
			}
		}
		last := c.Args.Args[len(c.Args.Args)-1]
		tail := syntax.Span{Start: last.Span().End, End: c.Args.Span().End - 1}
		if strings.TrimSpace(string(ctx.File.Src[tail.Start:tail.End])) == "," {
			addition = strings.TrimPrefix(addition, ",")
		}
		fixes = append(fixes, nativeFix(insert, addition, "Append this cookie header"))
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// cookieHeaderName returns the cookie name of a Set-Cookie field line, or ""
// when the line does not carry a name=value pair.
func cookieHeaderName(line string) string {
	_, pair, _ := strings.Cut(line, ":")
	pair, _, _ = strings.Cut(pair, ";")
	name, _, found := strings.Cut(pair, "=")
	if !found {
		return ""
	}
	return strings.TrimSpace(name)
}

// cookieHeaderRemovedBetween reports whether a header_remove() call that may
// drop the Set-Cookie headers sits between the two header() calls, on any
// branch.
func cookieHeaderRemovedBetween(ctx *analysis.Context, previous, c *syntax.FuncCall) bool {
	for _, reached := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		remove, ok := reached.Node.(*syntax.FuncCall)
		if !ok || NativeBuiltinName(ctx, remove) != "header_remove" {
			continue
		}
		if remove.Span().Start < previous.Span().End || remove.Span().Start >= c.Span().Start {
			continue
		}
		// Without an argument every header goes; an unknown name may be this one.
		target := CallArgument(remove.Args, 0, "name")
		field, known := NativeString(ctx, target)
		if target == nil || !known || strings.EqualFold(strings.TrimSpace(field), "set-cookie") {
			return true
		}
	}
	return false
}

// CheckSameSiteNoneWithoutSecure checks the independently specified contract.
func CheckSameSiteNoneWithoutSecure(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false
	var fixes []diagnostic.Fix

	if name != "setcookie" || ctx.PHP < phpversion.PHP73 {
		return
	}
	entries, known := NativeArrayEntries(ctx, CallArgument(c.Args, 2, "expires_or_options"))
	if !known {
		return
	}
	// PHP matches cookie option names without regard to case. Two spellings of
	// one option leave the effective value to entry order, so nothing is proven.
	options := make(map[string]syntax.Expr, len(entries))
	for key, value := range entries {
		key = strings.ToLower(key)
		if _, dup := options[key]; dup {
			return
		}
		options[key] = value
	}
	site, s := NativeString(ctx, options["s:samesite"])
	if !s || !strings.EqualFold(site, "none") {
		return
	}
	secure, exists := options["s:secure"]
	if !exists {
		report = true
	} else {
		truth, known := NativeTruth(ctx, secure)
		report = known && !truth
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckSessionCookieOptionsSetAfterStart checks the independently specified contract.
func CheckSessionCookieOptionsSetAfterStart(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false
	var fixes []diagnostic.Fix

	if name != "session_set_cookie_params" {
		return
	}
	start := nativeStateCall(ctx, c, "session")
	if start == nil || !NativeBuiltin(ctx, start, "session_start") {
		return
	}
	report = sessionStartKeepsActive(ctx, start) && !nativeGuard(ctx, c, func(cond syntax.Expr, truth bool) bool {
		return sessionInactiveWhen(ctx, cond, truth, start)
	})

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// sessionStartKeepsActive proves the start options leave the session open.
// read_and_close closes it again at once; unknown options prove nothing.
func sessionStartKeepsActive(ctx *analysis.Context, start *syntax.FuncCall) bool {
	options := CallArgument(start.Args, 0, "options")
	if options == nil {
		return start.Args == nil || len(start.Args.Args) == 0
	}
	entries, known := NativeArrayEntries(ctx, options)
	if !known {
		return false
	}
	flag, set := entries["s:read_and_close"]
	closing, known := NativeTruth(ctx, flag)
	return !set || (known && !closing)
}

// sessionInactiveWhen proves that cond having the given truth leaves no active
// session: the start call itself failed, or a session_status() test evaluated
// after the start excludes the active state.
func sessionInactiveWhen(ctx *analysis.Context, cond syntax.Expr, truth bool, start *syntax.FuncCall) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.FuncCall:
		return !truth && x == start
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && sessionInactiveWhen(ctx, x.Expr, !truth, start)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
			// A true conjunction and a false disjunction fix both operands.
			if truth == (x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd) {
				return sessionInactiveWhen(ctx, x.Left, truth, start) || sessionInactiveWhen(ctx, x.Right, truth, start)
			}
		case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
			equal := x.Op.Kind == syntax.TIsIdentical || x.Op.Kind == syntax.TIsEqual
			for _, pair := range [][2]syntax.Expr{{x.Left, x.Right}, {x.Right, x.Left}} {
				if expected, ok := astquery.BoolConst(syntax.UnwrapParens(pair[1])); ok {
					return sessionInactiveWhen(ctx, pair[0], truth == (expected == equal), start)
				}
				status, ok := syntax.UnwrapParens(pair[0]).(*syntax.FuncCall)
				if !ok || !NativeBuiltin(ctx, status, "session_status") || start.Span().End > x.Span().Start {
					continue
				}
				if value, known := NativeInt(ctx, pair[1]); known {
					return (truth == equal) != (value == 2) // PHP_SESSION_ACTIVE
				}
			}
		}
	}
	return false
}

// CheckSessionLockHeldDuringBlockingCall checks the independently specified contract.
func CheckSessionLockHeldDuringBlockingCall(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	if name != "sleep" && name != "usleep" && name != "curl_exec" {
		return
	}
	start := nativeStateCall(ctx, c, "session")
	if start == nil || !NativeBuiltin(ctx, start, "session_start") {
		return
	}
	if sessionStartKeepsLock(ctx, start) && sessionFilesHandler(ctx, start) {
		ctx.ReportNode(c, message)
	}
}

// sessionStartKeepsLock reports whether the session_start call provably leaves
// the session open: its options, when given, must be a literal array with
// constant string keys whose read_and_close entry is absent or known falsy.
func sessionStartKeepsLock(ctx *analysis.Context, start *syntax.FuncCall) bool {
	if start.Args == nil || len(start.Args.Args) == 0 {
		return true
	}
	options := NativeArray(ctx, CallArgument(start.Args, 0, "options"))
	if options == nil {
		return false
	}
	for _, item := range options.Items {
		key, known := NativeString(ctx, item.Key)
		if !known {
			return false
		}
		if closes, known := NativeTruth(ctx, item.Value); key == "read_and_close" && (closes || !known) {
			return false
		}
	}
	return true
}

// sessionFilesHandler proves that the session starts with the builtin files
// handler: the last handler selection written before session_start must be an
// unconditional ini_set of session.save_handler to 'files'. A later ini write
// with another or unknown value, an ini write to an unknown option, ini_restore,
// session_set_save_handler and a session_module_name setter all discard the proof.
func sessionFilesHandler(ctx *analysis.Context, start *syntax.FuncCall) bool {
	var last *syntax.FuncCall
	files := false
	for _, c := range ctx.Flow().Calls(syntax.EnclosingVariableScope(start)) {
		call, ok := c.Node.(*syntax.FuncCall)
		if !ok || call.Span().Start >= start.Span().Start || (last != nil && call.Span().Start < last.Span().Start) {
			continue
		}
		switch NativeBuiltinName(ctx, call) {
		case "ini_set", "ini_alter", "ini_restore":
			key, known := NativeString(ctx, CallArgument(call.Args, 0, "option"))
			if known && key != "session.save_handler" {
				continue
			}
			value, set := NativeString(ctx, CallArgument(call.Args, 1, "value"))
			last, files = call, known && set && value == "files"
		case "session_set_save_handler":
			last, files = call, false
		case "session_module_name":
			if call.Args != nil && len(call.Args.Args) > 0 {
				last, files = call, false
			}
		}
	}
	return files && NativeDominates(last, start)
}

// CheckSessionMutationAfterClose checks the independently specified contract.
func CheckSessionMutationAfterClose(ctx *analysis.Context, n syntax.Node, message string) {
	if !nativeSessionMutation(n) {
		return
	}
	state, known := ctx.Flow().GlobalStateBefore(n, "session")
	if known && (state.Operation == "session_write_close" || state.Operation == "session_commit" || state.Operation == "session_abort") {
		ctx.ReportNode(n, message)
	}
}

// CheckUntrustedForwardedClientAddress checks the independently specified contract.
func CheckUntrustedForwardedClientAddress(ctx *analysis.Context, n syntax.Node, message string) {
	binary, ok := n.(*syntax.Binary)
	if !ok {
		return
	}
	if binary.Op.Kind != syntax.TIsIdentical && binary.Op.Kind != syntax.TIsEqual {
		return
	}
	for _, operand := range []syntax.Expr{binary.Left, binary.Right} {
		fetch, ok := operand.(*syntax.ArrayDimFetch)
		if !ok {
			continue
		}
		if forwardedServerKey(ctx, fetch) == "HTTP_X_FORWARDED_FOR" {
			for p := binary.Parent(); p != nil; p = p.Parent() {
				if branch, ok := p.(*syntax.If); ok && branch.Cond != nil && syntax.UnwrapParens(branch.Cond) == binary && nativeProtectedBranch(ctx, branch.Body) {
					if !forwardedProxyGuarded(ctx, branch) {
						ctx.ReportNode(binary, message)
					}
					return
				}
			}
		}
	}
}

// forwardedServerKey returns the constant key of a $_SERVER element read, or
// "" for any other expression.
func forwardedServerKey(ctx *analysis.Context, fetch *syntax.ArrayDimFetch) string {
	server, ok := fetch.Var.(*syntax.Variable)
	key, known := NativeString(ctx, fetch.Dim)
	if !ok || !known || server.Name != "_SERVER" {
		return ""
	}
	return key
}

// forwardedProxyGuarded reports whether a condition that may validate the
// proxy boundary already decides whether branch runs: the condition of an
// enclosing if chain, or of an earlier if in an enclosing statement list one
// of whose arms leaves that list.
func forwardedProxyGuarded(ctx *analysis.Context, branch *syntax.If) bool {
	for n := syntax.Node(branch); n != nil && !syntax.IsVariableScope(n); n = n.Parent() {
		if outer, ok := n.(*syntax.If); ok && outer != branch && forwardedProxyChain(ctx, outer) {
			return true
		}
		if st, ok := n.(syntax.Stmt); ok {
			for prev, exists := astquery.PrevStmt(ctx.File, st); exists; prev, exists = astquery.PrevStmt(ctx.File, prev) {
				if guard, ok := prev.(*syntax.If); ok && forwardedChainLeaves(guard) && forwardedProxyChain(ctx, guard) {
					return true
				}
			}
		}
	}
	return false
}

// forwardedChainLeaves reports whether some arm of the if chain never falls
// through to the statement after it.
func forwardedChainLeaves(chain *syntax.If) bool {
	if syntax.Terminates(chain.Body) || (chain.Else != nil && syntax.Terminates(chain.Else.Body)) {
		return true
	}
	for _, arm := range chain.ElseIfs {
		if syntax.Terminates(arm.Body) {
			return true
		}
	}
	return false
}

// forwardedProxyChain reports whether any condition of the if chain may
// validate the proxy boundary.
func forwardedProxyChain(ctx *analysis.Context, chain *syntax.If) bool {
	if forwardedProxyTest(ctx, chain.Cond) {
		return true
	}
	for _, arm := range chain.ElseIfs {
		if forwardedProxyTest(ctx, arm.Cond) {
			return true
		}
	}
	return false
}

// forwardedProxyTest reports whether cond reads the connecting peer address
// or calls anything but a builtin function: either may be a trusted-proxy
// check this rule does not model.
func forwardedProxyTest(ctx *analysis.Context, cond syntax.Expr) bool {
	found := false
	syntax.Inspect(cond, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.ArrayDimFetch:
			found = found || forwardedServerKey(ctx, x) == "REMOTE_ADDR"
		case *syntax.FuncCall:
			function := ctx.Types().ResolveFunction(x)
			found = found || function == nil || !function.Builtin
		case *syntax.MethodCall, *syntax.StaticCall:
			found = true
		}
		return !found
	})
	return found
}

// nativeLengthFixSafe permits removing only arguments with no observable evaluation.
func nativeLengthFixSafe(call *syntax.FuncCall) bool {
	for _, node := range call.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || arg.ByRef || flowquery.MayHaveSideEffects(arg.Value) {
			return false
		}
	}
	return true
}

func nativeEmittedBody(ctx *analysis.Context, header *syntax.FuncCall, body syntax.Expr) bool {
	if body == nil {
		return false
	}
	if text, known := NativeString(ctx, body); known {
		ascii := true
		for _, c := range []byte(text) {
			if c >= 128 {
				ascii = false
			}
		}
		if ascii {
			return false
		}
	}
	if _, known := NativeInt(ctx, body); known {
		return false
	}
	if literal, ok := NativeValue(ctx, body).(*syntax.Literal); ok && literal.LitKind == syntax.LitFloat {
		return false
	}
	if constant, ok := NativeValue(ctx, body).(*syntax.ConstFetch); ok {
		name := strings.ToLower(GlobalConstName(ctx, constant))
		if name == "true" || name == "false" || name == "null" {
			return false
		}
	}
	for _, call := range nativePriorGlobal(ctx, header, "ob_start", "ini_set") {
		if NativeBuiltinName(ctx, call) == "ob_start" && !CallArgumentOmitted(call.Args, 0, "callback") {
			return false
		}
		if key, known := NativeString(ctx, CallArgument(call.Args, 0, "option")); known && key == "zlib.output_compression" {
			return false
		}
	}
	owner := syntax.EnclosingVariableScope(header)
	for _, call := range ctx.Flow().Calls(owner) {
		if call.Node.Span().Start > header.Span().End {
			return false
		}
	}
	outputs := 0
	matched := false
	syntax.InspectFile(ctx.File, func(node syntax.Node) bool {
		if syntax.EnclosingVariableScope(node) != owner || node.Span().Start < header.Span().End {
			return true
		}
		if _, ok := node.(*syntax.Print); ok {
			outputs++
			return true
		}
		// Template text between PHP tags is one more response chunk.
		if html, ok := node.(*syntax.InlineHTML); ok && html.Raw != "" {
			outputs++
			return true
		}
		echo, ok := node.(*syntax.Echo)
		if !ok {
			return true
		}
		outputs++
		if len(echo.Exprs) == 1 && NativeDominates(header, echo) && !nativeRepeatedOutput(echo, owner) && nativeSameExpression(ctx, body, echo.Exprs[0]) {
			matched = true
		}
		return true
	})
	return matched && outputs == 1
}

// nativeRepeatedOutput reports whether a loop inside owner may run the output
// statement more than once.
func nativeRepeatedOutput(output, owner syntax.Node) bool {
	for p := output.Parent(); p != nil && p != owner; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			return true
		}
	}
	return false
}

// nativeLengthCountsBytes reports whether the resolved mb_strlen encoding maps
// every byte to one character, so the call already measures bytes.
func nativeLengthCountsBytes(ctx *analysis.Context, call *syntax.FuncCall) bool {
	name, known := NativeString(ctx, CallArgument(call.Args, 1, "encoding"))
	if !known {
		return false
	}
	switch strings.ToUpper(name) {
	case "8BIT", "BINARY", "ASCII", "US-ASCII", "LATIN1",
		"ISO-8859-1", "ISO-8859-2", "ISO-8859-3", "ISO-8859-4", "ISO-8859-5", "ISO-8859-6", "ISO-8859-7", "ISO-8859-8",
		"ISO-8859-9", "ISO-8859-10", "ISO-8859-13", "ISO-8859-14", "ISO-8859-15", "ISO-8859-16",
		"WINDOWS-1251", "WINDOWS-1252", "WINDOWS-1254", "CP1251", "CP1252", "CP1254":
		return true
	}
	return false
}

func nativeLengthEncodingSafe(ctx *analysis.Context, call *syntax.FuncCall) bool {
	encoding, state := CallArgumentBinding(call.Args, 1, "encoding")
	if state == ArgumentMissing {
		return true
	}
	name, known := NativeString(ctx, encoding)
	if !known {
		return false
	}
	switch strings.ToUpper(name) {
	case "UTF-8", "UTF8", "UTF-16", "UTF-16LE", "UTF-16BE", "UTF-32", "UTF-32LE", "UTF-32BE":
		return true
	}
	return false
}

func nativeProtectedBranch(ctx *analysis.Context, body syntax.Stmt) bool {
	found := false
	syntax.Inspect(body, func(node syntax.Node) bool {
		if node != body && syntax.IsVariableScope(node) {
			return false
		}
		if call, ok := node.(*syntax.FuncCall); ok && nativeAnnotatedCall(ctx, call, "@custos-protected") {
			found = true
		}
		return true
	})
	return found
}
