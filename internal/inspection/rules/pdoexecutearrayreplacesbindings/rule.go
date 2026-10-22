// Package pdoexecutearrayreplacesbindings implements the native PdoExecuteArrayReplacesBindings inspection.
package pdoexecutearrayreplacesbindings

import (
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Omit the empty parameter array to retain existing bindings."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoExecuteArrayReplacesBindings" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	var fixes []diagnostic.Fix
	if !semanticquery.NativeMethod(ctx, c, "PDOStatement", "execute") {
		return
	}
	facts, known := semanticquery.PreparedSQL(ctx, c.Var)
	if !known || len(facts.Markers) == 0 {
		return
	}

	array := semanticquery.NativeArray(ctx, semanticquery.CallArgument(c.Args, 0, "params"))
	if array == nil || len(array.Items) != 0 || !pdoPriorBindingsRequired(ctx, c, facts) {
		return
	}
	if len(c.Args.Args) == 1 && semanticquery.ArgumentsRemovable(&syntax.FuncCall{Args: c.Args}) && semanticquery.CommentFreeSpan(ctx, c.Args.Span()) {
		fixes = append(fixes, semanticquery.ReplacementFix(c.Args.Span(), "()", "Retain previous bindings"))
	}
	ctx.ReportNode(c, message, fixes...)
}

// pdoPriorBindingsRequired proves that the statement has earlier
// bindValue/bindParam calls and that each one names a marker of the prepared
// SQL. A binding for an absent marker makes a bare execute() fail, so the empty
// array that discards it is not a mistake.
func pdoPriorBindingsRequired(ctx *analysis.Context, c *syntax.MethodCall, facts semanticquery.SQLFacts) bool {
	prior, historyKnown := semanticquery.NativeCallHistory(ctx, c, c.Var, "bindparam", "bindvalue").Certain()
	if !historyKnown {
		return false
	}
	for _, node := range prior {
		// Anything but a method call with a proven marker leaves the key empty,
		// which no marker matches.
		key := ""
		if binding, ok := node.(*syntax.MethodCall); ok {
			value := semanticquery.CallArgument(binding.Args, 0, "param")
			if name, known := semanticquery.NativeString(ctx, value); known {
				key = strings.TrimPrefix(name, ":")
			} else if position, known := semanticquery.NativeInt(ctx, value); known && position > 0 {
				key = strconv.FormatInt(position-1, 10)
			}
		}
		if !facts.Markers[key] {
			return false
		}
	}
	return len(prior) > 0
}
