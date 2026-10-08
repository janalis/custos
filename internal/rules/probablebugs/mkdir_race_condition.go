package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// mkdirRaceCondition reports mkdir() calls whose failure is not re-checked
// with is_dir().
type mkdirRaceCondition struct{}

func init() { register(mkdirRaceCondition{}) }

// Semantic marks the rule as needing the project symbol index.
func (mkdirRaceCondition) Semantic() {}

func (mkdirRaceCondition) ID() string { return "MkdirRaceCondition" }

func (mkdirRaceCondition) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (mkdirRaceCondition) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	// D1
	if _, name, ok := util.CallName(call); !ok || !strings.EqualFold(name, "mkdir") || call.Args == nil {
		return
	}
	if c := len(call.Args.Args); c < 1 || c > 3 {
		return
	}
	if _, ok := call.Args.Args[0].(*syntax.Arg); !ok { // mkdir(...) builds a closure
		return
	}
	if f := ctx.Types().ResolveFunction(call); f != nil {
		if !strings.EqualFold(strings.TrimPrefix(f.FQN, `\`), "mkdir") {
			return
		}
	} else if !ctx.IsGlobalFunctionCall(call, "mkdir") {
		return
	}
	if util.InTestContext(ctx, call) { // D2
		return
	}

	// D3: locate the target.
	inverted, silenced := false, false
	var target syntax.Node = call
	var cur syntax.Node = call
locate:
	for {
		p := cur.Parent()
		switch p := p.(type) {
		case *syntax.If: // an expression child of an if is its condition
			target = cur
			break locate
		case *syntax.Assign, *syntax.ExprStmt:
			target = cur
			break locate
		case *syntax.Paren:
			cur = p
		case *syntax.Unary:
			switch p.Op.Kind {
			case syntax.TExclaim:
				inverted = !inverted
			case syntax.TAt:
				silenced = true
			default:
				return
			}
			cur = p
		case *syntax.Binary:
			switch p.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
				target = cur
				break locate
			case syntax.TIsIdentical, syntax.TIsNotIdentical:
				other := p.Left
				if other == cur {
					other = p.Right
				}
				if v, ok := util.BoolConst(other); ok {
					if !v {
						inverted = !inverted
					}
					if p.Op.Kind == syntax.TIsNotIdentical {
						inverted = !inverted
					}
				}
			}
			cur = p
		default:
			return
		}
	}

	args := ctx.SpanText(argsInner(call.Args))
	dir, tempVar := mkdirDir(ctx, call)
	// Spell builtins so they reach the global functions even when the
	// namespace declares or imports same-named functions.
	mkdirFn := util.QualifiedBuiltinFor(ctx, "mkdir", call)
	isDirFn := util.QualifiedBuiltin(ctx, "is_dir", call.Span().Start)
	mk := mkdirFn + "(" + args + ")"
	isDir := isDirFn + "(" + dir + ")"
	if tempVar {
		mk = mkdirFn + "($concurrentDirectory = " + args + ")"
		isDir = isDirFn + "($concurrentDirectory)"
	}
	if silenced {
		// Keep the silence operator: dropping it would emit "File exists"
		// warnings in the very race the re-check handles.
		mk = "@" + mk
	}
	andForm := "!" + mk + " && !" + isDir
	orForm := mk + " || " + isDir
	andMsg := "Re-check with is_dir() after a failed mkdir: '!mkdir(" + args + ") && !is_dir(...)'."
	orMsg := "Re-check with is_dir() after a failed mkdir: 'mkdir(" + args + ") || is_dir(...)'."

	switch c := target.Parent().(type) {
	case *syntax.ExprStmt: // D4
		thrown := dir
		if tempVar {
			thrown = "$concurrentDirectory"
		}
		repl := "if (" + andForm + ") { throw new \\RuntimeException(" + util.QualifiedBuiltin(ctx, "sprintf", call.Span().Start) + "('Directory \"%s\" was not created', " + thrown + ")); }"
		ctx.ReportNode(c, "mkdir() outcome is ignored; use 'if (!mkdir("+args+") && !is_dir(...)) { ... }'.",
			analysis.Fix{Title: "Throw when the directory was not created", Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{{Span: c.Span(), NewText: repl}}
			}})
	case *syntax.If: // D5
		// custos: `if (!mkdir($d)) { …is_dir($d)… }` already re-checks in
		// the failure branch.
		if inverted && hasIsDirCall(ctx, c.Body, mkdirDirKeys(ctx, call)) {
			return
		}
		msg, repl := orMsg, orForm
		if inverted {
			msg, repl = andMsg, andForm
		}
		ctx.ReportNode(target, msg, analysis.Fix{Title: "Re-check with is_dir()", Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: target.Span(), NewText: repl}}
		}})
	case *syntax.Binary: // D6
		if _, ok := syntax.UnwrapParens(c.Right).(*syntax.Exit); ok {
			return
		}
		outer := c
		if c.Right == target {
			if pb, ok := c.Parent().(*syntax.Binary); ok {
				outer = pb
			}
		}
		if hasIsDirCall(ctx, outer.Right, mkdirDirKeys(ctx, call)) {
			return
		}
		and := outer.Op.Kind == syntax.TBooleanAnd || outer.Op.Kind == syntax.TAnd
		msg := orMsg
		if and {
			msg = andMsg
		}
		if c.Right != target {
			ctx.ReportNode(target, msg)
			return
		}
		left := ctx.Text(c.Left)
		repl := left + " || " + orForm
		if c.Op.Kind == syntax.TBooleanAnd || c.Op.Kind == syntax.TAnd {
			repl = left + " && " + andForm
		}
		ctx.ReportNode(target, msg, analysis.Fix{Title: "Re-check with is_dir()", Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: c.Span(), NewText: repl}}
		}})
	}
}

// argsInner is the span from after the opening parenthesis of l to the end
// of its last argument (l has at least one; a missing closing parenthesis in
// broken code is tolerated).
func argsInner(l *syntax.ArgList) syntax.Span {
	return syntax.Span{Start: l.Span().Start + 1, End: l.Args[len(l.Args)-1].Span().End}
}

// mkdirDir returns the first argument text and whether the temp-var form is
// needed (first argument neither a plain variable nor a quoted string).
func mkdirDir(ctx *analysis.Context, call *syntax.FuncCall) (string, bool) {
	a := call.Args.Args[0].(*syntax.Arg) // checked by Check
	switch v := a.Value.(type) {
	case *syntax.Variable:
		if v.NameExpr == nil {
			return ctx.Text(v), false
		}
	case *syntax.Literal:
		if v.LitKind == syntax.LitString {
			return ctx.Text(v), false
		}
	case *syntax.InterpolatedString:
		if !v.Backtick {
			return ctx.Text(v), false
		}
	}
	return ctx.Text(a.Value), true
}

// hasIsDirCall reports whether e contains an is_dir() call whose first
// argument is one of the directory expressions in keys (normalised text).
func hasIsDirCall(ctx *analysis.Context, e syntax.Node, keys []string) bool {
	found := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		call, ok := n.(*syntax.FuncCall)
		if !ok || !strings.EqualFold(util.CallLastName(call), "is_dir") || call.Args == nil || len(call.Args.Args) == 0 {
			return true
		}
		if a, ok := call.Args.Args[0].(*syntax.Arg); ok && a.Value != nil && !a.Unpack {
			k := mkdirNorm(ctx.Text(syntax.UnwrapParens(a.Value)))
			for _, want := range keys {
				if k == want {
					found = true
				}
			}
		}
		return !found
	})
	return found
}

// mkdirDirKeys returns the normalised texts an is_dir() re-check may use for
// mkdir's directory: the first argument itself and, when it is an assignment
// (`mkdir($d = expr)`), the assigned variable and the assigned value.
func mkdirDirKeys(ctx *analysis.Context, call *syntax.FuncCall) []string {
	a := call.Args.Args[0].(*syntax.Arg) // checked by Check
	v := syntax.UnwrapParens(a.Value)
	keys := []string{mkdirNorm(ctx.Text(v))}
	if as, ok := v.(*syntax.Assign); ok && as.Op.Kind == syntax.TEqual {
		keys = append(keys, mkdirNorm(ctx.Text(syntax.UnwrapParens(as.Var))), mkdirNorm(ctx.Text(syntax.UnwrapParens(as.Value))))
	}
	return keys
}

// mkdirNorm drops whitespace so formatting differences do not matter.
func mkdirNorm(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r':
			return -1
		}
		return r
	}, s)
}
