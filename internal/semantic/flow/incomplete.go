package flow

import (
	"cmp"
	"slices"

	"custos/internal/php/syntax"
)

// Gap describes a bounded analysis that could not establish its full contract.
// It contains source coordinates only and never retains syntax trees.
type Gap struct {
	Span   syntax.Span
	Reason string
}

func (r *scopeResult) gap(reason string, span syntax.Span) {
	if r.gaps == nil {
		r.gaps = map[string]syntax.Span{}
	}
	if prior, ok := r.gaps[reason]; !ok || span.Start < prior.Start {
		r.gaps[reason] = span
	}
}

const maxGaps = 128

// gapCollector retains the earliest unique gaps without allocating storage for
// every analyzed scope. The marker is added only after an actual overflow.
type gapCollector struct {
	items    []Gap
	overflow bool
}

func compareGap(a, b Gap) int {
	if order := cmp.Compare(a.Span.Start, b.Span.Start); order != 0 {
		return order
	}
	if order := cmp.Compare(a.Span.End, b.Span.End); order != 0 {
		return order
	}
	return cmp.Compare(a.Reason, b.Reason)
}

func (c *gapCollector) add(gap Gap) {
	position, found := slices.BinarySearchFunc(c.items, gap, compareGap)
	if found {
		return
	}
	if len(c.items) == maxGaps {
		c.overflow = true
		if position == maxGaps {
			return
		}
		copy(c.items[position+1:], c.items[position:len(c.items)-1])
		c.items[position] = gap
		return
	}
	if c.items == nil {
		c.items = make([]Gap, 0, maxGaps)
	}
	c.items = append(c.items, Gap{})
	copy(c.items[position+1:], c.items[position:len(c.items)-1])
	c.items[position] = gap
}

// Gaps lists incompleteness encountered by the queries actually evaluated.
// It does not force scopes that no enabled inspection needs to analyze.
func (e *Env) Gaps() []Gap {
	var collector gapCollector
	for scope, result := range e.scopes {
		if !result.complete && len(result.gaps) == 0 {
			span := e.file.Span()
			if scope != nil {
				span = scope.Span()
			}
			collector.add(Gap{Span: span, Reason: "flow-budget"})
		}
		for reason, span := range result.gaps {
			collector.add(Gap{Span: span, Reason: reason})
		}
	}
	if collector.overflow {
		collector.items[maxGaps-1] = Gap{Span: e.file.Span(), Reason: "status-limit"}
	}
	return collector.items
}
