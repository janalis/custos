package flow

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/php/syntax"
	"custos/internal/testing/testbudget"
)

// nestedLoops is a function of `depth` loops inside one another over
// `locals` variables. Each loop hands request data one level out, so the
// fixed point needs a number of rounds that grows with the depth.
func nestedLoops(locals, depth int) string {
	var src strings.Builder
	src.WriteString("function f($c){")
	for i := range locals {
		fmt.Fprintf(&src, "$v%d=%d;", i, i)
	}
	for i := range depth {
		fmt.Fprintf(&src, "while($c > %d){$v%d=$v%d . 'x'; if($c){$v%d=$_GET['q%d'] ?? $v%d;}", i, i, i+1, i+2, i, i)
	}
	src.WriteString(strings.Repeat("}", depth))
	src.WriteString("$z=$_GET['z']; probe('tool ' . $z);}")
	return src.String()
}

// switching is a long function the graph does not cover: a `switch` leaves
// nothing provable, which must not cost more than a function without one.
func switching(locals, blocks int) string {
	return strings.Replace(longFunction(locals, blocks), "$z=$_GET['z'];", "switch($cfg){case 1: $z=1; break; default: $z=$_GET['z'];}", 1)
}

// The work is bounded in bindings copied, which does not depend on the
// machine; the wall-clock limit only catches a runaway.
func TestDeepLoopsAndUncoveredScopesStayFast(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		tainted   bool
		copies    int
		limit     time.Duration
	}{
		{"nested loops", nestedLoops(150, 20), true, 900000, 5 * time.Second},
		{"deeply nested loops", nestedLoops(150, 40), true, MaxCopied, 0},
		{"long function with a switch", switching(200, 300), false, 500000, 5 * time.Second},
		{"wide arrays", wideArrays(20, 30, 300), true, 40000, 5 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			start := time.Now()
			if got := e.Tainted(argument(callsNamed(f, "probe")[0]), Shell); got != tc.tainted {
				t.Errorf("tainted=%v", got)
			}
			if d := time.Since(start); tc.limit > 0 && d > testbudget.Of(tc.limit) {
				t.Errorf("flow took %s", d)
			}
			if r := e.scope(e.file.Stmts[0].(*syntax.Function)); r.copies > tc.copies {
				t.Errorf("%d bindings copied", r.copies)
			}
		})
	}
}

// wideArrays is a function over arrays of literal keys carried through many
// branches: the elements belong to the array values and cost nothing more.
func wideArrays(arrays, keys, blocks int) string {
	var src strings.Builder
	src.WriteString("function f($c){")
	for i := range arrays {
		fmt.Fprintf(&src, "$a%d=[", i)
		for k := range keys {
			fmt.Fprintf(&src, "'k%d'=>%d,", k, k)
		}
		src.WriteString("];")
	}
	for i := range blocks {
		fmt.Fprintf(&src, "if($c > %d){$a%d['k%d']=$c;}", i, i%arrays, i%keys)
	}
	src.WriteString("$a0['z']=$_GET['z']; probe('tool ' . $a0['z']); check($a1['k1']);}")
	return src.String()
}

func TestWideArraysKeepTheirElements(t *testing.T) {
	e, f := environment(t, wideArrays(20, 30, 300), nil)
	if e.Tainted(argument(callsNamed(f, "check")[0]), Shell) || !e.Value(argument(callsNamed(f, "check")[0])).Complete {
		t.Fatal("clean element lost")
	}
}

// A scope that ran out of budget stops there: it does not keep visiting
// blocks whose facts nobody will read.
func TestExhaustedScopesStopAtOnce(t *testing.T) {
	e, _ := environment(t, switching(300, 600), nil)
	r := e.scope(e.file.Stmts[0].(*syntax.Function))
	if r.complete || !r.spent || r.copies > MaxCopied/2 {
		t.Fatalf("complete=%v spent=%v copies=%d", r.complete, r.spent, r.copies)
	}
	e, _ = environment(t, nestedLoops(300, 70), nil)
	r = e.scope(e.file.Stmts[0].(*syntax.Function))
	if r.complete || r.copies > MaxCopied+MaxCopied/8 {
		t.Fatalf("complete=%v copies=%d", r.complete, r.copies)
	}
	// Branches inside one statement are charged like branches between them.
	e, f := environment(t, "function f($c){"+manyLocals(2000)+strings.Repeat("$r="+strings.Repeat("$c && ", 30)+"$c;", 60)+"probe($r);}", nil)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("copies inside an expression were not charged")
	}
}

func manyLocals(n int) string {
	var src strings.Builder
	for i := range n {
		fmt.Fprintf(&src, "$l%d=%d;", i, i)
	}
	return src.String()
}

func TestSourcesStayOrderedAndBounded(t *testing.T) {
	at := func(start uint32) Source {
		return Source{Kind: "_GET", Span: syntax.Span{Start: start, End: start + 4}, Parameter: -1}
	}
	a, b := []Source{at(1), at(5), at(9)}, []Source{at(3), at(5), at(11)}
	joined, whole := union(a, b)
	if !whole || len(joined) != 5 {
		t.Fatalf("joined=%v", joined)
	}
	for i := 1; i < len(joined); i++ {
		if !before(joined[i-1], joined[i]) {
			t.Fatalf("out of order: %v", joined)
		}
	}
	// A list joined with part of itself, or with nothing, is that list.
	for _, part := range [][]Source{nil, joined, joined[1:3], {at(11)}, append([]Source(nil), joined...)} {
		if same, whole := union(joined, part); !whole || &same[0] != &joined[0] || len(same) != len(joined) || cap(same) != len(same) {
			t.Fatalf("union with %v changed the list", part)
		}
	}
	if only, _ := union(nil, b); &only[0] != &b[0] || cap(only) != len(only) {
		t.Fatal("union with nothing copied the list")
	}
	// Two sources at one position differ by what else names them.
	first := Source{Kind: "parameter", Parameter: 0}
	for _, second := range []Source{{Kind: "parameter", Parameter: 1}, {Kind: "parameter", Path: "b.php"}, {Kind: "session"}, {Kind: "parameter", Span: syntax.Span{End: 2}}} {
		if !before(first, second) || before(second, first) || before(first, first) {
			t.Fatalf("order of %+v and %+v", first, second)
		}
	}
	var many, more []Source
	for i := range uint32(MaxSources) {
		many = append(many, at(2*i))
		more = append(more, at(2*i+1))
	}
	if capped, whole := union(many, more); whole || len(capped) != MaxSources {
		t.Fatalf("%d sources, whole=%v", len(capped), whole)
	}
	// A full list still changes when an earlier source displaces a later one.
	full := frame{vars: map[string]Value{"x": {Sources: more}}, states: map[uint32]State{}}
	if changed, _ := full.absorb(frame{vars: map[string]Value{"x": {Sources: many}}}); !changed || full.vars["x"].Sources[0] != many[0] {
		t.Fatalf("full list unchanged: %+v", full.vars["x"])
	}
}

// The summaries of a file depend on a snapshot only through the contracts
// read while computing them.
func TestSummariesKnowTheContractsTheyUsed(t *testing.T) {
	e, f := environment(t, `function wrap($c){return inner($c);} function plain($c){return strlen($c);} function inner($c){return 1;}`, nil)
	inner := func(kind string) *Snapshot {
		return NewSnapshot(&File{Path: "inner.php", Summaries: map[string]Summary{"inner": {Return: Value{Complete: true, Sources: []Source{{Kind: kind, Parameter: -1}}}}}})
	}
	env, own := Summarize(f, e.types, inner("_GET"))
	if !own.consulted["inner"] || env.consulted != nil || len(own.Summaries) != 3 {
		t.Fatalf("consulted=%v summaries=%v", own.consulted, own.Summaries)
	}
	if !tainted(own.Summaries["wrap"].Return) {
		t.Fatalf("wrap=%+v", own.Summaries["wrap"])
	}
	// The environment answers for the file without analysing it again.
	if len(env.scopes) == 0 || len(env.Calls(f.Stmts[0].(*syntax.Function))) != 1 {
		t.Fatal("environment lost its scopes")
	}
	for _, tc := range []struct {
		name          string
		earlier, next *Snapshot
		current       bool
	}{
		{"same contract", inner("_GET"), inner("_GET"), true},
		{"changed contract", inner("_GET"), inner("_POST"), false},
		{"new contract", NewSnapshot(), inner("_GET"), false},
		{"removed contract", inner("_GET"), nil, false},
		{"unrelated contract", inner("_GET"), NewSnapshot(inner("_GET").files["inner.php"], &File{Path: "o.php", Summaries: map[string]Summary{"other": {}}}), true},
	} {
		if got := own.Current(tc.earlier, tc.next); got != tc.current {
			t.Errorf("%s: current=%v", tc.name, got)
		}
	}
	base := NewSnapshot(own)
	other := &File{Path: own.Path, Summaries: map[string]Summary{"wrap": {}}}
	for _, tc := range []struct {
		name     string
		snapshot *Snapshot
		file     *File
		holds    bool
	}{
		{"no snapshot", nil, own, false},
		{"same summaries", base, own, true},
		{"other summaries", base, other, false},
		{"unknown file", base, &File{Path: "new.php"}, false},
		{"overlaid file", base.WithFile(other), other, true},
		{"file under an overlay", base.WithFile(&File{Path: "new.php"}), own, true},
		{"replaced under an overlay", base.WithFile(other), own, false},
	} {
		if got := tc.snapshot.Holds(tc.file); got != tc.holds {
			t.Errorf("%s: holds=%v", tc.name, got)
		}
	}
}

func BenchmarkNestedLoops(b *testing.B) {
	e, f := environment(b, nestedLoops(150, 20), nil)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if !New(f, e.types, nil).Tainted(x, Shell) {
			b.Fatal("request data lost")
		}
	}
}

func BenchmarkSwitchFunction(b *testing.B) {
	e, f := environment(b, switching(200, 300), nil)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if New(f, e.types, nil).Tainted(x, Shell) {
			b.Fatal("uncovered scope proved a fact")
		}
	}
}
