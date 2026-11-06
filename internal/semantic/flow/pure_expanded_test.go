package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestExpandedPureResults(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($v){if($v==='drop'){return false;}return true;}probe(f('drop'));`, constantScalar{k: 'b'}},
		{`function f($v){if($v===1){$x=2;}elseif($v===2){$x=3;}else{$x=4;}return $x+1;}probe(f(2));`, constantScalar{k: 'i', i: 4}},
		{`function f($v){return $v ? 1 : 2;}probe(f(false));`, constantScalar{k: 'i', i: 2}},
		{`function f($v){return $v ?: 'fallback';}probe(f('value'));`, constantScalar{k: 's', s: "value"}},
		{`function f($v){return $v && (9223372036854775807+1);}probe(f(false));`, constantScalar{k: 'b'}},
		{`function f($v){return $v || (9223372036854775807+1);}probe(f(true));`, constantScalar{k: 'b', b: true}},
		{`function f($v){return $v ?? 'fallback';}probe(f(null));`, constantScalar{k: 's', s: "fallback"}},
		{`function f($v){return $v ?? (9223372036854775807+1);}probe(f('kept'));`, constantScalar{k: 's', s: "kept"}},
		{`function f($v){return strlen($v)===0;}probe(f(''));`, constantScalar{k: 'b', b: true}},
		{`function f($v){return trim($v);}probe(f(" \tvalue\n\x00"));`, constantScalar{k: 's', s: "value"}},
		{`function f($v){return strlen(string:$v);}probe(f('é'));`, constantScalar{k: 'i', i: 2}},
		{`function f($v=['safe']){return $v[0];}probe(f());`, constantScalar{k: 's', s: "safe"}},
		{`function f($v=['01'=>'a','b']){return $v[0];}probe(f());`, constantScalar{k: 's', s: "b"}},
		{`function f($v=['+1'=>'a','b']){return $v[0];}probe(f());`, constantScalar{k: 's', s: "b"}},
		{`function f($v=['-0'=>'a','b']){return $v[0];}probe(f());`, constantScalar{k: 's', s: "b"}},
		{`function f($v=['nested'=>['safe']]){return $v['nested'][0];}probe(f());`, constantScalar{k: 's', s: "safe"}},
		{`function f($v){$a=['chosen'=>$v,'ignored'=>'other'];return $a['chosen'];}probe(f('safe'));`, constantScalar{k: 's', s: "safe"}},
		{`function add($x,$y=1){return $x+$y;}function f($v){return add(...[$v]);}probe(f(2));`, constantScalar{k: 'i', i: 3}},
		{`function sub($x,$y){return $x-$y;}function f($v){return sub(...['y'=>1,'x'=>$v]);}probe(f(5));`, constantScalar{k: 'i', i: 4}},
		{`function sub($x,$y){return $x-$y;}function f($v){return sub(...[$v],y:1);}probe(f(5));`, constantScalar{k: 'i', i: 4}},
		{`function sub($x,$y){return $x-$y;}function f($v){$pack=[$v,1];return sub(...$pack);}probe(f(5));`, constantScalar{k: 'i', i: 4}},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		scalar, known := knownScalar(value)
		if !known || scalar != tc.want {
			t.Errorf("%s: scalar=%+v known=%v want=%+v", tc.source, scalar, known, tc.want)
		}
	}
}

func TestExpandedPureProvenance(t *testing.T) {
	for _, tc := range []struct {
		source string
		taint  bool
	}{
		{`function first($v){return $v[0];}probe(first(['safe',$_GET['x']]));`, false},
		{`function first($v){return $v[0];}function wrap($v){return first(...[$v]);}probe(wrap(['safe',$_GET['x']]));`, false},
		{`function first($v){return $v[0];}probe(first([$_GET['x'],'safe']));`, true},
		{`function first($v=['safe']){return $v[0];}$cb=first(...);probe($cb());`, false},
		{`function f($v){if($v==='drop'){return false;}return true;}$a=array_filter(['drop','keep'],'f');probe($a[1]);`, false},
	} {
		e, file := summarized(t, tc.source)
		at := argument(callsNamed(file, "probe")[0])
		if v := e.Value(at); !v.Complete || e.Tainted(at, HTML) != tc.taint {
			t.Errorf("%s: %+v", tc.source, v)
		}
	}
}

func TestExpandedPureRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($v){return strlen($v);}probe(f(42));`,
		`function f($v){return trim($v,'a..z');}probe(f('x'));`,
		`function f($v){return strlen(nope:$v);}probe(f('x'));`,
		`function f($v){return strlen(...[$v]);}probe(f('x'));`,
		`function f($v){return $v[2];}probe(f(['safe']));`,
		`function f($v=['01'=>'a','b']){return $v[2];}probe(f());`,
		`function f($v=[1.5=>'safe']){return $v[1];}probe(f());`,
		`function f($v=[1+1=>'safe']){return $v[2];}probe(f());`,
		`function f($v){return $v[0];}probe(f(opaque()));`,
		`function f($v){return $v[0];}$a=['safe'];$a[0]=opaque();probe(f($a));`,
		`function f($v){return [$v=>1];}probe(f('x'));`,
		`function f($v){return [&$v];}probe(f('x'));`,
		`function f($v){return [...$v];}probe(f(['x']));`,
		`function f($v){return [0=>'a',0=>$v];}probe(f('x'));`,
		`function f($v){return [-1=>$v];}probe(f('x'));`,
		`function f($v){return $v ? opaque() : 1;}probe(f(false));`,
		`function id($v){return $v;}function f($v){return id(...opaque());}probe(f(1));`,
		`function id($v){return $v;}function f($v){return id(...['v'=>$v],v:2);}probe(f(1));`,
		`function id($v){return $v;}function f($v){return id(v:$v,...[2]);}probe(f(1));`,
	} {
		e, file := summarized(t, source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if _, known := knownScalar(value); known {
			t.Errorf("unsafe exact result: %s", source)
		}
	}
}

func BenchmarkExpandedPureContracts(b *testing.B) {
	source := `function f($v,$options=['yes'=>1,'no'=>0]){if(strlen(trim($v))===0){return $options['no'];}return $options['yes'];}function wrap($v){return f(...[$v]);}probe(wrap('value'));`
	e, file := summarized(b, source)
	at := argument(callsNamed(file, "probe")[0])
	b.ReportAllocs()
	for b.Loop() {
		New(file, e.types, e.snapshot, Options{}).Value(at)
	}
}

func TestExpandedPureCrossFileEdit(t *testing.T) {
	defs := `<?php function choose($v){if(strlen(trim($v))===0){return true;}return false;}function wrap($v){return choose(...[$v]);}`
	caller := `<?php function run(){$a=array_filter([' '],'wrap');probe($a);}`
	ix := stubsIndexForResource(t, defs, caller)
	makeFile := func(path, source string) *syntax.File {
		return syntax.Parse(path, []byte(source), syntax.Options{Version: phpversion.Max})
	}
	definitions := makeFile("definitions.php", defs)
	first := NewSnapshot(Extract(definitions, ix, phpversion.Max, nil, Options{}))
	file := makeFile("caller.php", caller)
	env, published := Summarize(file, resourceTypes(file, ix), first, Options{})
	before := env.Value(argument(callsNamed(file, "probe")[0]))
	if before.elements == nil || !before.elements.whole || len(before.elements.items) != 1 {
		t.Fatal("cross-file selected entry lost")
	}
	edited := makeFile("definitions.php", strings.Replace(defs, "===0", "!==0", 1))
	next := first.WithFile(Extract(edited, ix, phpversion.Max, nil, Options{}))
	if published.Current(first, next) {
		t.Fatal("native/branch/pack contract edit did not invalidate caller")
	}
	after := NewLocal(file, resourceTypes(file, ix), next, Options{}, nil).Value(argument(callsNamed(file, "probe")[0]))
	if after.elements == nil || !after.elements.whole || len(after.elements.items) != 0 {
		t.Fatal("edited predicate retained old entry")
	}
}

func TestPureSpreadsRejectStaleParameterMetadata(t *testing.T) {
	old := `<?php function leaf($x){return $x;}`
	caller := `<?php function wrap($v){return leaf(...[$v]);}`
	beforeIndex := stubsIndexForResource(t, old, caller)
	options := syntax.Options{Version: phpversion.Max}
	snapshot := NewSnapshot(Extract(syntax.Parse("definitions.php", []byte(old), options), beforeIndex, phpversion.Max, nil, Options{}))
	for _, current := range []string{
		`<?php function leaf($x,$y){return $x;}`,
		`<?php function leaf(&$x){return $x;}`,
		`<?php function leaf(...$x){return $x;}`,
		`<?php function leaf($renamed){return $renamed;}`,
	} {
		ix := stubsIndexForResource(t, current, caller)
		file := syntax.Parse("caller.php", []byte(caller), options)
		env := NewLocal(file, resourceTypes(file, ix), snapshot, Options{}, nil)
		summary, _ := env.contract("wrap")
		if summary.pure != nil {
			t.Fatalf("stale signature established a composed contract: %s", current)
		}
	}
}
