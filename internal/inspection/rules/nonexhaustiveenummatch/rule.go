// Package nonexhaustiveenummatch implements the native
// NonExhaustiveEnumMatch inspection.
package nonexhaustiveenummatch

import (
	"slices"
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

type rule struct{}

// New returns the NonExhaustiveEnumMatch inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonExhaustiveEnumMatch" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMatch} }

const message = "Handle every reachable enum case."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP.Below(phpversion.PHP81) {
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
	// missing starts as every case and loses the ones an arm handles.
	missing := map[string]bool{}
	for name, constant := range class.Consts {
		if constant.Case {
			missing[name] = true
		}
	}
	for _, arm := range match.Arms {
		for _, cond := range arm.Conds {
			name := caseName(ctx, classes[0], cond)
			if name == "" {
				return
			}
			delete(missing, name)
		}
	}
	if len(missing) == 0 {
		return
	}
	// The cases the subject may still be there: comparisons and assignments
	// narrow them, unless a jump may land between those and the match.
	if names, ok := ctx.Types().Native().TypeOf(match.Cond).EnumCaseNames(classes[0]); ok && !semanticquery.NativeLabeled(ctx, match.Cond) {
		for name := range missing {
			if !slices.Contains(names, name) {
				delete(missing, name)
			}
		}
	}
	if len(missing) != 0 {
		ctx.ReportNode(match, message)
	}
}

func (rule) Semantic() {}

// caseName returns the case of enum class an `Enum::NAME` expression
// denotes, constants of the enum aliasing a case included; "" when e is
// anything else or the constant cannot be traced to a case.
func caseName(ctx *analysis.Context, class string, e syntax.Expr) string {
	if _, ok := e.(*syntax.ClassConstFetch); !ok {
		return ""
	}
	names, ok := ctx.Types().Native().TypeOf(e).EnumCaseNames(class)
	if !ok || len(names) != 1 {
		return ""
	}
	return names[0]
}
