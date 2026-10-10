package semanticquery

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

func nativePasswordParameter(ctx *analysis.Context, at syntax.Node, input syntax.Expr) bool {
	scope := syntax.EnclosingVariableScope(at)
	if scope == nil {
		return false
	}
	comment := index.DocComment(ctx.File, scope)
	value := ctx.Flow().Value(input)
	if !value.Complete {
		return false
	}
	for _, source := range value.Sources {
		if source.Kind != "parameter" {
			continue
		}
		params := syntax.VariableScopeParams(scope)
		name := params[source.Parameter].Var.Name
		for _, line := range strings.Split(comment, "\n") {
			fields := strings.Fields(strings.TrimSpace(strings.Trim(line, "/*")))
			for i, field := range fields {
				if field == "@param" && i+2 < len(fields) && fields[i+1] == "password-string" && fields[i+2] == "$"+name {
					return true
				}
			}
		}
	}
	return false
}

func nativeUploadMimeBranch(ctx *analysis.Context, call *syntax.FuncCall) bool {
	upload, known := nativeUploadEntry(ctx, CallArgument(call.Args, 0, "from"), "tmp_name")
	if !known {
		return false
	}
	for child, parent := syntax.Node(call), call.Parent(); parent != nil; child, parent = parent, parent.Parent() {
		if syntax.IsVariableScope(parent) {
			break
		}
		branch, ok := parent.(*syntax.If)
		if !ok || child != branch.Body {
			continue
		}
		comparison, ok := syntax.UnwrapParens(branch.Cond).(*syntax.Binary)
		if !ok || (comparison.Op.Kind != syntax.TIsIdentical && comparison.Op.Kind != syntax.TIsEqual) {
			continue
		}
		for i, operand := range []syntax.Expr{comparison.Left, comparison.Right} {
			entry, known := nativeUploadEntry(ctx, operand, "type")
			if !known || entry != upload {
				continue
			}
			other := comparison.Right
			if i == 1 {
				other = comparison.Left
			}
			if mime, known := NativeString(ctx, other); known && strings.Contains(mime, "/") {
				return true
			}
		}
	}
	return false
}

func nativeUploadEntry(ctx *analysis.Context, expr syntax.Expr, field string) (string, bool) {
	fetch, ok := expr.(*syntax.ArrayDimFetch)
	if !ok {
		return "", false
	}
	key, known := NativeString(ctx, fetch.Dim)
	if !known || key != field {
		return "", false
	}
	entry, ok := fetch.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return "", false
	}
	root, ok := entry.Var.(*syntax.Variable)
	if !ok || root.Name != "_FILES" {
		return "", false
	}
	return NativeString(ctx, entry.Dim)
}

func nativeHTMLText(ctx *analysis.Context, expr syntax.Expr) bool {
	// Only prove a single interpolation preceded entirely by known bytes.
	// Earlier dynamic output could open an attribute or raw-text element.
	var prefix strings.Builder
	var input syntax.Expr
	var collect func(syntax.Expr) bool
	collect = func(part syntax.Expr) bool {
		if binary, ok := part.(*syntax.Binary); ok && binary.Op.Kind == syntax.TDot {
			return collect(binary.Left) && collect(binary.Right)
		}
		if text, known := NativeString(ctx, part); known {
			if input == nil {
				prefix.WriteString(text)
			}
			return true
		}
		if input != nil {
			return false
		}
		input = part
		return true
	}
	return collect(expr) && input != nil && nativeHTMLTextPrefix(prefix.String()) && ctx.Flow().Tainted(input, flow.HTML)
}

func nativeHTMLTextPrefix(prefix string) bool {
	lower := strings.ToLower(prefix)
	if strings.Contains(lower, "<script") || strings.Contains(lower, "<style") {
		return false
	}
	tag, text := false, false
	var quote byte
	for i := 0; i < len(prefix); i++ {
		c := prefix[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if tag {
			switch c {
			case '\'', '"':
				quote = c
			case '>':
				tag, text = false, true
			case '<':
				return false
			}
			continue
		}
		if c == '<' {
			start := i + 1
			if start < len(prefix) && prefix[start] == '/' {
				start++
			}
			// Comments, declarations and incomplete/unknown tags cannot
			// establish text context. Only ordinary ASCII tag names do.
			if start >= len(prefix) || lower[start] < 'a' || lower[start] > 'z' {
				return false
			}
			tag = true
		}
	}
	return text && !tag
}

func nativePlainResponse(ctx *analysis.Context, at syntax.Node) bool {
	for _, header := range nativePriorGlobal(ctx, at, "header") {
		value, known := NativeString(ctx, CallArgument(header.Args, 0, "header"))
		if known && strings.HasPrefix(strings.ToLower(value), "content-type:") && !strings.Contains(strings.ToLower(value), "text/html") {
			return true
		}
	}
	return false
}

// CheckFastDigestUsedForPasswordStorage checks the independently specified contract.
func CheckFastDigestUsedForPasswordStorage(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}

	report := false

	if !nativeAnnotatedCall(ctx, c, "@custos-credential-store") {
		return
	}
	digest, ok := NativeValue(ctx, CallArgument(c.Args, 0, "")).(*syntax.FuncCall)
	if !ok {
		return
	}
	hash := NativeBuiltinName(ctx, digest)
	if hash != "md5" && hash != "sha1" && hash != "hash" {
		return
	}
	position := 0
	if hash == "hash" {
		position = 1
	}
	report = nativePasswordParameter(ctx, c, CallArgument(digest.Args, position, "data"))

	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckPasswordComparedWithFreshHash checks the independently specified contract.
func CheckPasswordComparedWithFreshHash(ctx *analysis.Context, n syntax.Node, message string) {
	binary, ok := n.(*syntax.Binary)
	if !ok {
		return
	}
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	if binary.Op.Kind != syntax.TIsIdentical && binary.Op.Kind != syntax.TIsEqual && binary.Op.Kind != syntax.TIsNotIdentical && binary.Op.Kind != syntax.TIsNotEqual {
		return
	}
	for i, expr := range []syntax.Expr{binary.Left, binary.Right} {
		call, ok := NativeValue(ctx, expr).(*syntax.FuncCall)
		if !ok || NativeBuiltinName(ctx, call) != "password_hash" {
			continue
		}
		password := CallArgument(call.Args, 0, "password")
		stored := binary.Right
		if i == 1 {
			stored = binary.Left
		}
		if password == nil {
			return
		}
		var fixes []diagnostic.Fix
		fixable := syntax.UnwrapParens(expr) == call && !flowquery.MayHaveSideEffects(password) && !flowquery.MayHaveSideEffects(stored)
		_, bound := ctx.Flow().BoundArguments(call)
		algorithm, constant := NativeValue(ctx, CallArgument(call.Args, 1, "algo")).(*syntax.ConstFetch)
		fixable = fixable && bound && constant && (GlobalConstName(ctx, algorithm) == "PASSWORD_DEFAULT" || GlobalConstName(ctx, algorithm) == "PASSWORD_BCRYPT")
		if options := CallArgument(call.Args, 2, "options"); options != nil {
			entries, known := NativeArrayEntries(ctx, options)
			fixable = fixable && known && len(entries) == 0
		}
		if call.Args != nil {
			for _, node := range call.Args.Args {
				arg, ok := node.(*syntax.Arg)
				if !ok || arg.Unpack || arg.ByRef || flowquery.MayHaveSideEffects(arg.Value) {
					fixable = false
				}
			}
		}
		if fixable && binary.Op.Kind == syntax.TIsIdentical && nativeCleanEdit(ctx, binary.Span()) {
			fixes = append(fixes, nativeFix(binary.Span(), nativeGlobalCallName(ctx, "password_verify", binary)+"("+ctx.Text(password)+", "+ctx.Text(stored)+")", "Verify the stored password hash"))
		}
		ctx.ReportNode(binary, message, fixes...)
		return
	}
}

// CheckUnescapedHTMLOutput checks the independently specified contract.
func CheckUnescapedHTMLOutput(ctx *analysis.Context, n syntax.Node, message string) {
	echo, ok := n.(*syntax.Echo)
	if !ok {
		return
	}
	if nativePlainResponse(ctx, echo) {
		return
	}
	for _, expr := range echo.Exprs {
		if nativeHTMLText(ctx, expr) && ctx.Flow().Tainted(expr, flow.HTML) {
			ctx.ReportNode(echo, message)
			return
		}
	}
}

// CheckUntrustedFilesystemPath checks the independently specified contract.
func CheckUntrustedFilesystemPath(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	switch name {
	case "readfile", "fopen", "file_get_contents", "file_put_contents", "unlink", "rmdir":
		report = ctx.Flow().Tainted(CallArgument(c.Args, 0, "filename"), flow.Path)
	}

	report = report || nativeWrapperSink(ctx, c, flow.Path)
	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckUntrustedHeaderValue checks the independently specified contract.
func CheckUntrustedHeaderValue(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	if name == "header" {
		value := CallArgument(c.Args, 0, "header")
		report = ctx.Flow().Tainted(value, flow.Header)

	}

	report = report || nativeWrapperSink(ctx, c, flow.Header)
	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckUntrustedNetworkDestination checks the independently specified contract.
func CheckUntrustedNetworkDestination(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	if name == "curl_exec" {
		handle := CallArgument(c.Args, 0, "handle")
		origin, ok := NativeValue(ctx, handle).(*syntax.FuncCall)
		if ok && NativeBuiltinName(ctx, origin) == "curl_init" {
			report = ctx.Flow().Tainted(CallArgument(origin.Args, 0, "url"), flow.URL)
		}
		for _, previous := range NativePriorCalls(ctx, c, handle, "curl_setopt") {
			option := previous.(*syntax.FuncCall)
			flag := CallArgument(option.Args, 1, "option")
			constant, ok := flag.(*syntax.ConstFetch)
			if ok && strings.EqualFold(constant.Name.Value, "CURLOPT_URL") {
				report = ctx.Flow().Tainted(CallArgument(option.Args, 2, "value"), flow.URL)
			}
		}
	}

	report = report || nativeWrapperSink(ctx, c, flow.URL, "curl_init", "curl_setopt")
	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckUntrustedShellCommand checks the independently specified contract.
func CheckUntrustedShellCommand(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	switch name {
	case "system", "exec", "shell_exec", "passthru", "popen":
		report = ctx.Flow().Tainted(CallArgument(c.Args, 0, "command"), flow.Shell)
	}

	report = report || nativeWrapperSink(ctx, c, flow.Shell)
	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckUntrustedSQLConstruction checks the independently specified contract.
func CheckUntrustedSQLConstruction(ctx *analysis.Context, n syntax.Node, message string) {
	if method, ok := n.(*syntax.MethodCall); ok {
		if NativeMethod(ctx, method, "PDO", "query") || NativeMethod(ctx, method, "PDO", "exec") || NativeMethod(ctx, method, "PDO", "prepare") {
			sql := CallArgument(method.Args, 0, "query")
			if sql == nil {
				sql = CallArgument(method.Args, 0, "statement")
			}
			if ctx.Flow().Tainted(sql, flow.SQL) {
				ctx.ReportNode(method, message)
			}
		}
		return
	}
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}

	report := false

	report = report || nativeWrapperSink(ctx, c, flow.SQL)
	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckUnvalidatedRedirectTarget checks the independently specified contract.
func CheckUnvalidatedRedirectTarget(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	if name == "header" {
		value := CallArgument(c.Args, 0, "header")
		location := false
		if binary, ok := value.(*syntax.Binary); ok && binary.Op.Kind == syntax.TDot {
			prefix, known := NativeString(ctx, binary.Left)
			location = known && strings.HasPrefix(strings.ToLower(prefix), "location:")
		}

		report = location && ctx.Flow().Tainted(value, flow.URL)

	}

	report = report || nativeWrapperSink(ctx, c, flow.URL, "header")
	if report {
		ctx.ReportNode(c, message)
	}
}

// CheckUploadClientMimeTrusted checks the independently specified contract.
func CheckUploadClientMimeTrusted(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	report = name == "move_uploaded_file" && nativeUploadMimeBranch(ctx, c)

	if report {
		ctx.ReportNode(c, message)
	}
}

// nativeWrapperSink proves an external source reaches a matching project wrapper sink.
func nativeWrapperSink(ctx *analysis.Context, c *syntax.FuncCall, context flow.Context, operations ...string) bool {
	for _, sink := range ctx.Flow().Sinks(c) {
		if len(operations) > 0 {
			matched := false
			for _, operation := range operations {
				if sink.Name == operation {
					matched = true
				}
			}
			if !matched {
				continue
			}
		}
		if sink.Context != context || !sink.Argument.Complete || sink.Argument.Safe&context != 0 {
			continue
		}
		for _, source := range sink.Argument.Sources {
			if source.Parameter < 0 {
				return true
			}
		}
	}
	return false
}
