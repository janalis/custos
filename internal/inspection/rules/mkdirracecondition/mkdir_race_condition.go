package mkdirracecondition

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// mkdirRaceCondition reports mkdir() calls whose failure is not re-checked
// with is_dir().
type mkdirRaceCondition struct{}

// Semantic marks the rule as needing the project symbol index.
func (mkdirRaceCondition) Semantic()  {}
func (mkdirRaceCondition) ID() string { return "MkdirRaceCondition" }
func (mkdirRaceCondition) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (mkdirRaceCondition) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	// D1
	if _, name, ok := astquery.CallName(call); !ok || !strings.EqualFold(name, "mkdir") || call.Args == nil {
		return
	}
	if c := len(call.Args.Args); c < 1 || c > 3 {
		return
	}
	// The directory is the first positional or the `directory:` argument.
	// mkdir(...) builds a closure and a spread hides the directory: neither
	// is reported.
	dirArg := semanticquery.CallArgument(call.Args, 0, "directory")
	if dirArg == nil {
		return
	}
	if f := ctx.Types().ResolveFunction(call); f != nil {
		if !strings.EqualFold(strings.TrimPrefix(f.FQN, `\`), "mkdir") {
			return
		}
	} else if !ctx.IsGlobalFunctionCall(call, "mkdir") {
		return
	}
	if semanticquery.InTestContext(ctx, call) { // D2
		return
	}
	if mkdirFreshName(ctx, dirArg) { // E6
		return
	}
	keys := mkdirDirKeys(ctx, dirArg)
	if mkdirUsedAsLock(ctx, call, keys) {
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
			case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
				other := p.Left
				if other == cur {
					other = p.Right
				}
				// mkdir() returns a boolean, so a loose comparison with a
				// boolean literal means the same as the strict one.
				v, ok := astquery.BoolConst(other)
				if !ok {
					return
				}
				if !v {
					inverted = !inverted
				}
				if p.Op.Kind == syntax.TIsNotIdentical || p.Op.Kind == syntax.TIsNotEqual {
					inverted = !inverted
				}
				cur = p
			default:
				// custos: any other operator (`xor`, `&`, `=== null`, …)
				// has an operand a rewrite of the call would drop.
				return
			}
		default:
			return
		}
	}

	inner := argsInner(call.Args)
	args := ctx.SpanText(inner)
	dir, tempVar := mkdirDir(ctx, dirArg)
	// Spell builtins so they reach the global functions even when the
	// namespace declares or imports same-named functions.
	mkdirFn := semanticquery.QualifiedBuiltinFor(ctx, "mkdir", call)
	isDirFn := semanticquery.QualifiedBuiltin(ctx, "is_dir", call.Span().Start)
	mk := mkdirFn + "(" + args + ")"
	isDir := isDirFn + "(" + dir + ")"
	if tempVar {
		// The assignment goes in front of the directory value, after a
		// `directory:` label and after arguments written before it.
		at := dirArg.Span().Start
		mk = mkdirFn + "(" + ctx.SpanText(syntax.Span{Start: inner.Start, End: at}) + "$concurrentDirectory = " + ctx.SpanText(syntax.Span{Start: at, End: inner.End}) + ")"
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
		// custos: `if (!is_dir($d)) { mkdir($d); } return is_dir($d);`
		// re-checks in a later statement.
		if laterIsDirCall(ctx, c, keys) || silenced && guardedByIsDir(ctx, c, keys) { // E7
			return
		}
		thrown := dir
		if tempVar {
			thrown = "$concurrentDirectory"
		}
		repl := "if (" + andForm + ") { throw new \\RuntimeException(" + semanticquery.QualifiedBuiltin(ctx, "sprintf", call.Span().Start) + "('Directory \"%s\" was not created', " + thrown + ")); }"
		if _, _, ok := astquery.StmtList(ctx.File, c); !ok {
			// custos: an unbraced body (`if ($c) mkdir($d); else …`): a
			// bare if would capture the outer else.
			repl = "{ " + repl + " }"
		}
		ctx.ReportNode(c, "mkdir() outcome is ignored; use 'if (!mkdir("+args+") && !is_dir(...)) { ... }'.",
			diagnostic.Fix{Title: "Throw when the directory was not created", Edits: func() []diagnostic.TextEdit {
				span := c.Span()
				if !strings.HasSuffix(ctx.SpanText(span), ";") {
					// The statement is ended by `?>`, which must stay.
					span.End = c.Expr.Span().End
				}
				return []diagnostic.TextEdit{{Span: span, NewText: repl}}
			}})
	case *syntax.If: // D5
		// custos: `if (!mkdir($d)) { …is_dir($d)… }` already re-checks in
		// the failure branch.
		if inverted && hasIsDirCall(ctx, c.Body, keys) {
			return
		}
		msg, repl := orMsg, orForm
		if inverted {
			msg, repl = andMsg, andForm
		}
		ctx.ReportNode(target, msg, diagnostic.Fix{Title: "Re-check with is_dir()", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: target.Span(), NewText: repl}}
		}})
	case *syntax.Binary: // D6
		if syntax.ExitInvocation(c.Right) {
			return
		}
		outer := c
		if c.Right == target {
			if pb, ok := c.Parent().(*syntax.Binary); ok {
				outer = pb
			}
		}
		if hasIsDirCall(ctx, outer.Right, keys) || mkdirRetryLoop(ctx, c, keys) {
			return
		}
		// custos: the re-check form follows the call's polarity, not the
		// operator: `A || !mkdir($d)` needs `!mkdir($d) && !is_dir($d)`
		// (upstream emitted the `||` form, inverting the condition).
		msg, form, formAnd := orMsg, orForm, false
		if inverted {
			msg, form, formAnd = andMsg, andForm, true
		}
		if c.Right != target {
			ctx.ReportNode(target, msg)
			return
		}
		and := c.Op.Kind == syntax.TBooleanAnd || c.Op.Kind == syntax.TAnd
		if and != formAnd {
			form = "(" + form + ")"
		}
		// The operator is kept as written: `$x = A or mkdir($d)` binds
		// differently from `||`.
		repl := ctx.Text(c.Left) + " " + ctx.SpanText(c.Op.Span) + " " + form
		ctx.ReportNode(target, msg, diagnostic.Fix{Title: "Re-check with is_dir()", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: c.Span(), NewText: repl}}
		}})
	}
}

// mkdirRetryLoop reports whether the condition containing c is a loop's
// whose is_dir() test of the same directory runs again on the next
// iteration (`while (!is_dir($d) && !@mkdir($d)) { usleep(…); }`): the race
// is already handled (custos).
func mkdirRetryLoop(ctx *analysis.Context, c syntax.Node, keys []string) bool {
	cond := c
	for {
		switch p := cond.Parent().(type) {
		case *syntax.Binary, *syntax.Unary, *syntax.Paren:
			cond = p
			continue
		case *syntax.While:
			return p.Cond == cond && hasIsDirCall(ctx, cond, keys)
		case *syntax.DoWhile:
			return p.Cond == cond && hasIsDirCall(ctx, cond, keys)
		}
		return false
	}
}

// mkdirUsedAsLock reports whether the outcome of call is tested and the
// function (or the file's top-level code) holding it also removes the same
// directory with rmdir(): mkdir()
// is then an atomic lock (`if (!@mkdir($lock)) return; …; rmdir($lock);`),
// and its failure must not be overridden by an is_dir() re-check (custos).
func mkdirUsedAsLock(ctx *analysis.Context, call *syntax.FuncCall, keys []string) bool {
	var n syntax.Node = call
	for {
		switch p := n.Parent().(type) {
		case *syntax.Paren, *syntax.Unary:
			n = p
			continue
		case *syntax.ExprStmt:
			return false // outcome ignored: a scratch directory, not a lock
		}
		break
	}
	if scope := syntax.EnclosingFuncLike(call); scope != nil {
		return hasCheckCall(ctx, scope, keys, rmdirOnly)
	}
	for _, st := range ctx.File.Stmts {
		if hasCheckCall(ctx, st, keys, rmdirOnly) {
			return true
		}
	}
	return false
}

var rmdirOnly = map[string]bool{"rmdir": true}

// laterIsDirCall reports whether a statement following s, in its own
// statement list or in an enclosing one of the same function, checks one of
// the directory expressions in keys (laterChecks).
func laterIsDirCall(ctx *analysis.Context, s syntax.Stmt, keys []string) bool {
	for s != nil {
		if list, i, ok := astquery.StmtList(ctx.File, s); ok {
			for _, next := range list[i+1:] {
				if hasCheckCall(ctx, next, keys, laterChecks) {
					return true
				}
				if assignsKey(ctx, next, keys) { // a later check names another directory
					return false
				}
			}
		}
		n := s.Parent()
		s = nil
		for ; n != nil && !syntax.IsFuncLike(n); n = n.Parent() {
			if st, ok := n.(syntax.Stmt); ok {
				s = st
				break
			}
		}
	}
	return false
}

// assignsKey reports whether n assigns (any operator) to one of the
// directory expressions in keys.
func assignsKey(ctx *analysis.Context, n syntax.Node, keys []string) bool {
	found := false
	syntax.Inspect(n, func(x syntax.Node) bool {
		if a, ok := x.(*syntax.Assign); ok {
			k := mkdirNorm(ctx.Text(syntax.UnwrapParens(a.Var)))
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

// argsInner is the span from after the opening parenthesis of l to the end
// of its last argument (l has at least one; a missing closing parenthesis in
// broken code is tolerated).
func argsInner(l *syntax.ArgList) syntax.Span {
	return syntax.Span{Start: l.Span().Start + 1, End: l.Args[len(l.Args)-1].Span().End}
}

// mkdirDir returns the directory argument text and whether the temp-var form
// is needed: the text is only repeated when evaluating it again is free of
// side effects (a plain variable, or a quoted string interpolating nothing
// but plain variables). Heredocs are never repeated.
func mkdirDir(ctx *analysis.Context, dirArg syntax.Expr) (string, bool) {
	text := ctx.Text(dirArg)
	switch v := dirArg.(type) {
	case *syntax.Variable:
		return text, v.NameExpr != nil
	case *syntax.Literal:
		return text, v.LitKind != syntax.LitString || strings.HasPrefix(text, "<<<")
	case *syntax.InterpolatedString:
		if v.Backtick || v.Heredoc {
			return text, true
		}
		for _, part := range v.Parts {
			switch part := part.(type) {
			case *syntax.StringPart:
			case *syntax.Variable:
				if part.NameExpr != nil {
					return text, true
				}
			default: // property, offset or call: may run user code
				return text, true
			}
		}
		return text, false
	}
	return text, true
}

// guardedByIsDir reports whether statement st is the body, or in the body,
// of an `if` whose condition checks the same directory with is_dir(): with
// `@`, `if (!is_dir($d)) { @mkdir($d); }` is a deliberate best effort (E7).
func guardedByIsDir(ctx *analysis.Context, st syntax.Stmt, keys []string) bool {
	p := st.Parent()
	if _, ok := p.(*syntax.Block); ok {
		p = p.Parent()
	}
	i, ok := p.(*syntax.If)
	return ok && hasIsDirCall(ctx, i.Cond, keys)
}

// hasIsDirCall reports whether e contains an is_dir() call whose first
// argument is one of the directory expressions in keys (normalised text).
func hasIsDirCall(ctx *analysis.Context, e syntax.Node, keys []string) bool {
	return hasCheckCall(ctx, e, keys, isDirOnly)
}

var (
	isDirOnly = map[string]bool{"is_dir": true}
	// laterChecks are the calls that, in a statement after an ignored
	// mkdir(), find out whether the directory exists (custos: realpath()
	// returns false for a missing directory, Kimai's doctor page).
	laterChecks = map[string]bool{"is_dir": true, "file_exists": true, "realpath": true, "is_writable": true, "is_writeable": true}
)

// hasCheckCall reports whether e contains a call to one of names whose
// first argument is one of the directory expressions in keys.
func hasCheckCall(ctx *analysis.Context, e syntax.Node, keys []string, names map[string]bool) bool {
	found := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		call, ok := n.(*syntax.FuncCall)
		if !ok || !names[strings.ToLower(astquery.CallLastName(call))] || call.Args == nil || len(call.Args.Args) == 0 {
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
// mkdir's directory: the directory argument itself and, when it is an
// assignment (`mkdir($d = expr)`), the assigned variable and the assigned value.
func mkdirDirKeys(ctx *analysis.Context, dirArg syntax.Expr) []string {
	v := syntax.UnwrapParens(dirArg)
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

// mkdirRandom lists the functions whose result makes a directory name that
// nobody else can have created first.
var mkdirRandom = map[string]bool{"random_bytes": true, "random_int": true, "uniqid": true, "mt_rand": true, "rand": true}

// mkdirFreshName implements E6: every value the directory argument can hold
// is built from a random value. Creating such a directory must fail when it
// already exists, so an is_dir() fallback would trust a planted directory.
func mkdirFreshName(ctx *analysis.Context, dirArg syntax.Expr) bool {
	values := []syntax.Expr{dirArg}
	if _, ok := syntax.UnwrapParens(dirArg).(*syntax.Variable); ok {
		vs, complete := semanticquery.LocalValuesComplete(ctx.Types(), dirArg)
		if !complete || len(vs) == 0 {
			return false
		}
		values = vs
	}
	for _, v := range values {
		random := false
		syntax.Inspect(v, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				if _, name, ok := astquery.CallName(c); ok && mkdirRandom[strings.ToLower(name)] {
					random = true
				}
			}
			return !random
		})
		if !random {
			return false
		}
	}
	return true
}
