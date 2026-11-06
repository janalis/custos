package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestPureQueryResults(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($v){return match($v){'drop'=>false,default=>true};}probe(f('drop'));`, constantScalar{k: 'b'}},
		{`function f($v){return match($v){1=>false,'1',2=>true,default=>false};}probe(f('1'));`, constantScalar{k: 'b', b: true}},
		{`function f($v){return match($v){default=>1,2=>3};}probe(f(2));`, constantScalar{k: 'i', i: 3}},
		{`function f($v){return match($v){0.0=>1,default=>2};}probe(f(-0.0));`, constantScalar{k: 'i', i: 1}},
		{`function f($v){return match($v){1=>2,default=>3};}probe(f(4));`, constantScalar{k: 'i', i: 3}},
		{`function f($r){return $r['label'] ?? 'fallback';}probe(f([]));`, constantScalar{k: 's', s: "fallback"}},
		{`function f($r){return $r['label'] ?? 'fallback';}probe(f(['label'=>null]));`, constantScalar{k: 's', s: "fallback"}},
		{`function f($r){return $r['label'] ?? 'fallback';}probe(f(['label'=>'kept']));`, constantScalar{k: 's', s: "kept"}},
		{`function f($r){return $r['a']['b'] ?? 'fallback';}probe(f([]));`, constantScalar{k: 's', s: "fallback"}},
		{`function f($r){return isset($r['a']);}probe(f(['a'=>null]));`, constantScalar{k: 'b'}},
		{`function f($r){return isset($r['a']);}probe(f(['a'=>1]));`, constantScalar{k: 'b', b: true}},
		{`function f($r){return isset($r['a'],$r['b']);}probe(f(['a'=>1]));`, constantScalar{k: 'b'}},
		{`function f($r){return array_key_exists('a',$r);}probe(f(['a'=>null]));`, constantScalar{k: 'b', b: true}},
		{`function f($r){return array_key_exists(key:'a',array:$r);}probe(f([]));`, constantScalar{k: 'b'}},
		{`final class K{static function blank($s){return strlen(trim($s))===0;}}function f($s){return K::blank($s);}probe(f(' '));`, constantScalar{k: 'b', b: true}},
		{`class K{final static function add($a,$b=1){return $a+$b;}}function f($s){return K::add(...[$s]);}probe(f(2));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){$sum=0;foreach($xs as $x){$sum=$sum+$x;}return $sum;}probe(f([1,2]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){$sum=7;foreach($xs as $x)$sum=$sum+$x;return $sum;}probe(f([]));`, constantScalar{k: 'i', i: 7}},
		{`function f($xs){$sum='';foreach($xs as $x){$sum=$sum.$x;}return $sum;}probe(f(['b'=>'B','a'=>'A']));`, constantScalar{k: 's', s: "BA"}},
	} {
		e, file := summarized(t, tc.source)
		for _, env := range []*Env{e, NewLocal(file, e.types, e.snapshot, Options{}, nil)} {
			got, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0])))
			if !known || got != tc.want {
				t.Errorf("%s local=%v: got=%+v known=%v want=%+v", tc.source, env.local != nil, got, known, tc.want)
			}
		}
	}
}

func TestPureQueryRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($v,$unused){return array_key_exists('x',$v);}probe(f([],new stdClass));`,
		`function f($r){return $r['missing'][strlen(1)]??'safe';}probe(f([]));`,
		`function f($v){return match($v){1=>2};}probe(f(3));`,
		`function f($v){return match($v){opaque()=>1,default=>2};}probe(f(3));`,
		`function f($v){return $v['missing'];}probe(f([]));`,
		`function f($v){return array_key_exists('x',$v);}probe(f(new stdClass));`,
		`function f($v){return $v['x']??'safe';}probe(f(opaque()));`,
		`function f($v){return array_key_exists(nope:'x',array:$v);}probe(f([]));`,
		`class K{static function go($x){return $x+1;}}function f($v){return K::go($v);}probe(f(1));`,
		`final class K{private static function go($x){return $x+1;}}function f($v){return K::go($v);}probe(f(1));`,
		`class K{static function go($x){return $x+1;}static function f($v){return static::go($v);}}probe(K::f(1));`,
		`function f($xs){$s=0;foreach($xs as &$x){$s=$s+$x;}return $s;}probe(f([1]));`,

		`function f($xs){$s=0;foreach($xs as $x){$s=$s+$x;opaque();}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){$s=$s+$x;}return $s;}probe(f([new stdClass]));`,
		`function f($xs){$s=0;foreach($xs as $x){$s=$s+$x;}return $s;}probe(f([[]]));`,
		`function f($xs){$s=0;foreach($xs as $x){$s=$s+$x;}return $s;}probe(f(opaque()));`,
		`function f($xs){$s=0;foreach($xs as $x){$s=$s+$x;}return $s;}probe(f([9223372036854775807,1]));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s: invented %+v", source, got)
		}
	}
}

func TestPureStaticCompositionRejectsStaleMetadata(t *testing.T) {
	old := `<?php final class K{static function leaf($x){return $x;}}`
	caller := `<?php function wrap($v){return K::leaf($v);}`
	ix := stubsIndexForResource(t, old, caller)
	options := syntax.Options{Version: phpversion.Max}
	snapshot := NewSnapshot(Extract(syntax.Parse("definitions.php", []byte(old), options), ix, phpversion.Max, nil, Options{}))
	for _, params := range []string{"$x,$y", "&$x", "...$x", "$renamed"} {
		current := strings.Replace(old, "($x)", "("+params+")", 1)
		ix := stubsIndexForResource(t, current, caller)
		file := syntax.Parse("caller.php", []byte(caller), options)
		env := NewLocal(file, resourceTypes(file, ix), snapshot, Options{}, nil)
		summary, _ := env.contract("wrap")
		if summary.pure != nil {
			t.Fatalf("stale static signature accepted: %s", params)
		}
	}
}
