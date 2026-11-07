package flow

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestArraySpreadCompleteness(t *testing.T) {
	for _, tc := range []struct {
		source   string
		complete bool
	}{
		{`probe([...[]]);`, true},
		{`$a=[1,2];probe([...$a]);`, true},
		{`probe([...[...[1]]]);`, true},
		{`probe([...['a'=>1], 'a'=>2]);`, true},
		{`probe([...unknown()]);`, false},
		{`$a=[1];$a[$key]=unknown();probe([...$a]);`, false},
		{`$a=[1];$b=&$a;unknown($b);probe([...$a]);`, false},
		{`$x=1;$a=[&$x];probe([...$a]);`, false},
		{`$a=[1];$a=unknown();probe([...$a]);`, false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, f := environment(t, `function probe($x){}`+tc.source, nil)
			if got := e.Value(argument(callsNamed(f, "probe")[0])).Complete; got != tc.complete {
				t.Fatalf("complete=%v want %v", got, tc.complete)
			}
		})
	}
}

func BenchmarkArraySpread(b *testing.B) {
	for b.Loop() {
		e, f := environment(b, `function probe($x){} $a=[1,2];probe([...[...$a],...$a]);`, nil)
		e.Value(argument(callsNamed(f, "probe")[0]))
	}
}

func TestArraySpreadVersionProof(t *testing.T) {
	for _, tc := range []struct {
		src     string
		version phpversion.Version
		want    bool
	}{
		{`[1]`, phpversion.PHP73, false},
		{`[1]`, phpversion.PHP74, true},
		{`['x'=>1]`, phpversion.PHP80, false},
		{`[...[1]]`, phpversion.PHP80, false},
		{`['x'=>1]`, phpversion.PHP81, true},
	} {
		e, f := environment(t, `function probe($x){} probe(`+tc.src+`);`, nil)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		e.types.PHP = tc.version
		if got := e.completeSpread(v); got != tc.want {
			t.Fatalf("%s at %v: %v", tc.src, tc.version, got)
		}
	}
}

func TestArraySpreadProvenance(t *testing.T) {
	for _, src := range []string{
		`probe([...[$_GET['v']]]);`,
		`$a=[$_GET['v']];probe([...$a]);`,
		`probe([...['x'=>$_GET['v']], 'x'=>'safe']);`,
	} {
		e, f := environment(t, `function probe($x){}`+src, nil)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		if !v.Complete || len(v.Sources) == 0 || v.Safe.Covers(Shell) {
			t.Fatalf("lost unsafe provenance: %s: %+v", src, v)
		}
	}
}
