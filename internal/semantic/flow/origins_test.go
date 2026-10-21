package flow

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func producerSwitch(count int, producer, guard func(int) string) string {
	var src strings.Builder
	src.WriteString(`function f($choice,$input,$pattern){switch($choice){`)
	for i := range count {
		if i == count-1 {
			src.WriteString("default:")
		} else {
			fmt.Fprintf(&src, "case %d:", i)
		}
		fmt.Fprintf(&src, "$v=%s;", producer(i))
		if guard != nil {
			src.WriteString(guard(i))
		}
		src.WriteString("break;")
	}
	src.WriteString(`}probe($v);}`)
	return src.String()
}

func TestProducerRetentionContracts(t *testing.T) {
	for _, count := range []int{5, 64, 65, 70} {
		e, f := environment(t, producerSwitch(count, func(int) string { return `preg_replace($pattern,'x',$input)` }, nil), nil)
		x := argument(callsNamed(f, "probe")[0])
		origins, single := e.Reaching(x)
		want := count
		if count > MaxOrigins {
			want = 1
		}
		if single || len(origins) != want || e.CompressedOrigins(x) != (count > MaxOrigins) {
			t.Fatalf("count=%d origins=%d single=%v compressed=%v", count, len(origins), single, e.CompressedOrigins(x))
		}
	}
	for _, producer := range []string{`project($input)`, `parse_url($input)`, `fopen($input,'r')`, `gzdecode($input)`, `$input->read()`} {
		e, f := environment(t, producerSwitch(70, func(int) string { return producer }, nil), nil)
		x := argument(callsNamed(f, "probe")[0])
		origins, _ := e.Reaching(x)
		if len(origins) != 70 || e.CompressedOrigins(x) {
			t.Fatalf("sensitive/unknown %s lost exact producers: %d", producer, len(origins))
		}
	}
}

func TestProducerRetentionGuardStates(t *testing.T) {
	e, f := environment(t, producerSwitch(70, func(int) string { return `preg_replace($pattern,'x',$input)` }, func(i int) string {
		if i == 69 {
			return ""
		}
		return `if($v===null){return;}`
	}), nil)
	x := argument(callsNamed(f, "probe")[0])
	origins, _ := e.Reaching(x)
	guarded, unguarded := false, false
	for _, o := range origins {
		guarded = guarded || o.NonNull
		unguarded = unguarded || !o.NonNull
	}
	if !guarded || !unguarded {
		t.Fatalf("late unguarded path disappeared: %+v", origins)
	}
	if e.ProducerBudgetExceeded(x) {
		t.Fatal("contract compression exhausted host budget")
	}
}

func TestNonCallableProducerAmbiguity(t *testing.T) {
	e, f := environment(t, producerSwitch(70, func(i int) string { return fmt.Sprint(i) }, nil), nil)
	x := argument(callsNamed(f, "probe")[0])
	origins, single := e.Reaching(x)
	if len(origins) != 0 || single || !e.CompressedOrigins(x) {
		t.Fatal("non-call path marker lost ambiguity")
	}
	if _, known := e.Resolve(x); known {
		t.Fatal("compressed literals established an exact value")
	}
}

func TestOriginCellsChargedOnce(t *testing.T) {
	e, f := environment(t, `probe(file_get_contents($path));probe(stream_get_contents($handle));`, nil)
	calls := callsNamed(f, "probe")
	a, b := argument(calls[0]), argument(calls[1])
	r := &scopeResult{complete: true, originTypes: e.types}
	set := origins(Value{Expr: a}, Value{Expr: b}, r)
	if r.retained != 3 {
		t.Fatalf("initial retained=%d", r.retained)
	}
	rebuilt := finishOrigins(&originSet{items: append([]Origin(nil), set.items...)}, r, set)
	if rebuilt != set || r.retained != 3 {
		t.Fatal("identical publication allocated retained cells")
	}

	again := origins(Value{origins: set}, Value{Expr: a}, r)
	if again != set || r.retained != 3 {
		t.Fatal("shared origin set charged again")
	}
	guarded := set.excluding(true, false, r)
	if r.retained != 6 || guarded.excluding(true, false, r) != guarded || r.retained != 6 {
		t.Fatal("unchanged guard charged origin cells")
	}
	raw := guarded.unguarded(r)
	if r.retained != 9 || raw.unguarded(r) != raw || r.retained != 9 {
		t.Fatal("unchanged invalidation charged origin cells")
	}
	original := frame{owner: r, vars: map[string]Value{"v": {origins: raw}}, states: map[uint32]State{}}
	copy := clone(&original)
	if copy.vars["v"].origins != raw || r.retained != 9 {
		t.Fatal("frame copy charged shared origins")
	}
	r.retained = MaxRetained - 1
	if next := origins(Value{Expr: a}, Value{Expr: b}, r); !next.full || !r.spent || r.complete {
		t.Fatal("origin exhaustion did not invalidate scope")
	}
}

func BenchmarkProducerRetention(b *testing.B) {
	for _, count := range []int{5, 64, 70} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			e, f := environment(b, producerSwitch(count, func(int) string { return `preg_replace($pattern,'x',$input)` }, nil), nil)
			x := argument(callsNamed(f, "probe")[0])
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				New(f, e.types, nil, Options{}).Reaching(x)
			}
		})
	}
}

func TestProducerHostBudgetExhaustion(t *testing.T) {
	e, f := environment(t, producerSwitch(1300, func(int) string { return `project($input)` }, nil), nil)
	x := argument(callsNamed(f, "probe")[0])
	if !e.ProducerBudgetExceeded(x) {
		t.Fatal("individually retained producers exceeded no host budget")
	}
	if origins, _ := e.Reaching(x); len(origins) != 0 {
		t.Fatal("exhausted scope exported producers")
	}
	if e.CompressedOrigins(x) {
		t.Fatal("host exhaustion confused with contract compression")
	}
}

func TestCompressedProducerStatesAndOrder(t *testing.T) {
	e, f := environment(t, producerSwitch(70, func(int) string { return `file_get_contents($input)` }, nil), nil)
	calls := callsNamed(f, "file_get_contents")
	build := func(reverse bool) (*originSet, *scopeResult) {
		r := &scopeResult{complete: true, originTypes: e.types}
		var value Value
		for at := range len(calls) {
			i := at
			if reverse {
				i = len(calls) - 1 - at
			}
			next := Value{Expr: calls[i], NonFalse: i&1 != 0, NonNull: i&2 != 0}
			joined := origins(value, next, r)
			value = Value{origins: joined}
		}
		return value.origins, r
	}
	forward, r := build(false)
	reverse, _ := build(true)
	if len(forward.items) != 4 || !sameOrigins(forward, reverse) {
		t.Fatalf("guard states/order differ: %+v %+v", forward.items, reverse.items)
	}
	guarded := forward.excluding(true, false, r)
	if len(guarded.items) != 2 {
		t.Fatalf("guard did not normalize states: %+v", guarded.items)
	}
	for _, o := range guarded.items {
		if !o.NonFalse {
			t.Fatal("guard state lost")
		}
	}
	raw := guarded.unguarded(r)
	if len(raw.items) != 1 || raw.items[0].Expr != calls[0] || raw.items[0].NonFalse || raw.items[0].NonNull {
		t.Fatal("invalidation lost canonical unguarded witness")
	}
}

func TestCompressedProducerSummariesReleaseSyntax(t *testing.T) {
	src := strings.Replace(producerSwitch(70, func(int) string { return `file_get_contents($input)` }, nil), `probe($v);}`, `system($v);return $v;}`, 1)
	e, f := environment(t, src, nil)
	summary := Extract(f, e.types.Index, e.types.PHP, nil, Options{}).Summaries["f"]
	if summary.Return.Expr != nil || summary.Return.origins != nil || len(summary.Sinks) == 0 {
		t.Fatal("compressed return syntax escaped")
	}
	for _, sink := range summary.Sinks {
		if sink.Argument.Expr != nil || sink.Argument.origins != nil {
			t.Fatal("compressed sink syntax escaped")
		}
	}
}

func TestOriginPublisherPreservesCoLocatedProofStates(t *testing.T) {
	e, f := environment(t, `probe(file_get_contents($input)->read());probe(@project($input));`, nil)
	outer := argument(callsNamed(f, "probe")[0]).(*syntax.MethodCall)
	inner := outer.Var
	r := &scopeResult{complete: true, originTypes: e.types}
	compact := finishOrigins(&originSet{compressed: true, items: []Origin{
		{Expr: outer}, {Expr: inner, NonFalse: true, NonNull: true}, {Expr: inner, NonFalse: true}, {Expr: inner, NonNull: true}, {Expr: inner},
	}}, r)
	if len(compact.items) != 5 || compact.items[4].Expr != outer {
		t.Fatal("co-located call identities or proof states were lost")
	}
	wrapped := argument(callsNamed(f, "probe")[1])
	if !callOrigin(wrapped) {
		t.Fatal("error suppression hid opaque call origin")
	}
	unowned := finishOrigins(&originSet{compressed: true, items: []Origin{{Expr: wrapped}, {Expr: inner}}}, nil)
	if len(unowned.items) != 2 {
		t.Fatal("missing semantic owner compressed unresolved contracts")
	}
	exhausted := &scopeResult{complete: true, originTypes: e.types, retained: MaxRetained}
	if exhausted.producers(Value{Expr: inner}) != nil || !exhausted.spent {
		t.Fatal("producer export escaped its exhausted owner")
	}
}

func TestOriginPublicationCapacity(t *testing.T) {
	_, f := environment(t, `probe(file_get_contents($path));probe(stream_get_contents($handle));probe(preg_replace($p,$r,$s));`, nil)
	calls := callsNamed(f, "probe")
	items := make([]Origin, 3, 10)
	for i := range items {
		items[i] = Origin{Expr: argument(calls[i])}
	}
	published := finishOrigins(&originSet{items: items}, nil)
	if cap(published.items) != len(published.items) {
		t.Fatal("publication retained unused builder capacity")
	}
	exact := make([]Origin, len(published.items))
	copy(exact, published.items)
	reused := finishOrigins(&originSet{items: exact}, nil)
	if &reused.items[0] != &exact[0] {
		t.Fatal("exact-capacity publication copied immutable builder")
	}
}
