package flow

import (
	"slices"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestPureCompositionRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($xs){return array_map(function($x){return $x;yield 1;},$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function($x){return $x;yield from [];},$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function($x){return $x;static $old=1;},$xs)[0];}probe(f([1]));`,

		`function f($xs){return implode($xs);}probe(f(['a']));`,
		`function f($xs){return implode($xs,',');}probe(f(['a']));`,
		`function f($xs){return implode(',',$xs,'extra');}probe(f(['a']));`,
		`function f($xs){return implode(',',$xs);}probe(f([1]));`,
		`function f($xs){return implode(',',$xs);}probe(f([opaque()]));`,
		`function f($xs){return implode(',',$xs);}probe(f([new stdClass]));`,
		`function f($x){return explode('',$x)[0];}probe(f('a'));`,
		`function f($x){return explode(',',$x,2)[0];}probe(f('a,b'));`,
		`function f($x){return explode(',',$x)[0];}probe(f(1));`,
		`function f($x){return explode(',',$x)[0];}probe(f($_GET['x']));`,
		`function f($xs){return array_column($xs,null)[0];}probe(f([['v'=>1]]));`,
		`function f($xs){return array_column($xs,true)[0];}probe(f([['v'=>1]]));`,
		`function f($xs){return array_column($xs,'v')[0];}probe(f([['other'=>1]]));`,
		`function f($xs){return array_column($xs,'v')[0];}probe(f([new stdClass]));`,
		`function f($xs){return array_column($xs,'v')[0];}probe(f([1]));`,
		`function f($xs){return array_column($xs,'v')[0];}probe(f([['v'=>new stdClass]]));`,
		`function f($xs){return count(array_merge($xs));}probe(f([1]));`,
		`function f($xs){return count(array_merge($xs,$xs,$xs));}probe(f([1]));`,
		`function f($xs){return count(array_merge(...[$xs,$xs]));}probe(f([1]));`,
		`function f($xs){return count(array_merge(arrays:$xs,arrays2:$xs));}probe(f([1]));`,
		`function f($xs){return count(array_merge($xs,1));}probe(f([1]));`,
		`function f($xs){return count(array_merge($xs,$xs));}probe(f([new stdClass]));`,
		`function f($xs,$n){return array_map(function($x)use($n){return $x+$n;},$xs)[0];}probe(f([1],2));`,
		`function f($xs){return array_map(function(&$x){return $x;},$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function &($x){return $x;},$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function(...$x){return 1;},$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function($x){static $n=1;return $x+$n;},$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function($x){return $this->v;},$xs)[0];}probe(f([1]));`,
		`function f($xs,$n){return array_map(function($x){return $x+$n;},$xs)[0];}probe(f([1],2));`,
		`function f($xs){return array_map(function($x){return opaque();},$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function($x){yield $x;},$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map(function($x){return f([$x]);},$xs)[0];}probe(f([1]));`,
		`function f($xs){$cb=function($x){return $x+1;};return array_map($cb,$xs)[0];}probe(f([1]));`,
		`function f($xs){foreach($xs as [$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[true]]));`,
		`function f($xs){foreach($xs as [$keep,$v]){if($keep){return 'fixed';}}return 'none';}probe(f([[true]]));`,
		`function f($xs){foreach($xs as [$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[false]]));`,
		`function f($xs){foreach($xs as [$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[$_GET['x'],'no'],[true,'yes']]));`,
		`function f($xs){foreach($xs as [$keep,$v]){if($keep){return $v;}}return $v;}probe(f([]));`,
		`function f($xs){foreach($xs as $keep=>[$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[true,'yes']]));`,
		`function f($xs){foreach($xs as [&$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[true,'yes']]));`,
		`function f($xs){foreach($xs as [$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[true,'yes'],[new stdClass]]));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s invented %+v", source, got)
		}
	}
}

func TestPureCompositionVersions(t *testing.T) {
	for _, tc := range []struct {
		source  string
		version phpversion.Version
		valid   bool
	}{
		{`function f($xs){return array_column($xs,'v');}`, phpversion.PHP54, false},
		{`function f($xs){return array_column($xs,'v');}`, phpversion.PHP55, true},
		{`function f($xs){return array_map(function($x){return $x;},$xs);}`, phpversion.PHP53, true},
		{`function f($xs){return array_map(static function($x){return $x;},$xs);}`, phpversion.PHP53, false},
		{`function f($xs){return array_map(static function($x){return $x;},$xs);}`, phpversion.PHP54, true},
		{`function f($xs){return array_map(function(int $x){return $x;},$xs);}`, phpversion.PHP56, false},
		{`function f($xs){return array_map(function($x):int{return $x;},$xs);}`, phpversion.PHP56, false},
		{`function f($xs){return array_map(function(array $x){return $x;},$xs);}`, phpversion.PHP53, true},
		{`function f($xs){foreach($xs as list($keep,$v)){if($keep){return $v;}}return 'none';}`, phpversion.PHP54, false},
		{`function f($xs){foreach($xs as list($keep,$v)){if($keep){return $v;}}return 'none';}`, phpversion.PHP55, true},
	} {
		e, file := environmentAt(t, tc.source, tc.version)
		out := Extract(file, e.types.Index, tc.version, nil, Options{})
		if (out.Summaries["f"].pure != nil) != tc.valid {
			t.Errorf("%v %s", tc.version, tc.source)
		}
	}
}

func TestPureStringCompositionSources(t *testing.T) {
	e, _ := environment(t, ``, nil)
	text := scalarValue(constantScalar{k: 's', s: "a,b"})
	text.Sources = []Source{{Kind: "input", Parameter: 0}}
	text.Safe = HTML
	separator := scalarValue(constantScalar{k: 's', s: ","})
	separator.Sources = []Source{{Kind: "separator", Parameter: 1}}
	separator.Safe = HTML
	array := func(values ...Value) Value {
		set := &elementSet{whole: true, items: map[string]Value{}, width: len(values)}
		for i, value := range values {
			key := string(rune('a' + i))
			set.items[key] = value
			set.order = append(set.order, key)
		}
		return Value{Complete: true, plain: true, deep: true, elements: set}
	}
	split := &pureExpr{kind: pureSplit, left: &pureSlots[0], right: &pureSlots[1]}
	got, ok := e.evalPure(split, []Value{separator, text}, &scopeResult{}, MaxDepth)
	if !ok || len(got.elements.items) != 2 {
		t.Fatal("split failed")
	}
	for _, value := range got.elements.items {
		if !slices.Equal(value.Sources, text.Sources) || value.Safe != 0 {
			t.Fatal("split provenance or escaping proof")
		}
	}
	join := &pureExpr{kind: pureJoin, left: &pureSlots[0], right: &pureSlots[1]}
	for _, tc := range []struct {
		values  []Value
		sources int
	}{{nil, 0}, {[]Value{text}, 1}, {[]Value{text, text}, 2}} {
		got, ok := e.evalPure(join, []Value{separator, array(tc.values...)}, &scopeResult{}, MaxDepth)
		if !ok || len(got.Sources) != tc.sources || got.Safe != 0 {
			t.Fatal("join source selection or escaping")
		}
	}
	// Do not silently truncate an origin union or accept an oversized input.
	many := text
	for i := 0; i <= MaxSources; i++ {
		many.Sources = append(many.Sources, Source{Kind: "extra", Parameter: i + 2})
	}
	if _, ok := e.evalPure(join, []Value{separator, array(many)}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("source union truncated")
	}
	if _, ok := evalPureSplit(",", many, &scopeResult{}); ok {
		t.Fatal("split source budget bypassed")
	}
	for _, tc := range []struct {
		separator string
		input     string
	}{{",", strings.Repeat("x", 4097)}, {",", strings.Repeat(",", pureLimit)}} {
		if _, ok := evalPureSplit(tc.separator, scalarValue(constantScalar{k: 's', s: tc.input}), &scopeResult{}); ok {
			t.Fatal("split size budget bypassed")
		}
	}
}

func TestPureCompositionDependencies(t *testing.T) {
	original, defs := closedFileEnvironment(t, "callbacks.php", `function cb($x){return $x+1;}`, nil)
	snapshot := NewSnapshot(Extract(defs, original.types.Index, phpversion.Max, nil, Options{}))
	consumer, file := closedFileEnvironment(t, "consumer.php", `function f($xs){return array_map(function($x){return cb($x);},$xs);}`, snapshot)
	consumer.types.Index.Add(index.Extract(defs))
	extracted := Extract(file, consumer.types.Index, phpversion.Max, snapshot, Options{})
	if extracted.Summaries["f"].pure == nil || !slices.Contains(extracted.Consulted(), "cb") {
		t.Fatal("closure dependency missing")
	}
	changed, updated := closedFileEnvironment(t, "callbacks.php", `function cb($x){return $x+2;}`, nil)
	next := NewSnapshot(Extract(updated, changed.types.Index, phpversion.Max, nil, Options{}))
	if extracted.Current(snapshot, next) {
		t.Fatal("changed closure dependency retained")
	}
}

func TestPureCompositionRuntimeEdges(t *testing.T) {
	e, _ := environment(t, ``, nil)
	text := scalarValue(constantScalar{k: 's', s: "x"})
	record := func(values map[string]Value, order []string) Value {
		return Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: values, order: order, width: len(order)}}
	}
	one := record(map[string]Value{"0": text}, []string{"0"})
	row := record(map[string]Value{"v": text}, []string{"v"})
	rows := record(map[string]Value{"0": row}, []string{"0"})
	literals := pureConstant
	column := &pureExpr{kind: pureColumn, right: literals(constantScalar{k: 's', s: "v"})}
	if _, ok := e.evalArrayTransform(column, rows, nil, &scopeResult{operations: MaxTransfers - 1}, MaxDepth); ok {
		t.Fatal("column work budget bypassed")
	}
	if _, ok := e.evalArrayTransform(column, rows, nil, &scopeResult{copies: MaxCopied}, MaxDepth); ok {
		t.Fatal("column copy budget bypassed")
	}
	merge := &pureExpr{kind: pureMerge, right: &pureSlots[0]}
	if _, ok := e.evalArrayTransform(merge, one, []Value{one}, &scopeResult{copies: MaxCopied}, MaxDepth); ok {
		t.Fatal("merge copy budget bypassed")
	}
	join := &pureExpr{kind: pureJoin, left: &pureSlots[0], right: &pureSlots[1]}
	for _, r := range []scopeResult{{copies: MaxCopied}, {operations: MaxTransfers - 3}} {
		if _, ok := e.evalPure(join, []Value{text, one}, &r, MaxDepth); ok {
			t.Fatal("join budget bypassed")
		}
	}
	large := scalarValue(constantScalar{k: 's', s: strings.Repeat("x", 4097)})
	if _, ok := e.evalPure(join, []Value{text, record(map[string]Value{"0": large}, []string{"0"})}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("join output size bypassed")
	}
	left, right := text, text
	for i := 0; i <= MaxSources; i++ {
		source := Source{Kind: "source", Parameter: i}
		if i < MaxSources/2 {
			left.Sources = append(left.Sources, source)
		} else {
			right.Sources = append(right.Sources, source)
		}
	}
	if _, ok := e.evalPure(join, []Value{text, record(map[string]Value{"0": left, "1": right}, []string{"0", "1"})}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("join source union bypassed")
	}
	for _, r := range []scopeResult{{copies: MaxCopied}, {operations: MaxTransfers}} {
		if _, ok := evalPureSplit(",", text, &r); ok {
			t.Fatal("split budget bypassed")
		}
	}
	c := pureCompiler{env: e}
	if c.column(&syntax.FuncCall{}, nil) != nil || c.stringTransform(&syntax.FuncCall{}, "implode", nil) != nil {
		t.Fatal("missing metadata accepted")
	}
	if c.inlineProgram(&syntax.ArrowFunction{Params: make([]*syntax.Param, pureLimit+1)}) != nil {
		t.Fatal("parameter budget bypassed")
	}
	old, _ := environmentAt(t, ``, phpversion.PHP56)
	if (&pureCompiler{env: old}).inlineClosure(&syntax.Closure{Params: []*syntax.Param{nil}}) != nil {
		t.Fatal("malformed legacy parameter accepted")
	}
}

func TestPureCompositionMalformedPrograms(t *testing.T) {
	e, _ := environment(t, ``, nil)
	c := pureCompiler{env: e}
	call := &syntax.FuncCall{Args: &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{ByRef: true}, &syntax.Arg{}}}}
	if c.arrayMerge(call, []index.Param{{Name: "arrays", Variadic: true}}) != nil {
		t.Fatal("reference merge accepted")
	}
	for _, scope := range []syntax.Node{&syntax.Closure{}, &syntax.ArrowFunction{}} {
		if c.inlineProgram(scope) != nil {
			t.Fatal("missing callable body accepted")
		}
	}
}
