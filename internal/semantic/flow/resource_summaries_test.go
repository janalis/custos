package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func TestNamedResourceContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		closed bool
	}{
		{`function openIt(){return tmpfile();}$h=openIt();fclose($h);probe($h);`, true},
		{`final class R{static function open(){return tmpfile();}}$cb=R::open(...);$h=$cb();fclose($h);probe($h);`, true},
		{`function openIt():mixed{return tmpfile();}$h=openIt();fclose($h);probe($h);`, true},
		{`function openIt(){return tmpfile();}$a=openIt();$b=openIt();fclose($a);probe($b);`, false},
		{`function pass($h){return $h;}$h=tmpfile();fclose($h);$a=pass($h);probe($a);`, true},
		{`function pass($h){$copy=$h;return $copy;}$h=tmpfile();fclose($h);$a=pass($h);probe($a);`, true},
		{`function pass($h):mixed{return $h;}$h=tmpfile();fclose($h);$a=pass($h);probe($a);`, true},
		{`function pass($h){return $h;}function caller($h){fclose($h);$a=pass($h);probe($a);}`, true},
		{`function openIt(){return tmpfile();}function wrap(){return openIt();}$h=wrap();fclose($h);probe($h);`, true},
		{`function openIt(){return tmpfile();}$cb=openIt(...);$h=$cb();fclose($h);probe($h);`, true},
		{`function openIt():int{return (int)tmpfile();}$h=openIt();fclose($h);probe($h);`, false},
		{`function openIt(){return (int)tmpfile();}$h=openIt();fclose($h);probe($h);`, false},
		{`function &openIt(){$h=tmpfile();return $h;}$h=openIt();fclose($h);probe($h);`, false},
		{`function pass(array $h){return $h;}$h=tmpfile();fclose($h);$a=pass($h);probe($a);`, false},
		{`function pass($h){opaque($h);return $h;}$h=tmpfile();fclose($h);$a=pass($h);probe($a);`, false},
		{`function openIt(){$h=tmpfile();fclose($h);return $h;}$h=openIt();probe($h);`, false},
		{`function openIt(){$h=tmpfile();if(!fclose($h)){throw new Exception;}return $h;}$h=openIt();probe($h);`, true},
	} {
		e, f := summarized(t, tc.source)
		call := callsNamed(f, "probe")[0]
		x := argument(call)
		if _, closed := e.StateBefore(call, x, "fclose"); closed != tc.closed {
			t.Fatalf("%s: %+v closed=%v", tc.source, e.Value(x), closed)
		}
	}
}

func TestCrossFileResourceDependency(t *testing.T) {
	definitions := `<?php function openIt(){return tmpfile();}`
	ix := stubsIndexForResource(t, definitions, `<?php function caller(){$h=openIt();fclose($h);probe($h);}`)
	defs := syntax.Parse("definitions.php", []byte(definitions), syntax.Options{Version: phpversion.Max})
	snapshot := NewSnapshot(Extract(defs, ix, phpversion.Max, nil, Options{}))
	caller := syntax.Parse("caller.php", []byte(`<?php function caller(){$h=openIt();fclose($h);probe($h);}`), syntax.Options{Version: phpversion.Max})
	e, out := Summarize(caller, resourceTypes(caller, ix), snapshot, Options{})
	call := callsNamed(caller, "probe")[0]
	if _, closed := e.StateBefore(call, argument(call), "fclose"); !closed {
		t.Fatal("cross-file factory identity lost")
	}
	replacement := syntax.Parse("definitions.php", []byte(`<?php function openIt(){return 'safe';}`), syntax.Options{Version: phpversion.Max})
	next := snapshot.WithFile(Extract(replacement, ix, phpversion.Max, nil, Options{}))
	if out.Current(snapshot, next) {
		t.Fatal("resource-contract change did not invalidate caller")
	}
}

func stubsIndexForResource(t testing.TB, defs, caller string) *index.Index {
	t.Helper()
	ix := index.New(stubs.Index())
	for path, src := range map[string]string{"definitions.php": defs, "caller.php": caller} {
		f := syntax.Parse(path, []byte(src), syntax.Options{Version: phpversion.Max})
		ix.Add(index.Extract(f))
	}
	return ix
}

func resourceTypes(f *syntax.File, ix *index.Index) *infer.Env {
	return infer.NewEnv(f, names.New(f), ix, phpversion.Max, infer.Options{})
}

func TestGeneratorsHaveNoResourceContract(t *testing.T) {
	e, _ := summarized(t, `function openIt(){yield tmpfile();}`)
	summary, _ := e.contract("openit")
	if summary.resource != nil {
		t.Fatal("generator acquired resource-return contract")
	}
}

func TestResourcePassProofBoundaries(t *testing.T) {
	for _, source := range []string{
		`function pass($h){$a=&$h;return $a;}`,
		`function pass($h){$h.='x';return $h;}`,
		`function pass($h){` + strings.Repeat(`$a=$h;`, 130) + `return $h;}`,
	} {
		e, _ := summarized(t, source)
		summary, _ := e.contract("pass")
		if summary.resource != nil {
			t.Fatal("impure or over-budget pass gained a contract")
		}
	}
	e, file := environment(t, `probe(1);`, nil)
	at := callsNamed(file, "probe")[0]
	for _, args := range [][]Value{nil, {{Complete: false}}, {{Complete: true, Identity: 1}}} {
		if _, _, known := e.instantiateResource(at, &resourceReturn{parameter: 0}, args, &frame{}, &scopeResult{}); known {
			t.Fatal("unknown caller resource accepted")
		}
	}
}
