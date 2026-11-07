package flow

import (
	"slices"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestPureBindingVersions(t *testing.T) {
	for _, tc := range []struct {
		source  string
		version phpversion.Version
		valid   bool
	}{
		{`function f($xs){list($a)=$xs;return $a;}`, phpversion.PHP53, true},
		{`function f($xs){[$a]=$xs;return $a;}`, phpversion.PHP70, false},
		{`function f($xs){[$a]=$xs;return $a;}`, phpversion.PHP71, true},
		{`function f($xs){list('x'=>$a)=$xs;return $a;}`, phpversion.PHP70, false},
		{`function f($xs){list('x'=>$a)=$xs;return $a;}`, phpversion.PHP71, true},
		{`function f($xs){return array_column($xs,'v','id');}`, phpversion.PHP54, false},
		{`function f($xs){return array_column($xs,'v','id');}`, phpversion.PHP55, true},
	} {
		e, file := environmentAt(t, tc.source, tc.version)
		if got := Extract(file, e.types.Index, tc.version, nil, Options{}).Summaries["f"].pure != nil; got != tc.valid {
			t.Errorf("%s %v valid=%v", tc.source, tc.version, got)
		}
	}
}

func TestPureBindingProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($xs){return array_column($xs,'v','id');}probe(f([['id'=>'x','v'=>$_GET['v']],['id'=>'x','v'=>'safe']])['x']);`, false},
		{`function f($xs){return array_column($xs,'v','id');}probe(f([['id'=>'x','v'=>'safe'],['id'=>'x','v'=>$_GET['v']]])['x']);`, true},
		{`function f($k,$v){return array_combine($k,$v);}probe(f(['x','x'],[$_GET['v'],'safe'])['x']);`, false},
		{`function f($k,$v){return array_combine($k,$v);}probe(f(['x','x'],['safe',$_GET['v']])['x']);`, true},
		{`function f($a,$b){return array_replace($a,$b);}probe(f(['x'=>$_GET['v']],['x'=>'safe'])['x']);`, false},
		{`function f($a,$b){return array_replace($a,$b);}probe(f(['x'=>'safe'],['x'=>$_GET['v']])['x']);`, true},
		{`function f($xs){[$a,$b]=$xs;return $a;}probe(f(['safe',$_GET['v']]));`, false},
		{`function f($xs){[$a,$b]=$xs;return $b;}probe(f(['safe',$_GET['v']]));`, true},
		{`function f($xs){$out=[];foreach($xs as $x){$v=$x['v'];$out[]=$v;}return $out;}probe(f([['v'=>'safe','unused'=>$_GET['v']]])[0]);`, false},
		{`function f($xs){$out=[];foreach($xs as $x){$v=$x['v'];$out[]=$v;}return $out;}probe(f([['v'=>$_GET['v']]])[0]);`, true},
	} {
		e, file := summarized(t, tc.source)
		v := e.Value(argument(callsNamed(file, "probe")[0]))
		if got := len(v.Sources) != 0; got != tc.tainted {
			t.Errorf("%s tainted=%v", tc.source, got)
		}
	}
}

func TestPureBindingRuntimeLimits(t *testing.T) {
	e, _ := environment(t, ``, nil)
	text := scalarValue(constantScalar{k: 's', s: "x"})
	record := func(values ...Value) Value {
		set := &elementSet{whole: true, items: map[string]Value{}}
		for i, v := range values {
			key := string(rune('0' + i))
			putWinningEntry(set, key, v)
		}
		return winningRecord(set)
	}
	one := record(text)
	combine := &pureExpr{kind: pureCombine, left: &pureSlots[0], right: &pureSlots[1]}
	sourced := text
	sourced.Sources = []Source{{Kind: "input", Parameter: 0}}
	for _, slots := range [][]Value{{record(sourced), one}, {record(scalarValue(constantScalar{k: 'b'})), one}, {one, record()}, {record(), record()}} {
		if _, ok := e.evalPure(combine, slots, &scopeResult{}, MaxDepth); ok {
			t.Fatal("unsafe combine accepted")
		}
	}
	row := winningRecord(&elementSet{whole: true, items: map[string]Value{"v": text, "id": sourced}, order: []string{"v", "id"}})
	column := &pureExpr{kind: pureColumn, left: &pureSlots[0], right: pureConstant(constantScalar{k: 's', s: "v"}), other: pureConstant(constantScalar{k: 's', s: "id"})}
	if _, ok := e.evalPure(column, []Value{record(row)}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("source-bearing index accepted")
	}
	row.elements.items["id"] = text
	for _, r := range []scopeResult{{copies: MaxCopied}, {operations: MaxTransfers - 3}} {
		if _, ok := e.evalPure(column, []Value{record(row)}, &r, MaxDepth); ok {
			t.Fatal("column budget bypassed")
		}
	}
	destruct := &pureExpr{kind: pureDestructure, left: &pureSlots[0], right: pureConstant(constantScalar{k: 'i', i: 1}), destruct: []pureField{{name: "a", key: "0"}}}
	for _, tc := range []struct {
		slots []Value
		r     scopeResult
	}{{nil, scopeResult{}}, {[]Value{one}, scopeResult{copies: MaxCopied}}, {[]Value{record()}, scopeResult{}}} {
		if _, ok := e.evalPure(destruct, tc.slots, &tc.r, MaxDepth); ok {
			t.Fatal("destructure failure bypassed")
		}
	}
	prepare := &pureExpr{kind: pureReduce, left: &pureSlots[0], right: pureConstant(constantScalar{k: 'i'}), other: pureConstant(constantScalar{k: 'i'}), prepared: true, args: []*pureExpr{&pureSlots[9]}}
	if _, ok := e.evalPure(prepare, []Value{one}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("unused preparation read bypassed")
	}
	c := pureCompiler{env: e}
	if c.pairedArray(&syntax.FuncCall{}, "array_combine", nil) != nil {
		t.Fatal("missing metadata accepted")
	}
	params := []index.Param{{Name: "array"}, {Name: "replacements", Variadic: true}}
	if c.pairedArray(&syntax.FuncCall{}, "array_replace", params) != nil {
		t.Fatal("missing operands accepted")
	}
}

func TestPureBindingCompilerLimits(t *testing.T) {
	for _, tc := range []struct{ source, native string }{
		{`function f($xs){return array_column(nope:$xs);}`, "array_column"},
		{`function f($xs){return array_combine(nope:$xs,values:$xs);}`, "array_combine"},
	} {
		e, file := environment(t, tc.source, nil)
		c := pureCompiler{env: e}
		c.bind("xs", &pureSlots[0])
		call := callsNamed(file, tc.native)[0]
		params, _, _ := e.declaration(call)
		if c.native(call, tc.native, params) != nil {
			t.Fatal("invalid role accepted")
		}
	}
	for _, source := range []string{
		`function f($xs){[$a]=$xs;opaque();return $a;}`,
		`function f($xs){$s=0;foreach($xs as $x){$n+=1;$s+=$n;}return $s;}`,
		`function f($xs){$s=0;foreach($xs as $x){if($x){}else{} $s+=1;}return $s;}`,
		`function f($a,$b,$c,$d,$e,$f,$g,$h,$n){$s=0;foreach($a as $x){$n=$x;$s+=$n;}return $s;}`,
	} {
		e, file := environment(t, source, nil)
		if Extract(file, e.types.Index, e.types.PHP, nil, Options{}).Summaries["f"].pure != nil {
			t.Fatalf("invalid program accepted %s", source)
		}
	}
	e, file := environment(t, `function f($xs,$v){[$a]=$xs;return $v;}`, nil)
	var assign *syntax.Assign
	var ret *syntax.Return
	syntax.InspectFile(file, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Assign:
			assign = n
		case *syntax.Return:
			ret = n
		}
		return true
	})
	c := pureCompiler{env: e}
	c.bind("xs", &pureSlots[0])
	large := pureSlots[1]
	large.cost = pureLimit - 2
	c.bind("v", &large)
	if c.destructure(assign, []syntax.Stmt{ret}, MaxDepth) != nil {
		t.Fatal("field cost bypassed")
	}
	e, file = environment(t, `function f($xs,$v){$s=0;foreach($xs as $x){$n=$v;$s+=1;}return $s;}`, nil)
	var loop *syntax.Foreach
	syntax.InspectFile(file, func(n syntax.Node) bool {
		if n, ok := n.(*syntax.Foreach); ok {
			loop = n
		}
		return true
	})
	c = pureCompiler{env: e}
	c.bind("xs", &pureSlots[0])
	large.cost = pureLimit
	c.bind("v", &large)
	if c.reduction("s", pureConstant(constantScalar{k: 'i'}), loop, []syntax.Stmt{&syntax.Return{Expr: &syntax.Variable{Name: "s"}}}) != nil {
		t.Fatal("preparation cost bypassed")
	}
}

func TestPureBindingDependencies(t *testing.T) {
	original, defs := closedFileEnvironment(t, "prepare.php", `function cb($x){return trim($x);}`, nil)
	snapshot := NewSnapshot(Extract(defs, original.types.Index, phpversion.Max, nil, Options{}))
	consumer, file := closedFileEnvironment(t, "consumer.php", `function f($xs){$out=[];foreach($xs as $x){$v=cb($x);$out[]=$v;}return $out;}`, snapshot)
	consumer.types.Index.Add(index.Extract(defs))
	extracted := Extract(file, consumer.types.Index, phpversion.Max, snapshot, Options{})
	if extracted.Summaries["f"].pure == nil || !slices.Contains(extracted.Consulted(), "cb") {
		t.Fatal("preparation dependency missing")
	}
	changed, updated := closedFileEnvironment(t, "prepare.php", `function cb($x){return $x;}`, nil)
	next := NewSnapshot(Extract(updated, changed.types.Index, phpversion.Max, nil, Options{}))
	if extracted.Current(snapshot, next) {
		t.Fatal("changed preparation dependency retained")
	}
}

func TestPureBindingShapeAndSourceBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	scalar := scalarValue(constantScalar{k: 's', s: "value"})
	record := func(items map[string]Value, order ...string) Value {
		return winningRecord(&elementSet{whole: true, items: items, order: order})
	}
	keys := record(map[string]Value{"0": scalarValue(constantScalar{k: 's', s: "a"}), "1": scalarValue(constantScalar{k: 's', s: "b"})}, "0", "1")
	a, b := scalar, scalar
	for i := 0; i <= MaxSources; i++ {
		source := Source{Kind: "input", Parameter: i}
		if i < MaxSources/2 {
			a.Sources = append(a.Sources, source)
		} else {
			b.Sources = append(b.Sources, source)
		}
	}
	values := record(map[string]Value{"0": a, "1": b}, "0", "1")
	// Build the complete input without relying on its aggregate source union;
	// winning values must still enforce the result union bound.
	values.Complete = true
	for _, kind := range []uint8{pureCombine, pureArrayReplace} {
		left, right := keys, values
		if kind == pureArrayReplace {
			left = record(map[string]Value{})
			right = values
		}
		expr := &pureExpr{kind: kind, left: &pureSlots[0], right: &pureSlots[1]}
		if _, ok := e.evalPure(expr, []Value{left, right}, &scopeResult{}, MaxDepth); ok {
			t.Fatal("winning source union truncated")
		}
	}
	optional := keys
	set := *keys.elements
	set.optional = true
	optional.elements = &set
	expr := &pureExpr{kind: pureCombine, left: &pureSlots[0], right: &pureSlots[1]}
	if _, ok := e.evalPure(expr, []Value{optional, keys}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("optional key shape accepted")
	}
	for _, r := range []scopeResult{{copies: MaxCopied}, {operations: MaxTransfers - 3}} {
		if _, ok := e.evalPure(expr, []Value{keys, keys}, &r, MaxDepth); ok {
			t.Fatal("combined budget bypassed")
		}
	}
}

func TestPureTransformCombinedDepth(t *testing.T) {
	e, _ := environment(t, ``, nil)
	nested := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{}, depth: MaxDepth - 1}}
	keys := winningRecord(&elementSet{whole: true, items: map[string]Value{"0": scalarValue(constantScalar{k: 's', s: "a"}), "1": scalarValue(constantScalar{k: 's', s: "b"})}, order: []string{"0", "1"}})
	values := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{"0": nested, "1": nested}, order: []string{"0", "1"}}}
	for _, kind := range []uint8{pureCombine, pureArrayReplace} {
		expr := &pureExpr{kind: kind, left: &pureSlots[0], right: &pureSlots[1]}
		result, ok := e.evalPure(expr, []Value{keys, values}, &scopeResult{}, MaxDepth)
		key := "a"
		if kind == pureArrayReplace {
			key = "0"
		}
		if !ok || result.elements.items[key].elements != nil {
			t.Fatal("combined width bypassed nested depth bound")
		}
	}
}
