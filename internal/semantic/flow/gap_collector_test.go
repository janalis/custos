package flow

import (
	"fmt"
	"reflect"
	"testing"

	"custos/internal/php/syntax"
)

func TestGapCollectorBounds(t *testing.T) {
	for _, count := range []int{0, 1, 127, 128, 129, 10000} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var forward, reverse gapCollector
			gap := func(i int) Gap {
				return Gap{Span: syntax.Span{Start: uint32(i), End: uint32(i + 1)}, Reason: "dynamic-code"}
			}
			for i := range count {
				forward.add(gap(i))
				forward.add(gap(i))
				reverse.add(gap(count - 1 - i))
			}
			if !reflect.DeepEqual(forward, reverse) {
				t.Fatal("order affected retained status")
			}
			want := min(count, maxGaps)
			if len(forward.items) != want || cap(forward.items) > maxGaps || forward.overflow != (count > maxGaps) {
				t.Fatalf("items=%d cap=%d overflow=%v", len(forward.items), cap(forward.items), forward.overflow)
			}
			for i, g := range forward.items {
				if g != gap(i) {
					t.Fatalf("gap %d: %+v", i, g)
				}
			}
		})
	}
}

func TestGapCollectorTies(t *testing.T) {
	a := Gap{Span: syntax.Span{Start: 1, End: 2}, Reason: "a"}
	b := Gap{Span: syntax.Span{Start: 1, End: 2}, Reason: "b"}
	c := Gap{Span: syntax.Span{Start: 1, End: 3}, Reason: "a"}
	var collector gapCollector
	for _, gap := range []Gap{c, b, a, b, c, a} {
		collector.add(gap)
	}
	if !reflect.DeepEqual(collector.items, []Gap{a, b, c}) {
		t.Fatal(collector.items)
	}
}

func BenchmarkGapCollector(b *testing.B) {
	for _, count := range []int{128, 10000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var collector gapCollector
				for i := range count {
					collector.add(Gap{Span: syntax.Span{Start: uint32(i)}, Reason: "dynamic-code"})
				}
			}
		})
	}
}

func TestGapBudgetFallback(t *testing.T) {
	for _, kind := range []string{"transfer", "retention"} {
		for _, fileScope := range []bool{false, true} {
			t.Run(fmt.Sprint(kind, fileScope), func(t *testing.T) {
				env, file := environment(t, `function example(){}`, nil)
				result := &scopeResult{complete: true}
				if kind == "transfer" {
					reserve(result, MaxTransfers+1)
				} else {
					retain(result, MaxRetained+1)
				}
				var scope syntax.Node
				if !fileScope {
					scope = file.Stmts[0]
				}
				env.scopes[scope] = result
				gaps := env.Gaps()
				span := file.Span()
				if scope != nil {
					span = scope.Span()
				}
				if len(gaps) != 1 || gaps[0] != (Gap{Span: span, Reason: "flow-budget"}) {
					t.Fatal(gaps)
				}
			})
		}
	}
}

func TestGapKeepsEarliestReason(t *testing.T) {
	var result scopeResult
	result.gap("dynamic-code", syntax.Span{Start: 6})
	result.gap("dynamic-code", syntax.Span{Start: 3})
	result.gap("dynamic-code", syntax.Span{Start: 9})
	if got := result.gaps["dynamic-code"].Start; got != 3 {
		t.Fatal(got)
	}
}

func BenchmarkFlowGapRecording(b *testing.B) {
	var result scopeResult
	span := syntax.Span{Start: 1, End: 2}
	result.gap("unresolved-dispatch", span)
	b.ReportAllocs()
	for b.Loop() {
		result.gap("unresolved-dispatch", span)
	}
}
