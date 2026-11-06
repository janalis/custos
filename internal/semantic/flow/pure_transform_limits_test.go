package flow

import (
	"slices"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestPureTransformRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($xs){return count(array_slice($xs));}probe(f([1]));`,
		`function f($xs){return count(array_values());}probe(f([1]));`,
		`function f($xs){return count(array_keys());}probe(f([1]));`,

		`function f($xs,$n){return array_map(fn($x)=>$x+$n,$xs)[0];}probe(f([1],2));`,
		`function f($xs){return array_map(fn($x)=>$this->v,$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(fn(&$x)=>$x,$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(fn(...$xs)=>1,$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(fn($x)=>opaque(),$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(fn($x)=>f([$x]),$xs)[0];}probe(f([1]));`,
		`function f($xs){$cb=fn($x)=>$x+1;return array_map($cb,$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function($x){return $x+1;},$xs)[0];}probe(f([1]));`,
		`function f($xs){return count(array_map(fn()=>1,$xs));}probe(f([1]));`,
		`function f($xs){return array_map(fn($x,$y)=>$x+$y,$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(fn($x)=>new stdClass,$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(fn($x)=>yield $x,$xs)[0];}probe(f([1]));`,
		`function add($c,$x){return $c+$x;}function f($xs){return array_reduce($xs,'add');}probe(f([1]));`,
		`function add($c,$x){return $c+$x;}function f($xs){return array_reduce($xs,'add',[]);}probe(f([1]));`,
		`function add(&$c,$x){return $c+$x;}function f($xs){return array_reduce($xs,'add',0);}probe(f([1]));`,
		`function f($xs){return array_reduce($xs,'missing',0);}probe(f([1]));`,
		`function f($xs){return array_reduce($xs,'strlen',0);}probe(f([1]));`,
		`function f($xs){return array_reduce($xs,fn($c,$x)=>$c+$x,0);}probe(f([1]));`,
		`function add($c,$x){return $x;}function f($xs){return array_reduce($xs,'add',0);}probe(f([$_GET['v'],'fixed']));`,
		`function add($c,$x){return [];}function f($xs){return array_reduce($xs,'add',0);}probe(f([1]));`,
		`function add($c,$x){return $c+$x;}function f($xs){return array_reduce($xs,'add',0);}probe(f([new stdClass]));`,
		`function f($xs){return count(array_slice($xs,'1'));}probe(f([1,2]));`,
		`function f($xs){return count(array_slice($xs,0,'1'));}probe(f([1,2]));`,
		`function f($xs){return count(array_slice($xs,0,null,1));}probe(f([1,2]));`,
		`function f($xs,$n){return count(array_slice($xs,$n));}probe(f([1],$_GET['v']));`,
		`function f($xs){return count(array_slice($xs,0));}probe(f([new stdClass]));`,
		`function f($xs){return count(array_keys($xs,1));}probe(f([1]));`,
		`function f($xs){return count(array_keys($xs,1,false));}probe(f([1]));`,
		`function f($xs){return count(array_keys($xs,1,true));}probe(f([$_GET['v'],1]));`,
		`function f($xs){return count(array_keys($xs,[1],true));}probe(f([[new stdClass]]));`,
		`function f($xs){return count(array_keys(array:$xs,strict:true));}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as [$a,$b]){$s+=$a;}return $s;}probe(f([[1]]));`,
		`function f($xs){$s=0;foreach($xs as [$a,$b]){if($a){break;}$s+=$a;}return $s;}probe(f([[1]]));`,
		`function f($xs){$s=0;foreach($xs as [[$a]]){$s+=$a;}return $s;}probe(f([[[1]]]));`,
		`function f($xs){$s=0;foreach($xs as [&$a]){$s+=$a;}return $s;}probe(f([[1]]));`,
		`function f($xs){$s=0;foreach($xs as [$s]){$s+=1;}return $s;}probe(f([[1]]));`,
		`function f($xs){$s=0;foreach($xs as [$a,$a]){$s+=$a;}return $s;}probe(f([[1,2]]));`,
		`function f($xs){$s=0;foreach($xs as $a=>[$a]){$s+=$a;}return $s;}probe(f([[1]]));`,
		`function f($xs){$s=0;foreach($xs as [$a]){$s+=$a;}return $a;}probe(f([[1]]));`,
		`function f($xs){$s=0;foreach($xs as ['a'=>$a]){$s+=$a;}return $s;}probe(f([[]]));`,
		`function f($xs){$s=0;foreach($xs as [$a]){$s+=$a;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as [$a]){$s+=$a;}return $s;}probe(f([[new stdClass]]));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s: invented %+v", source, got)
		}
	}
}

func TestPureTransformVersions(t *testing.T) {
	for _, tc := range []struct {
		source  string
		version phpversion.Version
		valid   bool
	}{
		{`function f($xs){return array_map(fn($x)=>$x,$xs);}`, phpversion.PHP73, false},
		{`function f($xs){return array_map(fn($x)=>$x,$xs);}`, phpversion.PHP74, true},
		{`function f($xs){$s=0;foreach($xs as list($a)){$s+=$a;}return $s;}`, phpversion.PHP54, false},
		{`function f($xs){$s=0;foreach($xs as list($a)){$s+=$a;}return $s;}`, phpversion.PHP55, true},
		{`function f($xs){$s=0;foreach($xs as [$a]){$s+=$a;}return $s;}`, phpversion.PHP70, false},
		{`function f($xs){$s=0;foreach($xs as [$a]){$s+=$a;}return $s;}`, phpversion.PHP71, true},
		{`function f($xs){$s=0;foreach($xs as list('a'=>$a)){$s+=$a;}return $s;}`, phpversion.PHP70, false},
	} {
		e, file := environmentAt(t, tc.source, tc.version)
		out := Extract(file, e.types.Index, tc.version, nil, Options{})
		if (out.Summaries["f"].pure != nil) != tc.valid {
			t.Errorf("version %v %s", tc.version, tc.source)
		}
	}
}

func TestPureTransformRuntimeBudgets(t *testing.T) {
	e, _ := environment(t, ``, nil)
	num := scalarValue(constantScalar{k: 'i', i: 1})
	record := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{"0": num}, order: []string{"0"}}}
	lit := pureConstant
	slice := &pureExpr{kind: pureSlice, left: &pureSlots[0], right: lit(constantScalar{k: 'i'}), other: lit(constantScalar{k: 'n'}), args: []*pureExpr{lit(constantScalar{k: 'b'})}}
	keys := &pureExpr{kind: pureFilteredKeys, left: &pureSlots[0], right: lit(constantScalar{k: 'i', i: 1}), other: lit(constantScalar{k: 'b', b: true})}
	fold := &pureExpr{kind: pureFold, left: &pureSlots[0], right: lit(constantScalar{k: 'i'}), other: &pureSlots[0]}
	for _, tc := range []struct {
		expr  *pureExpr
		slots []Value
		r     scopeResult
	}{
		{slice, []Value{record}, scopeResult{copies: MaxCopied}},
		{keys, []Value{record}, scopeResult{copies: MaxCopied}},
		{keys, []Value{record}, scopeResult{operations: MaxTransfers - 4}},
		{fold, nil, scopeResult{}},
		{fold, []Value{record}, scopeResult{operations: MaxTransfers - 3}},
		{&pureExpr{kind: pureInline}, nil, scopeResult{}},
	} {
		if _, ok := e.evalPure(tc.expr, tc.slots, &tc.r, MaxDepth); ok {
			t.Fatal("invalid or exhausted transform accepted")
		}
	}
	if bindPureFields([]pureField{{key: "0"}}, record, make([]Value, 4), 0, &scopeResult{operations: MaxTransfers}) {
		t.Fatal("field work budget bypassed")
	}
	if bindPureFields([]pureField{{key: "0"}}, num, make([]Value, 4), 0, &scopeResult{}) {
		t.Fatal("scalar row accepted")
	}
	compiler := pureCompiler{env: e, inlineDepth: MaxDepth}
	if compiler.inlineArrow(&syntax.ArrowFunction{}) != nil {
		t.Fatal("inline depth bypassed")
	}
	if compiler.fold(&syntax.FuncCall{}, nil) != nil || compiler.slice(&syntax.FuncCall{}, nil) != nil {
		t.Fatal("missing native metadata accepted")
	}
	// Defaults cannot become callable values in any other argument role.
	for _, src := range []string{`function f($xs){return array_map('trim',[fn($x)=>$x])[0];}probe(f([1]));`, `function f($xs){return array_map(fn($x)=>$x,fn($x)=>$x)[0];}probe(f([1]));`} {
		env, file := summarized(t, src)
		if _, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Fatal("callable value escaped")
		}
	}
}

func TestPureTransformCompilerEdges(t *testing.T) {
	e, _ := environment(t, ``, nil)
	c := pureCompiler{env: e}
	for _, value := range []syntax.Expr{
		&syntax.Variable{NameExpr: &syntax.Variable{Name: "x"}},
		&syntax.Array{Short: true, Items: make([]*syntax.ArrayItem, pureLimit+1)},
		&syntax.List{Items: []*syntax.ArrayItem{nil}},
		&syntax.List{Items: []*syntax.ArrayItem{{Key: &syntax.Literal{LitKind: syntax.LitInt, Raw: "0"}}}},
		&syntax.Array{Short: true, Items: []*syntax.ArrayItem{{Key: &syntax.Variable{Name: "k"}, Value: &syntax.Variable{Name: "v"}}}},
	} {
		if fields, _ := c.loopFields(value, "s"); fields != nil {
			t.Fatal("invalid destructuring accepted")
		}
	}
	duplicate := &syntax.ArrowFunction{Params: []*syntax.Param{{Var: &syntax.Variable{Name: "x"}}, {Var: &syntax.Variable{Name: "x"}}}}
	if c.inlineArrow(duplicate) != nil {
		t.Fatal("duplicate arrow parameter accepted")
	}
	for _, tc := range []struct {
		source, native string
		cost           int
	}{
		{`function f($xs,$v){return array_slice($xs,$v);}`, "array_slice", pureLimit},
		{`function f($xs,$v){return array_slice($xs,0,null,$v);}`, "array_slice", pureLimit},
		{`function f($xs,$v){return array_keys($v);}`, "array_keys", pureLimit},
	} {
		env, file := environment(t, tc.source, nil)
		call := callsNamed(file, tc.native)[0]
		params, _, _ := env.declaration(call)
		compiler := pureCompiler{env: env}
		compiler.bind("xs", &pureSlots[0])
		huge := pureSlots[1]
		huge.cost = tc.cost
		compiler.bind("v", &huge)
		var got *pureExpr
		if tc.native == "array_slice" {
			got = compiler.slice(call, params)
		} else {
			got = compiler.collectionNative(call, tc.native, params)
		}
		if got != nil {
			t.Fatal("expanded collection budget bypassed")
		}
	}
}

func TestPureTransformDependencies(t *testing.T) {
	for _, tc := range []struct{ definition, changed, caller string }{
		{`function cb($c,$x){return $c+$x;}`, `function cb($c,$x){return $c-$x;}`, `function f($xs){return array_reduce($xs,'cb',0);}`},
		{`function cb($x){return $x+1;}`, `function cb($x){return $x+2;}`, `function f($xs){return array_map(fn($x)=>cb($x),$xs);}`},
	} {
		original, defs := closedFileEnvironment(t, "callbacks.php", tc.definition, nil)
		snapshot := NewSnapshot(Extract(defs, original.types.Index, phpversion.Max, nil, Options{}))
		consumer, file := closedFileEnvironment(t, "consumer.php", tc.caller, snapshot)
		consumer.types.Index.Add(index.Extract(defs))
		extracted := Extract(file, consumer.types.Index, phpversion.Max, snapshot, Options{})
		if extracted.Summaries["f"].pure == nil || !slices.Contains(extracted.Consulted(), "cb") {
			t.Fatal("callback dependency missing")
		}
		changed, updated := closedFileEnvironment(t, "callbacks.php", tc.changed, nil)
		next := NewSnapshot(Extract(updated, changed.types.Index, phpversion.Max, nil, Options{}))
		if extracted.Current(snapshot, next) {
			t.Fatal("callback change did not invalidate wrapper")
		}
	}
}
