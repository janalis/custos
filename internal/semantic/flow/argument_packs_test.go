package flow

import (
	"fmt"
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestArgumentPacks(t *testing.T) {
	for _, tc := range []struct {
		name, decl, src string
		known           bool
		tainted         []int
	}{
		{"positional", "function target($x, $y) {}", "target(...['safe', $_GET['v']]);", true, []int{1}},
		{"named", "function target($x, $y) {}", "target(...['y'=>$_GET['v'], 'x'=>'safe']);", true, []int{1}},
		{"saved", "function target($x, $y) {}", "$a=['safe', $_GET['v']]; target(...$a);", true, []int{1}},
		{"empty named key", "function target($x) {}", "target(...[''=>$_GET['v']]);", false, nil},
		{"empty variadic key", "function target(...$x) {}", "target(...[''=>$_GET['v']]);", true, []int{0}},
		{"successive named packs", "function target($x,$y) {}", "target(...['y'=>$_GET['v']], ...['x'=>'safe']);", true, []int{1}},
		{"captured before later assignment", "function target($x,$y) {}", "$a=[$_GET['v']]; target(...$a, y: $a=[]);", true, []int{0}},
		{"empty", "function target() {}", "target(...[]);", true, nil},
		{"variadic", "function target($x, ...$rest) {}", "target('safe', 'also safe', $_GET['v']);", true, []int{1}},
		{"empty variadic", "function target(...$rest) {}", "target();", true, nil},
		{"named variadic", "function target(...$rest) {}", "target(extra: $_GET['v']);", true, []int{0}},
		{"mixed", "function target($x, ...$rest) {}", "target(...['safe'], extra: $_GET['v']);", true, []int{1}},
		{"unknown", "function target($x) {}", "target(...unknown());", false, nil},
		{"optional doc", "function target($x) {}", "/** @var array{0?: string} $a */ $a=unknown(); target(...$a);", false, nil},
		{"reference pack", "function target($x) {}", "$v=$_GET['v']; target(...[&$v]);", false, nil},
		{"reference parameter", "function target(&$x) {}", "target(...[$_GET['v']]);", false, nil},
		{"reference variadic", "function target(&...$x) {}", "target($_GET['v']);", false, nil},
		{"missing", "function target($x, $y) {}", "target(...['safe']);", false, nil},
		{"extra", "function target($x) {}", "target(...['safe', 'extra']);", false, nil},
		{"unknown named", "function target($x='safe') {}", "target(other: $_GET['v']);", false, nil},
		{"duplicate", "function target($x) {}", "target('safe', ...['x'=>$_GET['v']]);", false, nil},
		{"duplicate variadic", "function target(...$x) {}", "target(...['extra'=>'safe'], ...['extra'=>$_GET['v']]);", false, nil},
		{"numeric after named", "function target($x, ...$rest) {}", "target(...['x'=>'safe', 0=>$_GET['v']]);", false, nil},
		{"mutated", "function target($x) {}", "$a=['safe']; $a[$_GET['k']]=$_GET['v']; target(...$a);", false, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.decl+tc.src, nil)
			calls := callsNamed(f, "target")
			args, ok := e.BoundArguments(calls[len(calls)-1])
			if ok != tc.known {
				t.Fatalf("bound=%v want %v: %+v", ok, tc.known, args)
			}
			for i, arg := range args {
				want := false
				for _, p := range tc.tainted {
					want = want || p == i
				}
				if got := len(arg.Sources) > 0 && !arg.Safe.Covers(Shell); got != want {
					t.Errorf("argument %d taint %v want %v", i, got, want)
				}
			}
		})
	}
}

func TestArgumentPackVersions(t *testing.T) {
	for _, tc := range []struct {
		src     string
		version phpversion.Version
		known   bool
	}{
		{"target(...['safe']);", phpversion.PHP55, false},
		{"target(...['safe']);", phpversion.PHP56, true},
		{"target(...['x'=>'safe']);", phpversion.PHP74, false},
		{"target(...['x'=>'safe']);", phpversion.PHP80, true},
		{"target(x: 'safe');", phpversion.PHP74, false},
		{"target(...[], x: 'safe');", phpversion.PHP80, false},
		{"target(...[], x: 'safe');", phpversion.PHP81, true},
	} {
		e, f := environment(t, "function target($x){}"+tc.src, nil)
		e.types.PHP = tc.version
		if _, ok := e.BoundArguments(callsNamed(f, "target")[0]); ok != tc.known {
			t.Errorf("%s at %s: %v", tc.src, tc.version, ok)
		}
	}
}

func TestArgumentPackSummaryProvenance(t *testing.T) {
	src := `function forward($x) { return $x; }
 function variadic(...$xs) { foreach ($xs as $x) { system($x); } }
 probe(forward(...[$_GET['v']])); variadic('safe', $_GET['v']);`
	e, f := environment(t, src, nil)
	snapshot := NewSnapshot(Extract(f, e.types.Index, phpversion.Max, nil, Options{}))
	e = New(f, e.types, snapshot, Options{})
	if !e.Tainted(argument(callsNamed(f, "probe")[0]), Shell) {
		t.Fatal("unpacked return lost provenance")
	}
	sinks := e.Sinks(callsNamed(f, "variadic")[0])
	if len(sinks) != 1 || len(sinks[0].Argument.Sources) == 0 {
		t.Fatalf("variadic sinks: %+v", sinks)
	}
}

func TestArgumentPackBudgets(t *testing.T) {
	for _, n := range []int{65, 257} {
		var params, args strings.Builder
		for i := range n {
			if i > 0 {
				params.WriteByte(',')
				args.WriteByte(',')
			}
			fmt.Fprintf(&params, "$p%d", i)
			args.WriteString("1")
		}
		e, f := environment(t, "function target("+params.String()+"){} target("+args.String()+");", nil)
		if _, ok := e.BoundArguments(callsNamed(f, "target")[0]); ok {
			t.Fatal("oversized binding accepted")
		}
	}
}

func BenchmarkArgumentPacks(b *testing.B) {
	for _, src := range []string{`target('safe', $_GET['v']);`, `target(...['safe', $_GET['v']]);`, `target('safe', ...[$_GET['v']]);`} {
		b.Run(src, func(b *testing.B) {
			e, f := environment(b, "function target($x,...$rest){}"+src, nil)
			call := callsNamed(f, "target")[0]
			e.BoundArguments(call)
			b.ReportAllocs()
			for b.Loop() {
				e.BoundArguments(call)
			}
		})
	}
}

func TestExpandedArgumentBudget(t *testing.T) {
	pack := "[" + strings.Repeat("1,", 32) + "]"
	e, f := environment(t, "function target(...$x){} target(..."+strings.TrimSuffix(strings.Repeat(pack+",...", 9), ",...")+");", nil)
	if _, ok := e.BoundArguments(callsNamed(f, "target")[0]); ok {
		t.Fatal("expanded argument budget was ignored")
	}
}
