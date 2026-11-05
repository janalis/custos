package flow

import "testing"

func TestExtendedPredicateContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		empty  bool
	}{
		{`function zero(int $v){return $v;}$xs=array_filter(['00'],'zero');probe($xs);`, true},
		{`function factory(){return fn(int $v)=>$v;}$xs=array_filter(['00'],factory());probe($xs);`, true},
		{`function factory(){return function(int $v){return $v;};}$xs=array_filter(['00'],factory());probe($xs);`, true},
		{`function zero(int $v):string{return $v;}$xs=array_filter(['00'],'zero');probe($xs);`, true},
		{`declare(strict_types=1);function zero(int $v):string{return $v;}$xs=array_filter(['00'],'zero');probe($xs);`, false},
		{`function zero(int $v){return $v;}$cb=zero(...);$xs=array_filter(['00'],$cb);probe($xs);`, true},
		{`final class C{static function zero(int $v){return $v;}}$xs=array_filter(['00'],C::zero(...));probe($xs);`, true},
		{`function drop($v){}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, true},
		{`function drop($v){return;}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, true},
		{`function drop($v):void{}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, true},
		{`function drop($v){if($v){return false;}}$xs=array_filter([$_GET['x']],'drop');probe($xs);`, false},
		{`function factory(){return function($v){};}$xs=array_filter([$_GET['x']],factory());probe($xs);`, true},
		{`$xs=array_filter([$_GET['x']],function($v){});probe($xs);`, true},
		{`$sum=array_reduce([1,-1],fn($c,$v)=>$c+$v,0);$xs=array_filter([$_GET['x']],fn()=>$sum);probe($xs);`, true},
	} {
		e, f := summarized(t, tc.source)
		x := argument(callsNamed(f, "probe")[0])
		v := e.Value(x)
		empty := v.elements != nil && v.elements.whole && !v.elements.optional && len(v.elements.items) == 0
		if empty != tc.empty {
			t.Fatalf("%s: empty=%v value=%+v", tc.source, empty, v)
		}
	}
}

func TestExtendedResourceContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		closed bool
	}{
		{`function make(){return ['h'=>tmpfile()];}function pass($r){return $r;}$r=pass(make());fclose($r['h']);probe($r['h']);`, true},
		{`function pass(array $r):array{return $r;}$r=pass(['h'=>tmpfile()]);fclose($r['h']);probe($r['h']);`, true},
		{`$r=['h'=>tmpfile()];$cb=fn()=>$r;$a=$cb();fclose($a['h']);probe($r['h']);`, true},
		{`$r=['h'=>tmpfile()];$cb=function()use($r){$h=$r['h'];if(!fclose($h)){throw new Exception;}return $r;};$a=$cb();probe($r['h']);`, true},
		{`function make(){return fn()=>['h'=>tmpfile()];}$cb=make();$r=$cb();fclose($r['h']);probe($r['h']);`, true},
		{`function make(){return fn()=>['h'=>tmpfile()];}$cb=make();$a=$cb();$b=$cb();fclose($a['h']);probe($b['h']);`, false},
		{`function make(){return function(){$h=tmpfile();return ['h'=>$h,'alias'=>$h];};}$cb=make();$r=$cb();fclose($r['h']);probe($r['alias']);`, true},
		{`function make(){return fn()=>['h'=>tmpfile()];}$xs=array_map(make(),[1,2]);fclose($xs[0]['h']);probe($xs[1]['h']);`, false},
		{`function pass($r){opaque($r);return $r;}$r=pass(['h'=>tmpfile()]);fclose($r['h']);probe($r['h']);`, false},
		{`function pass(&$r){return $r;}$r=['h'=>tmpfile()];$a=pass($r);fclose($a['h']);probe($r['h']);`, false},
	} {
		e, f := summarized(t, tc.source)
		at := callsNamed(f, "probe")[0]
		if _, closed := e.StateBefore(at, argument(at), "fclose"); closed != tc.closed {
			t.Fatalf("%s: closed=%v value=%+v", tc.source, closed, e.Value(argument(at)))
		}
	}
}

func TestScalarAccumulatorContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`$sum=array_reduce([1,-1],fn($c,$v)=>$c+$v,0);probe($sum);`, constantScalar{k: 'i'}},
		{`$sum=array_reduce(['a','b'],fn($c,$v)=>$c.$v,'');probe($sum);`, constantScalar{k: 's', s: "ab"}},
		{`$sum=array_reduce([2,3],fn($c,$v)=>$c*$v,1);probe($sum);`, constantScalar{k: 'i', i: 6}},
		{`$cb=fn(int $v)=>-$v;probe($cb('2'));`, constantScalar{k: 'i', i: -2}},
		{`$cb=fn(int $v)=>+$v;probe($cb('2'));`, constantScalar{k: 'i', i: 2}},
		{`$cb=fn(float $v)=>-$v;probe($cb(2));`, constantScalar{k: 'f', f: -2}},
	} {
		e, f := summarized(t, tc.source)
		got, known := knownScalar(e.Value(argument(callsNamed(f, "probe")[0])))
		if !known || !sameScalar(&got, &tc.want) {
			t.Fatalf("%s: scalar=%+v known=%v", tc.source, got, known)
		}
	}
}

func TestExtendedContractRefusals(t *testing.T) {
	for _, source := range []string{
		`function pass($r){return (array)$r;}$r=pass(['h'=>tmpfile()]);fclose($r['h']);probe($r['h']);`,
		`function pass($r){return $r;}$r=['h'=>new stdClass];$a=pass($r);fclose($a['h']);probe($r['h']);`,
		`$r=['h'=>new stdClass];$cb=fn()=>$r;$a=$cb();fclose($a['h']);probe($r['h']);`,
		`function pass($r,$x){return $r;}$r=['h'=>tmpfile()];$a=pass($r,new stdClass);fclose($a['h']);probe($r['h']);`,
		`function pass($r):object{return $r;}$r=['h'=>tmpfile()];$a=pass($r);fclose($a['h']);probe($r['h']);`,
		`function pass($r){return $r;}$r=['h'=>tmpfile()];opaque($r);$a=pass($r);fclose($a['h']);probe($r['h']);`,
	} {
		e, f := summarized(t, source)
		at := callsNamed(f, "probe")[0]
		if _, closed := e.StateBefore(at, argument(at), "fclose"); closed {
			t.Fatalf("unsafe resource forwarding: %s", source)
		}
	}
}

func TestForwardingRejectsOtherParameterConversions(t *testing.T) {
	e, f := summarized(t, `function pass($r,int $other){return $r;}$r=['h'=>tmpfile()];$a=pass($r,1.5);fclose($a['h']);probe($r['h']);`)
	at := callsNamed(f, "probe")[0]
	if _, closed := e.StateBefore(at, argument(at), "fclose"); closed {
		t.Fatal("record crossed a potentially warning-producing parameter conversion")
	}
}

func TestResourceInputProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		source string
		closed bool
	}{
		{`$unused=tmpfile();$r=['h'=>new stdClass];$cb=fn()=>$r;$a=$cb();fclose($a['h']);probe($r['h']);`, false},
		{`$r=['h'=>tmpfile()];$outer=function()use($r){$inner=fn()=>$r;return $inner();};$a=$outer();fclose($a['h']);probe($r['h']);`, false},
	} {
		e, f := summarized(t, tc.source)
		at := callsNamed(f, "probe")[0]
		if _, closed := e.StateBefore(at, argument(at), "fclose"); closed != tc.closed {
			t.Fatalf("%s: closed=%v", tc.source, closed)
		}
	}
	e, f := summarized(t, `$unused=tmpfile();$cb=fn($x)=>$x;probe($cb(unknown()));`)
	if v := e.Value(argument(callsNamed(f, "probe")[0])); v.Complete {
		t.Fatal("unknown seed acquired resource proof")
	}
}

func TestBorrowedParameterRecordForwarding(t *testing.T) {
	for _, source := range []string{
		`function pass($r){return $r;}function caller($h){fclose($h);$r=['h'=>$h];$a=pass($r);probe($a['h']);}`,
		`function caller($h){fclose($h);$r=['h'=>$h];$cb=fn()=>$r;$a=$cb();probe($a['h']);}`,
	} {
		e, f := summarized(t, source)
		at := callsNamed(f, "probe")[0]
		if _, closed := e.StateBefore(at, argument(at), "fclose"); !closed {
			t.Fatalf("borrowed record history lost: %s", source)
		}
	}
}
