// Package parsestrkeynormalizationmismatch implements the native ParseStrKeyNormalizationMismatch inspection.
package parsestrkeynormalizationmismatch

import (
	"net/url"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Read the normalized query parameter key."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ParseStrKeyNormalizationMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	fetch := n.(*syntax.ArrayDimFetch)
	if !astquery.StableArrayRead(ctx.File, fetch) {
		return
	}
	if boundByReference(fetch) {
		return
	}
	for p := fetch.Parent(); p != nil; p = p.Parent() {
		switch x := p.(type) {
		case *syntax.Isset, *syntax.Empty:
			return
		case *syntax.Binary:
			if x.Op.Kind == syntax.TCoalesce {
				return
			}
		}
		if _, ok := p.(syntax.Stmt); ok {
			break
		}
	}
	text, key, known := querySource(ctx, fetch, fetch)
	if !known || !strings.ContainsAny(key, ". ") {
		return
	}
	pairs, known := queryPairs(text)
	if !known {
		return
	}
	normalized := strings.NewReplacer(".", "_", " ", "_").Replace(key)
	if _, exists := pairs[normalized]; exists {
		if _, literal := fetch.Dim.(*syntax.Literal); literal {
			quoted := "'" + strings.ReplaceAll(strings.ReplaceAll(normalized, "\\", "\\\\"), "'", "\\'") + "'"
			ctx.ReportNode(fetch, message, astquery.ReplaceFix(fetch.Dim.Span(), quoted))
		} else {
			ctx.ReportNode(fetch, message)
		}
	}
}

// boundByReference reports whether the fetch, or a deeper dimension reached
// through it, is bound by reference: the value of a reference assignment, a
// by-reference array item or the source of a by-reference foreach. Such a
// binding creates the dimension instead of reading it. A fetch that is only
// the key of another access (`$x = &$cache[$q["a.b"]]`) is still a read.
func boundByReference(fetch *syntax.ArrayDimFetch) bool {
	node := syntax.Node(fetch)
	for {
		outer, ok := node.Parent().(*syntax.ArrayDimFetch)
		if !ok || outer.Var != node {
			break
		}
		node = outer
	}
	switch x := node.Parent().(type) {
	case *syntax.Assign:
		return x.ByRef && x.Value == node
	case *syntax.ArrayItem:
		return x.ByRef && x.Value == node
	case *syntax.Foreach:
		return x.ByRef && x.Expr == node
	}
	return false
}

func querySource(ctx *analysis.Context, at syntax.Node, fetch *syntax.ArrayDimFetch) (string, string, bool) {
	variable, ok := fetch.Var.(*syntax.Variable)
	if !ok {
		return "", "", false
	}
	key, known := semanticquery.NativeString(ctx, fetch.Dim)
	if !known {
		return "", "", false
	}
	prior := flowquery.NativePriorStatements(ctx.File, at)
	if len(prior) == 0 {
		return "", "", false
	}
	statement, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	// A nested closure or arrow function owns its own variables: the name may
	// be a parameter or an unrelated local there.
	if !ok || syntax.EnclosingVariableScope(statement) != syntax.EnclosingVariableScope(fetch) {
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
