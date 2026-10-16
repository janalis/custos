package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

func TestAssertionFlowOption(t *testing.T) {
	for _, tc := range []struct {
		name, src        string
		enabled, guarded bool
	}{
		{"default", `$v=file_get_contents($p);assert($v!==false);probe($v);`, false, false},
		{"enabled", `$v=file_get_contents($p);assert($v!==false);probe($v);`, true, true},
		{"named", `$v=file_get_contents($p);assert(assertion:$v!==false);probe($v);`, true, true},
		{"alias", `use function assert as check;$v=file_get_contents($p);check($v!==false);probe($v);`, true, true},
		{"shadow", `namespace N;function assert($v){}$v=file_get_contents($p);assert($v!==false);probe($v);`, true, false},
		{"dynamic", `$v=file_get_contents($p);$fn='assert';$fn($v!==false);probe($v);`, true, false},
		{"spread", `$v=file_get_contents($p);assert(...[$v!==false]);probe($v);`, true, false},
		{"callable", `$v=file_get_contents($p);$fn=assert(...);probe($v);`, true, false},
		{"wrongname", `$v=file_get_contents($p);assert(other:$v!==false);probe($v);`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base, f := environment(t, tc.src, nil)
			env := New(f, base.types, nil, Options{Assertions: tc.enabled})
			if got := env.Excludes(argument(callsNamed(f, "probe")[0]), "false"); got != tc.guarded {
				t.Fatalf("guarded=%v", got)
			}
		})
	}
}

func TestOptionalAssertionArgumentEffects(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, src := range []string{
			`$v=$_GET['x'];assert(($v=htmlspecialchars($v))!=='');probe($v);`,
			`$v=$_GET['x'];assert(($v=(int)$v)>=0);probe($v);`,
		} {
			base, f := environment(t, src, nil)
			types := infer.NewEnv(f, base.types.Names, base.types.Index, phpversion.Max, infer.Options{Assertions: enabled})
			x := argument(callsNamed(f, "probe")[0])
			types.TypeOf(x) // prior type queries cannot supply a disabled assertion proof
			env := New(f, types, nil, Options{Assertions: enabled})
			if tainted := env.Tainted(x, HTML); tainted == enabled {
				t.Fatalf("enabled=%v tainted=%v: %s", enabled, tainted, src)
			}
		}
	}
	base, f := environment(t, `$v='safe';$fn='assert';$fn(($v=$_GET['x'])!=='');probe($v);`, nil)
	if !New(f, base.types, nil).Tainted(argument(callsNamed(f, "probe")[0]), HTML) {
		t.Fatal("dynamic assert arguments always evaluate")
	}
}

func TestAssertionSnapshotIdentity(t *testing.T) {
	off := NewSnapshot()
	on := NewSnapshotWithOptions(Options{Assertions: true})
	if off.Equal(on) {
		t.Fatal("empty snapshots with different assumptions must differ")
	}
	empty := &File{Path: "empty.php", Summaries: map[string]Summary{}, options: Options{Assertions: true}}
	held := NewSnapshot(empty)
	if empty.Current(off, on) || !empty.Current(on, on) || off.Holds(empty) || !held.Holds(empty) {
		t.Fatal("empty-file option identity lost")
	}
	if held.WithoutFile(empty.Path).Equal(off) {
		t.Fatal("removal lost enabled mode")
	}
	if !off.WithFile(empty).Equal(on) {
		t.Fatal("overlay did not replace incompatible mode")
	}
	f := syntax.Parse("source.php", []byte(`<?php function source(){ $v=$_GET['x'];assert(($v=htmlspecialchars($v))!=='');return $v;}`), syntax.Options{})
	base, _ := environment(t, `probe(1);`, nil)
	enabled := Extract(f, base.types.Index, phpversion.Max, on, Options{Assertions: true})
	source, _ := NewSnapshot(enabled).lookup("source")
	if source.Return.Safe&HTML == 0 {
		t.Fatal("enabled source not escaped")
	}
}

func TestLegacyAssertionClobbersScope(t *testing.T) {
	for _, tc := range []struct {
		condition string
		php       phpversion.Version
		complete  bool
	}{
		{`'$n=5'`, phpversion.PHP74, false},
		{`$n`, phpversion.PHP74, false},
		{`$n>0`, phpversion.PHP74, true},
		{`is_int($n)`, phpversion.PHP74, true},
		{`'$n=5'`, phpversion.PHP84, true},
	} {
		for _, enabled := range []bool{false, true} {
			base, f := environment(t, `function f(int $n){assert(`+tc.condition+`);probe($n);}`, nil)
			types := infer.NewEnv(f, base.types.Names, base.types.Index, tc.php, infer.Options{Assertions: enabled})
			env := New(f, types, nil, Options{Assertions: enabled})
			v := env.Value(argument(callsNamed(f, "probe")[0]))
			if v.Complete != tc.complete {
				t.Fatalf("condition=%s php=%v enabled=%v complete=%v", tc.condition, tc.php, enabled, v.Complete)
			}
		}
	}
}

func BenchmarkAssertionArguments(b *testing.B) {
	for _, enabled := range []bool{false, true} {
		name := "disabled"
		if enabled {
			name = "enabled"
		}
		b.Run(name, func(b *testing.B) {
			base, f := environment(b, `$v=$_GET['x'];assert(($v=htmlspecialchars($v))!=='');probe($v);`, nil)
			x := argument(callsNamed(f, "probe")[0])
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				New(f, base.types, nil, Options{Assertions: enabled}).Value(x)
			}
		})
	}
}

func TestAssertionConditionRewriteRemainsTainted(t *testing.T) {
	base, f := environment(t, `function f(int $x){assert(is_int($x)&&($x=$_GET['x']));probe($x);}`, nil)
	env := New(f, base.types, nil, Options{Assertions: true})
	if !env.Tainted(argument(callsNamed(f, "probe")[0]), HTML) {
		t.Fatal("assertion condition hid later raw rewrite")
	}
}

func TestAssertionSnapshotModeMismatchCannotSupplySafety(t *testing.T) {
	source := syntax.Parse("source.php", []byte(`<?php function source(){ $v=$_GET['x'];assert(($v=htmlspecialchars($v))!=='');return $v;}`), syntax.Options{})
	base, _ := environment(t, `probe(1);`, nil)
	base.types.Index.Add(index.Extract(source))
	enabled := NewSnapshot(Extract(source, base.types.Index, phpversion.Max, nil, Options{Assertions: true}))
	caller := syntax.Parse("caller.php", []byte(`<?php probe(source());`), syntax.Options{})
	types := infer.NewEnv(caller, names.New(caller), base.types.Index, phpversion.Max)
	value := argument(callsNamed(caller, "probe")[0])
	if New(caller, types, enabled, Options{Assertions: true}).Value(value).Safe&HTML == 0 {
		t.Fatal("matching enabled snapshot lost sanitizer")
	}
	if New(caller, types, enabled).Value(value).Safe&HTML != 0 {
		t.Fatal("incompatible enabled snapshot supplied sanitizer")
	}
}
