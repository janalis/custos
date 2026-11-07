package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestClosureResourceProjection(t *testing.T) {
	for _, tc := range []struct {
		source string
		closed bool
	}{
		{`$open=fn()=>fopen('x','rb');$h=$open();fclose($h);probe($h);`, true},
		{`$open=fn()=>tmpfile();$h=$open();fclose($h);probe($h);`, true},
		{`$open=fn()=>fopen('x','rb');$a=$open();$b=$open();fclose($a);probe($b);`, false},
		{`$h=fopen('x','rb');$peek=fn()=>$h;fclose($h);$alias=$peek();probe($alias);`, true},
		{`$h=fopen('x','rb');$pass=fn($x)=>$x;fclose($h);$alias=$pass($h);probe($alias);`, true},
		{`$h=fopen('x','rb');$close=fn()=>fclose($h);$close();probe($h);`, false},
		{`$h=fopen('x','rb');$close=function()use($h){if(!fclose($h)){return;}return;};$close();probe($h);`, false},
		{`$open=function(){$h=fopen('x','rb');fclose($h);return $h;};$h=$open();probe($h);`, false},
		{`$open=function(){$h=fopen('x','rb');if(!fclose($h)){throw new Exception;}return $h;};$h=$open();probe($h);`, true},
		{`$open=fn()=>new stdClass;$h=$open();probe($h);`, false},
		{`$open=function(){$h=fopen('x','rb');fclose($h);if($h){return $h;}return $h;};$h=$open();probe($h);`, false},

		{`$h=fopen('x','rb');fclose($h);$peek=function()use($h){$h=tmpfile();return $h;};$other=$peek();probe($h);`, true},
		{`$h=fopen('x','rb');fclose($h);$peek=function()use($h){$h=tmpfile();return $h;};$other=$peek();probe($other);`, false},
	} {
		e, f := summarized(t, tc.source)
		probe := callsNamed(f, "probe")[0]
		x := argument(probe)
		_, closed := e.StateBefore(probe, x, "fclose")
		if closed != tc.closed {
			t.Fatalf("%s: value=%+v closed=%v", tc.source, e.Value(x), closed)
		}
	}
}

func TestClosureResourceConservativeIdentities(t *testing.T) {
	// Foreign allocation identities never become caller resources.
	at := &syntax.FuncCall{}
	result := Value{Identity: 9}
	projectClosureResources(at, &result, nil, &frame{}, &scopeResult{}, &scopeResult{})
	if result.Identity != 0 {
		t.Fatal("foreign identity escaped")
	}
}

func BenchmarkClosureResources(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		e, f := summarized(b, `$open=fn()=>tmpfile();$a=$open();$b=$open();fclose($a);probe($b);`)
		probe := callsNamed(f, "probe")[0]
		e.StateBefore(probe, argument(probe), "fclose")
	}
}
