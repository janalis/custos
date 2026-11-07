package infer

import (
	"reflect"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// maxFixpoints bounds the recursive inferences an Env re-runs under an
// assumption; beyond it a cycle stays unknown and is reported as before.
const maxFixpoints = 64

type undoKind uint8

const (
	undoCache undoKind = iota
	undoBase
	undoChain
	undoBody
	undoProp
	undoCapture
	undoAssert
	undoGeneric
	undoFunction
	undoBuiltin
)

// undo names one type-dependent cache entry written while a type query was
// active, so that results derived from a recursion assumption can be retracted.
type undo struct {
	kind undoKind
	expr syntax.Expr
	call *syntax.FuncCall
	v    *syntax.Variable
	span syntax.Span
	text string
}

// recursionHit is a read of an expression whose type was being inferred.
type recursionHit struct {
	head syntax.Expr
	span syntax.Span
}

func (e *Env) remember(u undo) {
	if e.typing > 0 {
		e.journal = append(e.journal, u)
	}
}

func (e *Env) rollback(mark int) {
	for i := len(e.journal) - 1; i >= mark; i-- {
		u := e.journal[i]
		switch u.kind {
		case undoCache:
			delete(e.cache, u.expr)
		case undoBase:
			delete(e.bases, u.v)
		case undoChain:
			delete(e.chains, u.expr)
		case undoBody:
			delete(e.bodies, u.span)
		case undoProp:
			delete(e.props, u.span)
		case undoCapture:
			delete(e.captures, captureKey{u.expr, u.text})
		case undoAssert:
			delete(e.asserts, u.expr)
		case undoGeneric:
			delete(e.generics, u.text)
		case undoFunction:
			delete(e.functions, u.call)
		default:
			delete(e.builtinCalls, u.call)
		}
	}
	e.journal = e.journal[:mark]
}

// recursion records a read of x while x is being inferred. Under a fixpoint
// assumption the read sees the assumed type; otherwise it is unknown and the
// gap is reported unless the cycle later converges.
func (e *Env) recursion(x syntax.Expr) (types.Type, bool) {
	if t, ok := e.assumed[x]; ok {
		return t, true
	}
	if e.heads == nil {
		e.heads = map[syntax.Expr]bool{}
	}
	e.heads[x] = true
	e.hits = append(e.hits, recursionHit{head: x, span: x.Span()})
	return types.Unknown, false
}

// fixpoint resolves a recursion headed by x, whose first inference read x
// as unknown. Re-inferring x with that first type assumed and getting it
// back proves it covers every value the cycle can produce; the reads made
// under the assumption are then kept and the gap is not reported. Otherwise
// every result derived from the assumption is retracted and x is inferred as
// before. Nested heads inside a re-inference keep the plain behavior.
func (e *Env) fixpoint(x syntax.Expr, first types.Type, journal, hits int) types.Type {
	delete(e.heads, x)
	if first.IsUnknown() || e.fixing > 0 || e.fixpoints >= maxFixpoints {
		e.settle(x, hits, false)
		return first
	}
	e.fixpoints++
	e.fixing++
	defer func() { e.fixing-- }()
	e.rollback(journal)
	e.hits = e.hits[:hits]
	if e.assumed == nil {
		e.assumed = map[syntax.Expr]types.Type{}
	}
	e.assumed[x] = first
	seeded := e.inferNullsafe(x)
	delete(e.assumed, x)
	delete(e.heads, x)
	if reflect.DeepEqual(seeded, first) {
		e.settle(x, hits, true)
		return seeded
	}
	e.rollback(journal)
	e.hits = e.hits[:hits]
	plain := e.inferNullsafe(x)
	delete(e.heads, x)
	e.settle(x, hits, false)
	return plain
}

// settle reports or drops the recursion hits on x since mark; hits on other
// expressions still being inferred stay pending for them.
func (e *Env) settle(x syntax.Expr, mark int, converged bool) {
	kept := e.hits[:mark]
	for _, hit := range e.hits[mark:] {
		switch {
		case hit.head != x:
			kept = append(kept, hit)
		case !converged:
			e.noteGap(gapRecursion, hit.span)
		}
	}
	e.hits = kept
}
