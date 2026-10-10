package semanticquery

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
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
	for p := c.Parent(); p != nil; p = p.Parent() {
		block, ok := p.(*syntax.Block)
		if !ok {
			continue
		}
		for _, s := range block.Stmts {
			branch, ok := s.(*syntax.If)
			if !ok || branch.Span().End > c.Span().Start || branch.Else != nil || len(branch.ElseIfs) > 0 || syntax.Terminates(branch.Body) {
				continue
			}
			unauthorized, ok := branch.Cond.(*syntax.Unary)
			if !ok || unauthorized.Op.Kind != syntax.TExclaim {
				continue
			}
			if _, ok := unauthorized.Expr.(*syntax.Variable); !ok {
				continue
			}
			found := false
			syntax.Inspect(branch.Body, func(child syntax.Node) bool {
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
			if found {
				return true
			}
		}
	}
	return false
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
	_, report = ctx.Flow().GlobalStateBefore(c, "committed-output")

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
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
	if known && status == 204 && len(echo.Exprs) > 0 {
		for _, expr := range echo.Exprs {
			if content, known := NativeString(ctx, expr); known && content != "" {
				ctx.ReportNode(echo, message)
				return
			}
		}
	}
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
	if old == value {
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
	site, s := NativeString(ctx, entries["s:samesite"])
	if !s || !strings.EqualFold(site, "none") {
		return
	}
	secure, exists := entries["s:secure"]
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
	state, known := ctx.Flow().GlobalStateBefore(c, "session")
	report = known && state.Operation == "session_start"

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckSessionLockHeldDuringBlockingCall checks the independently specified contract.
func CheckSessionLockHeldDuringBlockingCall(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false
	var fixes []diagnostic.Fix

	if name != "sleep" && name != "usleep" && name != "curl_exec" {
		return
	}
	state, known := ctx.Flow().GlobalStateBefore(c, "session")
	if !known || state.Operation != "session_start" {
		return
	}
	for _, prior := range nativePriorGlobal(ctx, c, "ini_set") {
		key, k := NativeString(ctx, CallArgument(prior.Args, 0, "option"))
		value, v := NativeString(ctx, CallArgument(prior.Args, 1, "value"))
		if k && v && key == "session.save_handler" && value == "files" {
			report = true
		}
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
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
		server, ok := fetch.Var.(*syntax.Variable)
		key, known := NativeString(ctx, fetch.Dim)
		if ok && known && server.Name == "_SERVER" && key == "HTTP_X_FORWARDED_FOR" {
			for p := binary.Parent(); p != nil; p = p.Parent() {
				if branch, ok := p.(*syntax.If); ok && branch.Cond != nil && syntax.UnwrapParens(branch.Cond) == binary && nativeProtectedBranch(ctx, branch.Body) {
					ctx.ReportNode(binary, message)
					return
				}
			}
		}
	}
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
		if NativeBuiltinName(ctx, call) == "ob_start" && CallArgument(call.Args, 0, "callback") != nil {
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
		echo, ok := node.(*syntax.Echo)
		if !ok {
			return true
		}
		outputs++
		if len(echo.Exprs) == 1 && NativeDominates(header, echo) && nativeSameExpression(ctx, body, echo.Exprs[0]) {
			matched = true
		}
		return true
	})
	return matched && outputs == 1
}

func nativeLengthEncodingSafe(ctx *analysis.Context, call *syntax.FuncCall) bool {
	encoding := CallArgument(call.Args, 1, "encoding")
	if encoding == nil {
		return true
	}
	name, known := NativeString(ctx, encoding)
	if !known {
		return false
	}
	switch strings.ToUpper(name) {
	case "UTF-8", "UTF8", "ASCII", "ISO-8859-1", "UTF-16", "UTF-16LE", "UTF-16BE", "UTF-32", "UTF-32LE", "UTF-32BE":
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
