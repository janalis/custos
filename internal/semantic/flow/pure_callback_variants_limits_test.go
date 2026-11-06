package flow

import (
	"slices"
	"testing"

	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestPureCallbackVariantRefusals(t *testing.T) {
	for _, source := range []string{
		`function keep($x){return true;}function keep($x){return false;}function f($xs){return count(array_filter($xs,'keep',0));}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){$s=2;}$s+=$x;}return $s;}probe(f([1]));`,
		`function keep($v){return true;}function f($xs){return count(array_filter($xs,'keep',3));}probe(f([1]));`,
		`function keep($v){return true;}function f($xs,$mode){return count(array_filter($xs,'keep',$mode));}probe(f([1],0));`,
		`function keep($v){return true;}function f($xs){return count(array_filter($xs,'keep','0'));}probe(f([1]));`,
		`function keep($v,$k){return $v;}function f($xs){return count(array_filter($xs,'keep',1));}probe(f([$_GET['v']]));`,
		`function keep($k){return $k==='x';}function f($xs){return count(array_filter($xs,'keep',2));}probe(f([new stdClass]));`,
		`function add($a,$b){return $a+$b;}function f($a,$b){return array_map('add',$a,$b)[0];}probe(f([1],[2,3]));`,
		`function add($a,$b){return $a+$b;}function f($a,$b){return array_map('add',$a,$b)[0];}probe(f([1],opaque()));`,
		`function add($a,$b){return $a+$b;}function f($a,$b){return array_map('add',$a,$b,$b)[0];}probe(f([1],[2]));`,
		`function add($a,$b){return $a+$b;}function f($a,$b){return array_map(callback:'add',array:$a,arrays:$b)[0];}probe(f([1],[2]));`,
		`function add($a,$b){return $a+$b;}function f($a,$b){return array_map(...['add',$a,$b])[0];}probe(f([1],[2]));`,
		`function add(&$a,$b){return $a+$b;}function f($a,$b){return array_map('add',$a,$b)[0];}probe(f([1],[2]));`,
		`function f($xs){return array_map('strlen',$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map('trim',$xs)[0];}probe(f([null]));`,
		`function f($xs){return array_map('strtolower',$xs)[0];}probe(f(['YES']));`,
		`function f($xs){return array_map('strlen',$xs,$xs)[0];}probe(f(['yes']));`,
		`function f($xs){return count(array_filter($xs,'strlen'));}probe(f(['yes']));`,
		`function f($xs){return array_map('trim',$xs)[0];}probe(f([new stdClass]));`,
		`function f($xs){return array_map('strlen',$xs)[0];}probe(f([opaque()]));`,
		`function f($xs){$s=0;foreach($xs as $x){if($x===null){break 2;}$s+=$x;}return $s;}probe(f([1,null]));`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){break $x;}$s+=$x;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){break;}$s+=$x;}return $s;}probe(f([$_GET['v'],1]));`,
		`function f($xs){$s=0;foreach($xs as &$x){if($x){break;}$s+=$x;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){break;}$s+=$x;$s+=1;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){break;}$s+=$x;}return $x;}probe(f([1]));`,
		`function f($x){return match($x){[$x['missing']]=>'no',[1]=>'yes',default=>'none'};}probe(f([1]));`,
		`function f($x,$v){return match($x){[$v]=>'no',[1]=>'yes',default=>'none'};}probe(f([1],$_GET['v']));`,
		`function f($x,$v){return in_array($x,[[$v],[1]],true);}probe(f([1],$_GET['v']));`,
		`function f($x){return in_array($x,[[1]],false);}probe(f([1]));`,
		`function f($x){return in_array($x,[[1]],true);}probe(f([new stdClass]));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s: invented %+v", source, got)
		}
	}
}

func TestPureCallbackVariantVersions(t *testing.T) {
	for _, source := range []string{
		`function keep($k){return true;}function f($xs){return array_filter($xs,'keep',0);}`,
		`function keep($k){return true;}function f($xs){return array_filter($xs,'keep',ARRAY_FILTER_USE_KEY);}`,
	} {
		e, file := environmentAt(t, source, phpversion.PHP55)
		out := Extract(file, e.types.Index, phpversion.PHP55, nil, Options{})
		if out.Summaries["f"].pure != nil {
			t.Fatal("legacy filter mode accepted")
		}
	}
	e, file := environmentAt(t, `function keep($k){return true;}function f($xs){return array_filter($xs,'keep',ARRAY_FILTER_USE_KEY);}`, phpversion.PHP56)
	out := Extract(file, e.types.Index, phpversion.PHP56, nil, Options{})
	if out.Summaries["f"].pure == nil {
		t.Fatal("PHP56 key mode refused")
	}
}

func TestPureCallbackVariantBudgets(t *testing.T) {
	e, _ := environment(t, ``, nil)
	array := func(v Value) Value {
		return Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{"x": v}, order: []string{"x"}}}
	}
	value := scalarValue(constantScalar{k: 'i', i: 1})
	record := array(value)
	expr := &pureExpr{kind: pureHigher, left: &pureSlots[0], right: &pureSlots[0], other: &pureSlots[1]}
	for _, tc := range []struct {
		slots []Value
		r     scopeResult
	}{
		{[]Value{record}, scopeResult{}},
		{[]Value{record, value}, scopeResult{}},
		{[]Value{record, array(value)}, scopeResult{copies: MaxCopied}},
	} {
		if _, ok := e.evalPure(expr, tc.slots, &tc.r, MaxDepth); ok {
			t.Fatal("invalid second map input or budget accepted")
		}
	}
	if _, known := pureQueryIdentity(record, record, &scopeResult{operations: MaxTransfers}, MaxDepth); known {
		t.Fatal("query budget bypassed")
	}
	if _, known := pureQueryIdentity(record, record, &scopeResult{}, 1); known {
		t.Fatal("query depth bypassed")
	}
	if _, known := pureQueryIdentity(Value{}, value, &scopeResult{}, MaxDepth); known {
		t.Fatal("unknown payload compared")
	}
	// Exhaustion must not be replaced by a later structural mismatch or match.
	for _, kind := range []uint8{pureMember, pureMatch} {
		query := &pureExpr{kind: kind, left: &pureSlots[0], right: &pureSlots[1], other: pureConstant(constantScalar{k: 'b', b: true}), args: []*pureExpr{&pureSlots[0], pureConstant(constantScalar{k: 'b', b: true})}}
		if _, ok := e.evalPure(query, []Value{record, array(record)}, &scopeResult{operations: MaxTransfers - 3}, MaxDepth); ok {
			t.Fatal("exhausted strict query accepted")
		}
	}
	// Both query expressions are evaluated before any result can be published.
	query := &pureExpr{kind: pureMatch, left: &pureSlots[0], args: []*pureExpr{&pureSlots[1], pureConstant(constantScalar{k: 'b', b: true})}}
	if _, ok := e.evalPure(query, []Value{record}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("missing match test ignored")
	}
}

func TestPureArrayMatchMissingSubject(t *testing.T) {
	e, _ := environment(t, ``, nil)
	if _, ok := e.evalPure(&pureExpr{kind: pureMatch, left: &pureSlots[0]}, nil, &scopeResult{}, MaxDepth); ok {
		t.Fatal("missing subject ignored")
	}
}

func TestPureCallbackVariantsCrossFileDependency(t *testing.T) {
	for _, tc := range []struct{ definition, changed, caller string }{
		{`function cb($x,$y){return $x+$y;}`, `function cb($x,$y){return $x-$y;}`, `function f($x,$y){return array_map('cb',$x,$y);}probe(f([1],[2]));`},
		{`function cb($k){return $k==='x';}`, `function cb($k){return $k==='y';}`, `function f($x){return array_filter($x,'cb',2);}probe(f(['x'=>1]));`},
		{`function cb($v,$k){return $k==='x';}`, `function cb($v,$k){return $k==='y';}`, `function f($x){return array_filter($x,'cb',1);}probe(f(['x'=>1]));`},
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
			t.Fatal("changed callback did not invalidate wrapper")
		}
	}
}
