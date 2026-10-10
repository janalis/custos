// Package enumfromuncheckedexternalvalue implements the EnumFromUncheckedExternalValue inspection.
package enumfromuncheckedexternalvalue

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "EnumFromUncheckedExternalValue" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KStaticCall} }

const message = "Handle unknown enum backing values explicitly."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	call := n.(*syntax.StaticCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "from") {
		return
	}
	class := ctx.Index().Class(semanticquery.StaticCallClass(ctx, call.Class), ctx.PHP)
	if class == nil || class.Kind != syntax.KindEnum {
		return
	}
	if !ctx.Index().IsSubtype(class.FQN, "BackedEnum", ctx.PHP) {
		return
	}
	value := semanticquery.CallArgument(call.Args, 0, "value")
	fetch, ok := value.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	source, ok := fetch.Var.(*syntax.Variable)
	if !ok || (source.Name != "_GET" && source.Name != "_POST" && source.Name != "_REQUEST" && source.Name != "_COOKIE") {
		return
	}
	for p := call.Parent(); p != nil; p = p.Parent() {
		if guard, ok := p.(*syntax.If); ok && guard.Body.Span().Contains(call.Span()) {
			stmts, listed := syntax.StmtListOf(guard.Body)
			if listed && len(stmts) > 0 && stmts[0].Span().Contains(call.Span()) {
				check, _ := semanticquery.GlobalCall(ctx, syntax.UnwrapParens(guard.Cond), "in_array")
				if check != nil && astquery.Equivalent(ctx.File, semanticquery.CallArgument(check.Args, 0, "needle"), value) {
					strict, known := astquery.BoolConst(semanticquery.CallArgument(check.Args, 2, "strict"))
					if strict && known {
						allowed := semanticquery.NativeArray(ctx, semanticquery.CallArgument(check.Args, 1, "haystack"))
						if allowed != nil && len(allowed.Items) > 0 {
							valid := true
							for _, item := range allowed.Items {
								found := false
								for _, constant := range class.Consts {
									if constant.Case && constant.Value == ctx.Text(item.Value) {
										found = true
									}
								}
								valid = valid && found
							}
							if valid {
								return
							}
						}
					}
				}
			}
		}
		if tr, ok := p.(*syntax.Try); ok {
			for _, catch := range tr.Catches {
				for _, name := range catch.Types {
					fq := ctx.Names().Class(name.Value, name.Span().Start)
					if fq == "ValueError" || fq == "Error" || fq == "Throwable" {
						return
					}
				}
			}
		}
	}
	ctx.ReportNode(call, message)
}

func (rule) Semantic() {}
