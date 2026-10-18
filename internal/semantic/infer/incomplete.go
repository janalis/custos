package infer

import (
	"cmp"
	"slices"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// Gap identifies a limit encountered by an evaluated type query. It retains no AST.
type Gap struct {
	Span   syntax.Span
	Reason string
}

type inferenceGapKind uint8

const (
	gapRecursion inferenceGapKind = iota
	gapInferenceBudget
	gapArgumentBudget
)

func (e *Env) gapRoot() *Env {
	if e.gapOwner != nil {
		return e.gapOwner
	}
	return e
}

// noteGap keeps one location per fixed reason: storage is bounded by the three
// internal identifiers, independent of the number of queries or derived Envs.
func (e *Env) noteGap(kind inferenceGapKind, span syntax.Span) {
	root := e.gapRoot()
	if root.gaps == nil {
		root.gaps = map[inferenceGapKind]syntax.Span{}
	}
	prior, found := root.gaps[kind]
	if !found || span.Start < prior.Start || span.Start == prior.Start && span.End < prior.End {
		root.gaps[kind] = span
	}
}

// Gaps lists encountered inference limits in source order. Native and option
// variants share the originating environment's status, but keep separate caches.
func (e *Env) Gaps() []Gap {
	root := e.gapRoot()
	var out []Gap
	names := [...]string{"inference-recursion", "inference-budget", "argument-budget"}
	for kind, span := range root.gaps {
		out = append(out, Gap{Span: span, Reason: names[kind]})
	}
	slices.SortFunc(out, func(a, b Gap) int {
		if order := cmp.Compare(a.Span.Start, b.Span.Start); order != 0 {
			return order
		}
		if order := cmp.Compare(a.Span.End, b.Span.End); order != 0 {
			return order
		}
		return cmp.Compare(a.Reason, b.Reason)
	})
	return out
}

func (e *Env) bindingBudget(args *syntax.ArgList, params []index.Param) bool {
	if len(params) <= maxBoundParams && (args == nil || len(args.Args) <= maxWrittenArgs) {
		return false
	}
	span := syntax.Span{}
	if args != nil {
		span = args.Span()
	}
	e.noteGap(gapArgumentBudget, span)
	return true
}

func (e *Env) boundArgument(args *syntax.ArgList, params []index.Param, target int) (syntax.Expr, argumentBindingState) {
	if e.bindingBudget(args, params) {
		return nil, bindingUnknown
	}
	return bindArgument(args, params, target, e.PHP)
}
