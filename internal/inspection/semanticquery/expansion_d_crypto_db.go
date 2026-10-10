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
	events := nativeContractEvents(ctx, syntax.EnclosingVariableScope(a))
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
	for _, event := range events {
		if event.Span().Start <= start || event.Span().End >= end {
			continue
		}
		switch e := event.(type) {
		case *syntax.Assign:
			if astquery.MentionsVariable(e.Var, av.Name) || (e.ByRef && astquery.MentionsVariable(e.Value, av.Name)) {
				return false
			}
		case *syntax.Unset:
			if astquery.MentionsVariable(e, av.Name) {
				return false
			}
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

func expansionDFailureBranch(ctx *analysis.Context, e syntax.Expr) *syntax.If {
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
			e = p
		case *syntax.Assign:
			e = p
		case *syntax.Unary:
			if p.Op.Kind != syntax.TExclaim {
				return nil
			}
			e = p
		case *syntax.Binary:
			if p.Op.Kind != syntax.TIsEqual || !expansionDFalse(p.Right) {
				return nil
			}
			e = p
		case *syntax.If:
			neg, ok := syntax.UnwrapParens(e).(*syntax.Unary)
			loose, lok := syntax.UnwrapParens(e).(*syntax.Binary)
			if (!ok || neg.Op.Kind != syntax.TExclaim) && (!lok || loose.Op.Kind != syntax.TIsEqual) {
				return nil
			}
			found := false
			syntax.Inspect(p.Body, func(n syntax.Node) bool {
				if syntax.IsVariableScope(n) {
					return false
				}
				if t, ok := n.(*syntax.Throw); ok && NativeCallbackReachable(t) {
					found = true
				}
				if r, ok := n.(*syntax.Return); ok && expansionDFalse(r.Expr) && NativeCallbackReachable(r) {
					found = true
				}
				return true
			})
			if found {
				return p
			}
			return nil
		default:
			return nil
		}
	}
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
		// Any write to the requested-size parameter invalidates the contract proof.
		changed := false
		syntax.Inspect(body, func(event syntax.Node) bool {
			if event.Span().Start >= r.Span().Start || syntax.IsVariableScope(event) && event != body {
				return false
			}
			switch a := event.(type) {
			case *syntax.Assign:
				changed = changed || astquery.MentionsVariable(a.Var, size.Name)
			case *syntax.IncDec:
				changed = changed || astquery.MentionsVariable(a.Var, size.Name)
			}
			return true
		})
		if changed {
			return
		}
	}
	ctx.ReportNode(c, message)
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
	c := n.(*syntax.FuncCall)
	if !NativeBuiltin(ctx, c, "curl_exec") || expansionDFailureBranch(ctx, c) == nil {
		return
	}
	enabled, known := NativeTruth(ctx, expansionDCurlOption(ctx, c, CallArgument(c.Args, 0, "handle"), "CURLOPT_RETURNTRANSFER"))
	if known && enabled {
		ctx.ReportNode(c, message)
	}
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
	branch := expansionDFailureBranch(ctx, c)
	if branch == nil {
		return
	}
	inspected := false
	syntax.Inspect(branch.Body, func(n syntax.Node) bool {
		if f, ok := n.(*syntax.FuncCall); ok && NativeBuiltin(ctx, f, "socket_last_error") {
			inspected = true
		}
		return true
	})
	if inspected {
		return
	}
	var latest *syntax.FuncCall
	for _, event := range expansionDCalls(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if !ok || p.Span().End >= c.Span().Start || !expansionDSame(ctx, CallArgument(p.Args, 0, "socket"), CallArgument(c.Args, 0, "socket")) {
			continue
		}
		if NativeBuiltin(ctx, p, "socket_set_block") || NativeBuiltin(ctx, p, "socket_set_nonblock") {
			latest = p
		}
	}
	if latest != nil && NativeBuiltin(ctx, latest, "socket_set_nonblock") && NativeCallSucceeded(ctx, latest, c) {
		ctx.ReportNode(c, message)
	}
}

func CheckSocketReadZeroRejected(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if NativeBuiltin(ctx, c, "socket_read") && expansionDConst(ctx, CallArgument(c.Args, 2, "mode"), "PHP_BINARY_READ") && expansionDFailureBranch(ctx, c) != nil {
		ctx.ReportNode(c, message)
	}
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
	found := false
	syntax.Inspect(branch.Body, func(n syntax.Node) bool {
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
	if found {
		ctx.ReportNode(c, message)
	}
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
	e := NativeConditionUse(ctx, c, true)
	if e == nil {
		return
	}
	branch, ok := e.Parent().(*syntax.If)
	if !ok {
		return
	}
	found := false
	syntax.Inspect(branch.Body, func(n syntax.Node) bool {
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

func CheckPdoBoundValueAssumedLive(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !NativeMethod(ctx, c, "PDOStatement", "execute") || CallArgument(c.Args, 0, "params") != nil {
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
			if expansionDSame(ctx, CallArgument(binding.Args, 0, "param"), CallArgument(later.Args, 0, "param")) {
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
	if a := NativeArray(ctx, CallArgument(construct.Args, 3, "options")); a != nil {
		for _, item := range a.Items {
			if cc, ok := item.Key.(*syntax.ClassConstFetch); ok && ExpansionCConstant(ctx, cc, "PDO", "ATTR_EMULATE_PREPARES") {
				emulation = item.Value
			}
		}
	}
	for _, prior := range expansionDMethodCalls(ctx, c, c.Var, "PDO", "setAttribute") {
		if cc, ok := CallArgument(prior.Args, 0, "attribute").(*syntax.ClassConstFetch); ok && ExpansionCConstant(ctx, cc, "PDO", "ATTR_EMULATE_PREPARES") {
			emulation = CallArgument(prior.Args, 1, "value")
		}
	}
	enabled, known := NativeTruth(ctx, emulation)
	if !known || enabled {
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
	if !ExpansionDPristine(ctx, group.Var, fetch) {
		return
	}
	all, ok := NativeLocalValue(ctx, group.Var).(*syntax.MethodCall)
	if !ok || !NativeMethod(ctx, all, "PDOStatement", "fetchAll") {
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
	statement, ok := NativeLocalValue(ctx, all.Var).(*syntax.MethodCall)
	if !ok || (!NativeMethod(ctx, statement, "PDO", "query") && !NativeMethod(ctx, statement, "PDO", "prepare")) {
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
	if source == nil || !NativeMethod(ctx, source, "PDOStatement", "fetch") || !expansionDSame(ctx, produced, assign.Value) {
		return
	}
	if override := CallArgument(source.Args, 0, "mode"); override != nil {
		mode, known := NativeContractInt(ctx, override)
		if !known || mode != 0 {
			return
		}
	}
	modes := expansionDMethodCalls(ctx, source, source.Var, "PDOStatement", "setFetchMode")
	if len(modes) == 0 {
		return
	}
	latest := modes[len(modes)-1]
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
	if NativeMethod(ctx, c, "PDO", "exec") && expansionDFailureBranch(ctx, c) != nil {
		ctx.ReportNode(c, message)
	}
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
		prior := expansionDMethodCalls(ctx, escape, escape.Var, "mysqli", "set_charset")
		if len(prior) == 0 {
			continue
		}
		before, known := NativeString(ctx, CallArgument(prior[len(prior)-1].Args, 0, "charset"))
		if !known {
			continue
		}
		changes := expansionDMethodCalls(ctx, c, c.Var, "mysqli", "set_charset")
		for _, change := range changes {
			if change.Span().Start <= escape.Span().End || !expansionDMethodSucceeded(ctx, change, c) {
				continue
			}
			after, known := NativeString(ctx, CallArgument(change.Args, 0, "charset"))
			if known && !strings.EqualFold(before, after) {
				ctx.ReportNode(c, message)
				return
			}
		}
	}
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
		if !ok || check.Span().Start < p.Span().End || check.Span().End >= c.Span().Start || !NativeMethod(ctx, check, "mysqli", "more_results") || !expansionDSame(ctx, check.Var, c.Var) {
			continue
		}
		loop, ok := check.Parent().(*syntax.While)
		if !ok || loop.Cond != check {
			continue
		}
		drained := false
		syntax.Inspect(loop.Body, func(n syntax.Node) bool {
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
	active := false
	for _, prior := range expansionDMethodCalls(ctx, c, c.Var, class, "exec", "beginTransaction", "commit", "rollBack") {
		if NativeMethod(ctx, prior, class, "beginTransaction") {
			active = true
			continue
		}
		if NativeMethod(ctx, prior, class, "commit") || NativeMethod(ctx, prior, class, "rollBack") {
			active = false
			continue
		}
		sql, known := NativeString(ctx, CallArgument(prior.Args, 0, "statement"))
		if !known {
			return
		}
		tokens, known := expansionDSQL(sql)
		if !known {
			return
		}
		active, _ = expansionDTransaction(tokens, active)
	}
	sql, known := NativeString(ctx, CallArgument(c.Args, 0, "statement"))
	if !known {
		return
	}
	tokens, known := expansionDSQL(sql)
	if !known {
		return
	}
	_, bad := expansionDTransaction(tokens, active)
	if bad {
		ctx.ReportNode(c, message)
	}
}
