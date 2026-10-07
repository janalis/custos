package probablebugs

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// suspiciousReturn reports a `return` in a `finally` block whose `try` block
// itself returns or throws.
type suspiciousReturn struct{}

func init() { register(suspiciousReturn{}) }

const suspiciousReturnMsg = "Returning from 'finally' discards the try block's return value or exception."

func (suspiciousReturn) ID() string { return "SuspiciousReturn" }

func (suspiciousReturn) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KReturn}
}

func (suspiciousReturn) Check(ctx *analysis.Context, n syntax.Node) {
	if n.Span().Len() == 0 {
		return
	}
	for p := n.Parent(); p != nil; p = p.Parent() { // D2
		if util.IsFuncLike(p) {
			return
		}
		fin, ok := p.(*syntax.Finally)
		if !ok {
			continue
		}
		try, ok := fin.Parent().(*syntax.Try)
		if !ok || try.Body == nil {
			return
		}
		if containsReturnOrThrow(try.Body) { // D3
			ctx.ReportNode(n, suspiciousReturnMsg)
		}
		return
	}
}

// containsReturnOrThrow searches n at any depth for a return or throw,
// without entering nested functions, closures, arrow functions or classes
// (their returns/throws do not leave the try block).
func containsReturnOrThrow(n syntax.Node) bool {
	found := false
	syntax.Inspect(n, func(c syntax.Node) bool {
		if found {
			return false
		}
		if util.IsFuncLike(c) {
			return false
		}
		switch c.(type) {
		case *syntax.ClassLike:
			return false
		case *syntax.Return, *syntax.Throw:
			found = true
			return false
		}
		return true
	})
	return found
}
