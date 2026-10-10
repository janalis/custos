// Package nonexhaustiveenummatch implements the NonExhaustiveEnumMatch inspection.
package nonexhaustiveenummatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonExhaustiveEnumMatch" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMatch} }

const message = "Handle every reachable enum case."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	match := n.(*syntax.Match)
	for _, arm := range match.Arms {
		if arm.Conds == nil {
			return
		}
	}
	classes := ctx.TypeOf(match.Cond).Classes()
	if len(classes) != 1 {
		return
	}
	class := ctx.Index().Class(strings.TrimPrefix(classes[0], `\`), ctx.PHP)
	if class == nil || class.Kind != syntax.KindEnum {
		return
	}
	covered := map[string]bool{}
	reachable := ""
	if value, ok := semanticquery.NativeValue(ctx, match.Cond).(*syntax.ClassConstFetch); ok {
		if name, ok := value.Name.(*syntax.Identifier); ok && semanticquery.StaticCallClass(ctx, value.Class) == class.FQN {
			if constant := class.Consts[name.Value]; constant != nil && constant.Case {
				reachable = name.Value
			}
		}
	}
	for _, arm := range match.Arms {
		for _, cond := range arm.Conds {
			fetch, ok := cond.(*syntax.ClassConstFetch)
			if !ok {
				return
			}
			id, ok := fetch.Name.(*syntax.Identifier)
			if !ok || semanticquery.StaticCallClass(ctx, fetch.Class) != class.FQN {
				return
			}
			covered[id.Value] = true
		}
	}
	for name, constant := range class.Consts {
		if constant.Case && (reachable == "" || reachable == name) && !covered[name] {
			ctx.ReportNode(match, message)
			return
		}
	}
}

func (rule) Semantic() {}
