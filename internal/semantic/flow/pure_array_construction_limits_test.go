package flow

import (
	"math"
	"slices"
	"strconv"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func constructionRecord(values ...Value) Value {
	set := &elementSet{whole: true, width: len(values), items: map[string]Value{}}
	for i, value := range values {
		key := strconv.Itoa(i)
		set.items[key] = value
		set.order = append(set.order, key)
	}
	return Value{Complete: true, plain: true, deep: true, elements: set}
}

func TestPureArrayConstructionRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($n){return array_fill(-1,$n,'v')[0];}probe(f(2));`,
		`function f($n){return count(array_fill(0,$n,'v'));}probe(f(0));`,
		`function f($n){return count(array_fill(0,$n,'v'));}probe(f(-1));`,
		`function f($n){return count(array_fill(0,$n,'v'));}probe(f(65));`,
		`function f($n){return count(array_fill(9223372036854775807,$n,'v'));}probe(f(2));`,
		`function f($n){return count(array_fill(0,$n,'v'));}probe(f('2'));`,
		`function f($n){return count(array_fill(0,$n,'v'));}probe(f(true));`,
		`function f($v){return array_fill(0,1,$v)[0];}probe(f(new stdClass));`,
		`function f($k){return array_fill_keys($k,'v')[0];}probe(f([null]));`,
		`function f($k){return array_fill_keys($k,'v')[0];}probe(f([false]));`,
		`function f($k){return array_fill_keys($k,'v')[0];}probe(f([1.5]));`,
		`function f($k){return array_fill_keys($k,'v')[0];}probe(f([$_GET['k']]));`,
		`function f($k,$v){return count(array_fill_keys($k,$v['missing']));}probe(f([],[]));`,
		`function f($k,$v){return count(array_fill_keys($k,$v));}probe(f([],new stdClass));`,
		`function f($xs){return array_flip($xs)['x'];}probe(f([false,'x']));`,
		`function f($xs){return array_flip($xs)['x'];}probe(f([$_GET['v'],'x']));`,
		`function f($xs){return count(array_chunk($xs,0));}probe(f([]));`,
		`function f($xs){return count(array_chunk($xs,-1));}probe(f([1]));`,
		`function f($xs){return count(array_chunk($xs,'1'));}probe(f([1]));`,
		`function f($xs){return count(array_chunk($xs,1,1));}probe(f([1]));`,
		`function f($xs,$v){return count(array_chunk($xs,1,$v['missing']));}probe(f([],[]));`,
		`function f($xs){return array_sum($xs);}probe(f([1.5]));`,
		`function f($xs){return array_sum($xs);}probe(f(['1']));`,
		`function f($xs){return array_sum($xs);}probe(f([true]));`,
		`function f($xs){return array_sum($xs);}probe(f([null]));`,
		`function f($xs){return array_sum($xs);}probe(f([9223372036854775807,1,-1]));`,
		`function f($xs){return array_sum($xs);}probe(f([-9223372036854775807,-2,2]));`,
		`function f($xs){return array_sum($xs);}probe(f([opaque()]));`,
		`function f($xs){return array_sum($xs);}probe(f([new stdClass]));`,
		`function f($xs){return count(array_fill(nope:$xs));}probe(f(1));`,
		`function f($xs){return array_fill(0,$xs)[0];}probe(f(1));`,
		`function f($xs){return array_fill_keys($xs)[0];}probe(f(['x']));`,
		`function f($xs){return array_flip()[0];}probe(f([]));`,
		`function f($xs){return array_sum($xs,$xs);}probe(f([1]));`,
		`function f($xs){return count(array_chunk($xs));}probe(f([1]));`,
		`function f($xs){return array_sum(&$xs);}probe(f([1]));`,
		`function f($xs){return array_sum($xs);} $v=1;probe(f([&$v]));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s invented %+v", source, got)
		}
	}
}

func TestPureArrayConstructionRuntimeBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	integer := func(i int64) Value { return scalarValue(constantScalar{k: 'i', i: i}) }
	text := scalarValue(constantScalar{k: 's', s: "x"})
	array := constructionRecord(text, text)
	fill := &pureExpr{kind: pureFill, left: &pureSlots[0], right: &pureSlots[1], other: &pureSlots[2]}
	keys := &pureExpr{kind: pureFillKeys, left: &pureSlots[0], right: &pureSlots[1]}
	flip := &pureExpr{kind: pureFlip, left: &pureSlots[0]}
	chunk := &pureExpr{kind: pureChunk, left: &pureSlots[0], right: &pureSlots[1], other: &pureSlots[2]}
	sum := &pureExpr{kind: pureSum, left: &pureSlots[0]}
	sourcedKey := text
	sourcedKey.Sources = []Source{{Kind: "input", Parameter: 0}}
	sourcedInt := integer(1)
	sourcedInt.Sources = sourcedKey.Sources
	deep := constructionRecord(text)
	deep.elements.depth = MaxDepth - 1
	veryDeep := deep
	ds := *deep.elements
	ds.depth = MaxDepth
	veryDeep.elements = &ds
	invalid := text
	invalid.Complete = false
	many := text
	for i := 0; i <= MaxSources; i++ {
		many.Sources = append(many.Sources, Source{Kind: "input", Parameter: i})
	}
	optional := array
	os := *array.elements
	os.optional = true
	optional.elements = &os
	cases := []struct {
		expr  *pureExpr
		slots []Value
		r     scopeResult
	}{
		{fill, []Value{integer(0)}, scopeResult{}},
		{fill, []Value{integer(0), integer(1)}, scopeResult{}},
		{fill, []Value{sourcedInt, integer(1), text}, scopeResult{}},
		{fill, []Value{integer(0), sourcedInt, text}, scopeResult{}},
		{fill, []Value{integer(0), integer(2), text}, scopeResult{copies: MaxCopied}},
		{fill, []Value{integer(0), integer(2), text}, scopeResult{operations: MaxTransfers - 4}},
		{fill, []Value{integer(0), integer(2), deep}, scopeResult{}},
		{fill, []Value{integer(0), integer(1), many}, scopeResult{}},
		{keys, []Value{array}, scopeResult{}},
		{keys, []Value{constructionRecord(sourcedKey), text}, scopeResult{}},
		{keys, []Value{array, invalid}, scopeResult{}},
		{keys, []Value{array, deep}, scopeResult{}},
		{keys, []Value{constructionRecord(), many}, scopeResult{}},
		{keys, []Value{array, text}, scopeResult{copies: MaxCopied}},
		{keys, []Value{optional, text}, scopeResult{}},
		{flip, []Value{constructionRecord(sourcedKey)}, scopeResult{}},
		{flip, []Value{array}, scopeResult{copies: MaxCopied}},
		{chunk, []Value{array}, scopeResult{}},
		{chunk, []Value{array, integer(1)}, scopeResult{}},
		{chunk, []Value{array, sourcedInt, scalarValue(constantScalar{k: 'b'})}, scopeResult{}},
		{chunk, []Value{array, integer(1), scalarValue(constantScalar{k: 'b'})}, scopeResult{copies: MaxCopied}},
		{chunk, []Value{constructionRecord(deep), integer(1), scalarValue(constantScalar{k: 'b'})}, scopeResult{}},
		{chunk, []Value{constructionRecord(veryDeep), integer(1), scalarValue(constantScalar{k: 'b'})}, scopeResult{}},
		{chunk, []Value{constructionRecord(invalid), integer(1), scalarValue(constantScalar{k: 'b'})}, scopeResult{}},
		{sum, []Value{constructionRecord(integer(1))}, scopeResult{copies: MaxCopied}},
		{sum, []Value{constructionRecord(integer(math.MaxInt64), integer(1))}, scopeResult{}},
		{sum, []Value{constructionRecord(integer(math.MinInt64), integer(-1))}, scopeResult{}},
		{sum, []Value{constructionRecord(many)}, scopeResult{}},
	}
	for i, tc := range cases {
		if _, ok := e.evalPure(tc.expr, tc.slots, &tc.r, MaxDepth); ok {
			t.Errorf("case%d accepted unsafe input/budget", i)
		}
	}
	ints := constructionRecord(sourcedInt, integer(2))
	result, ok := e.evalPure(sum, []Value{ints}, &scopeResult{}, MaxDepth)
	if !ok || len(result.Sources) != 1 || result.Safe != everyContext {
		t.Fatal("sum lost numeric provenance/safety")
	}
	a, b := integer(1), integer(2)
	for i := 0; i <= MaxSources; i++ {
		source := Source{Kind: "input", Parameter: i}
		if i < MaxSources/2 {
			a.Sources = append(a.Sources, source)
		} else {
			b.Sources = append(b.Sources, source)
		}
	}
	if _, ok := e.evalPure(sum, []Value{constructionRecord(a, b)}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("sum source union truncated")
	}
	// Overflow of a chunk's value provenance must refuse the complete construction.
	if _, ok := e.evalPure(chunk, []Value{constructionRecord(a, b), integer(2), scalarValue(constantScalar{k: 'b'})}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("chunk source union truncated")
	}
}

func TestPureArrayConstructionCompilerBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	c := pureCompiler{env: e}
	if c.arrayConstruction(&syntax.FuncCall{}, "opaque", nil) != nil || c.arrayConstruction(&syntax.FuncCall{}, "array_fill", nil) != nil {
		t.Fatal("invalid native metadata accepted")
	}
	for _, name := range []string{"array_flip", "array_fill", "array_fill_keys", "array_chunk", "array_sum"} {
		source := `function f($xs){return ` + name + `($xs);}`
		if name == "array_fill" {
			source = `function f($xs){return array_fill(0,1,$xs);}`
		}
		if name == "array_fill_keys" {
			source = `function f($xs){return array_fill_keys(['x'],$xs);}`
		}
		if name == "array_chunk" {
			source = `function f($xs){return array_chunk($xs,1);}`
		}
		e, file := environment(t, source, nil)
		call := callsNamed(file, name)[0]
		params, _, _ := e.declaration(call)
		for _, depth := range []bool{false, true} {
			c := pureCompiler{env: e}
			huge := pureSlots[0]
			if depth {
				huge.depth = MaxDepth
			} else {
				huge.cost = pureLimit
			}
			c.bind("xs", &huge)
			if c.arrayConstruction(call, name, params) != nil {
				t.Fatal("expanded construction bound bypassed")
			}
		}
	}
}

func TestPureArrayConstructionVersions(t *testing.T) {
	for _, tc := range []struct {
		source  string
		version phpversion.Version
		valid   bool
	}{
		{`function f($xs){return array_sum($xs);}`, phpversion.PHP53, true},
		{`function f($xs){return array_flip($xs);}`, phpversion.PHP53, true},
		{`function f($xs){return array_fill(0,1,$xs);}`, phpversion.PHP53, true},
		{`function f($xs){return array_fill_keys(array('x'),$xs);}`, phpversion.PHP53, true},
		{`function f($xs){return array_chunk($xs,1);}`, phpversion.PHP53, true},
		{`function f($xs){return array_sum(array:$xs);}`, phpversion.PHP74, false},
		{`function f($xs){return array_sum(array:$xs);}`, phpversion.PHP80, true},
	} {
		e, file := environmentAt(t, tc.source, tc.version)
		if got := Extract(file, e.types.Index, tc.version, nil, Options{}).Summaries["f"].pure != nil; got != tc.valid {
			t.Errorf("%s %v valid=%v", tc.source, tc.version, got)
		}
	}
}

func TestPureArrayConstructionDependencies(t *testing.T) {
	original, defs := closedFileEnvironment(t, "value.php", `function cb($x){return trim($x);}`, nil)
	snapshot := NewSnapshot(Extract(defs, original.types.Index, phpversion.Max, nil, Options{}))
	consumer, file := closedFileEnvironment(t, "consumer.php", `function f($x){return array_fill(0,1,cb($x));}`, snapshot)
	consumer.types.Index.Add(index.Extract(defs))
	extracted := Extract(file, consumer.types.Index, phpversion.Max, snapshot, Options{})
	if extracted.Summaries["f"].pure == nil || !slices.Contains(extracted.Consulted(), "cb") {
		t.Fatal("construction dependency missing")
	}
	changed, updated := closedFileEnvironment(t, "value.php", `function cb($x){return $x;}`, nil)
	next := NewSnapshot(Extract(updated, changed.types.Index, phpversion.Max, nil, Options{}))
	if extracted.Current(snapshot, next) {
		t.Fatal("changed value dependency retained")
	}
}

func TestPureArrayConstructionCombinedShape(t *testing.T) {
	e, _ := environment(t, ``, nil)
	text := scalarValue(constantScalar{k: 's', s: "x"})
	number := scalarValue(constantScalar{k: 'i', i: 1})
	var large []Value
	for i := 0; i <= pureLimit; i++ {
		large = append(large, text)
	}
	for _, kind := range []uint8{pureFillKeys, pureFlip, pureSum} {
		expr := &pureExpr{kind: kind, left: &pureSlots[0], right: &pureSlots[1]}
		if _, ok := e.evalPure(expr, []Value{constructionRecord(large...), text}, &scopeResult{}, MaxDepth); ok {
			t.Fatal("shape bound bypassed")
		}
	}
	expr := &pureExpr{kind: pureChunk, left: &pureSlots[0], right: &pureSlots[1], other: pureConstant(constantScalar{k: 'b'})}
	if _, ok := e.evalPure(expr, []Value{constructionRecord(large[:pureLimit/2+1]...), number}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("combined outer and inner shape bound bypassed")
	}
}
