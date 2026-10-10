// Package bcryptpasswordtruncation implements the native BcryptPasswordTruncation inspection.
package bcryptpasswordtruncation

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Keep bcrypt input within its supported byte length."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "BcryptPasswordTruncation" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "password_hash") {
		return
	}
	algorithm, ok := syntax.UnwrapParens(semanticquery.CallArgument(call.Args, 1, "algo")).(*syntax.ConstFetch)
	if !ok || semanticquery.GlobalConstName(ctx, algorithm) != "PASSWORD_BCRYPT" {
		return
	}
	password := semanticquery.CallArgument(call.Args, 0, "password")
	size, known := length(ctx, password)
	if known && size > 72 {
		ctx.ReportNode(password, message)
	}
}

func length(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if s, ok := semanticquery.NativeString(ctx, e); ok {
		return int64(len(s)), true
	}
	call, ok := semanticquery.NativeValue(ctx, e).(*syntax.FuncCall)
	if !ok {
		return 0, false
	}
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "random_bytes":
		n, ok := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 0, "length"))
		return n, ok && n > 0
	case "str_repeat":
		s, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "string"))
		n, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 1, "times"))
		if ok && known && n >= 0 && n <= 1048576 {
			return int64(len(s)) * n, true
		}
	}
	return 0, false
}
