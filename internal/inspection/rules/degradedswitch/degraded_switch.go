package degradedswitch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// degradedSwitch reports switch statements with a single case and/or only a
// default branch.
type degradedSwitch struct{}

func (degradedSwitch) ID() string               { return "DegradedSwitch" }
func (degradedSwitch) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KSwitch} }
func (degradedSwitch) Check(ctx *analysis.Context, n syntax.Node) {
	sw := n.(*syntax.Switch)
	cases, hasDefault := 0, false
	for _, c := range sw.Cases {
		if c.Cond == nil {
			hasDefault = true
		} else {
			cases++
		}
	}
	var msg string
	switch {
	case cases == 0 && hasDefault: // D1
		msg = "This switch only has a default branch; keep just its body."
	case cases == 1 && !hasDefault: // D2
		msg = "This switch has a single case; an 'if' is clearer."
	case cases == 1 && hasDefault: // D3
		msg = "This switch has a single case and a default; an 'if'/'else' is clearer."
	default:
		return
	}
	s := sw.Span()
	ctx.Report(syntax.Span{Start: s.Start, End: s.Start + uint32(len("switch"))}, msg)
}
