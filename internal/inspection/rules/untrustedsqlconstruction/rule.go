// Package untrustedsqlconstruction implements the native UntrustedSqlConstruction inspection.
package untrustedsqlconstruction

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

const message = "Bind untrusted values as SQL parameters."

// maxDepth bounds the structural walk through definitions and operands.
const maxDepth = 16

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedSqlConstruction" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall, syntax.KFuncCall} }

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if method, ok := n.(*syntax.MethodCall); ok {
		if pdoSink(ctx, method) {
			sql := semanticquery.CallArgument(method.Args, 0, "query")
			if sql == nil {
				sql = semanticquery.CallArgument(method.Args, 0, "statement")
			}
			if sql != nil && tainted(ctx, sql, maxDepth) {
				ctx.ReportNode(method, message)
			}
		}
		return
	}
	if wrapperSink(ctx, n) {
		ctx.ReportNode(n, message)
	}
}

// pdoSink proves a call to the builtin PDO::query, PDO::exec or PDO::prepare,
// including inherited ones. All three exist on every supported PHP version, so
// a stub entry that only describes a later signature still identifies the sink.
func pdoSink(ctx *analysis.Context, c *syntax.MethodCall) bool {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok {
		return false
	}
	switch strings.ToLower(id.Value) {
	case "query", "exec", "prepare":
	default:
		return false
	}
	classes := ctx.Types().Native().TypeOf(c.Var).Classes()
	if len(classes) != 1 || !ctx.Index().IsSubtype(classes[0], "PDO", ctx.PHP) {
		return false
	}
	m := ctx.Index().FindMethod(classes[0], id.Value, ctx.PHP)
	if m == nil {
		m = ctx.Index().FindMethod(classes[0], id.Value, 0)
	}
	return m != nil && m.Builtin
}

// wrapperSink proves an external source reaches a SQL sink inside a project wrapper.
func wrapperSink(ctx *analysis.Context, call syntax.Node) bool {
	for _, sink := range ctx.Flow().Sinks(call) {
		if sink.Context != flow.SQL || !sink.Argument.Complete || sink.Argument.Safe&flow.SQL != 0 {
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

// tainted refines the flow verdict structurally: text is injectable only when
// some part contributing characters to it is. Unrecognized shapes keep the
// flow verdict, so the walk can only remove reports for proven-safe parts.
func tainted(ctx *analysis.Context, x syntax.Expr, depth int) bool {
	x = syntax.UnwrapParens(x)
	if !ctx.Flow().Tainted(x, flow.SQL) {
		return false
	}
	if depth == 0 {
		return true
	}
	depth--
	switch n := x.(type) {
	case *syntax.Variable:
		if value, known := ctx.Flow().Resolve(n); known && value != x && !aliased(ctx, n) {
			return tainted(ctx, value, depth)
		}
	case *syntax.InterpolatedString:
		for _, part := range n.Parts {
			if tainted(ctx, part, depth) {
				return true
			}
		}
		return false
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TDot, syntax.TCoalesce:
			return tainted(ctx, n.Left, depth) || tainted(ctx, n.Right, depth)
		case syntax.TPlus:
			// Two arrays form a union that keeps request text; one number does not.
			return !number(ctx, n.Left) && !number(ctx, n.Right)
		}
		return !number(ctx, x)
	case *syntax.Unary:
		return !number(ctx, x)
	case *syntax.Ternary:
		then := n.Then
		if then == nil {
			then = n.Cond
		}
		return tainted(ctx, then, depth) || tainted(ctx, n.Else, depth)
	case *syntax.Match:
		// The subject only selects an arm; the result is one of the arm bodies.
		for _, arm := range n.Arms {
			if tainted(ctx, arm.Body, depth) {
				return true
			}
		}
		return false
	case *syntax.Array:
		for _, item := range n.Items {
			if item == nil || item.Unpack || item.ByRef || tainted(ctx, item.Value, depth) {
				return true
			}
		}
		return false
	case *syntax.FuncCall:
		return taintedCall(ctx, n, depth)
	}
	return true
}

// taintedCall handles builtins whose result text comes from known arguments.
func taintedCall(ctx *analysis.Context, c *syntax.FuncCall, depth int) bool {
	name := semanticquery.NativeBuiltinName(ctx, c)
	if numberBuiltins[name] || scalarBuiltins[name] {
		return false
	}
	var parts []syntax.Expr
	switch name {
	case "implode", "join":
		parts = []syntax.Expr{semanticquery.CallArgument(c.Args, 0, "separator"), semanticquery.CallArgument(c.Args, 1, "array")}
	case "array_fill":
		parts = []syntax.Expr{semanticquery.CallArgument(c.Args, 2, "value")}
	case "str_repeat":
		parts = []syntax.Expr{semanticquery.CallArgument(c.Args, 0, "string")}
	default:
		return true
	}
	for _, part := range parts {
		if part == nil || tainted(ctx, part, depth) {
			return true
		}
	}
	return false
}

// number proves an expression yields an int or float, or throws, whatever its
// operands hold: numeric literals, arithmetic and counting builtins.
func number(ctx *analysis.Context, x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal:
		return n.LitKind == syntax.LitInt || n.LitKind == syntax.LitFloat
	case *syntax.Unary:
		switch n.Op.Kind {
		case syntax.TPlus, syntax.TMinus, syntax.TIntCast, syntax.TDoubleCast:
			return true
		}
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TMinus, syntax.TMul, syntax.TDiv, syntax.TMod, syntax.TPow:
			return true
		case syntax.TPlus:
			return number(ctx, n.Left) || number(ctx, n.Right)
		}
	case *syntax.FuncCall:
		return numberBuiltins[semanticquery.NativeBuiltinName(ctx, n)]
	}
	return false
}

// numberBuiltins return an int or float, never false, text or an array.
var numberBuiltins = map[string]bool{
	"count": true, "sizeof": true, "strlen": true, "mb_strlen": true, "iconv_strlen": true,
	"substr_count": true, "mb_substr_count": true, "strcmp": true, "strcasecmp": true,
	"strncmp": true, "strncasecmp": true, "strnatcmp": true, "strnatcasecmp": true,
	"ord": true, "crc32": true, "levenshtein": true, "similar_text": true,
	"strspn": true, "strcspn": true, "intdiv": true, "array_sum": true, "array_product": true,
}

// scalarBuiltins may also return a boolean: no argument text survives either.
var scalarBuiltins = map[string]bool{
	"strpos": true, "stripos": true, "strrpos": true, "strripos": true,
	"mb_strpos": true, "mb_stripos": true, "mb_strrpos": true, "mb_strripos": true,
	"strtotime": true, "in_array": true, "array_key_exists": true, "key_exists": true,
	"str_contains": true, "str_starts_with": true, "str_ends_with": true,
	"is_numeric": true, "is_int": true, "is_float": true, "is_string": true, "is_array": true,
	"is_bool": true, "is_null": true, "is_scalar": true,
	"ctype_digit": true, "ctype_alnum": true, "ctype_alpha": true, "ctype_xdigit": true,
	"ctype_lower": true, "ctype_upper": true, "ctype_space": true, "ctype_punct": true,
}
