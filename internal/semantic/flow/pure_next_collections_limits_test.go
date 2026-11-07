package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestPureNextCollectionsRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($xs){$s=0;foreach($xs as $x){$s+=$x;}return $x;}probe(f([1]));`,
		`function f($xs,$x){$s=0;foreach($xs as $x){$s+=$x;}return $x;}probe(f([],9));`,
		`function f($xs){$s=0;foreach($xs as $k=>$x){$s+=$x;}return $k;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){$s+=$x;}return $s+opaque();}probe(f([1]));`,
		`function f($xs){$out=[];foreach($xs as $x){if($x){$out[]=$x;}else{$out[]=1;}}return $out;}probe(f([1])[0]);`,
		`function f($xs){$out=[];foreach($xs as $x){if($x){$out[]=$x;}elseif($x===null){$out[]=1;}}return $out;}probe(f([1])[0]);`,
		`function f($xs){$out=[];foreach($xs as $x){if($x){$out[]=$x;$out[]=1;}}return $out;}probe(f([1])[0]);`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){$s+=$x;}}return $s;}probe(f([1]));`,
		`function f($xs){$out=[];foreach($xs as $x){if(opaque()){$out[]=$x;}}return $out;}probe(f([1])[0]);`,
		`function f($xs){$out=[];foreach($xs as $x){if($x!==null){$out[]=$x;}}return count($out);}probe(f([$_GET['x']]));`,
		`function f($xs){return array_values($xs)[0];}probe(f([new stdClass]));`,
		`function f($xs){return array_keys($xs,'yes')[0];}probe(f(['x'=>'yes']));`,
		`function f($xs){return array_keys($xs,strict:true)[0];}probe(f(['x'=>'yes']));`,
		`function f($xs){return array_key_first($xs);}probe(f(opaque()));`,
		`function f($v){return is_string($v);}probe(f($_GET['x']));`,
		`function f($v){return is_int($v);}probe(f(new stdClass));`,
		`function f($v){return substr($v,'1');}probe(f('abc'));`,
		`function f($v){return substr($v,1,null);}probe(f('abc'));`,
		`function f($v){return substr($v,1);}probe(f(123));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s: invented %+v", source, got)
		}
	}
}

func TestPureNextCollectionsVersions(t *testing.T) {
	for _, tc := range []struct {
		source string
		php    phpversion.Version
		want   constantScalar
		known  bool
	}{
		{`function f($v){return substr($v,-1,-3);}probe(f('abc'));`, phpversion.PHP56, constantScalar{k: 's'}, true},
		{`function f($v){return substr($v,-1,-99);}probe(f('abc'));`, phpversion.PHP74, constantScalar{k: 'b'}, true},
		{`function f($v){return substr($v,99);}probe(f('abc'));`, phpversion.PHP74, constantScalar{k: 'b'}, true},
		{`function f($v){return substr($v,2,-2);}probe(f('abc'));`, phpversion.PHP74, constantScalar{k: 'b'}, true},
		{`function f($v){return substr($v,3);}probe(f('abc'));`, phpversion.PHP74, constantScalar{k: 's'}, true},
		{`function f($v){return substr($v,0,0);}probe(f('abc'));`, phpversion.PHP74, constantScalar{k: 's'}, true},
		{`function f($v){return substr($v,0);}probe(f(''));`, phpversion.PHP53, constantScalar{k: 'b'}, true},
		{`function f($v){return substr($v,1);}probe(f('abc'));`, phpversion.PHP53, constantScalar{k: 's', s: "bc"}, true},
		{`function f($v){return array_key_first($v);}probe(f(['x'=>1]));`, phpversion.PHP72, constantScalar{}, false},
	} {
		e, file := environmentAt(t, tc.source, tc.php)
		snapshot := NewSnapshot(Extract(file, e.types.Index, tc.php, nil, Options{}))
		env := New(file, e.types, snapshot, Options{})
		got, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0])))
		if known != tc.known || known && got != tc.want {
			t.Errorf("%s php=%v got=%+v known=%v want=%+v", tc.source, tc.php, got, known, tc.want)
		}
	}
}

func TestPureNextNativeBudgets(t *testing.T) {
	e, _ := environment(t, ``, nil)
	compiler := pureCompiler{env: e}
	if compiler.collectionNative(&syntax.FuncCall{}, "array_values", nil) != nil {
		t.Fatal("missing metadata accepted")
	}
	slot := &pureExpr{kind: pureSlot, slot: 0}
	one := pureConstant(constantScalar{k: 'i', i: 1})
	text := pureConstant(constantScalar{k: 's', s: "abc"})
	for _, tc := range []struct {
		expr  *pureExpr
		slots []Value
		r     scopeResult
	}{
		{&pureExpr{kind: pureLet, left: slot, right: one}, nil, scopeResult{}},
		{&pureExpr{kind: pureLet, left: one, right: one}, nil, scopeResult{copies: MaxCopied}},
		{&pureExpr{kind: pureTypeTest, left: slot}, []Value{{Complete: true}}, scopeResult{}},
		{&pureExpr{kind: pureTypeTest, left: slot}, []Value{{Complete: true, Identity: 1}}, scopeResult{}},
		{&pureExpr{kind: pureOrderedArray, left: slot}, []Value{{Complete: true}}, scopeResult{}},
		{&pureExpr{kind: pureOrderedArray, left: &pureExpr{kind: pureArray, args: []*pureExpr{text, one}}}, nil, scopeResult{copies: MaxCopied - 1}},
		{&pureExpr{kind: pureSubstring, left: text, right: slot}, nil, scopeResult{}},
		{&pureExpr{kind: pureSubstring, left: text, right: one, other: slot}, nil, scopeResult{}},
		{&pureExpr{kind: pureSubstring, left: pureConstant(constantScalar{k: 's', s: strings.Repeat("a", 4097)}), right: one}, nil, scopeResult{}},
	} {
		if _, ok := e.evalPure(tc.expr, tc.slots, &tc.r, MaxDepth); ok {
			t.Errorf("unsafe/budget result: %+v", tc.expr)
		}
	}
}

func TestPureNextCompilerBudgets(t *testing.T) {
	e, file := environment(t, `function f($xs,$g){$out=[];foreach($xs as $x){if($g){$out[]=$x;}}return $out;}`, nil)
	var loop *syntax.Foreach
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if n, ok := node.(*syntax.Foreach); ok {
			loop = n
		}
		return true
	})
	compiler := pureCompiler{env: e, nextSlot: 2}
	compiler.bind("xs", &pureSlots[0])
	guard := pureConstant(constantScalar{k: 'b', b: true})
	guard.cost = pureLimit
	compiler.bind("g", guard)
	ret := &syntax.Return{Expr: &syntax.Variable{Name: "out"}}
	if compiler.reduction("out", &pureExpr{kind: pureArray, cost: 1, depth: 1}, loop, []syntax.Stmt{ret}) != nil {
		t.Fatal("filter cost bypassed")
	}
	e, file = environment(t, `function f($v){return array_values($v);}`, nil)
	compiler = pureCompiler{env: e}
	huge := pureSlots[0]
	huge.cost = pureLimit
	compiler.bind("v", &huge)
	call := callsNamed(file, "array_values")[0]
	params, _, _ := e.declaration(call)
	if compiler.collectionNative(call, "array_values", params) != nil {
		t.Fatal("native child cost bypassed")
	}
	e, file = environment(t, `function f($xs){$s=0;foreach($xs as $x){$s=$x;}return $s+1;}`, nil)
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if n, ok := node.(*syntax.Foreach); ok {
			loop = n
		}
		return true
	})
	compiler = pureCompiler{env: e, nextSlot: 1, work: pureLimit - 3}
	compiler.bind("xs", &pureSlots[0])
	var returned *syntax.Return
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if n, ok := node.(*syntax.Return); ok {
			returned = n
		}
		return true
	})
	if compiler.reduction("s", pureConstant(constantScalar{k: 'i'}), loop, []syntax.Stmt{returned}) != nil {
		t.Fatal("tail compiler budget bypassed")
	}
}

func TestPureSliceProvenanceAndDiscardedLoop(t *testing.T) {
	e, _ := environment(t, ``, nil)
	value := scalarValue(constantScalar{k: 's', s: "abc"})
	value.Sources = []Source{{Kind: "request", Parameter: -1}}
	value.Safe = allContexts
	expr := &pureExpr{kind: pureSubstring, left: &pureSlots[0], right: pureConstant(constantScalar{k: 'i', i: 1})}
	got, ok := e.evalPure(expr, []Value{value}, &scopeResult{}, MaxDepth)
	scalar, known := knownScalar(got)
	if !ok || !known || scalar.s != "bc" || len(got.Sources) != 1 || got.Safe != 0 || value.Safe != allContexts {
		t.Fatalf("slice lost provenance or reused escaping: %+v", got)
	}
	loop := &pureExpr{kind: pureReduce, left: &pureSlots[0], right: pureConstant(constantScalar{k: 'i'}), other: pureConstant(constantScalar{k: 'i'})}
	tail := &pureExpr{kind: pureLet, left: loop, right: pureConstant(constantScalar{k: 'i', i: 1})}
	if _, ok := e.evalPure(tail, []Value{{}}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("discarded loop was not evaluated")
	}
}
