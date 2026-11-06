package flow

import (
	"math"
	"slices"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

func TestPureCallableProgramRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($xs){return array_map(null,$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map('missing',$xs)[0];}probe(f([1]));`,
		`function inc(&$x){return $x+1;}function f($xs){return array_map('inc',$xs)[0];}probe(f([1]));`,
		`function inc($x,...$rest){return $x+1;}function f($xs){return array_map('inc',$xs)[0];}probe(f([1]));`,
		`function inc($x,$y){return $x+$y;}function f($xs){return array_map('inc',$xs)[0];}probe(f([1]));`,
		`function f($xs){return array_map('f',$xs)[0];}probe(f([1]));`,
		`function inc($x){return $x+1;}function f($xs){return array_map('inc',$xs,$xs)[0];}probe(f([1]));`,
		`function keep($x){return opaque();}function f($xs){return count(array_filter($xs,'keep'));}probe(f([1]));`,
		`function keep($x){return $x;}function f($xs){return count(array_filter($xs,'keep'));}probe(f([$_GET['x']]));`,
		`function keep($x){return true;}function f($xs){return count(array_filter($xs,'keep'));}probe(f([new stdClass]));`,
		`function f($xs){foreach($xs as $x){if($x){return $x;}}return 0;}probe(f([$_GET['x'],1]));`,
		`function f($xs){foreach($xs as &$x){if($x){return $x;}}return 0;}probe(f([1]));`,
		`function f($xs){foreach($xs as $x){if($x){return $x;}else{return 0;}}return 0;}probe(f([1]));`,
		`function f($xs){foreach($xs as $x){if($x){return opaque();}}return 0;}probe(f([1]));`,
		`function f($xs){foreach($xs as $x){if($x){return $x;}}return $x;}probe(f([]));`,
		`function f($xs){$s=0;foreach($xs as $x){if($x===null){continue 2;}$s+=$x;}return $s;}probe(f([null,1]));`,
		`function f($xs){$s=0;foreach($xs as $x){if(opaque()){continue;}$s+=$x;}return $s;}probe(f([1]));`,
		`function f($xs){$out=[];foreach($xs as $x){if($x===null){continue;}if(true){$out[]=$x;}}return count($out);}probe(f([null]));`,
		`function f($x){return (int)$x;}probe(f('01'));`,
		`function f($x){return (int)$x;}probe(f('9223372036854775808'));`,
		`function f($x){return (int)$x;}probe(f(1.5));`,
		`function f($x){return (string)$x;}probe(f([]));`,
		`function f($x){return (bool)$x;}probe(f(new stdClass));`,
		`function f($xs){return $xs===['x'=>1];}probe(f(['x'=>$_GET['x']]));`,
		`function f($xs){return $xs===['x'=>1];}probe(f(['x'=>new stdClass]));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s: invented %+v", source, got)
		}
	}
}

func TestPureCallableProgramLimits(t *testing.T) {
	e, _ := environment(t, ``, nil)
	slot := &pureSlots[0]
	for _, tc := range []struct {
		expr  *pureExpr
		slots []Value
		r     scopeResult
	}{
		{&pureExpr{kind: pureFind, left: slot}, nil, scopeResult{}},
		{&pureExpr{kind: pureFind, left: &pureExpr{kind: pureArray}}, nil, scopeResult{copies: MaxCopied}},
		{&pureExpr{kind: pureHigher, left: slot}, []Value{{Complete: true}}, scopeResult{}},
		{&pureExpr{kind: pureHigher, left: &pureExpr{kind: pureArray}}, nil, scopeResult{copies: MaxCopied + 1}},
		{&pureExpr{kind: pureCast, left: slot, op: syntax.TIntCast}, []Value{{Complete: true}}, scopeResult{}},
	} {
		if _, ok := e.evalPure(tc.expr, tc.slots, &tc.r, MaxDepth); ok {
			t.Fatal("unknown/budget result proven")
		}
	}
	array := func(values ...Value) Value {
		set := &elementSet{whole: true, items: map[string]Value{}, width: len(values)}
		for i, value := range values {
			key := string(rune('a' + i))
			set.items[key] = value
			set.order = append(set.order, key)
		}
		return Value{Complete: true, plain: true, deep: true, elements: set}
	}
	number := scalarValue(constantScalar{k: 'i', i: 1})
	nan := scalarValue(constantScalar{k: 'f', f: math.NaN()})
	if equal, known := pureStrictRecord(array(nan), array(nan), &scopeResult{}, MaxDepth); !known || equal {
		t.Fatal("NaN identity was invented")
	}
	if _, known := pureStrictRecord(array(number), array(number), &scopeResult{}, 1); known {
		t.Fatal("depth limit bypassed")
	}
	if _, known := pureStrictRecord(array(number), array(number), &scopeResult{operations: MaxTransfers}, MaxDepth); known {
		t.Fatal("work limit bypassed")
	}
}

func TestPureCallableCompilationLimits(t *testing.T) {
	for _, source := range []string{
		`function f($xs){foreach($xs as $x){if($x){return $x;}}opaque();}`,
		`function f($xs){foreach($xs as $x){if($x){opaque();}}return 0;}`,
		`function f($xs){foreach($xs as $x){if($x){return $x;}$y=1;}return 0;}`,
		`function f($xs){foreach($xs as $_GET){if(true){return 1;}}return 0;}`,
		`function f($xs){foreach($xs as $x=>$x){if($x){return $x;}}return 0;}`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){continue $x;}$s+=$x;}return $s;}`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){continue 2;}$s+=$x;}return $s;}`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){continue;}$s+=$x;$s+=1;}return $s;}`,
	} {
		e, file := environment(t, source, nil)
		_, out := Summarize(file, e.types, nil, Options{})
		if out.Summaries["f"].pure != nil {
			t.Errorf("unsupported program: %s", source)
		}
	}
	e, _ := environment(t, ``, nil)
	if (&pureCompiler{env: e}).higher(&syntax.FuncCall{}, "array_map", nil) != nil {
		t.Fatal("missing roles accepted")
	}
}

func TestPureCallableSourceAndBudgetEdges(t *testing.T) {
	e, _ := environment(t, ``, nil)
	value := scalarValue(constantScalar{k: 'i', i: 12})
	value.Sources = []Source{{Kind: "request", Parameter: -1}}
	stringValue := scalarValue(constantScalar{k: 's', s: "12"})
	stringValue.Sources = value.Sources
	for _, op := range []syntax.TokenKind{syntax.TStringCast, syntax.TIntCast, syntax.TBoolCast} {
		got, ok := evalPureCast(&pureExpr{op: op}, value)
		if !ok || len(got.Sources) != 1 || got.Safe&(URL|Path) != 0 {
			t.Fatal("cast lost provenance or validated a destination")
		}
	}
	stringBinding := &pureContract{params: []pureParam{{typ: types.String}}}
	converted, valid := e.pureBind(stringBinding, []Value{value}, true)
	if !valid || len(converted[0].Sources) != 1 || converted[0].Safe&HTML == 0 || converted[0].Safe&(URL|Path) != 0 {
		t.Fatal("numeric string conversion proof")
	}
	binding := &pureContract{params: []pureParam{{typ: types.Int}}}
	bound, ok := e.pureBind(binding, []Value{stringValue}, true)
	if !ok || len(bound[0].Sources) != 1 || bound[0].Safe&HTML == 0 || bound[0].Safe&(URL|Path) != 0 {
		t.Fatal("callback conversion proof")
	}
	binary := &pureExpr{kind: pureBinary, op: syntax.TPlus, left: &pureSlots[0], right: pureConstant(constantScalar{k: 'i', i: 1})}
	got, ok := e.evalPure(binary, []Value{value}, &scopeResult{}, MaxDepth)
	if !ok || len(got.Sources) != 1 || got.Safe&(URL|Path) != 0 {
		t.Fatal("arithmetic lost callback provenance")
	}
	// Incomplete payloads stay unknown; a budget gap must not be overridden by
	// a later known difference.
	set := func(v Value) Value {
		return Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{"a": v}, order: []string{"a"}}}
	}
	if _, known := pureStrictRecord(set(Value{}), set(value), &scopeResult{}, MaxDepth); known {
		t.Fatal("incomplete field compared")
	}
	higher := &pureExpr{kind: pureHigher, left: &pureSlots[0], right: &pureSlots[1]}
	if _, ok := e.evalPure(higher, []Value{set(value)}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("unknown callback return accepted")
	}
	if _, ok := e.evalPureHigher(higher, set(value), nil, &scopeResult{copies: MaxCopied}, MaxDepth); ok {
		t.Fatal("copy budget bypassed")
	}
	// Both sides must be evaluated before array identity is decided.
	compare := &pureExpr{kind: pureBinary, op: syntax.TIsIdentical, left: &pureExpr{kind: pureArray}, right: &pureSlots[0]}
	if _, ok := e.evalPure(compare, nil, &scopeResult{}, MaxDepth); ok {
		t.Fatal("unproven RHS ignored")
	}
}

func TestPureHigherLegacyTypes(t *testing.T) {
	for _, source := range []string{
		`function inc(int $x){return $x+1;}function f($xs){return array_map('inc',$xs);}`,
		`function inc($x):int{return $x+1;}function f($xs){return array_map('inc',$xs);}`,
	} {
		e, file := environmentAt(t, source, phpversion.PHP56)
		out := Extract(file, e.types.Index, phpversion.PHP56, nil, Options{})
		if out.Summaries["f"].pure != nil {
			t.Fatal("legacy scalar type accepted")
		}
	}
}

func TestPureCallableMetadataAndExpansion(t *testing.T) {
	source := `function inc($x){return $x;}function f($xs){return array_map('inc',$xs);}`
	e, file := environment(t, source, nil)
	call := callsNamed(file, "array_map")[0]
	params, _, _ := e.declaration(call)
	one := pureConstant(constantScalar{k: 'i', i: 1})
	huge := *one
	huge.cost = pureLimit
	for _, contract := range []*pureContract{
		{expr: one},
		{params: []pureParam{{name: "wrong"}}, expr: one},
		{params: []pureParam{{name: "x"}}, expr: &huge},
	} {
		e.snapshot = NewSnapshot(&File{Summaries: map[string]Summary{"inc": {pure: contract, Fixed: true}}})
		compiler := pureCompiler{env: e}
		compiler.bind("xs", &pureSlots[0])
		if compiler.higher(call, "array_map", params) != nil {
			t.Fatal("stale or over-budget callback accepted")
		}
	}
	e, file = environment(t, `function f($xs,$fallback){foreach($xs as $x){if($x){return $x;}}return $fallback;}`, nil)
	var loop *syntax.Foreach
	var fallback *syntax.Return
	syntax.InspectFile(file, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.Foreach:
			loop = n
		case *syntax.Return:
			if plainVariable(n.Expr) == "fallback" {
				fallback = n
			}
		}
		return true
	})
	compiler := pureCompiler{env: e, nextSlot: 2}
	compiler.bind("xs", &pureSlots[0])
	compiler.bind("fallback", &huge)
	if compiler.find(loop, []syntax.Stmt{fallback}) != nil {
		t.Fatal("search fallback expansion bypassed budget")
	}
	for _, num := range []syntax.Expr{&syntax.Variable{Name: "x"}, &syntax.Literal{LitKind: syntax.LitInt, Raw: "2"}} {
		e, file = environment(t, `function f($xs){$s=0;foreach($xs as $x){if($x){continue;}$s+=$x;}return $s;}`, nil)
		var returned *syntax.Return
		syntax.InspectFile(file, func(node syntax.Node) bool {
			switch n := node.(type) {
			case *syntax.Foreach:
				loop = n
			case *syntax.Continue:
				n.Num = num
			case *syntax.Return:
				returned = n
			}
			return true
		})
		compiler = pureCompiler{env: e, nextSlot: 1}
		compiler.bind("xs", &pureSlots[0])
		if compiler.reduction("s", one, loop, []syntax.Stmt{returned}) != nil {
			t.Fatal("unsupported continue target accepted")
		}
	}
}

func TestPureHigherCrossFileDependency(t *testing.T) {
	original, defs := closedFileEnvironment(t, "callbacks.php", `function inc($x){return $x+1;}`, nil)
	snapshot := NewSnapshot(Extract(defs, original.types.Index, phpversion.Max, nil, Options{}))
	consumer, source := closedFileEnvironment(t, "caller.php", `function f($xs){return array_map('inc',$xs)[0];}probe(f([1]));`, snapshot)
	consumer.types.Index.Add(index.Extract(defs))
	extracted := Extract(source, consumer.types.Index, phpversion.Max, snapshot, Options{})
	if !slices.Contains(extracted.Consulted(), "inc") {
		t.Fatal("callback dependency missing")
	}
	changed, updated := closedFileEnvironment(t, "callbacks.php", `function inc($x){return $x+2;}`, nil)
	next := NewSnapshot(Extract(updated, changed.types.Index, phpversion.Max, nil, Options{}))
	if extracted.Current(snapshot, next) {
		t.Fatal("changed callback did not invalidate wrapper")
	}
}

func TestPureScalarSourceBudget(t *testing.T) {
	e, _ := environment(t, ``, nil)
	left, right := scalarValue(constantScalar{k: 'i', i: 1}), scalarValue(constantScalar{k: 'i', i: 1})
	for i := 0; i <= MaxSources; i++ {
		source := Source{Kind: "parameter", Parameter: i}
		if i < MaxSources/2 {
			left.Sources = append(left.Sources, source)
		} else {
			right.Sources = append(right.Sources, source)
		}
	}
	expr := &pureExpr{kind: pureBinary, op: syntax.TPlus, left: &pureSlots[0], right: &pureSlots[1]}
	if _, ok := e.evalPure(expr, []Value{left, right}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("truncated provenance authorized an exact result")
	}
}
