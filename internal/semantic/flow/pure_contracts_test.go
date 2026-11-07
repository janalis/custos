package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestPureScalarContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function add($c,$v){return $c+$v;}probe(array_reduce([1,-1],'add',0));`, constantScalar{k: 'i'}},
		{`function factory(){return fn($c,$v)=>$c+$v;}probe(array_reduce([1,-1],factory(),0));`, constantScalar{k: 'i'}},
		{`function factory(){return function($c,$v){return $c+$v;};}probe(array_reduce([2,3],factory(),0));`, constantScalar{k: 'i', i: 5}},
		{`function zero(int $v){return $v;}function wrap($v){return zero($v);}probe(wrap('00'));`, constantScalar{k: 'i'}},
		{`function add($x,$y=2){return $x+$y;}function wrap($v){$a=$v;return add(y:3,x:$a);}probe(wrap(2));`, constantScalar{k: 'i', i: 5}},
		{`function add($x,$y=2){return $x+$y;}function wrap($v){return add($v);}probe(wrap(2));`, constantScalar{k: 'i', i: 4}},
		{`function choose($unused,int $v=0){return $v;}probe(choose('input'));`, constantScalar{k: 'i'}},
		{`function choose($unused,int $v=0){return $v;}$cb=choose(...);probe($cb('input'));`, constantScalar{k: 'i'}},
		{`final class C{static function choose($unused,int $v=0){return $v;}}$cb=C::choose(...);probe($cb('input'));`, constantScalar{k: 'i'}},
		{`function factory(int $v){return fn()=>$v;}$a=factory('00');$b=factory(1);probe($a());probe($b());`, constantScalar{k: 'i'}},
		{`function factory(int $v){return fn($x)=>$v+$x;}$a=factory('2');probe($a(3));`, constantScalar{k: 'i', i: 5}},
		{`function factory(){ $v=2;return function($x)use($v){return $v*$x;};}probe((factory())(3));`, constantScalar{k: 'i', i: 6}},
		{`function neg($v){return -$v;}probe(neg(2));`, constantScalar{k: 'i', i: -2}},
		{`function pos($v){return +$v;}probe(pos(2.0));`, constantScalar{k: 'f', f: 2}},
		{`function text($v){return $v.'!';}probe(text('ok'));`, constantScalar{k: 's', s: "ok!"}},
		{`function no($v){return !$v;}probe(no(0));`, constantScalar{k: 'b', b: true}},
		{`function choose($v=null){return $v;}probe(choose());`, constantScalar{k: 'n'}},
		{`function choose($unused,int $v=0){return $v;}$cb=choose(...);probe($cb(...['unused'=>'x']));`, constantScalar{k: 'i'}},
		{`function choose($unused,int $v=0){return $v;}$cb=choose(...);probe($cb(unused:'x'));`, constantScalar{k: 'i'}},
	} {
		e, f := summarized(t, tc.source)
		calls := callsNamed(f, "probe")
		got, known := knownScalar(e.Value(argument(calls[0])))
		if !known || !sameScalar(&got, &tc.want) {
			t.Errorf("%s: got=%+v known=%v", tc.source, got, known)
		}
		if len(calls) > 1 {
			v, ok := knownScalar(e.Value(argument(calls[1])))
			if !ok || v.k != 'i' || v.i != 1 {
				t.Fatal("factory bindings shared")
			}
		}
	}
}

func TestPurePredicateContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		empty  bool
	}{
		{`function drop($v){if($v)return false;return null;}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, true},
		{`$xs=array_filter([$_GET['x']],function($v){if($v)return false;return null;});probe($xs);`, true},
		{`function drop($v){if($v)return false;}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, true},
		{`function drop($v):string{if($v)return false;return 0;}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, true},
		{`function drop($v){if($v)return false;return 1;}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, false},
		{`function zero(int $v){return $v;}function wrap($v){return zero($v);}$xs=array_filter(['00'],'wrap');probe($xs);`, true},
		{`function choose($unused,int $v=0){return $v;}$xs=array_filter([$_GET['x']],'choose');probe($xs);`, true},
		{`function factory(int $v){return fn()=>$v;}$xs=array_filter([$_GET['x']],factory(0));probe($xs);`, true},
		{`function factory(int $v){return function()use($v){return $v;};}$xs=array_filter([$_GET['x']],factory(0));probe($xs);`, true},
		{`function drop($v):string{if($v)return -0.0;return 0;}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, false},
		{`function drop($v){if($v)return false;return null;}probe(drop($_GET['x'])===false);`, false},
	} {
		e, f := summarized(t, tc.source)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		empty := v.elements != nil && v.elements.whole && !v.elements.optional && len(v.elements.items) == 0
		if empty != tc.empty {
			t.Errorf("%s: empty=%v value=%+v", tc.source, empty, v)
		}
	}
}

func TestPureRecordComposition(t *testing.T) {
	for _, tc := range []struct {
		source string
		closed bool
	}{
		{`function pass($r){return $r;}function wrap($r){return pass($r);}$r=['h'=>tmpfile()];$a=wrap($r);fclose($a['h']);probe($r['h']);`, true},
		{`function pass(array $r):array{return $r;}function wrap(array $r):array{return pass($r);}$r=['h'=>tmpfile()];$a=wrap($r);fclose($a['h']);probe($r['h']);`, true},
		{`function pass($r){return $r;}function wrap($r){opaque($r);return pass($r);}$r=['h'=>tmpfile()];$a=wrap($r);fclose($a['h']);probe($r['h']);`, false},
		{`function pass(&$r){return $r;}function wrap($r){return pass($r);}$r=['h'=>tmpfile()];$a=wrap($r);fclose($a['h']);probe($r['h']);`, false},
	} {
		e, f := summarized(t, tc.source)
		at := callsNamed(f, "probe")[0]
		if _, closed := e.StateBefore(at, argument(at), "fclose"); closed != tc.closed {
			t.Errorf("%s: closed=%v", tc.source, closed)
		}
	}
}

func TestPureWrapperEditInvalidates(t *testing.T) {
	defs := `<?php function zero(int $v){return $v;}function wrap($v){return zero($v);}`
	caller := `<?php function caller(){$xs=array_filter(['00'],'wrap');probe($xs);}`
	ix := stubsIndexForResource(t, defs, caller)
	file := syntax.Parse("definitions.php", []byte(defs), syntax.Options{Version: phpversion.Max})
	first := NewSnapshot(Extract(file, ix, phpversion.Max, nil, Options{}))
	use := syntax.Parse("caller.php", []byte(caller), syntax.Options{Version: phpversion.Max})
	_, published := Summarize(use, resourceTypes(use, ix), first, Options{})
	edited := syntax.Parse("definitions.php", []byte(`<?php function zero(int $v){return $v+1;}function wrap($v){return zero($v);}`), syntax.Options{Version: phpversion.Max})
	next := first.WithFile(Extract(edited, ix, phpversion.Max, nil, Options{}))
	if published.Current(first, next) {
		t.Fatal("wrapper edit did not invalidate caller")
	}
	e := New(use, resourceTypes(use, ix), next, Options{})
	v := e.Value(argument(callsNamed(use, "probe")[0]))
	if v.elements != nil && !v.elements.optional && len(v.elements.items) == 0 {
		t.Fatal("old wrapper fact survived")
	}
}

func TestPureAssignmentComposition(t *testing.T) {
	for _, tc := range []struct {
		source string
		known  bool
	}{
		{`function f($v){$x=$v+1;return $x;}function g($v){return f($v);}probe(g(2));`, true},
		{`function f($v){$unused=+$v;return 0;}probe(f(new stdClass));`, true},
		{`function f($r,$v){$unused=+$v;return $r;}$r=['h'=>tmpfile()];$a=f($r,new stdClass);fclose($a['h']);probe($r['h']);`, false},
	} {
		e, f := summarized(t, tc.source)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		_, known := knownScalar(v)
		if known != tc.known {
			t.Errorf("%s: known=%v", tc.source, known)
		}
	}
}

func TestForwardRecordFallback(t *testing.T) {
	body := `return $r;`
	for _, source := range []string{
		`function pass($r,$unused=[]){` + body + `}$r=['h'=>tmpfile()];$a=pass($r,1);fclose($a['h']);probe($r['h']);`,
		`function pass($r,$unused=new stdClass){` + body + `}$r=['h'=>tmpfile()];$a=pass($r,1);fclose($a['h']);probe($r['h']);`,
		`final class C{static function pass($r,$unused=new stdClass){` + body + `}}$cb=C::pass(...);$r=['h'=>tmpfile()];$a=$cb($r,1);fclose($a['h']);probe($r['h']);`,
		`final class C{static function pass($r,$unused=[]){` + body + `}}$cb=C::pass(...);$r=['h'=>tmpfile()];$a=$cb($r,1);fclose($a['h']);probe($r['h']);`,
	} {
		e, f := summarized(t, source)
		at := callsNamed(f, "probe")[0]
		if _, closed := e.StateBefore(at, argument(at), "fclose"); !closed {
			t.Fatalf("exact forwarding fallback lost: %s", source)
		}
	}
}

func TestTruthyAgreementDoesNotImplyLiteralEquality(t *testing.T) {
	e, f := summarized(t, `function choose($v){if($v)return false;return null;}probe(choose($_GET['x']));probe(choose($_GET['x'])===false);`)
	calls := callsNamed(f, "probe")
	v := e.Value(argument(calls[0]))
	truth, known := predicateTruth(v)
	if !known || truth {
		t.Fatal("agreeing falsy returns lost")
	}
	if _, known := knownScalar(v); known {
		t.Fatal("different falsy values collapsed to a literal")
	}
	if _, known := knownScalar(e.Value(argument(calls[1]))); known {
		t.Fatal("truth agreement established strict equality")
	}
}
