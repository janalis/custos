package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestRecordResources(t *testing.T) {
	for _, tc := range []struct {
		source string
		closed bool
	}{
		{`function factory(){return ['stream'=>tmpfile()];}$r=factory();$h=$r['stream'];fclose($h);probe($r['stream']);`, true},
		{`function factory(){$h=tmpfile();return ['a'=>$h,'nested'=>['b'=>$h]];}$r=factory();fclose($r['a']);probe($r['nested']['b']);`, true},
		{`function factory(){return ['a'=>tmpfile(),'b'=>tmpfile()];}$r=factory();fclose($r['a']);probe($r['b']);`, false},
		{`function factory(){return ['stream'=>tmpfile()];}$a=factory();$b=factory();fclose($a['stream']);probe($b['stream']);`, false},
		{`function factory(){return ['stream'=>new stdClass];}$a=factory();fclose($a['stream']);probe($a['stream']);`, false},
		{`function factory(){return ['stream'=>tmpfile()];}$a=factory();fclose($a['stream']);opaque($a);probe($a['stream']);`, false},
		{`function wrap($h){return ['stream'=>$h];}$h=tmpfile();fclose($h);$r=wrap($h);probe($r['stream']);`, true},
		{`function wrap($h){fclose($h);return ['stream'=>$h];}$h=tmpfile();$r=wrap($h);probe($r['stream']);`, false},
		{`$factory=fn()=>['stream'=>tmpfile()];$r=$factory();fclose($r['stream']);probe($r['stream']);`, true},
		{`final class R{static function make(){return ['stream'=>tmpfile()];}}$r=R::make();fclose($r['stream']);probe($r['stream']);`, true},
		{`final class R{static function make(){return ['stream'=>tmpfile()];}}$cb=R::make(...);$r=$cb();fclose($r['stream']);probe($r['stream']);`, true},
		{`function factory(){return ['stream'=>tmpfile()];}$cb=factory(...);$r=$cb();fclose($r['stream']);probe($r['stream']);`, true},
		{`$xs=array_map(fn($x)=>['stream'=>tmpfile()],[1,2]);fclose($xs[0]['stream']);probe($xs[1]['stream']);`, false},
		{`function factory(){return ['stream'=>tmpfile()];}$r=factory();$r['stream']='safe';fclose($r['stream']);probe($r['stream']);`, false},
	} {
		e, f := summarized(t, tc.source)
		call := callsNamed(f, "probe")[0]
		if _, got := e.StateBefore(call, argument(call), "fclose"); got != tc.closed {
			t.Fatalf("%s: closed=%v value=%+v", tc.source, got, e.Value(argument(call)))
		}
	}
}

func TestRecordResourceBoundaries(t *testing.T) {
	for _, tc := range []struct {
		source string
		closed bool
	}{
		{`$h=new stdClass;fclose($h);$r=['h'=>$h];probe($r['h']);`, false},
		{`function wrap($h){return ['h'=>$h];}function caller($h){fread($h,1);$r=wrap($h);fclose($r['h']);probe($r['h']);}`, true},
		{`function factory(){$h=tmpfile();opaque($h);return ['h'=>$h];}$r=factory();fclose($r['h']);probe($r['h']);`, false},
		{`function factory(){$h=tmpfile();return ['h'=>$h,'a'=>$h];}$r=factory();opaque($r);fclose($r['h']);probe($r['a']);`, false},
		{`function factory(){$h=tmpfile();return ['h'=>$h];}$r=factory();fclose($r['h']);array_map(fn($x)=>opaque($r),[1]);probe($r['h']);`, false},
		{`function factory(){$h=tmpfile();fclose($h);return ['h'=>$h];}$r=factory();probe($r['h']);`, false},
		{`function factory(){$h=tmpfile();if(!fclose($h)){throw new Exception;}return ['h'=>$h];}$r=factory();probe($r['h']);`, true},
		{`$h=tmpfile();fclose($h);$cb=fn()=>['h'=>$h];$r=$cb();probe($r['h']);`, true},
		{`$h=tmpfile();$cb=function()use($h){fclose($h);return ['h'=>$h];};$r=$cb();probe($r['h']);`, false},
		{`$h=tmpfile();$cb=function()use($h){if(!fclose($h)){throw new Exception;}return ['h'=>$h];};$r=$cb();probe($r['h']);`, true},
		{`function wrap($h){return ['h'=>$h];}$r=wrap('safe');fclose($r['h']);probe($r['h']);`, false},
		{`function factory(){return ['h'=>tmpfile()];}$r=factory();$h=$r['h'];fclose($h);opaque($r);probe($h);`, false},
	} {
		e, f := summarized(t, tc.source)
		at := callsNamed(f, "probe")[0]
		if _, got := e.StateBefore(at, argument(at), "fclose"); got != tc.closed {
			t.Fatalf("%s: closed=%v", tc.source, got)
		}
	}
}

func TestRecordContractBounds(t *testing.T) {
	e := &Env{}
	at := &syntax.FuncCall{}
	first := e.recordIdentity(at, 1)
	if first == 0 || first >= bufferID || e.recordIdentity(at, 1) != first {
		t.Fatal("unstable or reserved identity")
	}
	for i := 2; i <= 4096; i++ {
		if id := e.recordIdentity(at, i); id == 0 || id == first {
			t.Fatal("record identities collided")
		}
	}
	if e.recordIdentity(at, 4097) != 0 {
		t.Fatal("record-site budget exceeded")
	}
	set := &elementSet{items: map[string]Value{}, whole: true, width: 1}
	set.put("h", Value{Complete: true, record: &resourceReturn{fresh: true, token: 4097}})
	value, states := e.instantiateRecord(at, Value{Complete: true, elements: set}, nil, &frame{}, &scopeResult{})
	if len(states) != 0 || value.elements.items["h"].storedIdentity != 0 || set.items["h"].record == nil {
		t.Fatal("exhausted record instantiated or mutated template")
	}
	for _, optional := range []bool{false, true} {
		child := &elementSet{items: map[string]Value{}, whole: optional, optional: optional, width: 1}
		child.put("h", Value{Complete: true, storedIdentity: 1})
		parent := &elementSet{items: map[string]Value{}, whole: true, width: 1}
		parent.put("child", Value{Complete: true, elements: child})
		out := recordContracts(Value{elements: parent}, &scopeResult{allocations: map[uint32]bool{1: true}}, nil)
		if out.elements.hasResources {
			t.Fatal("uncertain record published an identity")
		}
	}
}

func TestRecordPassGuards(t *testing.T) {
	for _, source := range []string{
		`function wrap($h){$a=&$h;return ['h'=>$a];}`,
		`function wrap($h){$a=[&$h];return ['h'=>$h];}`,
		`function wrap($h){$a=[...$h];return ['h'=>$h];}`,
	} {
		e, f := summarized(t, source)
		if pureRecordPass(syntax.VariableScopeBody(f.Stmts[0])) {
			t.Fatal("unsafe record pass accepted")
		}
		sum, _ := e.contract("wrap")
		if sum.Return.elements != nil && sum.Return.elements.hasResources {
			t.Fatal("unsafe borrowed record published")
		}
	}
	e, f := summarized(t, `$cb=function(){opaque();return ['h'=>tmpfile()];};$r=$cb();probe($r['h']);`)
	if e.Value(argument(callsNamed(f, "probe")[0])).Identity != 0 {
		t.Fatal("opaque closure exported record identity")
	}
}

func TestRecordAndPredicateSnapshotChanges(t *testing.T) {
	defs := `<?php function factory(){return ['h'=>tmpfile()];}function reject($v){return false;}function unwrap(){$r=factory();return $r['h'];}`
	caller := `<?php function caller(){$r=factory();fclose($r['h']);probe($r['h']);$xs=array_filter([$_GET['x']],'reject');probeFilter($xs);}`
	ix := stubsIndexForResource(t, defs, caller)
	file := syntax.Parse("definitions.php", []byte(defs), syntax.Options{Version: phpversion.Max})
	published := Extract(file, ix, phpversion.Max, nil, Options{})
	var check func(Value)
	check = func(v Value) {
		if v.Expr != nil || v.Identity != 0 || v.storedIdentity != 0 || v.origins != nil || v.targets != nil || v.returned != nil {
			t.Fatal("evaluation state escaped snapshot")
		}
		if v.elements != nil {
			for _, entry := range v.elements.items {
				check(entry)
			}
		}
	}
	for _, summary := range published.Summaries {
		check(summary.Return)
	}
	snapshot := NewSnapshot(published)
	use := syntax.Parse("caller.php", []byte(caller), syntax.Options{Version: phpversion.Max})
	e, out := Summarize(use, resourceTypes(use, ix), snapshot, Options{})
	at := callsNamed(use, "probe")[0]
	if _, closed := e.StateBefore(at, argument(at), "fclose"); !closed {
		t.Fatal("cross-file record identity lost")
	}
	filtered := argument(callsNamed(use, "probeFilter")[0])
	if e.Tainted(filtered, HTML) {
		t.Fatal("cross-file literal predicate lost")
	}
	edited := syntax.Parse("definitions.php", []byte(`<?php function factory(){return ['h'=>'safe'];}function reject($v){return true;}`), syntax.Options{Version: phpversion.Max})
	next := snapshot.WithFile(Extract(edited, ix, phpversion.Max, nil, Options{}))
	if out.Current(snapshot, next) {
		t.Fatal("record or scalar contract edit did not invalidate caller")
	}
	updated := New(use, resourceTypes(use, ix), next, Options{})
	if _, closed := updated.StateBefore(at, argument(at), "fclose"); closed || !updated.Tainted(filtered, HTML) {
		t.Fatal("edited contracts reused stale facts")
	}
}
