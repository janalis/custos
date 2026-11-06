package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestPureRecordProgramRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($v){$k=null;return [$k=>$v][0];}probe(f('yes'));`,
		`function f($x){return empty(opaque());}probe(f(1));`,
		`function f($x){return [opaque()=>1];}probe(f(1));`,
		`function f($xs){$a=0;foreach($xs as $x){$a+=$x;}if($a){}foreach($xs as $x){$a+=$x;}return $a;}probe(f([1]));`,
		`function f($x){return $x?'yes':'no';}probe(f(new stdClass));`,
		`function f($x){return empty($x);}probe(f(tmpfile()));`,
		`function f($x,$k){return empty($x[$k]);}probe(f([],opaque()));`,
		`function f($x){return [$x=>'yes'][0];}probe(f(true));`,
		`function f($x){return [$x=>'yes'][0];}probe(f(null));`,
		`function f($x){return [$x=>'yes'][0];}probe(f(1.5));`,
		`function f($x){return [$x=>'yes'][0];}probe(f($_GET['x']));`,
		`function f($x){return [$x=>'yes'][0];}probe(f(-1));`,
		`function f($x){return [$x=>'yes'][0];}probe(f(9223372036854775807));`,
		`function f($x){return [$x=>opaque()]['x'];}probe(f('x'));`,
		`function f($xs){$out=[];foreach($xs as $x){$out[$x['id']]=$x['v'];}return $out;}probe(f([['id'=>$_GET['id'],'v'=>'yes']])['x']);`,
		`function f($xs){$out=[];foreach($xs as $x){$out[$x['id']]=$x['v'];}return $out;}probe(f([['id'=>null,'v'=>'yes']])['x']);`,
		`function f($xs){$out=[];foreach($xs as $x){$out[opaque()]=$x;}return $out;}probe(f([1])['x']);`,
		`function f($x){return ([]+$x)[0];}probe(f(1));`,
		`function f($x){return (['x'=>'left']+opaque())['x'];}probe(f([]));`,
		`function f($x){return (['x'=>'left']+$x)['x'];}probe(f(['x'=>new stdClass]));`,
		`function f($xs){$a=0;foreach($xs as $x){$a+=$x;}$a=0;foreach($xs as $x){$a+=$x;}return $a;}probe(f([1]));`,
		`function f($xs){$a=0;foreach($xs as $x){$a+=$x;}$b=$x;foreach($xs as $x){$b+=$x;}return $b;}probe(f([]));`,
		`function f($xs){$a=0;foreach($xs as $x){$a+=$x;}$b=0;foreach($xs as $x){$b+=$x;}$c=0;foreach($xs as $x){$c+=$x;}return $c;}probe(f([1]));`,
		`function f($xs){$a=0;foreach($xs as $x){$a+=$x;}opaque();foreach($xs as $x){$a+=$x;}return $a;}probe(f([1]));`,
		`function f($xs){$a=0;foreach($xs as $x){$a+=$x;}$b=0;opaque();return $a;}probe(f([1]));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s: invented %+v", source, got)
		}
	}
}

func TestPureRecordProgramVersions(t *testing.T) {
	for _, tc := range []struct {
		source string
		php    phpversion.Version
		known  bool
		want   constantScalar
	}{
		{`function f($x){return empty($x);}probe(f(array()));`, phpversion.PHP53, true, constantScalar{k: 'b', b: true}},
		{`function f($x){return empty(array());}probe(f(1));`, phpversion.PHP53, false, constantScalar{}},
		{`function f($x){return empty(array());}probe(f(1));`, phpversion.PHP55, true, constantScalar{k: 'b', b: true}},
		{`function f($k){$r=array($k=>'yes');return $r['x'];}probe(f('x'));`, phpversion.PHP53, true, constantScalar{k: 's', s: "yes"}},
		{`function f($k){return [$k=>'yes']['x'];}probe(f('x'));`, phpversion.PHP53, false, constantScalar{}},
	} {
		e, file := environmentAt(t, tc.source, tc.php)
		snapshot := NewSnapshot(Extract(file, e.types.Index, tc.php, nil, Options{}))
		got, known := knownScalar(New(file, e.types, snapshot, Options{}).Value(argument(callsNamed(file, "probe")[0])))
		if known != tc.known || known && got != tc.want {
			t.Errorf("%s php=%v got=%+v known=%v", tc.source, tc.php, got, known)
		}
	}
}

func TestPureRecordProgramBudgets(t *testing.T) {
	e, _ := environment(t, ``, nil)
	one := pureConstant(constantScalar{k: 'i', i: 1})
	slot := &pureSlots[0]
	for _, expr := range []*pureExpr{
		{kind: pureEmpty, left: slot},
		{kind: pureChoice, left: slot, right: one, other: one},
		{kind: pureShort, op: syntax.TBooleanAnd, left: one, right: slot},
		{kind: pureBinary, op: syntax.TPlus, left: &pureExpr{kind: pureArray}, right: slot},
	} {
		if _, ok := e.evalPure(expr, nil, &scopeResult{}, MaxDepth); ok {
			t.Fatal("unproven input evaluated")
		}
	}
	array := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{"x": scalarValue(constantScalar{k: 'i'})}, order: []string{"x"}}}
	if _, ok := pureArrayUnion(array, array, &scopeResult{copies: MaxCopied}); ok {
		t.Fatal("union copy limit bypassed")
	}
	if putPureKey(&array, "y", scalarValue(constantScalar{k: 'i'}), &scopeResult{copies: MaxCopied}) {
		t.Fatal("keyed write copy limit bypassed")
	}
	if _, ok := pureArrayUnion(Value{}, array, &scopeResult{}); ok {
		t.Fatal("unknown union accepted")
	}
	if _, ok := e.evalPure(&pureExpr{kind: pureSpreadArray, args: []*pureExpr{slot, one}}, nil, &scopeResult{}, MaxDepth); ok {
		t.Fatal("unknown key evaluated")
	}
}

func TestPureKeyedBuilderBudgets(t *testing.T) {
	e, file := environment(t, `function f($xs,$key){$out=[];foreach($xs as $x){$out[$key]=$x;}return $out;}`, nil)
	compiler := pureCompiler{env: e, nextSlot: 2}
	compiler.bind("xs", &pureSlots[0])
	key := pureConstant(constantScalar{k: 's', s: "x"})
	key.cost = pureLimit
	compiler.bind("key", key)
	var loop *syntax.Foreach
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if n, ok := node.(*syntax.Foreach); ok {
			loop = n
		}
		return true
	})
	if compiler.reduction("out", &pureExpr{kind: pureArray, cost: 1, depth: 1}, loop, []syntax.Stmt{&syntax.Return{Expr: &syntax.Variable{Name: "out"}}}) != nil {
		t.Fatal("key expansion limit bypassed")
	}
	one := pureConstant(constantScalar{k: 'i', i: 1})
	input := &pureExpr{kind: pureArray, args: []*pureExpr{pureConstant(constantScalar{k: 's', s: "0"}), one}}
	for _, args := range [][]*pureExpr{nil, {nil, pureConstant(constantScalar{k: 's', s: "key"})}} {
		expr := &pureExpr{kind: pureBuild, left: input, right: &pureExpr{kind: pureArray}, other: one, args: args}
		if _, ok := e.evalPure(expr, nil, &scopeResult{copies: MaxCopied - 3}, MaxDepth); ok {
			t.Fatal("builder copy budget bypassed")
		}
	}
}
