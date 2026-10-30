package incrementdecrementoperationequivalent

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// incrementDecrementOperationEquivalent suggests ++/-- for `$n += 1`,
// `$n = $n - 1` and friends.
type incrementDecrementOperationEquivalent struct{}

func (incrementDecrementOperationEquivalent) ID() string {
	return "IncrementDecrementOperationEquivalent"
}
func (incrementDecrementOperationEquivalent) Semantic() {}

func (incrementDecrementOperationEquivalent) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KAssign}
}

func incDecIsOne(e syntax.Expr) bool {
	l, ok := e.(*syntax.Literal)
	return ok && l.LitKind == syntax.LitInt && l.Raw == "1"
}

// incDecValueDiscarded reports whether the value of assignment a is thrown
// away: it is a whole expression statement or an init/step expression of a
// `for` header. Only there do the prefix and postfix forms agree.
func incDecValueDiscarded(a *syntax.Assign) bool {
	parent, child := astquery.ParentSkipParens(a)
	switch p := parent.(type) {
	case *syntax.ExprStmt:
		return true
	case *syntax.For:
		for _, list := range [][]syntax.Expr{p.Init, p.Loop} {
			for _, e := range list {
				if e == child {
					return true
				}
			}
		}
	}
	return false
}

// incDecImplicitNull reports whether target is `$this->p` or `self::$p` /
// `static::$p` for a property the enclosing class declares without a type
// and without a non-null default, and (instance properties) its constructor
// does not initialise: such a property starts out as null.
func incDecImplicitNull(target syntax.Expr) bool {
	name, static := "", false
	switch t := target.(type) {
	case *syntax.PropertyFetch:
		v, ok1 := t.Var.(*syntax.Variable)
		id, ok2 := t.Name.(*syntax.Identifier)
		if !ok1 || !ok2 || v.Name != "this" {
			return false
		}
		name = id.Value
	case *syntax.StaticPropertyFetch:
		c, ok1 := t.Class.(*syntax.Name)
		v, ok2 := t.Name.(*syntax.Variable)
		if !ok1 || !ok2 || (!strings.EqualFold(c.Value, "self") && !strings.EqualFold(c.Value, "static")) {
			return false
		}
		name, static = v.Name, true
	default:
		return false
	}
	class := astquery.ClassScope(target)
	if class == nil {
		return false
	}
	for _, mem := range class.Members {
		p, ok := mem.(*syntax.Property)
		if !ok || p.Type != nil || p.Modifiers.Has(syntax.TStatic) != static {
			continue
		}
		for _, it := range p.Props {
			if it.Var.Name == name {
				return (it.Default == nil || syntax.IsNullConst(syntax.UnwrapParens(it.Default))) &&
					(static || !flowquery.CtorAssignsProperty(class, name))
			}
		}
	}
	return false
}

func (incrementDecrementOperationEquivalent) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Assign)
	op := ""
	switch a.Op.Kind {
	case syntax.TPlusEqual: // D1
		if incDecIsOne(a.Value) {
			op = "++"
		}
	case syntax.TMinusEqual: // D2
		if incDecIsOne(a.Value) {
			op = "--"
		}
	case syntax.TEqual:
		if a.ByRef {
			return
		}
		// `T = T ± 1` evaluates T twice, `++T` once: a target with side
		// effects is not equivalent (custos diverges).
		b, ok := a.Value.(*syntax.Binary)
		if !ok || flowquery.MayHaveSideEffects(a.Var) {
			return
		}
		switch b.Op.Kind {
		case syntax.TPlus: // D3
			if (incDecIsOne(b.Left) && astquery.EquivalentFoldNames(ctx.File, b.Right, a.Var)) ||
				(incDecIsOne(b.Right) && astquery.EquivalentFoldNames(ctx.File, b.Left, a.Var)) {
				op = "++"
			}
		case syntax.TMinus: // D4
			if incDecIsOne(b.Right) && astquery.EquivalentFoldNames(ctx.File, b.Left, a.Var) {
				op = "--"
			}
		}
	}
	if op == "" {
		return
	}
	// `+ 1` on a string or bool is arithmetic (or a TypeError), `++` on a
	// string is alphanumeric and on a bool a no-op (custos diverges).
	operand := a.Var
	if b, ok := a.Value.(*syntax.Binary); ok && a.Op.Kind == syntax.TEqual {
		operand = b.Left
		if incDecIsOne(b.Left) && b.Op.Kind == syntax.TPlus {
			operand = b.Right
		}
	}
	t := ctx.TypeOf(operand)
	if t.Has("string") || t.Has("bool") || t.Has("true") || t.Has("false") {
		return
	}
	// `null - 1` is -1 but `--` leaves null untouched (custos diverges): a
	// nullable operand, or one read under `@` (possibly unset), is skipped.
	if op == "--" {
		if u, ok := a.Parent().(*syntax.Unary); t.Has("null") || (ok && u.Op.Kind == syntax.TAt) {
			return
		}
	}
	if dim, ok := a.Var.(*syntax.ArrayDimFetch); ok { // D5
		t := ctx.TypeOf(dim.Var)
		if t.IsUnknown() || t.Has("string") {
			return
		}
		arr := false
		for _, at := range t.Atoms() {
			if at == "array" || strings.HasSuffix(at, "[]") {
				arr = true
			}
		}
		if !arr {
			return
		}
	}
	// The postfix form yields the old value, the assignment the new one:
	// it is only offered where the value is discarded (custos diverges).
	target := ctx.Text(a.Var)
	repl := op + target
	if !ctx.Bool("PREFER_PREFIX_STYLE") && incDecValueDiscarded(a) {
		repl = target + op
	}
	span := a.Span()
	msg := "Use '" + repl + "' instead."
	// A property that starts out as null may still be null here: `--` is
	// suggested but not applied.
	if op == "--" && t.IsUnknown() && incDecImplicitNull(a.Var) {
		ctx.Report(span, msg)
		return
	}
	// `$a-$b-=1` must not become `$a---$b` (lexed as `$a-- - $b`).
	text := repl
	if repl[0] == op[0] && ctx.File.Src[span.Start-1] == op[0] {
		text = " " + repl
	}
	ctx.Report(span, msg, diagnostic.Fix{
		Title: "Use the " + op + " operator",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: text}} },
	})
}
