// Package duplicatematchcondition implements the DuplicateMatchCondition inspection.
package duplicatematchcondition

import (
	"fmt"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DuplicateMatchCondition" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMatch} }

const message = "Remove the repeated match condition."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP80 {
		return
	}
	seen := map[string]bool{}
	match := n.(*syntax.Match)
	for index, arm := range match.Arms {
		for _, condition := range arm.Conds {
			key := ""
			value := semanticquery.NativeValue(ctx, condition)
			if fetch, ok := condition.(*syntax.ClassConstFetch); ok {
				value = fetch
			}
			switch value := value.(type) {
			case *syntax.Literal:
				if value.LitKind == syntax.LitString {
					text, ok := semanticquery.NativeString(ctx, value)
					if ok {
						key = "string:" + text
					}
				} else {
					key = fmt.Sprintf("%d:%s", value.LitKind, value.Raw)
				}
			case *syntax.ClassConstFetch:
				name, ok := value.Name.(*syntax.Identifier)
				class := semanticquery.StaticCallClass(ctx, value.Class)
				if ok && class != "" && ctx.Index().FindConst(class, name.Value, ctx.PHP) != nil {
					key = "case:" + class + "::" + name.Value
				}
			case *syntax.ConstFetch:
				name := strings.ToLower(value.Name.Value)
				if name == "true" || name == "false" || name == "null" {
					key = name
				}
			}
			if key != "" {
				if seen[key] {
					if len(arm.Conds) == 1 {
						removal := arm.Span()
						if next, ok := astquery.NextSignificant(ctx.File, removal.End); ok && next.Kind == syntax.TComma {
							removal.End = next.End
						} else if index > 0 {
							if previous, ok := astquery.NextSignificant(ctx.File, match.Arms[index-1].Span().End); ok && previous.Kind == syntax.TComma {
								removal.Start = previous.Start
							}
						}
						text := ctx.SpanText(removal)
						if !strings.Contains(text, "/*") && !strings.Contains(text, "//") && !strings.Contains(text, "#") {
							fix := astquery.ReplaceFix(removal, "")
							fix.Title = "Remove the unreachable match arm"
							ctx.ReportNode(condition, message, fix)
						} else {
							ctx.ReportNode(condition, message)
						}
					} else {
						ctx.ReportNode(condition, message)
					}
				}
				seen[key] = true
			}
		}
	}
}

func (rule) Semantic() {}
