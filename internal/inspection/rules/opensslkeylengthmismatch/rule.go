// Package opensslkeylengthmismatch implements the native OpenSslKeyLengthMismatch inspection.
package opensslkeylengthmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Provide the cipher-required key length."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OpenSslKeyLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, call)
	if name != "openssl_encrypt" && name != "openssl_decrypt" {
		return
	}
	cipher, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "cipher_algo"))
	if !known {
		return
	}
	parts := strings.Split(strings.ToLower(cipher), "-")
	if len(parts) != 3 || parts[0] != "aes" {
		return
	}
	expected := int64(0)
	switch parts[1] {
	case "128":
		expected = 16
	case "192":
		expected = 24
	case "256":
		expected = 32
	default:
		return
	}
	switch parts[2] {
	case "cbc", "gcm", "ccm", "ctr", "ecb", "cfb", "cfb1", "cfb8", "ofb":
	default:
		return
	}
	key := semanticquery.CallArgument(call.Args, 2, "passphrase")
	size, ok := length(ctx, key)
	if ok && size != expected && !replaceable(ctx, call, key) {
		ctx.ReportNode(key, message)
	}
}

// replaceable reports whether a variable key can receive another value
// through a path local flow does not model: a by-reference closure capture or
// array item, an extract() in its scope, or — for a variable that
// shares storage with other code — any later call that may run user code,
// or $GLOBALS access.
func replaceable(ctx *analysis.Context, call *syntax.FuncCall, key syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(key).(*syntax.Variable)
	if !ok {
		return false
	}
	value, _ := ctx.Flow().Resolve(key)
	proven := value.Span().End
	scope := syntax.EnclosingVariableScope(key)
	same := func(e syntax.Expr) bool {
		v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
		return ok && v.Name == variable.Name
	}
	// File-level variables, by-reference parameters and by-reference captures
	// of the enclosing closure share storage with code outside the scope.
	shared, found := scope == nil, false
	for _, param := range syntax.VariableScopeParams(scope) {
		shared = shared || param.ByRef && same(param.Var)
	}
	if closure, ok := scope.(*syntax.Closure); ok {
		for _, use := range closure.Uses {
			shared = shared || use.ByRef && same(use.Var)
		}
	}
	late := func(n syntax.Node) bool { return n != call && n.Span().End > proven }
	visit := func(n syntax.Node) bool {
		if found || n != scope && syntax.IsVariableScope(n) {
			if closure, ok := n.(*syntax.Closure); ok {
				for _, use := range closure.Uses {
					found = found || use.ByRef && same(use.Var)
				}
			}
			return false
		}
		switch n := n.(type) {
		case *syntax.Global:
			for _, v := range n.Vars {
				shared = shared || same(v)
				found = found || same(v) && late(n)
			}
		case *syntax.ArrayItem:
			found = n.ByRef && same(n.Value)
		case *syntax.FuncCall:
			found = semanticquery.NativeBuiltinName(ctx, n) == "extract" || shared && late(n) && !inert(ctx, n)
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
			found = shared && late(n)
		case *syntax.Variable:
			found = shared && n.Name == "GLOBALS"
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}

// inert reports whether n calls a builtin that runs no user code: one whose
// signature has no callable parameter. Such a call cannot replace a variable
// shared with other code.
func inert(ctx *analysis.Context, n *syntax.FuncCall) bool {
	if semanticquery.NativeBuiltinName(ctx, n) == "" {
		return false
	}
	for _, param := range ctx.Types().ResolveFunction(n).Params {
		if strings.Contains(strings.ToLower(param.Type), "callable") {
			return false
		}
	}
	return true
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
