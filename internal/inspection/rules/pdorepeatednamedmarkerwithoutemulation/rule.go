// Package pdorepeatednamedmarkerwithoutemulation implements PdoRepeatedNamedMarkerWithoutEmulation.
package pdorepeatednamedmarkerwithoutemulation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a distinct marker for each native parameter."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoRepeatedNamedMarkerWithoutEmulation" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDO", "prepare") || !semanticquery.PDODriver(ctx, c.Var, "mysql") {
		return
	}
	var emulation syntax.Expr
	construct := semanticquery.NativeConstruction(ctx, c.Var, "PDO")
	proven := construct.Span().End
	if a := semanticquery.NativeArray(ctx, semanticquery.CallArgument(construct.Args, 3, "options")); a != nil {
		for _, item := range a.Items {
			if cc, ok := item.Key.(*syntax.ClassConstFetch); ok && semanticquery.ExpansionCConstant(ctx, cc, "PDO", "ATTR_EMULATE_PREPARES") {
				emulation = item.Value
			}
		}
	}
	for _, prior := range semanticquery.EarlierMethodCalls(ctx, c, c.Var, "PDO", "setAttribute") {
		if cc, ok := semanticquery.CallArgument(prior.Args, 0, "attribute").(*syntax.ClassConstFetch); ok && semanticquery.ExpansionCConstant(ctx, cc, "PDO", "ATTR_EMULATE_PREPARES") {
			emulation, proven = semanticquery.CallArgument(prior.Args, 1, "value"), prior.Span().End
		}
	}
	enabled, known := semanticquery.NativeTruth(ctx, emulation)
	if !known || enabled || pdoEmulationMayChange(ctx, c, proven) {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "query"))
	if !known {
		return
	}
	tokens, known := semanticquery.SQLTokens(sql)
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
		if cc, ok := semanticquery.CallArgument(m.Args, 0, "attribute").(*syntax.ClassConstFetch); ok {
			if name, ok := cc.Name.(*syntax.Identifier); ok && name.Value != "ATTR_EMULATE_PREPARES" && !strings.HasSuffix(name.Value, "ATTR_DIRECT_QUERY") {
				return true
			}
			if enabled, known := semanticquery.NativeTruth(ctx, semanticquery.CallArgument(m.Args, 1, "value")); known && !enabled {
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
