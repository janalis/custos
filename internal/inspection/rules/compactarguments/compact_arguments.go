package compactarguments

import (
	"math"
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// compactArguments reports compact() names that never appeared as a variable
// earlier in the enclosing function.
type compactArguments struct{}

func (compactArguments) ID() string               { return "CompactArguments" }
func (compactArguments) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (compactArguments) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "compact") || call.Args == nil || len(call.Args.Args) == 0 { // D1, D2
		return
	}
	scope := syntax.EnclosingFuncLike(call) // D3
	for scope != nil {
		if _, arrow := scope.(*syntax.ArrowFunction); !arrow {
			break
		}
		scope = syntax.EnclosingFuncLike(scope) // arrow functions share the parent scope
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
		content, _, ok := astquery.QuotedStringRaw(arg.Value)
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

	// D5: every `$name` token of the scope before the call (parameters
	// included). The first offset of each name is indexed once per scope.
	vars := compactScopeVars(ctx, scope)
	callStart := call.Span().Start
	if vars.dynamic < callStart { // E5
		return
	}
	known := map[string]bool{}
	for _, c := range cands {
		if off, ok := vars.first[c.name]; ok && off < callStart {
			known[c.name] = true
		}
	}
	for _, p := range syntax.FuncLikeParams(scope) {
		if p.Var != nil && p.Var.Name != "" {
			known[p.Var.Name] = true
		}
	}
	if compactHasThis(call) {
		known["this"] = true
	}

	for _, c := range cands { // D6
		if !known[c.name] {
			ctx.Report(c.span, "Variable '$"+c.name+"' may be undefined when compact() runs.")
		}
	}
}

// compactVars is the variable index of one scope.
type compactVars struct {
	first   map[string]uint32 // offset of the first `$name` token per name
	dynamic uint32            // start of the first construct defining arbitrary variables
}

// compactScopeVars indexes the `$name` tokens of scope and its first
// dynamic definer (memoised per scope).
func compactScopeVars(ctx *analysis.Context, scope syntax.Node) *compactVars {
	sp := scope.Span()
	key := strconv.FormatUint(uint64(sp.Start), 10)
	return ctx.Memo(key, func() any {
		vars := &compactVars{first: map[string]uint32{}, dynamic: math.MaxUint32}
		f := ctx.File
		for i := astquery.TokenIndex(f, sp.Start); i < len(f.Tokens); i++ {
			t := f.Tokens[i]
			if t.Start >= sp.End {
				break
			}
			if t.Kind == syntax.TVariable && t.End-t.Start > 1 {
				name := string(ctx.Src[t.Start+1 : t.End])
				if _, ok := vars.first[name]; !ok {
					vars.first[name] = t.Start
				}
			}
		}
		syntax.Inspect(scope, func(n syntax.Node) bool {
			if n != scope && syntax.IsFuncLike(n) {
				// Nested functions and closures fill their own symbol
				// table; arrow functions are looked through like in D3.
				_, arrow := n.(*syntax.ArrowFunction)
				return arrow
			}
			if compactDefinesAnyVariable(ctx, n) {
				vars.dynamic = min(vars.dynamic, n.Span().Start)
			}
			return true
		})
		return vars
	}).(*compactVars)
}

// compactDefinesAnyVariable reports whether n can create variables whose
// names are not written in the source: extract(), one-argument parse_str(),
// include/require, eval and variable variables.
func compactDefinesAnyVariable(ctx *analysis.Context, n syntax.Node) bool {
	switch x := n.(type) {
	case *syntax.Include, *syntax.Eval:
		return true
	case *syntax.Variable:
		return x.NameExpr != nil && !astquery.IsStaticPropName(x)
	case *syntax.FuncCall:
		switch ctx.GlobalFunctionName(x) {
		case "extract":
			return true
		case "parse_str":
			return x.Args != nil && len(x.Args.Args) == 1
		}
	}
	return false
}

// compactHasThis reports whether `$this` exists where call runs: in a
// non-static method, or in non-static closures and arrow functions nested
// in one.
func compactHasThis(call *syntax.FuncCall) bool {
	for p := call.Parent(); p != nil; p = p.Parent() {
		switch f := p.(type) {
		case *syntax.Closure:
			if f.Static {
				return false
			}
		case *syntax.ArrowFunction:
			if f.Static {
				return false
			}
		case *syntax.Method:
			return !f.Modifiers.Has(syntax.TStatic)
		case *syntax.Function:
			return false
		}
	}
	return false
}
