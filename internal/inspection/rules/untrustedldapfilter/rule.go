// Package untrustedldapfilter implements the native UntrustedLdapFilter inspection.
package untrustedldapfilter

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape request data for LDAP filter syntax."

// escapeFilter is the value of LDAP_ESCAPE_FILTER; LDAP_ESCAPE_DN alone (2)
// leaves the filter metacharacters `*`, `(` and `)` untouched.
const escapeFilter = 1

// serverGenerated lists $_SERVER entries the SAPI or the server configuration
// produces; a client cannot place filter syntax in them.
var serverGenerated = map[string]bool{
	"REQUEST_TIME":       true,
	"REQUEST_TIME_FLOAT": true,
	"SERVER_ADDR":        true,
	"SERVER_PORT":        true,
	"REMOTE_PORT":        true,
	"SERVER_SOFTWARE":    true,
	"GATEWAY_INTERFACE":  true,
	"SERVER_PROTOCOL":    true,
	"DOCUMENT_ROOT":      true,
	"SCRIPT_FILENAME":    true,
	"argc":               true,
}

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedLdapFilter" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "ldap_search", "ldap_list", "ldap_read":
	default:
		return
	}
	filter := semanticquery.CallArgument(call.Args, 2, "filter")
	if untrusted(ctx, filter, 0) {
		ctx.ReportNode(filter, message)
	}
}

func request(ctx *analysis.Context, x *syntax.ArrayDimFetch) bool {
	v, ok := x.Var.(*syntax.Variable)
	if !ok {
		return false
	}
	switch v.Name {
	case "_GET", "_POST", "_REQUEST", "_COOKIE", "_SERVER":
	default:
		return false
	}
	if !semanticquery.NativeRequestUnwritten(ctx, x) {
		return false
	}
	key, known := key(ctx, x.Dim)
	return known && (v.Name != "_SERVER" || !serverGenerated[key])
}

// key returns a statically known offset: a quoted string, another
// literal, or the bare word of simple interpolation (`"$_GET[user]"`).
func key(ctx *analysis.Context, dim syntax.Expr) (string, bool) {
	if s, ok := semanticquery.NativeString(ctx, dim); ok {
		return s, true
	}
	if l, ok := dim.(*syntax.Literal); ok {
		return l.Raw, true
	}
	return "", false
}

func untrusted(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	if depth > 16 {
		return false
	}
	switch x := semanticquery.NativeValue(ctx, e).(type) {
	case *syntax.ArrayDimFetch:
		return request(ctx, x)
	case *syntax.Binary:
		return x.Op.Kind == syntax.TDot && (untrusted(ctx, x.Left, depth+1) || untrusted(ctx, x.Right, depth+1))
	case *syntax.InterpolatedString:
		// A backtick command sends its parts to the shell, not to the filter.
		for _, part := range x.Parts {
			if !x.Backtick && untrusted(ctx, part, depth+1) {
				return true
			}
		}
	case *syntax.FuncCall:
		// ldap_escape protects a filter only when its flags are omitted or
		// zero (everything is escaped) or include LDAP_ESCAPE_FILTER.
		if !semanticquery.NativeBuiltin(ctx, x, "ldap_escape") {
			return false
		}
		flags := semanticquery.CallArgument(x.Args, 2, "flags")
		if flags == nil {
			return false
		}
		if mask, known := semanticquery.NativeInt(ctx, flags); known && mask == 0 {
			return false
		}
		if set, proven := semanticquery.NativeFlag(ctx, flags, escapeFilter); !proven || set {
			return false
		}
		return untrusted(ctx, semanticquery.CallArgument(x.Args, 0, "value"), depth+1)
	}
	return false
}
