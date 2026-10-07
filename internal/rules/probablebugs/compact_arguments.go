package probablebugs

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// compactArguments reports compact() names that never appeared as a variable
// earlier in the enclosing function.
type compactArguments struct{}

func init() { register(compactArguments{}) }

func (compactArguments) ID() string { return "CompactArguments" }

func (compactArguments) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (compactArguments) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "compact") || call.Args == nil || len(call.Args.Args) == 0 { // D1, D2
		return
	}
	scope := util.EnclosingFuncLike(call) // D3
	for scope != nil {
		if _, arrow := scope.(*syntax.ArrowFunction); !arrow {
			break
		}
		scope = util.EnclosingFuncLike(scope) // arrow functions share the parent scope
	}
	if scope == nil {
		return
	}

	// D4: candidates, keeping the last occurrence of each name.
	type candidate struct {
		name string
		span syntax.Span
	}
	var cands []candidate
	for _, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Unpack {
			continue
		}
		content, _, ok := util.QuotedStringRaw(arg.Value)
		if !ok || content == "" {
			continue
		}
		for i := range cands {
			if cands[i].name == content {
				cands = append(cands[:i], cands[i+1:]...)
				break
			}
		}
		cands = append(cands, candidate{content, arg.Value.Span()})
	}
	if len(cands) == 0 {
		return
	}

	// D5: every `$name` token of the scope before the call (parameters included).
	known := map[string]bool{}
	f := ctx.File
	callStart := call.Span().Start
	for i := util.TokenIndex(f, scope.Span().Start); i < len(f.Tokens); i++ {
		t := f.Tokens[i]
		if t.Start >= callStart {
			break
		}
		if t.Kind == syntax.TVariable && t.End-t.Start > 1 {
			known[string(ctx.Src[t.Start+1:t.End])] = true
		}
	}
	for _, p := range util.FuncLikeParams(scope) {
		if p.Var != nil && p.Var.Name != "" {
			known[p.Var.Name] = true
		}
	}

	for _, c := range cands { // D6
		if !known[c.name] {
			ctx.Report(c.span, "Variable '$"+c.name+"' may be undefined when compact() runs.")
		}
	}
}
