// Package priorityqueueextractionshapemismatch implements the native PriorityQueueExtractionShapeMismatch inspection.
package priorityqueueextractionshapemismatch

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the priority queue extraction mode."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PriorityQueueExtractionShapeMismatch" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	f := n.(*syntax.ArrayDimFetch)
	key, ok := semanticquery.NativeString(ctx, f.Dim)
	if !ok || (key != "data" && key != "priority") {
		return
	}
	call, ok := semanticquery.NativeValue(ctx, f.Var).(*syntax.MethodCall)
	if !ok {
		return
	}
	id, ok := call.Name.(*syntax.Identifier)
	if !ok || (!strings.EqualFold(id.Value, "extract") && !strings.EqualFold(id.Value, "current") && !strings.EqualFold(id.Value, "top")) || !semanticquery.NativeMethod(ctx, call, "SplPriorityQueue", id.Value) || semanticquery.NativeConstruction(ctx, call.Var, "SplPriorityQueue") == nil {
		return
	}
	mode := int64(1)
	inserted := false
	priorityScalar := true
	proven := semanticquery.NativePriorCalls(ctx, call, call.Var, "setExtractFlags", "insert")
	for i, n := range append(proven, possibleCalls(ctx, call, proven)...) {
		c := n.(*syntax.MethodCall)
		name := c.Name.(*syntax.Identifier).Value
		if strings.EqualFold(name, "setExtractFlags") {
			var known bool
			mode, known = semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 0, "flags"))
			// A mode change that only may have run leaves the mode unknown.
			if !known || i >= len(proven) {
				return
			}
		} else {
			priority := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 1, "priority"))
			switch priority.(type) {
			case *syntax.Literal, *syntax.ConstFetch:
			default:
				priorityScalar = false
			}
			v := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "value"))
			switch v.(type) {
			case *syntax.Literal, *syntax.ConstFetch:
				// Only an insertion proven to run shows the queue holds a value.
				inserted = inserted || i < len(proven)
			default:
				return
			}
		}
	}
	if inserted && (mode == 1 || (mode == 2 && priorityScalar)) {
		ctx.ReportNode(f, message)
	}
}

// possibleCalls returns the setExtractFlags and insert calls written on the
// queue variable that are not proven to run before the extraction yet may:
// calls in a branch, a loop, a short-circuit operand or a nested closure that
// precede it, and any such call inside a loop that also contains it.
func possibleCalls(ctx *analysis.Context, at *syntax.MethodCall, proven []syntax.Node) []syntax.Node {
	scope := syntax.EnclosingVariableScope(at)
	// End of the outermost loop around the extraction: later calls inside it
	// run before the next iteration's extraction.
	reach := at.Span().Start
	for p := at.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			reach = p.Span().End
		}
	}
	receiver := ctx.Text(at.Var)
	var out []syntax.Node
	visit := func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			return true
		}
		if syntax.IsVariableScope(n) {
			return n == scope
		}
		c, ok := n.(*syntax.MethodCall)
		if !ok || slices.Contains(proven, n) || c.Span().Start > reach || ctx.Text(c.Var) != receiver {
			return true
		}
		if id, ok := c.Name.(*syntax.Identifier); ok && (strings.EqualFold(id.Value, "setExtractFlags") || strings.EqualFold(id.Value, "insert")) {
			out = append(out, c)
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return out
}
