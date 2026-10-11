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

// CheckUntrustedFilesystemPath checks the independently specified contract.
func CheckUntrustedFilesystemPath(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name := NativeBuiltinName(ctx, c)
	report := false

	switch name {
	case "readfile", "fopen", "file_get_contents", "file_put_contents", "unlink":
		report = ctx.Flow().Tainted(CallArgument(c.Args, 0, "filename"), flow.Path)
	case "rmdir":
		report = ctx.Flow().Tainted(CallArgument(c.Args, 0, "directory"), flow.Path)
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
