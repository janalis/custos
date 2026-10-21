package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

// TaintedWrapperSink reports whether c passes a value from outside the
// project to a sink of context inside the function it calls; with
// operations, only sinks of those names count.
func TaintedWrapperSink(ctx *analysis.Context, c *syntax.FuncCall, context flow.Context, operations ...string) bool {
	for _, sink := range ctx.Flow().Sinks(c) {
		if len(operations) > 0 {
			matched := false
			for _, operation := range operations {
				if sink.Name == operation {
					matched = true
				}
			}
			if !matched {
				continue
			}
		}
		if sink.Context != context || !sink.Argument.Complete || sink.Argument.Safe&context != 0 {
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
