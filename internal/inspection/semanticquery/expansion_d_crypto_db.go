package semanticquery

import (
	"net/url"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// expansionDSame proves a stable local variable even when its value comes from
// a parameter. Unknown calls, references and writes invalidate the relation.
func expansionDSame(ctx *analysis.Context, a, b syntax.Expr) bool {
	if a == nil || b == nil {
		return false
	}
	facts := nativeContractScope(ctx, syntax.EnclosingVariableScope(a))
	events := facts.nodes
	if events == nil {
		return false
	}
	at := syntax.Node(b)
	if a.Span().Start > b.Span().Start {
		at = a
	}
	if !ExpansionDUnaliased(ctx, a, at) || !ExpansionDUnaliased(ctx, b, at) {
		return false
	}
	if NativeSameValue(ctx, a, b) {
		return true
	}
	if x, xk := NativeString(ctx, a); xk {
		if y, yk := NativeString(ctx, b); yk && x == y {
			return true
		}
	}
	if x, xk := NativeContractInt(ctx, a); xk {
		if y, yk := NativeContractInt(ctx, b); yk && x == y {
			return true
		}
	}
	av, aok := syntax.UnwrapParens(a).(*syntax.Variable)
	bv, bok := syntax.UnwrapParens(b).(*syntax.Variable)
	if !aok || !bok || av.Name != bv.Name || syntax.EnclosingVariableScope(a) != syntax.EnclosingVariableScope(b) {
		return false
	}
	start, end := a.Span().End, b.Span().Start
	if start > end {
		start, end = b.Span().End, a.Span().Start
	}
	// Only a write of the variable itself separates the two reads; using it
	// as an index or a property name of another target does not.
	if facts.localWrites(ctx, syntax.EnclosingVariableScope(a)).rebound(av.Name, start, end) {
		return false
	}
	for _, event := range events {
		if event.Span().Start <= start || event.Span().End >= end {
			continue
		}
		switch e := event.(type) {
		case *syntax.FuncCall:
			if NativeBuiltinName(ctx, e) == "" && astquery.MentionsVariable(e.Args, av.Name) {
				return false
			}
		case *syntax.MethodCall:
			if astquery.MentionsVariable(e.Args, av.Name) {
				return false
			}
		}
	}
	return true
}

func expansionDCalls(ctx *analysis.Context, at syntax.Node) []syntax.Node {
	return nativeContractEvents(ctx, syntax.EnclosingVariableScope(at))
}

func expansionDByteLength(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if text, ok := NativeString(ctx, e); ok {
		return int64(len(text)), true
	}
	c, ok := NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok {
		return 0, false
	}
	switch NativeBuiltinName(ctx, c) {
	case "random_bytes":
		size, known := NativeContractInt(ctx, CallArgument(c.Args, 0, "length"))
		return size, known && size > 0
	case "str_repeat":
		text, known := NativeString(ctx, CallArgument(c.Args, 0, "string"))
		count, ok := NativeContractInt(ctx, CallArgument(c.Args, 1, "times"))
		if known && ok && count >= 0 && count <= 1<<30 {
			return int64(len(text)) * count, true
		}
	}
	return 0, false
}

func expansionDOrigin(ctx *analysis.Context, e syntax.Expr, names ...string) *syntax.FuncCall {
	if e == nil || !ExpansionDUnaliased(ctx, e, e) {
		return nil
	}
	c, ok := NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok {
		return nil
	}
	for _, name := range names {
		if NativeBuiltin(ctx, c, name) {
			return c
		}
	}
	return nil
}

func CheckHashContextUsedAfterFinalization(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	switch NativeBuiltinName(ctx, c) {
	case "hash_update", "hash_copy", "hash_final":
	default:
		return
	}
	h := CallArgument(c.Args, 0, "context")
	for _, prior := range expansionDCalls(ctx, c) {
		p, ok := prior.(*syntax.FuncCall)
		if ok && NativeBuiltin(ctx, p, "hash_final") && NativeDominates(p, c) && expansionDSame(ctx, CallArgument(p.Args, 0, "context"), h) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

func CheckHashFileFailureUsedAsDigest(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	var inputs []syntax.Expr
	switch c := n.(type) {
	case *syntax.FuncCall:
		if !NativeBuiltin(ctx, c, "file_put_contents") {
			return
		}
		inputs = []syntax.Expr{CallArgument(c.Args, 1, "data")}
	case *syntax.Binary:
		if c.Op.Kind != syntax.TIsIdentical && c.Op.Kind != syntax.TIsNotIdentical {
			return
		}
		if _, ok := NativeString(ctx, c.Right); ok {
			inputs = append(inputs, c.Left)
		}
		if _, ok := NativeString(ctx, c.Left); ok {
			inputs = append(inputs, c.Right)
		}
	}
	for _, e := range inputs {
		if expansionDOrigin(ctx, e, "hash_file") != nil && !NativeSentinelGuard(ctx, e, "false") {
			ctx.ReportNode(n, message)
			return
		}
	}
}

type expansionDOutputKey struct{ scope syntax.Node }

func expansionDOutputs(ctx *analysis.Context, at syntax.Node) []syntax.Node {
	scope := syntax.EnclosingVariableScope(at)
	return ctx.File.Memo(expansionDOutputKey{scope}, func() any {
		var out []syntax.Node
		visit := func(n syntax.Node) bool {
			if n != scope && syntax.IsVariableScope(n) {
				return false
			}
			switch n.(type) {
			case *syntax.Echo, *syntax.FuncCall:
				out = append(out, n)
			}
			return true
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		if len(out) > 512 {
			return []syntax.Node(nil)
		}
		return out
	}).([]syntax.Node)
}

func CheckOpenSslSigningFailureIgnored(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !NativeBuiltin(ctx, c, "openssl_sign") {
		return
	}
	if _, ok := c.Parent().(*syntax.ExprStmt); !ok {
		return
	}
	signature := CallArgument(c.Args, 1, "signature")
	for _, event := range expansionDOutputs(ctx, c) {
		if !NativeDominates(c, event) {
			continue
		}
		var inputs []syntax.Expr
		switch use := event.(type) {
		case *syntax.FuncCall:
			switch NativeBuiltinName(ctx, use) {
			case "base64_encode", "bin2hex":
				inputs = []syntax.Expr{CallArgument(use.Args, 0, "string")}
			case "file_put_contents":
				inputs = []syntax.Expr{CallArgument(use.Args, 1, "data")}
			}
		case *syntax.Echo:
			inputs = use.Exprs
		}
		for _, input := range inputs {
			if expansionDSame(ctx, signature, input) {
				ctx.ReportNode(c, message)
				return
			}
		}
	}
}

func CheckOpenSslEncryptionStatusUsedAsCiphertext(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	var inputs []syntax.Expr
	switch c := n.(type) {
	case *syntax.FuncCall:
		switch NativeBuiltinName(ctx, c) {
		case "base64_encode", "bin2hex":
			inputs = []syntax.Expr{CallArgument(c.Args, 0, "string")}
		case "file_put_contents":
			inputs = []syntax.Expr{CallArgument(c.Args, 1, "data")}
		}
	case *syntax.Echo:
		inputs = c.Exprs
	}
	for _, e := range inputs {
		origin := expansionDOrigin(ctx, e, "openssl_public_encrypt", "openssl_private_encrypt")
		if origin != nil {
			output, ok := CallArgument(origin.Args, 1, "encrypted_data").(*syntax.Variable)
			if !ok {
				continue
			}
			if status, ok := e.(*syntax.Variable); ok && status.Name == output.Name {
				continue
			}
			ctx.ReportNode(n, message)
			return
		}
	}
}

func CheckPasswordVerifyArgumentsReversed(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP55 || !NativeBuiltin(ctx, c, "password_verify") {
		return
	}
	origin := expansionDOrigin(ctx, CallArgument(c.Args, 0, "password"), "password_hash")
	if origin != nil && expansionDSame(ctx, CallArgument(origin.Args, 0, "password"), CallArgument(c.Args, 1, "hash")) {
		ctx.ReportNode(c, message)
	}
}

func CheckSodiumSecretboxKeyLengthMismatch(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP72 || (!NativeBuiltin(ctx, c, "sodium_crypto_secretbox") && !NativeBuiltin(ctx, c, "sodium_crypto_secretbox_open")) {
		return
	}
	length, known := expansionDByteLength(ctx, CallArgument(c.Args, 2, "key"))
	if known && length != 32 {
		ctx.ReportNode(c, message)
	}
}

func CheckSodiumSecretboxNonceReused(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP72 || !NativeBuiltin(ctx, c, "sodium_crypto_secretbox") {
		return
	}
	plaintext, known := NativeString(ctx, CallArgument(c.Args, 0, "message"))
	if !known {
		return
	}
	nonce := CallArgument(c.Args, 1, "nonce")
	key := CallArgument(c.Args, 2, "key")
	events := expansionDCalls(ctx, c)
	var latestMutation uint32
	for _, event := range events {
		m, ok := event.(*syntax.FuncCall)
		if !ok || m.Span().End >= c.Span().Start {
			continue
		}
		switch NativeBuiltinName(ctx, m) {
		case "sodium_increment", "sodium_add", "sodium_memzero":
			argument := "string"
			if NativeBuiltinName(ctx, m) == "sodium_add" {
				argument = "string1"
			}
			changed := CallArgument(m.Args, 0, argument)
			if m.Span().End > latestMutation && (astquery.Equivalent(ctx.File, nonce, changed) || astquery.Equivalent(ctx.File, key, changed)) {
				latestMutation = m.Span().End
			}
		}
	}
	for _, event := range events {
		p, ok := event.(*syntax.FuncCall)
		if !ok || p.Span().End < latestMutation || !NativeBuiltin(ctx, p, "sodium_crypto_secretbox") || !NativeDominates(p, c) {
			continue
		}
		previous, known := NativeString(ctx, CallArgument(p.Args, 0, "message"))
		if known && previous != plaintext && expansionDSame(ctx, CallArgument(p.Args, 1, "nonce"), CallArgument(c.Args, 1, "nonce")) && expansionDSame(ctx, CallArgument(p.Args, 2, "key"), CallArgument(c.Args, 2, "key")) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

func CheckDetachedSignaturePassedToCombinedVerifier(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP >= phpversion.PHP72 && NativeBuiltin(ctx, c, "sodium_crypto_sign_open") && expansionDOrigin(ctx, CallArgument(c.Args, 0, "signed_message"), "sodium_crypto_sign_detached") != nil {
		ctx.ReportNode(c, message)
	}
}

func CheckSodiumKdfContextLengthMismatch(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP72 || !NativeBuiltin(ctx, c, "sodium_crypto_kdf_derive_from_key") {
		return
	}
	length, known := expansionDByteLength(ctx, CallArgument(c.Args, 2, "context"))
	if known && length != 8 {
		ctx.ReportNode(c, message)
	}
}

func CheckSecretstreamPushAfterFinalTag(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	const name = "sodium_crypto_secretstream_xchacha20poly1305_push"
	if ctx.PHP < phpversion.PHP72 || !NativeBuiltin(ctx, c, name) {
		return
	}
	for _, event := range expansionDCalls(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if !ok || !NativeBuiltin(ctx, p, name) || !NativeDominates(p, c) || !expansionDSame(ctx, CallArgument(p.Args, 0, "state"), CallArgument(c.Args, 0, "state")) {
			continue
		}
		tag, known := NativeContractInt(ctx, CallArgument(p.Args, 3, "tag"))
		if known && tag == 3 {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// Keep URL parsing bound to Go's parser; PHP contracts are checked separately.
func expansionDAbsoluteURL(text string) bool {
	u, err := url.Parse(text)
	return err == nil && u.IsAbs() && u.Host != "" && strings.Contains(text, "://")
}

func expansionDConst(ctx *analysis.Context, e syntax.Expr, name string) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	if !ok {
		return false
	}
	resolved := GlobalConstName(ctx, c)
	symbol := ctx.Index().Constant(resolved, ctx.PHP)
	return resolved == name && symbol != nil && symbol.Builtin
}

func expansionDCurlSetting(ctx *analysis.Context, c *syntax.FuncCall, option string) syntax.Expr {
	if NativeBuiltin(ctx, c, "curl_setopt") && expansionDConst(ctx, CallArgument(c.Args, 1, "option"), option) {
		return CallArgument(c.Args, 2, "value")
	}
	if NativeBuiltin(ctx, c, "curl_setopt_array") {
		a := NativeArray(ctx, CallArgument(c.Args, 1, "options"))
		if a == nil {
			return nil
		}
		for _, item := range a.Items {
			if expansionDConst(ctx, item.Key, option) {
				return item.Value
			}
		}
	}
	return nil
}

func expansionDCurlOption(ctx *analysis.Context, at syntax.Node, handle syntax.Expr, option string) syntax.Expr {
	var result syntax.Expr
	for _, event := range expansionDCalls(ctx, at) {
		c, ok := event.(*syntax.FuncCall)
		if !ok || c.Span().End >= at.Span().Start || !expansionDSame(ctx, CallArgument(c.Args, 0, "handle"), handle) {
			continue
		}
		if !NativeDominates(c, at) {
			return nil
		}
		if !NativeBuiltin(ctx, c, "curl_setopt") && !NativeBuiltin(ctx, c, "curl_setopt_array") {
			return nil
		}
		if value := expansionDCurlSetting(ctx, c, option); value != nil {
			result = value
		}
	}
	return result
}

func expansionDFalse(e syntax.Expr) bool {
	v, k := astquery.BoolConst(syntax.UnwrapParens(e))
	return k && !v
}

func CheckCurlHeaderOptionRequiresArray(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := expansionDCurlSetting(ctx, c, "CURLOPT_HTTPHEADER")
	if value == nil {
		return
	}
	v := NativeLocalValue(ctx, value)
	switch v.(type) {
	case *syntax.Literal, *syntax.ConstFetch:
		if t := ctx.TypeOf(value); t.OnlyOf("string", "int", "float", "bool", "null") {
			ctx.ReportNode(c, message)
		}
	}
}

func CheckCurlUploadDeclaredSizeMismatch(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	size := expansionDCurlSetting(ctx, c, "CURLOPT_INFILESIZE")
	if size == nil {
		return
	}
	wanted, known := NativeContractInt(ctx, size)
	if !known || wanted < 0 {
		return
	}
	handle := CallArgument(c.Args, 0, "handle")
	uploading, known := NativeTruth(ctx, expansionDCurlOption(ctx, c, handle, "CURLOPT_UPLOAD"))
	if !known || !uploading {
		return
	}
	source := expansionDCurlOption(ctx, c, handle, "CURLOPT_INFILE")
	open := expansionDOrigin(ctx, source, "fopen")
	if open == nil {
		return
	}
	path, known := NativeString(ctx, CallArgument(open.Args, 0, "filename"))
	if !known || !strings.HasPrefix(path, "data://text/plain,") {
		return
	}
	mode, known := NativeString(ctx, CallArgument(open.Args, 1, "mode"))
	if !known || (mode != "r" && mode != "rb") {
		return
	}
	for _, event := range expansionDCalls(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if !ok || p == open || p.Span().Start <= open.Span().End || p.Span().End >= c.Span().Start {
			continue
		}
		if expansionDSame(ctx, CallArgument(p.Args, 0, "stream"), source) && !NativeBuiltin(ctx, p, "curl_setopt") {
			return
		}
	}
	data, err := url.PathUnescape(strings.TrimPrefix(path, "data://text/plain,"))
	if err == nil && int64(len(data)) != wanted {
		ctx.ReportNode(c, message)
	}
}

func CheckCurlReadCallbackExceedsRequestedSize(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := expansionDCurlSetting(ctx, c, "CURLOPT_READFUNCTION")
	if value == nil {
		return
	}
	params, body := NativeCallback(ctx, value)
	if len(params) < 3 {
		return
	}
	returns, complete := NativeReturns(body)
	if !complete {
		return
	}
	for _, ret := range returns {
		r, ok := ret.(*syntax.FuncCall)
		if !ok || !NativeBuiltin(ctx, r, "str_repeat") {
			return
		}
		text, known := NativeString(ctx, CallArgument(r.Args, 0, "string"))
		if !known || len(text) != 1 {
			return
		}
		sum, ok := syntax.UnwrapParens(CallArgument(r.Args, 1, "times")).(*syntax.Binary)
		if !ok || sum.Op.Kind != syntax.TPlus {
			return
		}
		size, ok := syntax.UnwrapParens(sum.Left).(*syntax.Variable)
		extra, known := NativeContractInt(ctx, sum.Right)
		if !ok || size.Name != params[2].Var.Name || !known || extra <= 0 {
			return
		}
	}
	// Any possible write to the requested-size parameter, wherever it sits in
	// the callback (a loop may run it before a return), invalidates the proof.
	if expansionDCurlSizeRebound(ctx, body, params[2].Var.Name) {
		return
	}
	ctx.ReportNode(c, message)
}

// expansionDCurlSizeRebound reports whether the callback body may give its
// local variable $name another value: direct and destructuring assignments,
// ++/--, foreach targets, global/static/catch/unset, references taken by
// assignment, array item or closure capture, the variable passed as a call
// argument (the parameter may be by-reference), and constructs that write
// the local scope by name (variable-variables, extract, include, eval).
func expansionDCurlSizeRebound(ctx *analysis.Context, body syntax.Node, name string) bool {
	rebound := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if rebound {
			return false
		}
		if syntax.IsVariableScope(n) {
			if closure, ok := n.(*syntax.Closure); ok {
				for _, use := range closure.Uses {
					rebound = rebound || use.ByRef && use.Var.Name == name
				}
			}
			return false
		}
		switch x := n.(type) {
		case *syntax.Variable:
			rebound = x.NameExpr != nil
		case *syntax.Assign:
			rebound = astquery.MentionsVariable(x.Var, name) || x.ByRef && astquery.MentionsVariable(x.Value, name)
		case *syntax.IncDec:
			rebound = astquery.MentionsVariable(x.Var, name)
		case *syntax.Foreach:
			rebound = astquery.MentionsVariable(x.Key, name) || astquery.MentionsVariable(x.Value, name)
		case *syntax.Global, *syntax.StaticVar, *syntax.Unset:
			rebound = astquery.MentionsVariable(x, name)
		case *syntax.Catch:
			rebound = x.Var != nil && x.Var.Name == name
		case *syntax.ArrayItem:
			rebound = x.ByRef && astquery.MentionsVariable(x.Value, name)
		case *syntax.Arg:
			v, ok := syntax.UnwrapParens(x.Value).(*syntax.Variable)
			rebound = ok && v.Name == name
		case *syntax.Include, *syntax.Eval:
			rebound = true
		case *syntax.FuncCall:
			switch ctx.GlobalFunctionName(x) {
			case "extract", "parse_str", "mb_parse_str", "import_request_variables":
				rebound = true
			}
		}
		return !rebound
	})
	return rebound
}

func CheckCurlHeaderCallbackMissingByteCount(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := expansionDCurlSetting(ctx, c, "CURLOPT_HEADERFUNCTION")
	if value == nil {
		return
	}
	_, body := NativeCallback(ctx, value)
	block, ok := body.(*syntax.Block)
	if !ok || NativeGeneratorBody(block) {
		return
	}
	missing := !ExpansionDCallbackTerminates(ctx, block)
	syntax.Inspect(block, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if r, ok := n.(*syntax.Return); ok && r.Expr == nil && NativeCallbackReachable(r) {
			missing = true
		}
		return true
	})
	if missing {
		ctx.ReportNode(c, message)
	}
}

func CheckCurlResponseZeroRejected(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c, ok := n.(*syntax.FuncCall)
	if !ok || !NativeBuiltin(ctx, c, "curl_exec") || expansionDFailureBranch(c) == nil {
		return
	}
	if curlResponseReturned(ctx, c, CallArgument(c.Args, 0, "handle")) {
		ctx.ReportNode(c, message)
	}
}

// curlResponseReturned replays every dominating option write on the handle in
// source order. Only a state whose last relevant write is a truthy
// RETURNTRANSFER proves that curl_exec() yields the body.
func curlResponseReturned(ctx *analysis.Context, at syntax.Node, handle syntax.Expr) bool {
	proven := false
	for _, event := range expansionDCalls(ctx, at) {
		c, ok := event.(*syntax.FuncCall)
		if !ok || c.Span().End >= at.Span().Start || !expansionDSame(ctx, CallArgument(c.Args, 0, "handle"), handle) {
			continue
		}
		if !NativeDominates(c, at) {
			return false
		}
		switch NativeBuiltinName(ctx, c) {
		case "curl_setopt":
			proven = curlResponseWrite(ctx, proven, CallArgument(c.Args, 1, "option"), CallArgument(c.Args, 2, "value"))
		case "curl_setopt_array":
			options := NativeArray(ctx, CallArgument(c.Args, 1, "options"))
			if options == nil {
				proven = false
				continue
			}
			for _, item := range options.Items {
				proven = curlResponseWrite(ctx, proven, item.Key, item.Value)
			}
		default:
			return false
		}
	}
	return proven
}

// curlResponseWrite applies one option write. An option that is not a builtin
// constant may be anything, and a replaced write sink makes curl_exec() return
// a boolean again.
func curlResponseWrite(ctx *analysis.Context, proven bool, option, value syntax.Expr) bool {
	name := ""
	if c, ok := syntax.UnwrapParens(option).(*syntax.ConstFetch); ok {
		name = GlobalConstName(ctx, c)
		if symbol := ctx.Index().Constant(name, ctx.PHP); symbol == nil || !symbol.Builtin {
			name = ""
		}
	}
	switch name {
	case "CURLOPT_RETURNTRANSFER":
		enabled, known := NativeTruth(ctx, value)
		return known && enabled
	case "", "CURLOPT_FILE", "CURLOPT_WRITEFUNCTION":
		return false
	}
	return proven
}

func CheckCurlWholeURLEscapedAsComponent(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := expansionDCurlSetting(ctx, c, "CURLOPT_URL")
	origin := expansionDOrigin(ctx, value, "curl_escape")
	if origin == nil {
		return
	}
	text, known := NativeString(ctx, CallArgument(origin.Args, 1, "string"))
	if known && expansionDAbsoluteURL(text) {
		ctx.ReportNode(c, message)
	}
}

func CheckNonblockingSocketConnectPendingRejected(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !NativeBuiltin(ctx, c, "socket_connect") {
		return
	}
	branch := expansionDFailureBranch(c)
	socket := CallArgument(c.Args, 0, "socket")
	if branch == nil || nonblockingConnectInspected(ctx, branch.Body, socket) {
		return
	}
	var latest *syntax.FuncCall
	for _, event := range expansionDCalls(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if !ok || p.Span().End >= c.Span().Start || !expansionDSame(ctx, CallArgument(p.Args, 0, "socket"), socket) {
			continue
		}
		if NativeBuiltin(ctx, p, "socket_set_block") || NativeBuiltin(ctx, p, "socket_set_nonblock") {
			latest = p
		}
	}
	// The last mode switch must be the nonblocking one and must have run:
	// either its success guards the connect or it is a dominating statement.
	if latest != nil && NativeBuiltin(ctx, latest, "socket_set_nonblock") && (NativeCallSucceeded(ctx, latest, c) || NativeDominates(latest, c)) {
		ctx.ReportNode(c, message)
	}
}

// nonblockingConnectInspected reports whether the failure branch may look at
// the reason of the failure: a direct error query, or a call that is not a
// builtin function and can reach the socket through its receiver or arguments.
func nonblockingConnectInspected(ctx *analysis.Context, body syntax.Node, socket syntax.Expr) bool {
	var names []string
	syntax.Inspect(socket, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil {
			names = append(names, v.Name)
		}
		return true
	})
	reaches := func(parts ...syntax.Node) bool {
		for _, part := range parts {
			for _, name := range names {
				if astquery.MentionsVariable(part, name) {
					return true
				}
			}
		}
		return false
	}
	inspected := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch x := n.(type) {
		case *syntax.FuncCall:
			switch NativeBuiltinName(ctx, x) {
			case "socket_last_error", "socket_get_option", "socket_getopt":
				inspected = true
			case "":
				inspected = inspected || reaches(x.Name, x.Args)
			}
		case *syntax.MethodCall:
			inspected = inspected || reaches(x.Var, x.Args)
		case *syntax.StaticCall:
			inspected = inspected || reaches(x.Class, x.Args)
		}
		return !inspected
	})
	return inspected
}

func CheckSocketReadZeroRejected(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if NativeBuiltin(ctx, c, "socket_read") && socketReadBinaryMode(ctx, c) && expansionDFailureBranch(c) != nil {
		ctx.ReportNode(c, message)
	}
}

// socketReadBinaryMode proves the binary read mode: either the builtin constant
// is passed, or the mode is left to its default, which is PHP_BINARY_READ. A
// spread or a first-class callable placeholder may bind anything.
func socketReadBinaryMode(ctx *analysis.Context, c *syntax.FuncCall) bool {
	positional := 0
	for _, node := range c.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack {
			return false
		}
		if arg.Name == nil {
			positional++
		} else if arg.Name.Value == "mode" {
			positional = 3
		}
	}
	if positional < 3 {
		return true
	}
	return expansionDConst(ctx, CallArgument(c.Args, 2, "mode"), "PHP_BINARY_READ")
}

func CheckTLSHandshakePendingAccepted(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !NativeBuiltin(ctx, c, "stream_socket_enable_crypto") {
		return
	}
	b, ok := c.Parent().(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TIsNotIdentical || !expansionDFalse(b.Right) {
		return
	}
	branch, ok := b.Parent().(*syntax.If)
	if !ok {
		return
	}
	stream := CallArgument(c.Args, 0, "stream")
	var latest *syntax.FuncCall
	for _, event := range expansionDCalls(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if ok && p.Span().End < c.Span().Start && NativeBuiltin(ctx, p, "stream_set_blocking") && expansionDSame(ctx, CallArgument(p.Args, 0, "stream"), stream) {
			latest = p
		}
	}
	if latest == nil || !expansionDFalse(CallArgument(latest.Args, 1, "enable")) || !NativeCallSucceeded(ctx, latest, c) {
		return
	}
	stmts := []syntax.Stmt{branch.Body}
	if block, ok := branch.Body.(*syntax.Block); ok {
		stmts = block.Stmts
	}
	found := false
	for _, stmt := range stmts {
		// Statements after a strict completion guard only run once a later
		// handshake call returned true.
		if tlsHandshakeCompletionGuard(ctx, stmt, stream) {
			break
		}
		syntax.Inspect(stmt, func(n syntax.Node) bool {
			if syntax.IsVariableScope(n) {
				return false
			}
			switch n.(type) {
			case *syntax.If, *syntax.While, *syntax.For, *syntax.Foreach, *syntax.DoWhile, *syntax.Switch, *syntax.Try:
				return false
			}
			if p, ok := n.(*syntax.FuncCall); ok && NativeBuiltin(ctx, p, "fwrite") && expansionDSame(ctx, stream, CallArgument(p.Args, 0, "stream")) {
				found = true
			}
			return true
		})
	}
	if found {
		ctx.ReportNode(c, message)
	}
}

// tlsHandshakeCompletionGuard recognizes `if (<handshake> !== true) { leave }`:
// a terminating branch taken for every result of a further enabling handshake
// call on the stream except true. `!`, `!= true` and `!== true` all reject
// both false and the pending 0; the result may be assigned inside the test.
func tlsHandshakeCompletionGuard(ctx *analysis.Context, stmt syntax.Stmt, stream syntax.Expr) bool {
	guard, ok := stmt.(*syntax.If)
	if !ok || !syntax.Terminates(guard.Body) {
		return false
	}
	var tested syntax.Expr
	switch cond := syntax.UnwrapParens(guard.Cond).(type) {
	case *syntax.Unary:
		if cond.Op.Kind == syntax.TExclaim {
			tested = cond.Expr
		}
	case *syntax.Binary:
		if value, known := astquery.BoolConst(syntax.UnwrapParens(cond.Right)); known && value && (cond.Op.Kind == syntax.TIsNotIdentical || cond.Op.Kind == syntax.TIsNotEqual) {
			tested = cond.Left
		}
	}
	tested = syntax.UnwrapParens(tested)
	if a, ok := tested.(*syntax.Assign); ok && !a.ByRef && a.Op.Kind == syntax.TEqual {
		tested = syntax.UnwrapParens(a.Value)
	}
	c, ok := tested.(*syntax.FuncCall)
	if !ok || !NativeBuiltin(ctx, c, "stream_socket_enable_crypto") || !expansionDSame(ctx, stream, CallArgument(c.Args, 0, "stream")) {
		return false
	}
	enabling, known := NativeTruth(ctx, CallArgument(c.Args, 1, "enable"))
	return known && enabling
}

func CheckFtpPendingTransferAcceptedAsComplete(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	switch NativeBuiltinName(ctx, c) {
	case "ftp_nb_get", "ftp_nb_put", "ftp_nb_fget", "ftp_nb_fput", "ftp_nb_continue":
	default:
		return
	}
	body := ftpPendingAcceptedBody(c)
	if body == nil {
		return
	}
	// A branch that mentions FTP_FINISHED checks or drives the transfer to
	// completion itself; whether it does so correctly is not provable here.
	handled := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if k, ok := n.(*syntax.ConstFetch); ok && strings.EqualFold(GlobalConstName(ctx, k), "FTP_FINISHED") {
			handled = true
		}
		return !handled
	})
	if handled {
		return
	}
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch n.(type) {
		case *syntax.If, *syntax.While, *syntax.For, *syntax.Foreach, *syntax.DoWhile, *syntax.Switch, *syntax.Try:
			return false
		}
		if r, ok := n.(*syntax.Return); ok {
			v, k := astquery.BoolConst(r.Expr)
			found = found || k && v
		}
		if echo, ok := n.(*syntax.Echo); ok {
			for _, v := range echo.Exprs {
				text, k := NativeString(ctx, v)
				found = found || k && strings.EqualFold(strings.TrimSpace(text), "complete")
			}
		}
		return true
	})
	if found {
		ctx.ReportNode(c, message)
	}
}

// ftpPendingAcceptedBody returns the statement an `if` runs when the status
// returned by call is non-zero, so FTP_MOREDATA reaches it: the body of a
// truthy test, the `else` of a negated one. It returns nil when the call is
// not such a condition. Parentheses, assignments, `!` and comparisons with a
// boolean literal are followed; the status is an int, so a strict comparison
// of the status itself with a boolean has a fixed outcome and selects the
// branch that always runs.
func ftpPendingAcceptedBody(call *syntax.FuncCall) syntax.Stmt {
	// `ftp_nb_get(...)` builds a closure and starts no transfer.
	for _, arg := range call.Args.Args {
		if _, ok := arg.(*syntax.VariadicPlaceholder); ok {
			return nil
		}
	}
	var e syntax.Expr = call
	truthy, status := true, true
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
		case *syntax.Assign:
			if p.Value != e || p.Op.Kind != syntax.TEqual || p.ByRef {
				return nil
			}
		case *syntax.Unary:
			if p.Op.Kind != syntax.TExclaim {
				return nil
			}
			truthy, status = !truthy, false
		case *syntax.Binary:
			other := p.Right
			if other == e {
				other = p.Left
			}
			want, ok := astquery.BoolConst(other)
			if !ok {
				return nil
			}
			switch p.Op.Kind {
			case syntax.TIsEqual:
				truthy = truthy == want
			case syntax.TIsNotEqual:
				truthy = truthy != want
			case syntax.TIsIdentical:
				truthy = !status && truthy == want
			case syntax.TIsNotIdentical:
				truthy = status || truthy != want
			default:
				return nil
			}
			status = false
		case *syntax.If:
			switch {
			case truthy:
				return p.Body
			case p.Else != nil:
				return p.Else.Body
			}
			return nil
		default:
			return nil
		}
		e = e.Parent().(syntax.Expr)
	}
}

// expansionDSQL tokenizes a deliberately bounded SQL subset. Strings and
// comments never become keywords or placeholders; uncertain syntax is refused.
func expansionDSQL(sql string) ([]string, bool) {
	if len(sql) > 65536 {
		return nil, false
	}
	var out []string
	for i := 0; i < len(sql); {
		ch := sql[i]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			i++
			continue
		}
		if ch == '$' {
			return nil, false
		}
		if ch == '#' || (ch == '-' && i+1 < len(sql) && sql[i+1] == '-') {
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			continue
		}
		if ch == '/' && i+1 < len(sql) && sql[i+1] == '*' {
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 || strings.Contains(sql[i+2:i+2+end], "/*") {
				return nil, false
			}
			i += end + 4
			continue
		}
		if ch == '\'' || ch == '"' || ch == '`' {
			q := ch
			i++
			var text strings.Builder
			closed := false
			for i < len(sql) {
				if sql[i] == '\\' {
					if i+1 >= len(sql) {
						return nil, false
					}
					text.WriteByte(sql[i+1])
					i += 2
					continue
				}
				if sql[i] == q {
					if i+1 < len(sql) && sql[i+1] == q {
						text.WriteByte(q)
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				text.WriteByte(sql[i])
				i++
			}
			if !closed {
				return nil, false
			}
			prefix := "'"
			if q == '`' {
				prefix = "`"
			}
			out = append(out, prefix+text.String())
			continue
		}
		if ch == ':' && i+1 < len(sql) && sql[i+1] == ':' {
			out = append(out, "::")
			i += 2
			continue
		}
		if ch == ':' || ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' {
			start := i
			i++
			for i < len(sql) {
				c := sql[i]
				if c != '_' && !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') {
					break
				}
				i++
			}
			word := sql[start:i]
			if ch != ':' {
				word = strings.ToUpper(word)
			}
			out = append(out, word)
			continue
		}
		out = append(out, string(ch))
		i++
	}
	return out, true
}

func expansionDMethodSucceeded(ctx *analysis.Context, c *syntax.MethodCall, at syntax.Node) bool {
	accepts := func(e syntax.Expr, truth bool) bool {
		e = syntax.UnwrapParens(e)
		if e == c {
			return truth
		}
		if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim && syntax.UnwrapParens(u.Expr) == c {
			return !truth
		}
		if b, ok := e.(*syntax.Binary); ok && syntax.UnwrapParens(b.Left) == c {
			v, k := astquery.BoolConst(b.Right)
			return k && ((truth && (b.Op.Kind == syntax.TIsIdentical || b.Op.Kind == syntax.TIsEqual) && v) || (!truth && (b.Op.Kind == syntax.TIsIdentical || b.Op.Kind == syntax.TIsEqual) && !v))
		}
		return false
	}
	for p := at.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if branch, ok := p.(*syntax.If); ok && branch.Body.Span().Contains(at.Span()) && accepts(branch.Cond, true) {
			return true
		}
		if st, ok := p.(syntax.Stmt); ok {
			prev, exists := astquery.PrevStmt(ctx.File, st)
			guard, ok := prev.(*syntax.If)
			if exists && ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && accepts(guard.Cond, false) {
				return true
			}
		}
	}
	return false
}

func expansionDMethodCalls(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr, class string, names ...string) []*syntax.MethodCall {
	var out []*syntax.MethodCall
	for _, event := range expansionDCalls(ctx, at) {
		c, ok := event.(*syntax.MethodCall)
		if !ok || c.Span().End >= at.Span().Start || (!NativeDominates(c, at) && !expansionDMethodSucceeded(ctx, c, at)) || !expansionDSame(ctx, c.Var, receiver) {
			continue
		}
		for _, name := range names {
			if NativeMethod(ctx, c, class, name) {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

func expansionDPdoDriver(ctx *analysis.Context, e syntax.Expr, driver string) bool {
	construct := NativeConstruction(ctx, e, "PDO")
	if construct == nil {
		return false
	}
	dsn, known := NativeString(ctx, CallArgument(construct.Args, 0, "dsn"))
	return known && strings.HasPrefix(strings.ToLower(dsn), driver+":")
}

// pdoBoundMayRebind reports whether a later binding can target the placeholder
// of an earlier one. PDO accepts a named placeholder with or without its
// leading colon; a name or position that is not statically known may match.
func pdoBoundMayRebind(ctx *analysis.Context, a, b syntax.Expr) bool {
	an, as := NativeString(ctx, a)
	bn, bs := NativeString(ctx, b)
	if as && bs {
		return strings.TrimPrefix(an, ":") == strings.TrimPrefix(bn, ":")
	}
	ai, ak := NativeInt(ctx, a)
	bi, bk := NativeInt(ctx, b)
	if ak && bk {
		return ai == bi
	}
	return (!as && !ak) || (!bs && !bk)
}

func CheckPdoBoundValueAssumedLive(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	_, paramsState := CallArgumentBinding(c.Args, 0, "params")
	if !NativeMethod(ctx, c, "PDOStatement", "execute") || paramsState != ArgumentMissing {
		return
	}
	bindings := expansionDMethodCalls(ctx, c, c.Var, "PDOStatement", "bindvalue", "bindparam")
	for i, binding := range bindings {
		if !NativeMethod(ctx, binding, "PDOStatement", "bindValue") {
			continue
		}
		variable, ok := CallArgument(binding.Args, 1, "value").(*syntax.Variable)
		if !ok {
			return
		}
		typ := ctx.TypeOf(variable)
		if !typ.OnlyOf("int", "string", "float", "bool", "null") {
			continue
		}
		rebound := false
		for _, later := range bindings[i+1:] {
			if pdoBoundMayRebind(ctx, CallArgument(binding.Args, 0, "param"), CallArgument(later.Args, 0, "param")) {
				rebound = true
			}
		}
		if rebound {
			continue
		}
		for _, event := range expansionDCalls(ctx, c) {
			a, ok := event.(*syntax.Assign)
			if !ok || a.Span().Start <= binding.Span().End || a.Span().End >= c.Span().Start || !NativeDominates(a, c) {
				continue
			}
			v, ok := a.Var.(*syntax.Variable)
			if ok && v.Name == variable.Name && !a.ByRef {
				ctx.ReportNode(c, message)
				return
			}
		}
	}
}

// pdoEmulationMayChange reports whether a setAttribute call on the connection
// of prepare can turn statement emulation back on after offset proven, where
// native preparation was last established on every path. Such a call counts
// when it sits between that point and prepare, or after prepare inside a loop
// that repeats both without repeating the proof. Closure bodies are included:
// a captured connection is the same object. Only a call that names another
// attribute constant, or keeps emulation provably off, leaves the proof alone.
func pdoEmulationMayChange(ctx *analysis.Context, prepare *syntax.MethodCall, proven uint32) bool {
	receiver, ok := syntax.UnwrapParens(prepare.Var).(*syntax.Variable)
	if !ok {
		return false
	}
	scope := syntax.EnclosingVariableScope(prepare)
	changed := false
	visit := func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
		default:
			if n != scope && syntax.IsVariableScope(n) {
				return false
			}
		}
		m, ok := n.(*syntax.MethodCall)
		if changed || !ok || m == prepare || m.Span().Start < proven {
			return !changed
		}
		id, named := m.Name.(*syntax.Identifier)
		v, local := syntax.UnwrapParens(m.Var).(*syntax.Variable)
		if !named || !local || !strings.EqualFold(id.Value, "setAttribute") || v.Name != receiver.Name {
			return true
		}
		if cc, ok := CallArgument(m.Args, 0, "attribute").(*syntax.ClassConstFetch); ok {
			if name, ok := cc.Name.(*syntax.Identifier); ok && name.Value != "ATTR_EMULATE_PREPARES" && !strings.HasSuffix(name.Value, "ATTR_DIRECT_QUERY") {
				return true
			}
			if enabled, known := NativeTruth(ctx, CallArgument(m.Args, 1, "value")); known && !enabled {
				return true
			}
		}
		if m.Span().End <= prepare.Span().Start {
			changed = true
			return false
		}
		for p := m.Parent(); p != nil; p = p.Parent() {
			switch p.(type) {
			case *syntax.While, *syntax.For, *syntax.Foreach, *syntax.DoWhile:
				if p.Span().Contains(prepare.Span()) && p.Span().Start >= proven {
					changed = true
					return false
				}
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return changed
}

func CheckPdoRepeatedNamedMarkerWithoutEmulation(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !NativeMethod(ctx, c, "PDO", "prepare") || !expansionDPdoDriver(ctx, c.Var, "mysql") {
		return
	}
	var emulation syntax.Expr
	construct := NativeConstruction(ctx, c.Var, "PDO")
	proven := construct.Span().End
	if a := NativeArray(ctx, CallArgument(construct.Args, 3, "options")); a != nil {
		for _, item := range a.Items {
			if cc, ok := item.Key.(*syntax.ClassConstFetch); ok && ExpansionCConstant(ctx, cc, "PDO", "ATTR_EMULATE_PREPARES") {
				emulation = item.Value
			}
		}
	}
	for _, prior := range expansionDMethodCalls(ctx, c, c.Var, "PDO", "setAttribute") {
		if cc, ok := CallArgument(prior.Args, 0, "attribute").(*syntax.ClassConstFetch); ok && ExpansionCConstant(ctx, cc, "PDO", "ATTR_EMULATE_PREPARES") {
			emulation, proven = CallArgument(prior.Args, 1, "value"), prior.Span().End
		}
	}
	enabled, known := NativeTruth(ctx, emulation)
	if !known || enabled || pdoEmulationMayChange(ctx, c, proven) {
		return
	}
	sql, known := NativeString(ctx, CallArgument(c.Args, 0, "query"))
	if !known {
		return
	}
	tokens, known := expansionDSQL(sql)
	if !known {
		return
	}
	seen := map[string]bool{}
	for _, token := range tokens {
		if strings.HasPrefix(token, ":") && token != "::" && len(token) > 1 {
			if seen[token] {
				ctx.ReportNode(c, message)
				return
			}
			seen[token] = true
		}
	}
}

func expansionDProjection(sql string) ([]string, bool) {
	tokens, known := expansionDSQL(sql)
	if !known || len(tokens) < 2 || tokens[0] != "SELECT" {
		return nil, false
	}
	var columns []string
	start := 1
	for i := 1; ; i++ {
		if i < len(tokens) && tokens[i] != "FROM" && tokens[i] != "," {
			if tokens[i] == "*" || tokens[i] == "(" || tokens[i] == ")" || tokens[i] == ";" {
				return nil, false
			}
			continue
		}
		part := tokens[start:i]
		if len(part) == 0 {
			return nil, false
		}
		name := ""
		if len(part) == 1 && !strings.HasPrefix(part[0], "'") {
			name = part[0]
		} else if len(part) >= 3 && part[len(part)-2] == "AS" {
			name = part[len(part)-1]
		}
		if name == "" || strings.HasPrefix(name, "'") {
			return nil, false
		}
		name = strings.ToUpper(strings.TrimPrefix(name, "`"))
		for _, prior := range columns {
			if prior == name {
				return nil, false
			}
		}
		columns = append(columns, name)
		if i == len(tokens) || tokens[i] == "FROM" {
			return columns, true
		}
		start = i + 1
	}
}

// pdoGroupedRoot names the plain variable an array element path starts from,
// or "" when the path has another base.
func pdoGroupedRoot(e syntax.Expr) string {
	for {
		switch x := syntax.UnwrapParens(e).(type) {
		case *syntax.ArrayDimFetch:
			e = x.Var
		case *syntax.Variable:
			return x.Name
		default:
			return ""
		}
	}
}

// pdoGroupedRowsRewritten reports whether the grouped rows may be changed
// through a reference before at: a by-reference foreach over them, a
// reference taken to them, or a call that receives them in a by-reference
// parameter or whose signature is unknown. Such code can restore the column.
func pdoGroupedRowsRewritten(ctx *analysis.Context, rows syntax.Expr, at syntax.Node) bool {
	name := pdoGroupedRoot(rows)
	if name == "" {
		return false
	}
	scope := syntax.EnclosingVariableScope(at)
	found := false
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= at.Span().Start || (n != scope && syntax.IsVariableScope(n)) {
			return false
		}
		switch x := n.(type) {
		case *syntax.Foreach:
			found = x.ByRef && pdoGroupedRoot(x.Expr) == name
		case *syntax.Assign:
			found = x.ByRef && pdoGroupedRoot(x.Value) == name
		case *syntax.Arg:
			if pdoGroupedRoot(x.Value) == name {
				found = ArgumentMayWrite(ctx.Types(), x)
			}
		}
		return !found
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}

// pdoGroupedLegacyQuery proves a PDO::query() call for targets below PHP 8.0.
// The stub index keeps only the 8.0 declaration of that method, so there the
// builtin is looked up without a version; a project override is not builtin.
func pdoGroupedLegacyQuery(ctx *analysis.Context, c *syntax.MethodCall) bool {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "query") || !ctx.PHP.Below(phpversion.PHP80) {
		return false
	}
	classes := ctx.Types().Native().TypeOf(c.Var).Classes()
	if len(classes) != 1 || !ctx.Index().IsSubtype(classes[0], "PDO", ctx.PHP) {
		return false
	}
	m := ctx.Index().FindMethod(classes[0], "query", 0)
	return m != nil && m.Builtin
}

// pdoGroupedLegacyFetchAll accepts fetchAll() on the untyped result of a
// proven legacy PDO::query() call, which is a PDOStatement when it is used.
func pdoGroupedLegacyFetchAll(ctx *analysis.Context, c *syntax.MethodCall) bool {
	id, ok := c.Name.(*syntax.Identifier)
	return ok && strings.EqualFold(id.Value, "fetchAll") && len(ctx.Types().Native().TypeOf(c.Var).Classes()) == 0
}

func CheckPdoGroupedFetchReadsRemovedColumn(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	fetch := n.(*syntax.ArrayDimFetch)
	if !astquery.ArrayFetchRequiresRead(fetch) {
		return
	}
	column, known := NativeString(ctx, fetch.Dim)
	if !known {
		return
	}
	row, ok := fetch.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	group, ok := row.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	if !ExpansionDPristine(ctx, group.Var, fetch) || pdoGroupedRowsRewritten(ctx, group.Var, fetch) {
		return
	}
	all, ok := NativeLocalValue(ctx, group.Var).(*syntax.MethodCall)
	if !ok {
		return
	}
	statement, ok := NativeLocalValue(ctx, all.Var).(*syntax.MethodCall)
	if !ok {
		return
	}
	legacy := pdoGroupedLegacyQuery(ctx, statement)
	if !legacy && !NativeMethod(ctx, statement, "PDO", "query") && !NativeMethod(ctx, statement, "PDO", "prepare") {
		return
	}
	if !NativeMethod(ctx, all, "PDOStatement", "fetchAll") && !(legacy && pdoGroupedLegacyFetchAll(ctx, all)) {
		return
	}
	modeExpr := CallArgument(all.Args, 0, "mode")
	mode, known := NativeContractInt(ctx, modeExpr)
	if b, ok := syntax.UnwrapParens(modeExpr).(*syntax.Binary); ok && b.Op.Kind == syntax.TPlus {
		left, lk := NativeContractInt(ctx, b.Left)
		right, rk := NativeContractInt(ctx, b.Right)
		mode, known = left|right, lk && rk && left&right == 0
	}
	flagsByName := false
	if b, ok := syntax.UnwrapParens(modeExpr).(*syntax.Binary); ok && (b.Op.Kind == syntax.TPlus || b.Op.Kind == syntax.TBar) {
		flagsByName = (ExpansionCConstant(ctx, b.Left, "PDO", "FETCH_GROUP") && ExpansionCConstant(ctx, b.Right, "PDO", "FETCH_ASSOC")) || (ExpansionCConstant(ctx, b.Right, "PDO", "FETCH_GROUP") && ExpansionCConstant(ctx, b.Left, "PDO", "FETCH_ASSOC"))
	}
	if !flagsByName && (!known || mode != 65538) {
		return
	}
	sql, known := NativeString(ctx, CallArgument(statement.Args, 0, "query"))
	if !known {
		return
	}
	columns, known := expansionDProjection(sql)
	if known && len(columns) > 0 && columns[0] == strings.ToUpper(column) {
		ctx.ReportNode(fetch, message)
	}
}

// pdoFetchIntoMethod proves a PDOStatement method call. Targets below PHP 8.0
// declare no native return type for PDO::query(), so a receiver whose latest
// local write is a dominating assignment of that builtin call is accepted too.
func pdoFetchIntoMethod(ctx *analysis.Context, c *syntax.MethodCall, name string) bool {
	if NativeMethod(ctx, c, "PDOStatement", name) {
		return true
	}
	id, ok := c.Name.(*syntax.Identifier)
	receiver, simple := c.Var.(*syntax.Variable)
	if !ok || !simple || !strings.EqualFold(id.Value, name) || !ExpansionDUnaliased(ctx, receiver, c) {
		return false
	}
	var origin *syntax.MethodCall
	for _, event := range expansionDCalls(ctx, c) {
		write, ok := event.(*syntax.Assign)
		if !ok || write.Span().Start >= c.Span().Start || !astquery.MentionsVariable(write.Var, receiver.Name) {
			continue
		}
		origin = nil
		if target, ok := write.Var.(*syntax.Variable); ok && target.Name == receiver.Name && !write.ByRef && write.Op.Kind == syntax.TEqual && NativeDominates(write, c) {
			origin, _ = write.Value.(*syntax.MethodCall)
		}
	}
	if origin == nil {
		return false
	}
	// The stub index lists PDO::query() from PHP 8.0 only: look the factory up
	// without a target version and still reject user overrides.
	factory, ok := origin.Name.(*syntax.Identifier)
	classes := ctx.Types().Native().TypeOf(origin.Var).Classes()
	if !ok || !strings.EqualFold(factory.Value, "query") || len(classes) != 1 || !ctx.Index().IsSubtype(classes[0], "PDO", ctx.PHP) {
		return false
	}
	method := ctx.Index().FindMethod(classes[0], "query", 0)
	return method != nil && method.Builtin
}

// pdoFetchIntoMode returns the setFetchMode call that governs every fetch of
// the loop. Another call on the same statement that is conditional, or that
// runs inside a loop around the fetch, leaves the mode and its target unproven.
func pdoFetchIntoMode(ctx *analysis.Context, fetch *syntax.MethodCall, loop syntax.Node) *syntax.MethodCall {
	end := loop.Span().End
	for p := loop.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			end = p.Span().End
		}
	}
	var latest *syntax.MethodCall
	for _, event := range expansionDCalls(ctx, fetch) {
		c, ok := event.(*syntax.MethodCall)
		if !ok || c.Span().Start >= end {
			continue
		}
		id, ok := c.Name.(*syntax.Identifier)
		if !ok || !strings.EqualFold(id.Value, "setFetchMode") || !astquery.Equivalent(ctx.File, c.Var, fetch.Var) {
			continue
		}
		latest = nil
		if c.Span().End < fetch.Span().Start && (NativeDominates(c, fetch) || expansionDMethodSucceeded(ctx, c, fetch)) && pdoFetchIntoMethod(ctx, c, "setFetchMode") && expansionDSame(ctx, c.Var, fetch.Var) {
			latest = c
		}
	}
	return latest
}

// pdoFetchIntoSingleRow reports whether the loop is left unconditionally after
// the append, so at most one fetched row is ever retained.
func pdoFetchIntoSingleRow(ctx *analysis.Context, assign syntax.Node, loop *syntax.While) bool {
	for n := assign.Parent(); n != loop; n = n.Parent() {
		st, ok := n.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, ok := astquery.StmtList(ctx.File, st)
		if !ok {
			continue
		}
		for _, next := range list[i+1:] {
			switch x := next.(type) {
			case *syntax.Return:
				return true
			case *syntax.Break:
				if pdoFetchIntoBreaksLoop(x, loop) {
					return true
				}
			case *syntax.ExprStmt:
				if syntax.Terminates(x) {
					return true
				}
			}
		}
	}
	return false
}

func pdoFetchIntoBreaksLoop(st *syntax.Break, loop *syntax.While) bool {
	for p := st.Parent(); p != loop; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach, *syntax.Switch:
			return false
		}
	}
	return true
}

func CheckPdoFetchIntoRetainsSharedRows(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	assign := n.(*syntax.Assign)
	target, ok := assign.Var.(*syntax.ArrayDimFetch)
	if !ok || target.Dim != nil || assign.Op.Kind != syntax.TEqual {
		return
	}
	loop := assign.Parent()
	for loop != nil {
		if _, ok := loop.(*syntax.While); ok {
			break
		}
		if syntax.IsVariableScope(loop) {
			return
		}
		loop = loop.Parent()
	}
	if loop == nil {
		return
	}
	var source *syntax.MethodCall
	var produced syntax.Expr
	while := loop.(*syntax.While)
	syntax.Inspect(while.Cond, func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok {
			if v, ok := a.Var.(*syntax.Variable); ok {
				if use, ok := assign.Value.(*syntax.Variable); ok && use.Name == v.Name {
					source, _ = a.Value.(*syntax.MethodCall)
					produced = a.Var
				}
			}
		}
		return true
	})
	if source == nil || !pdoFetchIntoMethod(ctx, source, "fetch") || !expansionDSame(ctx, produced, assign.Value) {
		return
	}
	if override := CallArgument(source.Args, 0, "mode"); override != nil {
		mode, known := NativeContractInt(ctx, override)
		if !known || mode != 0 {
			return
		}
	}
	latest := pdoFetchIntoMode(ctx, source, while)
	if latest == nil || pdoFetchIntoSingleRow(ctx, assign, while) {
		return
	}
	mode, known := NativeContractInt(ctx, CallArgument(latest.Args, 0, "mode"))
	object := CallArgument(latest.Args, 1, "args")
	if known && mode == 9 && object != nil && len(ctx.TypeOf(object).Classes()) > 0 {
		ctx.ReportNode(assign, message)
	}
}

func CheckPdoExecZeroRejectedAsFailure(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if NativeMethod(ctx, c, "PDO", "exec") && pdoExecRejectsZero(ctx, c) {
		ctx.ReportNode(c, message)
	}
}

// pdoExecRejectsZero reports whether the falsy branch of a PDO::exec result
// check throws or returns false without first consulting the connection's
// error state, which is how code tells zero affected rows from a failure.
func pdoExecRejectsZero(ctx *analysis.Context, c *syntax.MethodCall) bool {
	var e syntax.Expr = c
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
			e = p
		case *syntax.Assign:
			e = p
		case *syntax.Unary:
			if p.Op.Kind != syntax.TExclaim {
				return false
			}
			e = p
		case *syntax.Binary:
			if p.Op.Kind != syntax.TIsEqual || !expansionDFalse(p.Right) {
				return false
			}
			e = p
		case *syntax.If:
			neg, ok := syntax.UnwrapParens(e).(*syntax.Unary)
			loose, lok := syntax.UnwrapParens(e).(*syntax.Binary)
			if (!ok || neg.Op.Kind != syntax.TExclaim) && (!lok || loose.Op.Kind != syntax.TIsEqual) {
				return false
			}
			return pdoExecBranchRejects(ctx, p.Body, c.Var)
		default:
			return false
		}
	}
}

func pdoExecBranchRejects(ctx *analysis.Context, body syntax.Stmt, receiver syntax.Expr) bool {
	// Start offset of the first errorCode()/errorInfo() read on the receiver.
	consulted, seen := uint32(0), false
	var exits []syntax.Node
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch x := n.(type) {
		case *syntax.MethodCall:
			id, ok := x.Name.(*syntax.Identifier)
			if ok && !seen && (strings.EqualFold(id.Value, "errorCode") || strings.EqualFold(id.Value, "errorInfo")) && astquery.Equivalent(ctx.File, x.Var, receiver) {
				consulted, seen = x.Span().Start, true
			}
		case *syntax.Throw:
			if NativeCallbackReachable(x) {
				exits = append(exits, x)
			}
		case *syntax.Return:
			if expansionDFalse(x.Expr) && NativeCallbackReachable(x) {
				exits = append(exits, x)
			}
		}
		return true
	})
	for _, exit := range exits {
		if !seen || consulted > exit.Span().Start || pdoExecStraightLine(body, exit) {
			return true
		}
	}
	return false
}

// pdoExecStraightLine reports whether exit is a statement of the branch block
// preceded only by plain expression or output statements, so nothing in the
// branch decides whether it runs. A braceless branch needs no case: an earlier
// error-state read there always shares a construct with the exit.
func pdoExecStraightLine(body syntax.Stmt, exit syntax.Node) bool {
	stmt := exit
	if _, ok := exit.(*syntax.Throw); ok {
		stmt = exit.Parent()
	}
	block, ok := body.(*syntax.Block)
	if !ok || stmt.Parent() != syntax.Node(block) {
		return false
	}
	for _, s := range block.Stmts {
		if syntax.Node(s) == stmt {
			break
		}
		switch s.(type) {
		case *syntax.ExprStmt, *syntax.Echo:
		default:
			return false
		}
	}
	return true
}

func CheckMysqliEscapedValueSurvivesCharsetChange(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !NativeMethod(ctx, c, "mysqli", "query") {
		return
	}
	var escaped []*syntax.MethodCall
	sql := CallArgument(c.Args, 0, "query")
	if sql == nil {
		return
	}
	syntax.Inspect(sql, func(n syntax.Node) bool {
		e, ok := n.(syntax.Expr)
		if !ok {
			return true
		}
		origin, ok := NativeLocalValue(ctx, e).(*syntax.MethodCall)
		if ok && NativeMethod(ctx, origin, "mysqli", "real_escape_string") {
			escaped = append(escaped, origin)
		}
		return true
	})
	for _, escape := range escaped {
		if !expansionDSame(ctx, c.Var, escape.Var) {
			continue
		}
		// Only the last selection on each side decides: an earlier change that
		// a later call overrides, restores or leaves unproven is not evidence.
		prior := expansionDLastCharsetCall(ctx, escape, escape.Var, 0)
		if prior == nil || (!NativeDominates(prior, escape) && !expansionDMethodSucceeded(ctx, prior, escape)) {
			continue
		}
		before, known := NativeString(ctx, CallArgument(prior.Args, 0, "charset"))
		if !known {
			continue
		}
		change := expansionDLastCharsetCall(ctx, c, c.Var, escape.Span().End)
		if change == nil || !expansionDMethodSucceeded(ctx, change, c) {
			continue
		}
		after, known := NativeString(ctx, CallArgument(change.Args, 0, "charset"))
		if known && !strings.EqualFold(before, after) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// expansionDLastCharsetCall returns the last mysqli::set_charset call on the
// receiver written after offset from and before at, whether or not it is
// proven to execute; callers decide what that call proves.
func expansionDLastCharsetCall(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr, from uint32) *syntax.MethodCall {
	var last *syntax.MethodCall
	for _, event := range expansionDCalls(ctx, at) {
		c, ok := event.(*syntax.MethodCall)
		if ok && c.Span().Start > from && c.Span().End < at.Span().Start && NativeMethod(ctx, c, "mysqli", "set_charset") && expansionDSame(ctx, c.Var, receiver) {
			last = c
		}
	}
	return last
}

func CheckMysqliMultiQueryResultsNotDrained(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !NativeMethod(ctx, c, "mysqli", "query") {
		return
	}
	prior := expansionDMethodCalls(ctx, c, c.Var, "mysqli", "multi_query")
	if len(prior) == 0 {
		return
	}
	p := prior[len(prior)-1]
	if !expansionDMethodSucceeded(ctx, p, c) {
		return
	}
	sql, known := NativeString(ctx, CallArgument(p.Args, 0, "query"))
	if !known {
		return
	}
	tokens, known := expansionDSQL(sql)
	if !known {
		return
	}
	count := 0
	hasTokens := false
	for _, token := range tokens {
		if token == ";" {
			if hasTokens {
				count++
				hasTokens = false
			}
		} else {
			hasTokens = true
		}
	}
	if hasTokens {
		count++
	}
	if count < 2 {
		return
	}
	for _, event := range expansionDCalls(ctx, c) {
		check, ok := event.(*syntax.MethodCall)
		if !ok || check.Span().Start < p.Span().End || check.Span().End >= c.Span().Start || (!NativeMethod(ctx, check, "mysqli", "more_results") && !NativeMethod(ctx, check, "mysqli", "next_result")) || !expansionDSame(ctx, check.Var, c.Var) {
			continue
		}
		// The probe must decide the loop: alone or as an operand of `&&`.
		cond := syntax.Node(check)
		for {
			parent := cond.Parent()
			if b, ok := parent.(*syntax.Binary); ok && b.Op.Kind != syntax.TBooleanAnd {
				break
			}
			if _, ok := parent.(*syntax.Binary); !ok {
				if _, ok := parent.(*syntax.Paren); !ok {
					break
				}
			}
			cond = parent
		}
		var loop syntax.Node
		switch l := cond.Parent().(type) {
		case *syntax.While:
			if l.Cond == cond {
				loop = l
			}
		case *syntax.DoWhile:
			if l.Cond == cond {
				loop = l
			}
		}
		// A query inside the loop runs before the drain completes.
		if loop == nil || loop.Span().End > c.Span().Start {
			continue
		}
		drained := false
		syntax.Inspect(loop, func(n syntax.Node) bool {
			if next, ok := n.(*syntax.MethodCall); ok && NativeMethod(ctx, next, "mysqli", "next_result") && expansionDSame(ctx, next.Var, c.Var) {
				drained = true
			}
			return true
		})
		if drained {
			return
		}
	}
	ctx.ReportNode(c, message)
}

func CheckPgReturnedRowsUsedAsAffectedRows(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !NativeBuiltin(ctx, c, "pg_num_rows") {
		return
	}
	value := CallArgument(c.Args, 0, "result")
	query := expansionDOrigin(ctx, value, "pg_query")
	if query == nil || !NativeSentinelGuard(ctx, value, "false") {
		return
	}
	sql, known := NativeString(ctx, CallArgument(query.Args, 1, "query"))
	if !known {
		sql, known = NativeString(ctx, CallArgument(query.Args, 0, "query"))
	}
	if !known {
		return
	}
	tokens, known := expansionDSQL(sql)
	if !known || len(tokens) == 0 {
		return
	}
	switch tokens[0] {
	case "INSERT", "UPDATE", "DELETE":
	default:
		return
	}
	for _, token := range tokens {
		if token == "RETURNING" || token == ";" {
			return
		}
	}
	ctx.ReportNode(c, message)
}

func CheckSqliteNondeterministicFunctionDeclaredDeterministic(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if ctx.PHP < phpversion.PHP71 || !NativeMethod(ctx, c, "SQLite3", "createFunction") {
		return
	}
	flag, known := NativeContractInt(ctx, CallArgument(c.Args, 3, "flags"))
	if !known || flag&2048 == 0 {
		return
	}
	_, body := NativeCallback(ctx, CallArgument(c.Args, 1, "callback"))
	returns, complete := NativeReturns(body)
	if !complete {
		return
	}
	for _, ret := range returns {
		found := false
		if p, ok := syntax.UnwrapParens(ret).(*syntax.FuncCall); ok {
			switch NativeBuiltinName(ctx, p) {
			case "random_int":
				found = !expansionDSame(ctx, CallArgument(p.Args, 0, "min"), CallArgument(p.Args, 1, "max"))
			case "random_bytes", "time", "microtime":
				found = true
			}
		}
		if found {
			ctx.ReportNode(c, message)
			return
		}
	}
}

func expansionDTransaction(tokens []string, active bool) (bool, bool) {
	start := 0
	violation := false
	for i := 0; i <= len(tokens); i++ {
		if i < len(tokens) && tokens[i] != ";" {
			continue
		}
		part := tokens[start:i]
		start = i + 1
		if len(part) == 0 {
			continue
		}
		switch part[0] {
		case "BEGIN":
			active = true
		case "COMMIT", "END", "ROLLBACK":
			active = false
		case "PRAGMA":
			if len(part) >= 4 && part[1] == "FOREIGN_KEYS" && (part[2] == "=" || part[2] == "(") {
				violation = violation || active
			}
		}
	}
	return active, violation
}

func CheckSqliteForeignKeySettingInsideTransaction(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	class := "SQLite3"
	if !NativeMethod(ctx, c, class, "exec") {
		class = "PDO"
		if !NativeMethod(ctx, c, class, "exec") || !expansionDPdoDriver(ctx, c.Var, "sqlite") {
			return
		}
	}
	// active is true only while an open transaction is proven; every unknown
	// state is tracked as "not open" so it can never produce a report.
	active := false
	for _, event := range expansionDCalls(ctx, c) {
		prior, ok := event.(*syntax.MethodCall)
		if !ok || prior.Span().End >= c.Span().Start || !expansionDSame(ctx, prior.Var, c.Var) {
			continue
		}
		// A call on another branch may or may not have run before the PRAGMA.
		certain := NativeDominates(prior, c) || expansionDMethodSucceeded(ctx, prior, c)
		if !certain && sqliteCallLeaves(prior, c) {
			// The path through this call never reaches the PRAGMA.
			continue
		}
		name := ""
		if id, ok := prior.Name.(*syntax.Identifier); ok && NativeMethod(ctx, prior, class, id.Value) {
			name = strings.ToLower(id.Value)
		}
		switch name {
		case "":
			// User-defined or dynamic method: it may run any SQL.
			active = false
		case "begintransaction":
			active = certain
		case "commit", "rollback", "close", "open":
			active = false
		case "exec", "query", "querysingle", "prepare":
			tokens, known := sqliteStatementTokens(ctx, prior, class, name)
			if name == "prepare" {
				// A prepared statement runs later, at a point not tracked here.
				opened, _ := expansionDTransaction(tokens, false)
				kept, _ := expansionDTransaction(tokens, true)
				if !known || opened || !kept {
					return
				}
				continue
			}
			next, _ := expansionDTransaction(tokens, active)
			active = known && next && (certain || active)
		}
	}
	tokens, known := sqliteStatementTokens(ctx, c, class, "exec")
	if !known {
		return
	}
	_, bad := expansionDTransaction(tokens, active)
	if bad {
		ctx.ReportNode(c, message)
	}
}

// sqliteStatementTokens tokenizes the literal SQL passed to a SQLite3 or PDO
// statement method. Only exec() is known to run a whole script: how much of a
// multi-statement string the other methods run differs between the two
// extensions, so such a string is unknown for them.
func sqliteStatementTokens(ctx *analysis.Context, call *syntax.MethodCall, class, method string) ([]string, bool) {
	param := "query"
	if class == "PDO" && method == "exec" {
		param = "statement"
	}
	sql, known := NativeString(ctx, CallArgument(call.Args, 0, param))
	if !known {
		return nil, false
	}
	tokens, known := expansionDSQL(sql)
	if !known {
		return nil, false
	}
	if method != "exec" {
		for i, token := range tokens {
			if token == ";" && i+1 < len(tokens) {
				return nil, false
			}
		}
	}
	return tokens, true
}

// sqliteCallLeaves reports whether the statement holding prior is followed, in
// a block that does not contain at, by plain expression statements and then a
// return, throw or exit: control then leaves the function without reaching at.
// A try statement on the way out could resume execution, so it proves nothing.
func sqliteCallLeaves(prior, at syntax.Node) bool {
	left := false
	for node := prior; node != nil && !node.Span().Contains(at.Span()); node = node.Parent() {
		if _, ok := node.(*syntax.Try); ok || syntax.IsFuncLike(node) {
			return false
		}
		block, ok := node.Parent().(*syntax.Block)
		if _, plain := node.(*syntax.ExprStmt); left || !ok || !plain || block.Span().Contains(at.Span()) {
			continue
		}
		for _, st := range block.Stmts {
			if st.Span().Start < node.Span().End {
				continue
			}
			if _, ok := st.(*syntax.Return); ok {
				left = true
				break
			}
			next, ok := st.(*syntax.ExprStmt)
			if !ok {
				break
			}
			if _, throws := syntax.UnwrapParens(next.Expr).(*syntax.Throw); throws || syntax.ExitInvocation(next.Expr) {
				left = true
				break
			}
		}
	}
	return left
}
