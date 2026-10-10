// Package parsedqueryvaluedecodedtwice implements the native ParsedQueryValueDecodedTwice inspection.
package parsedqueryvaluedecodedtwice

import (
	"net/url"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the query value without decoding it a second time."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ParsedQueryValueDecodedTwice" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "urldecode" && name != "rawurldecode" {
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "string")
	text, key, known := querySource(ctx, c, input)
	if !known {
		return
	}
	pairs, known := queryPairs(text)
	if !known {
		return
	}
	original, known := pairs[key]
	if !known {
		return
	}
	var decoded string
	var err error
	if name == "urldecode" {
		decoded, err = url.QueryUnescape(original)
	} else {
		decoded, err = url.PathUnescape(original)
	}
	if err == nil && decoded != original {
		ctx.ReportNode(c, message)
	}
}

func querySource(ctx *analysis.Context, at syntax.Node, input syntax.Expr) (string, string, bool) {
	fetch, ok := input.(*syntax.ArrayDimFetch)
	if !ok {
		return "", "", false
	}
	variable, ok := fetch.Var.(*syntax.Variable)
	if !ok {
		return "", "", false
	}
	key, known := semanticquery.NativeString(ctx, fetch.Dim)
	if !known {
		return "", "", false
	}
	switch parent := at.Parent().(type) {
	case *syntax.ExprStmt, *syntax.Return:
	case *syntax.Echo:
		if len(parent.Exprs) != 1 {
			return "", "", false
		}
	case *syntax.Assign:
		if _, ok := parent.Parent().(*syntax.ExprStmt); !ok {
			return "", "", false
		}
	default:
		return "", "", false
	}
	prior := flowquery.NativePriorStatements(ctx.File, at)
	if len(prior) == 0 {
		return "", "", false
	}
	statement, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return "", "", false
	}
	parse, ok := statement.Expr.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, parse, "parse_str") {
		return "", "", false
	}
	output, ok := semanticquery.CallArgument(parse.Args, 1, "result").(*syntax.Variable)
	if !ok || output.Name != variable.Name {
		return "", "", false
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(parse.Args, 0, "string"))
	return text, key, known && len(text) <= 32768
}

func queryPairs(text string) (map[string]string, bool) {
	pairs := map[string]string{}
	for _, part := range strings.Split(text, "&") {
		pieces := strings.SplitN(part, "=", 2)
		if len(pieces) != 2 {
			return nil, false
		}
		key, err := url.QueryUnescape(pieces[0])
		if err != nil || strings.ContainsAny(key, "[]\x00") || strings.TrimLeft(key, " \t\r\n") != key {
			return nil, false
		}
		normalized := strings.NewReplacer(".", "_", " ", "_").Replace(key)
		if _, duplicate := pairs[normalized]; duplicate {
			return nil, false
		}
		value, err := url.QueryUnescape(pieces[1])
		if err != nil {
			return nil, false
		}
		pairs[normalized] = value
	}
	return pairs, true
}
