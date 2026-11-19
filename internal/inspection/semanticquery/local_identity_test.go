package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

func TestOutputParameterCall(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`pcntl_waitpid($p,$status);probe($status);`, true},
		{`$r=pcntl_waitpid($p,$status);probe($status);`, true},
		{`pcntl_waitpid($p,$other);probe($status);`, false},
		{`if($flag){pcntl_waitpid($p,$status);}probe($status);`, false},
		{`pcntl_waitpid($p,$status);$status=10;probe($status);`, false},
		{`pcntl_waitpid($p,$status);mutate($status);probe($status);`, false},
		{`pcntl_waitpid($p,$status);pcntl_wifexited($status);probe($status);`, true},
		{`pcntl_waitpid($p,$status);$object->call($status);probe($status);`, false},
		{`pcntl_waitpid($p,$status);Other::call($status);probe($status);`, false},
		{`function check(){pcntl_waitpid($p,$status);probe($status);}`, true},
		{`function check(){pcntl_waitpid($p,$status);$status=2;probe($status);}`, false},
		{"pcntl_waitpid($p,$status);" + strings.Repeat("strlen(\"x\");", 300) + "probe($status);", false},
		{"pcntl_waitpid($p,$status);" + strings.Repeat("$noise=1;", 300) + "probe($status);", false},
		{`pcntl_waitpid($p,$status);function unrelated(){$status=2;}probe($status);`, true},
		{`pcntl_waitpid($p,$status);probe(10);`, false},
		{`other($p,$status);probe($status);`, false},
		{`pcntl_waitpid($p,$status);pcntl_waitpid($p,$status);probe($status);`, true},
		{`pcntl_waitpid($p,$status);unset($status);probe($status);`, false},
		{`pcntl_waitpid($p,$status);[$status]=[1];probe($status);`, false},
		{`pcntl_waitpid($p,$status);$status++;probe($status);`, false},
		{`pcntl_waitpid($p,$status);$status[0]=1;probe($status);`, false},
		{`pcntl_waitpid($p,$status);foreach($xs as $status){}probe($status);`, false},
		{`pcntl_waitpid($p,$status);$f=function()use(&$status){};probe($status);`, false},
		{`pcntl_waitpid($p,$status);$f=function()use($status){};probe($status);`, true},
		{`pcntl_waitpid($p,$status);while($x){probe($status);$status=1;}`, true},
		{`while($x){pcntl_waitpid($p,$status);probe($status);$status=1;}`, true},
		{`pcntl_waitpid($p,$status);while($x){probe($status);}$status=1;`, true},
		{`pcntl_waitpid($p,$status);probe($status);$status=1;`, true},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			got := OutputParameterCall(ctx, c, CallArgument(c.Args, 0, ""), []string{"pcntl_waitpid"}, 1, "status") != nil
			if got != tc.want {
				t.Fatalf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestSameStableExpr(t *testing.T) {
	probeNative(t, `probe($a,$a,$b);`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		a, b, d := CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, ""), CallArgument(c.Args, 2, "")
		if !SameStableExpr(ctx, a, b) || SameStableExpr(ctx, a, d) || SameStableExpr(ctx, nil, b) {
			t.Fatal("source identity mismatch")
		}
	})
}

func BenchmarkOutputParameterCall(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
			OutputParameterCall(ctx, c, CallArgument(c.Args, 0, ""), []string{"pcntl_waitpid"}, 1, "status")
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte(`<?php pcntl_waitpid($p,$status);probe($status);`)
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(syntax.Parse("bench.php", src, syntax.Options{}))
	}
}

func TestLocalIdentityProofs(t *testing.T) {
	for _, tc := range []struct {
		src  string
		same bool
	}{
		{"probe(1,1);", true},
		{"probe(1,2);", false},
		{"probe($a,$b);", false},
		{"function f($a){useValue($a);unset($a);probe($a);}", false},
		{"function f($a){useValue($a);$o->unknown($a);probe($a);}", false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			first := CallArgument(c.Args, 0, "a")
			second := CallArgument(c.Args, 1, "b")
			if second == nil {
				for _, ev := range ScopeCallEvents(ctx, c) {
					if prior, ok := ev.(*syntax.FuncCall); ok && ctx.Text(prior.Name) == "useValue" {
						second = CallArgument(prior.Args, 0, "")
					}
				}
			}
			if same := SameStableLocal(ctx, first, second); same != tc.same {
				t.Fatalf("%s: same=%v", tc.src, same)
			}
			if SameStableLocal(ctx, first, nil) {
				t.Fatal("nil proves identity")
			}
		})
	}
	probeNative(t, "probe(1);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		for _, x := range []syntax.Expr{nil, CallArgument(c.Args, 0, "")} {
			if _, ok := KnownByteLength(ctx, x); ok {
				t.Fatal("non-string has known length")
			}
		}
	})
}

func TestLocalIdentityBudgets(t *testing.T) {
	probeNative(t, "function f($a){"+strings.Repeat("other();", 513)+"probe($a,$a);}", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if SameStableLocal(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("exhausted scope establishes identity")
		}
	})
}

func BenchmarkMembersUnwrittenBefore(b *testing.B) {
	out := []string{}
	e, err := analysis.NewEngine([]analysis.Rule{extensionContractProbe{"pristine", &out}}, analysis.Config{Only: []string{"AbstractClassInstantiation"}})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte("<?php $x=new SimpleXMLElement('<r/>');strlen($x);"), syntax.Options{})
	e.Analyze(f)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out = out[:0]
		e.Analyze(f)
	}
}

func BenchmarkUnaliasedLocalAllocations(b *testing.B) {
	var source strings.Builder
	source.WriteString("<?php ")
	for i := range 100 {
		fmt.Fprintf(&source, "$v%d=new stdClass();strlen($v%d);", i, i)
	}
	out := []string{}
	e, err := analysis.NewEngine([]analysis.Rule{extensionContractProbe{"unaliased", &out}}, analysis.Config{Only: []string{"AbstractClassInstantiation"}})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte(source.String()), syntax.Options{})
	e.Analyze(f)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out = out[:0]
		e.Analyze(f)
	}
}

func BenchmarkLocalEvents(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) { LocalEvents(ctx, c) }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte("<?php " + strings.Repeat("$h=fopen($p,'r');probe($h);", 40))
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(syntax.Parse("bench.php", src, syntax.Options{}))
	}
}

func TestSameLocalObject(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$d=new DOMDocument();probe($d,$d);`, true},
		{`$d=new DOMDocument();$n=$d->createElement('x');probe($n,$n);`, true},
		{`probe($a,$b);`, false},
		{`$n=7;probe($n,$n);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := SameLocalObject(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}
