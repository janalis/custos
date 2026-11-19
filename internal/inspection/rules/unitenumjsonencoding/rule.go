// Package unitenumjsonencoding implements the native UnitEnumJsonEncoding
// inspection.
package unitenumjsonencoding

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

type rule struct{}

// New returns the UnitEnumJsonEncoding inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UnitEnumJsonEncoding" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Serialize this unit enum through an explicit representation."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP.Below(phpversion.PHP81) {
		return
	}
	call, _ := semanticquery.GlobalCall(ctx, n, "json_encode")
	if call == nil {
		return
	}
	value := semanticquery.CallArgument(call.Args, 0, "value")
	for _, cls := range ctx.TypeOf(value).Classes() {
		class := ctx.Index().Class(strings.TrimPrefix(cls, `\`), ctx.PHP)
		if class == nil || class.Kind != syntax.KindEnum || ctx.Index().IsSubtype(class.FQN, "JsonSerializable", ctx.PHP) {
			continue
		}
		if !ctx.Index().IsSubtype(class.FQN, "BackedEnum", ctx.PHP) {
			ctx.ReportNode(call, message)
			return
		}
	}
}

func (rule) Semantic() {}
