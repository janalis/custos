// Package globbraceexpansionmissingflag implements the native GlobBraceExpansionMissingFlag inspection.
package globbraceexpansionmissingflag

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Enable brace expansion for this glob pattern."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GlobBraceExpansionMissingFlag" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "glob") {
		return
	}
	p, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "pattern"))
	if !k || !brace(p) {
		return
	}
	f := semanticquery.CallArgument(c.Args, 1, "flags")
	if f != nil {
		bits, known := semanticquery.NativeContractInt(ctx, f)
		if !known || bits&128 != 0 || bits&1024 != 0 {
			return
		}
	}
	ctx.ReportNode(c, message)
}

func brace(p string) bool {
	start := -1
	comma := false
	for i := 0; i < len(p); i++ {
		switch p[i] {
		case '\\':
			i++
		case '{':
			start = i
			comma = false
		case ',':
			if start >= 0 {
				comma = true
			}
		case '}':
			if start >= 0 && comma {
				return true
			}
			start = -1
		}
	}
	return false
}
