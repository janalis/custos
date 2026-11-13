package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// Instanceof flaws of an if or elseif condition, as NotOptimalIfConditions
// reports them with REPORT_INSTANCE_OF_FLAWS. The decided type checks
// leave them to that rule, so both read them here.

type instanceofFlawsKey struct{ cond syntax.Expr }

// InstanceofFlaws returns the operands of condition cond (split by
// astquery.SplitCondition, two at least) that are instanceof flaws, in
// source order:
//   - under `&&`, a comparison (`==`, `!=`, `===`, `!==`) of the value the
//     first instanceof of the chain tests with a literal value (see
//     flawLiteral), except a loose comparison with a string when the
//     instanceof names a class with __toString();
//   - an instanceof another instanceof of the chain on the same value makes
//     redundant, both naming known classes: under `||` the more specific
//     one, under `&&` the broader one.
//
// Cached per condition.
func InstanceofFlaws(ctx *analysis.Context, cond syntax.Expr) []syntax.Expr {
	return ctx.Shared(instanceofFlawsKey{cond}, func() any {
		ops, op := astquery.SplitCondition(cond)
		if len(ops) < 2 {
			return []syntax.Expr(nil)
		}
		flawed := map[syntax.Expr]bool{}
		if op == syntax.TBooleanAnd {
			equalitiesNextToInstanceof(ctx, ops, flawed)
		}
		redundantInstanceofs(ctx, ops, op, flawed)
		var out []syntax.Expr
		for _, o := range ops {
			if flawed[o] {
				out = append(out, o)
			}
		}
		return out
	}).([]syntax.Expr)
}

// flawLiteral reports whether e (parentheses stripped) is a literal
// value: a scalar or string literal, null, true or false, an array
// literal, or any of them under prefix operators (`-1`, `!false`).
func flawLiteral(e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Literal, *syntax.InterpolatedString, *syntax.Array:
		return true
	case *syntax.ConstFetch:
		switch strings.ToLower(strings.TrimPrefix(x.Name.Value, `\`)) {
		case "null", "true", "false":
			return true
		}
	case *syntax.Unary:
		return flawLiteral(x.Expr)
	}
	return false
}

func equalitiesNextToInstanceof(ctx *analysis.Context, ops []syntax.Expr, flawed map[syntax.Expr]bool) {
	var test *syntax.Instanceof
	for _, o := range ops {
		if io, ok := o.(*syntax.Instanceof); ok {
			test = io
			break
		}
	}
	if test == nil {
		return
	}
	for _, o := range ops {
		b, ok := o.(*syntax.Binary)
		if !ok {
			continue
		}
		switch b.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
		default:
			continue
		}
		var other syntax.Expr
		switch {
		case astquery.EquivalentFoldNames(ctx.File, b.Left, test.Expr):
			other = b.Right
		case astquery.EquivalentFoldNames(ctx.File, b.Right, test.Expr):
			other = b.Left
		}
		// Comparing the object with another value (`$a instanceof X && $a !==
		// $b`) is an ordinary identity or equality check.
		if other != nil && flawLiteral(other) && !stringableEquality(ctx, test, b, other) {
			flawed[b] = true
		}
	}
}

// stringableEquality reports a loose comparison b with string literal
// other when the instanceof class converts to a string:
// `$r instanceof Status && $r == 'QUEUED'` compares the status text.
func stringableEquality(ctx *analysis.Context, test *syntax.Instanceof, b *syntax.Binary, other syntax.Expr) bool {
	if b.Op.Kind != syntax.TIsEqual && b.Op.Kind != syntax.TIsNotEqual {
		return false
	}
	switch x := syntax.UnwrapParens(other).(type) {
	case *syntax.Literal:
		if x.LitKind != syntax.LitString {
			return false
		}
	case *syntax.InterpolatedString:
	default:
		return false
	}
	name, ok := test.Class.(*syntax.Name)
	if !ok {
		return false
	}
	fqn := ctx.Names().Class(name.Value, name.Span().Start)
	return ctx.Index().FindMethod(fqn, "__toString", ctx.PHP) != nil
}

type instanceofEntry struct {
	node *syntax.Instanceof
	fqn  string
}

func redundantInstanceofs(ctx *analysis.Context, ops []syntax.Expr, op syntax.TokenKind, flawed map[syntax.Expr]bool) {
	var entries []instanceofEntry
	for _, o := range ops {
		io, ok := o.(*syntax.Instanceof)
		if !ok {
			continue
		}
		name, ok := io.Class.(*syntax.Name)
		if !ok {
			continue
		}
		fqn := ctx.Names().Class(name.Value, name.Span().Start)
		c := ctx.Index().Class(fqn, ctx.PHP)
		if c == nil {
			continue
		}
		entries = append(entries, instanceofEntry{io, c.FQN})
	}
	if len(entries) < 2 {
		return
	}
	used := make([]bool, len(entries))
	for i := range entries {
		if used[i] {
			continue
		}
		group := []int{i}
		used[i] = true
		for j := i + 1; j < len(entries); j++ {
			if !used[j] && astquery.EquivalentFoldNames(ctx.File, entries[i].node.Expr, entries[j].node.Expr) {
				group = append(group, j)
				used[j] = true
			}
		}
		for _, a := range group {
			anc := map[string]bool{}
			for _, c := range ctx.Index().Ancestors(entries[a].fqn, ctx.PHP) {
				anc[flawClassKey(c.FQN)] = true
			}
			for _, b := range group {
				k := flawClassKey(entries[b].fqn)
				if b == a || k == "datetimeinterface" && ctx.PHP < phpversion.PHP55 || !anc[k] {
					continue
				}
				// a's class is a subtype of b's: under || the specific
				// check a is subsumed, under && the broad check b is.
				redundant := a
				if op == syntax.TBooleanAnd {
					redundant = b
				}
				flawed[entries[redundant].node] = true
			}
		}
	}
}

func flawClassKey(fqn string) string { return strings.ToLower(strings.TrimPrefix(fqn, `\`)) }
