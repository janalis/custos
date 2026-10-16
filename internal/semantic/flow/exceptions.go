package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// exceptionProof caches the syntax of a block once. Entry checks are repeated
// because a loop back-edge may introduce an undefined value or a discarded
// object. A failed proof preserves the ordinary exceptional transfer.
type exceptionProof struct {
	safe   bool
	local  bool
	reads  []string
	writes []string
}

type exceptionBuilder struct {
	proof   exceptionProof
	written map[string]bool
	fuel    int
}

func (g *graph) exceptionProofs(scope syntax.Node) {
	if g.env.ticks() {
		return
	}
	g.proofs = map[int]*exceptionProof{}
	for i := range g.blocks {
		b := &g.blocks[i]
		if b.exceptional <= 0 {
			continue
		}
		build := exceptionBuilder{proof: exceptionProof{safe: true, local: scope != nil}, fuel: 256}
		if b.node != nil {
			build.proof.safe = build.node(b.node)
		}
		for j := len(b.nodes) - 1; build.proof.safe && j >= 0; j-- {
			build.proof.safe = build.node(b.nodes[j])
		}
		g.proofs[i] = &build.proof
	}
}

func (b *exceptionBuilder) node(n syntax.Node) bool {
	switch n := n.(type) {
	case *syntax.ExprStmt:
		return b.node(n.Expr)
	case *syntax.Assign:
		name := plainVariable(n.Var)
		if name == "" || !importable(name) || external(name) || n.ByRef || n.Op.Kind != syntax.TEqual || !b.expr(n.Value, true) {
			return false
		}
		if !b.written[name] {
			b.proof.writes = append(b.proof.writes, name)
			if b.written == nil {
				b.written = map[string]bool{}
			}
			b.written[name] = true
		}
		return true
	case syntax.Expr:
		// A switch label also performs a loose comparison with its subject.
		if label, ok := n.Parent().(*syntax.Case); ok && label.Cond == n {
			return false
		}
		return b.expr(n, false)
	}
	return false
}

// Only boolean control operations accept arbitrary defined local values. Other
// operators need a fully literal proof, avoiding magic coercions, zero division
// and version-dependent arithmetic/type errors.
func (b *exceptionBuilder) expr(x syntax.Expr, scalar bool) bool {
	if b.fuel == 0 {
		return false
	}
	b.fuel--
	switch n := x.(type) {
	case *syntax.Paren:
		return b.expr(n.Expr, scalar)
	case *syntax.Literal:
		return true
	case *syntax.ConstFetch:
		return constantName(n) != ""
	case *syntax.Variable:
		if scalar || n.NameExpr != nil || !importable(n.Name) || external(n.Name) {
			return false
		}
		if !b.written[n.Name] {
			b.proof.reads = append(b.proof.reads, n.Name)
		}
		return true
	case *syntax.Unary:
		if n.Op.Kind == syntax.TExclaim {
			return b.expr(n.Expr, false)
		}
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor:
			return b.expr(n.Left, false) && b.expr(n.Right, false)
		}
	}
	fuel := b.fuel
	_, known := constantValue(x, MaxDepth*8, &fuel)
	b.fuel = fuel
	return known
}

func (p *exceptionProof) holds(f frame) bool {
	if p == nil || !p.safe {
		return false
	}
	for _, name := range p.reads {
		v, exists := f.vars[name]
		_, shared := f.aliased[name]
		if !exists || !v.Complete || !v.plain || shared || f.wild {
			return false
		}
	}
	for _, name := range p.writes {
		old, exists := f.vars[name]
		_, shared := f.aliased[name]
		if !p.local || shared || f.wild || !exists && !f.definite || exists && !storedScalar(old) {
			return false
		}
	}
	return true
}

// A stored scalar cannot invoke a destructor when replaced. Do not infer this
// from plain/deep alone: arrays can own nested objects. Intervals prove ints;
// literal and scalar-cast producers prove their actual stored result.
func storedScalar(v Value) bool {
	if !v.Complete {
		return false
	}
	if v.interval != nil {
		return true
	}
	switch n := syntax.UnwrapParens(v.Expr).(type) {
	case *syntax.Literal:
		return true
	case *syntax.ConstFetch:
		return constantName(n) != ""
	case *syntax.Unary:
		switch n.Op.Kind {
		case syntax.TIntCast, syntax.TDoubleCast, syntax.TStringCast, syntax.TBoolCast, syntax.TExclaim:
			return true
		}
	}
	return false
}

type ticksKey struct{}

// Ticks may invoke arbitrary callbacks between otherwise inert statements.
// Any declaration disables pruning for the file, including inherited lexical
// declarations around a function. This immutable syntax fact is shared.
func (e *Env) ticks() bool {
	return e.file.Memo(ticksKey{}, func() any {
		found := false
		syntax.InspectFile(e.file, func(n syntax.Node) bool {
			if d, ok := n.(*syntax.Declare); ok {
				for _, item := range d.Items {
					if item.Key != nil && strings.EqualFold(item.Key.Value, "ticks") {
						found = true
						break
					}
				}
			}
			return !found
		})
		return found
	}).(bool)
}
