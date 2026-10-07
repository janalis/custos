package architecture

import (
	"strconv"

	"custos/internal/analysis"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// multipleReturnStatements reports methods with many return statements.
type multipleReturnStatements struct{}

func init() { register(multipleReturnStatements{}) }

func (multipleReturnStatements) ID() string { return "MultipleReturnStatements" }

func (multipleReturnStatements) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }

func (multipleReturnStatements) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	if m.Body == nil || m.Name == nil { // E2
		return
	}
	r := 0 // D2
	syntax.Inspect(m.Body, func(x syntax.Node) bool {
		if syntax.IsFuncLike(x) {
			return false
		}
		switch x.(type) {
		case *syntax.ClassLike:
			return false
		case *syntax.Return:
			r++
		}
		return true
	})
	msg := strconv.Itoa(r) + " return statements in this method; try to funnel them into a single exit."
	switch { // D3
	case r >= ctx.Int("SCREAM_THRESHOLD"):
		ctx.ReportSeverity(m.Name.Span(), meta.SeverityError, msg)
	case r >= ctx.Int("COMPLAIN_THRESHOLD"):
		ctx.Report(m.Name.Span(), msg)
	}
}
