// Package priorityqueueextractionshapemismatch implements the native PriorityQueueExtractionShapeMismatch inspection.
package priorityqueueextractionshapemismatch

import (
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
	history := semanticquery.NativeCallHistory(ctx, call, call.Var, "setExtractFlags", "insert")
	if !history.Complete {
		return
	}
	modeKnown := true
	for _, event := range history.Events {
		c := event.Node.(*syntax.MethodCall)
		name := c.Name.(*syntax.Identifier).Value
		if strings.EqualFold(name, "setExtractFlags") {
			value, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 0, "flags"))
			if event.Must {
				mode, modeKnown = value, known
			} else if !known || !modeKnown || value != mode {
				modeKnown = false
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
				inserted = inserted || event.Must
			default:
				return
			}
		}
	}
	if inserted && modeKnown && (mode == 1 || (mode == 2 && priorityScalar)) {
		ctx.ReportNode(f, message)
	}
}
