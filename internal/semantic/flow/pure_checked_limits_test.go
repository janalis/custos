package flow

import (
	"math"
	"slices"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestPureCheckedRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($a,$b){return $a/$b;}probe(f(7,2));`,
		`function f($a,$b){return $a/$b;}probe(f(1,0));`,
		`function f($a,$b){return $a%$b;}probe(f(1,0));`,
		`function f($a,$b){return $a/$b;}probe(f(-9223372036854775807-1,-1));`,
		`function f($a,$b){return $a/$b;}probe(f('6',3));`,
		`function f($a,$b){return $a%$b;}probe(f(1.5,1));`,
		`function f($a,$b){return $a&$b;}probe(f('a','b'));`,
		`function f($a){return ~$a;}probe(f('a'));`,
		`function f($a,$b){return $a|$b;}probe(f(true,1));`,
		`function f($a,$b){return $a^$b;}probe(f($_GET['x'],1));`,
		`function f($a,$b){return $a<<$b;}probe(f(1,-1));`,
		`function f($a,$b){return $a>>$b;}probe(f(1,64));`,
		`function f($a,$b){return $a<<$b;}probe(f(1,63));`,
		`function f($a,$b){return $a<<$b;}probe(f(-9223372036854775807-1,1));`,
		`function f($a,$b){return $a&$b['missing'];}probe(f(0,[]));`,
		`function f($a,$b){return $a|$b;}probe(f(0,new stdClass));`,
		`function f($x){return str_replace('','a',$x);}probe(f('x'));`,
		`function f($x){return str_replace(['a'],'b',$x);}probe(f('a'));`,
		`function f($x){return str_replace('a',['b'],$x);}probe(f('a'));`,
		`function f($x){return str_replace('a','b',$x);}probe(f(1));`,
		`function f($x){return str_replace('a','b',$x,$count);}probe(f('a'));`,
		`function f($x,$r){return str_replace('missing',$r,$x);}probe(f('same',$_GET['r']));`,
		`function f($x,$n){return str_repeat($x,$n);}probe(f('ab',-1));`,
		`function f($x,$n){return str_repeat($x,$n);}probe(f('ab','2'));`,
		`function f($x,$n){return str_repeat($x,$n);}probe(f('ab',9223372036854775807));`,
		`function f($x,$n){return str_repeat($x,$n);}probe(f(new stdClass,0));`,
		`function f($xs){return array_reverse($xs,1)[0];}probe(f([1]));`,
		`function f($xs){return array_reverse($xs,true,1)[0];}probe(f([1]));`,
		`function f($xs){return count(array_reverse($xs));}probe(f([new stdClass]));`,
		`function f($x){switch($x){case 1:return 'one';default:return 'other';}}probe(f('1'));`,
		`function f($x){switch($x){case 1:return 'one';default:return 'other';}}probe(f(true));`,
		`function f($x){switch($x){case 1:return 'one';default:return 'other';}}probe(f(1.0));`,
		`function f($x){switch($x){case true:return 'bool';case 1:return 'one';default:return 'other';}}probe(f(1));`,
		`function f($x){switch($x){case '1':return 'string';default:return 'other';}}probe(f(1));`,
		`function f($x,$test){switch($x){case $test:return 'first';case 1:return 'one';default:return 'other';}}probe(f(1,$_GET['test']));`,
		`function f($x){switch($x){case 1:return 'one';}}probe(f(1));`,
		`function f($x){switch($x){case 1:case 2:return 'one';default:return 'other';}}probe(f(1));`,
		`function f($x){switch($x){case 1:break;default:return 'other';}return 'after';}probe(f(1));`,
		`function f($x){switch($x){case 1:$x=2;return 'one';default:return 'other';}}probe(f(1));`,
		`function f($x){switch($x){case 1:return 'one';default:return 'other';}return 'after';}probe(f(1));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s invented %+v", source, got)
		}
	}
}

func TestPureCheckedValueSources(t *testing.T) {
	e, _ := environment(t, ``, nil)
	number := scalarValue(constantScalar{k: 'i', i: 6})
	number.Sources = []Source{{Kind: "left", Parameter: 0}}
	number.Safe = URL | Path
	other := scalarValue(constantScalar{k: 'i', i: 3})
	other.Sources = []Source{{Kind: "right", Parameter: 1}}
	for _, op := range []syntax.TokenKind{syntax.TDiv, syntax.TMod, syntax.TAmpersand, syntax.TBar, syntax.TCaret, syntax.TSl, syntax.TSr, syntax.TTilde} {
		expr := &pureExpr{kind: pureInteger, op: op, left: &pureSlots[0], right: &pureSlots[1]}
		value, ok := e.evalPure(expr, []Value{number, other}, &scopeResult{}, MaxDepth)
		want := 2
		if op == syntax.TTilde {
			want = 1
		}
		if !ok || len(value.Sources) != want || value.Safe&(URL|Path) != 0 || value.Safe&HTML == 0 {
			t.Fatalf("integer sources/safety %v", op)
		}
	}
	text := scalarValue(constantScalar{k: 's', s: "aba"})
	text.Sources = number.Sources
	text.Safe = HTML
	replacement := scalarValue(constantScalar{k: 's', s: "x"})
	replacement.Sources = other.Sources
	replacement.Safe = HTML
	search := scalarValue(constantScalar{k: 's', s: "a"})
	expr := &pureExpr{kind: pureReplace, left: &pureSlots[0], right: &pureSlots[1], other: &pureSlots[2]}
	value, ok := e.evalPure(expr, []Value{search, replacement, text}, &scopeResult{}, MaxDepth)
	if !ok || len(value.Sources) != 2 || value.Safe != 0 {
		t.Fatal("replacement sources or escaping")
	}
	search = scalarValue(constantScalar{k: 's', s: "missing"})
	value, ok = e.evalPure(expr, []Value{search, replacement, text}, &scopeResult{}, MaxDepth)
	if !ok || !slices.Equal(value.Sources, text.Sources) || value.Safe != 0 {
		t.Fatal("unused replacement source or escaping")
	}
	expr = &pureExpr{kind: pureRepeat, left: &pureSlots[0], right: &pureSlots[1]}
	for _, count := range []int64{0, 1, 2} {
		value, ok = e.evalPure(expr, []Value{text, scalarValue(constantScalar{k: 'i', i: count})}, &scopeResult{}, MaxDepth)
		if !ok || value.Safe != 0 || (count == 0 && len(value.Sources) != 0) || (count != 0 && !slices.Equal(value.Sources, text.Sources)) {
			t.Fatal("repeat sources or escaping")
		}
	}
	// Both input vectors can be valid individually while their union is too large.
	for i := 0; i <= MaxSources; i++ {
		source := Source{Kind: "origin", Parameter: i}
		if i < MaxSources/2 {
			number.Sources = append(number.Sources, source)
		} else {
			other.Sources = append(other.Sources, source)
		}
	}
	expr = &pureExpr{kind: pureInteger, op: syntax.TAmpersand, left: &pureSlots[0], right: &pureSlots[1]}
	if _, ok := e.evalPure(expr, []Value{number, other}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("integer origin union truncated")
	}
	text.Sources, replacement.Sources = number.Sources, other.Sources
	expr = &pureExpr{kind: pureReplace, left: &pureSlots[0], right: &pureSlots[1], other: &pureSlots[2]}
	search = scalarValue(constantScalar{k: 's', s: "a"})
	if _, ok := e.evalPure(expr, []Value{search, replacement, text}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("replacement origin union truncated")
	}
}

func TestPureCheckedLimits(t *testing.T) {
	e, _ := environment(t, ``, nil)
	integer := &pureExpr{kind: pureInteger, op: syntax.TDiv, left: &pureSlots[0], right: &pureSlots[1]}
	if _, ok := e.evalPure(integer, []Value{scalarValue(constantScalar{k: 'i', i: 1})}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("missing right operand ignored")
	}
	number := scalarValue(constantScalar{k: 'i', i: 1})
	if _, ok := e.evalPure(&pureExpr{kind: pureInteger, op: syntax.TPow, left: &pureSlots[0], right: &pureSlots[0]}, []Value{number}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("unsupported integer opcode")
	}
	record := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{"0": number}, order: []string{"0"}}}
	reverse := &pureExpr{kind: pureReverse, left: &pureSlots[0], right: pureConstant(constantScalar{k: 'b'})}
	for _, r := range []scopeResult{{copies: MaxCopied}, {operations: MaxTransfers - 3}} {
		if _, ok := e.evalPure(reverse, []Value{record}, &r, MaxDepth); ok {
			t.Fatal("reverse budget bypassed")
		}
	}
	repeat := &pureExpr{kind: pureRepeat, left: &pureSlots[0], right: pureConstant(constantScalar{k: 'i', i: 2})}
	text := scalarValue(constantScalar{k: 's', s: "x"})
	for _, r := range []scopeResult{{copies: MaxCopied}, {operations: MaxTransfers - 3}} {
		if _, ok := e.evalPure(repeat, []Value{text}, &r, MaxDepth); ok {
			t.Fatal("repeat budget bypassed")
		}
	}
	large := scalarValue(constantScalar{k: 's', s: strings.Repeat("x", 4097)})
	if _, ok := e.evalPure(repeat, []Value{large}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("repeat input size bypassed")
	}
	switchExpr := &pureExpr{kind: pureSwitch, left: &pureSlots[0], other: pureConstant(constantScalar{k: 'i'}), args: []*pureExpr{pureConstant(constantScalar{k: 'i', i: 1}), pureConstant(constantScalar{k: 'i', i: 7})}}
	if _, ok := e.evalPure(switchExpr, []Value{number}, &scopeResult{operations: MaxTransfers - 2}, MaxDepth); ok {
		t.Fatal("switch work budget bypassed")
	}
	if _, ok := e.evalPure(switchExpr, nil, &scopeResult{}, MaxDepth); ok {
		t.Fatal("unknown switch subject ignored")
	}
	compiler := pureCompiler{env: e}
	if compiler.byteTransform(&syntax.FuncCall{}, "str_replace", nil) != nil || compiler.reverse(&syntax.FuncCall{}, nil) != nil {
		t.Fatal("missing native metadata accepted")
	}
}

func TestPureCheckedCompilerEdges(t *testing.T) {
	e, _ := environment(t, ``, nil)
	c := pureCompiler{env: e}
	integer := func(raw string) *syntax.Literal { return &syntax.Literal{LitKind: syntax.LitInt, Raw: raw} }
	if c.integer(syntax.TDiv, nil, integer("1")) != nil || c.integer(syntax.TDiv, integer("1"), nil) != nil {
		t.Fatal("missing integer operand accepted")
	}
	huge := pureSlots[0]
	huge.cost = pureLimit
	c.bind("huge", &huge)
	if c.integer(syntax.TTilde, &syntax.Variable{Name: "huge"}, nil) != nil {
		t.Fatal("unary expansion budget bypassed")
	}
	ret := func(expr syntax.Expr) []syntax.Stmt { return []syntax.Stmt{&syntax.Return{Expr: expr}} }
	for _, node := range []*syntax.Switch{
		{Cond: nil},
		{Cond: integer("1"), Cases: make([]*syntax.Case, pureLimit+1)},
		{Cond: integer("1"), Cases: []*syntax.Case{{Cond: integer("1"), Stmts: []syntax.Stmt{&syntax.Break{}}}}},
		{Cond: integer("1"), Cases: []*syntax.Case{{Stmts: ret(&syntax.Variable{Name: "missing"})}}},
		{Cond: integer("1"), Cases: []*syntax.Case{{Stmts: ret(nil)}, {Stmts: ret(nil)}}},
		{Cond: integer("1"), Cases: []*syntax.Case{{Cond: &syntax.Variable{Name: "missing"}, Stmts: ret(nil)}, {Stmts: ret(nil)}}},
	} {
		if c.integerSwitch(node) != nil {
			t.Fatal("invalid switch accepted")
		}
	}
}

func TestPureCheckedReplaceBudgets(t *testing.T) {
	e, _ := environment(t, ``, nil)
	text := scalarValue(constantScalar{k: 's', s: "a"})
	expr := &pureExpr{kind: pureReplace, left: &pureSlots[0], right: &pureSlots[1], other: &pureSlots[2]}
	for _, r := range []scopeResult{{copies: MaxCopied}, {operations: MaxTransfers - 4}} {
		if _, ok := e.evalPure(expr, []Value{text, text, text}, &r, MaxDepth); ok {
			t.Fatal("replace budget bypassed")
		}
	}
	large := scalarValue(constantScalar{k: 's', s: strings.Repeat("b", 4096)})
	if _, ok := e.evalPure(expr, []Value{text, large, scalarValue(constantScalar{k: 's', s: "aaaa"})}, &scopeResult{}, MaxDepth); ok {
		t.Fatal("replacement expansion bypassed")
	}
	for _, source := range []string{`function f($x){return str_replace('a','b');}probe(f('a'));`, `function f($x){return str_repeat('a');}probe(f('a'));`} {
		env, file := summarized(t, source)
		if _, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Fatal("missing byte argument accepted")
		}
	}
	min := scalarValue(constantScalar{k: 'i', i: math.MinInt64})
	if got, ok := e.evalPure(&pureExpr{kind: pureInteger, op: syntax.TMod, left: &pureSlots[0], right: pureConstant(constantScalar{k: 'i', i: -1})}, []Value{min}, &scopeResult{}, MaxDepth); !ok {
		t.Fatal("safe minimum remainder refused")
	} else if scalar, _ := knownScalar(got); scalar.i != 0 {
		t.Fatal("minimum remainder wrong")
	}
}

func TestPureCheckedVersionsAndDependencies(t *testing.T) {
	source := `function f($x){switch($x){case 1:return 'one';default:return 'other';}}`
	e, file := environmentAt(t, source, phpversion.PHP53)
	out := Extract(file, e.types.Index, phpversion.PHP53, nil, Options{})
	if out.Summaries["f"].pure == nil {
		t.Fatal("legacy integer switch refused")
	}
	original, defs := closedFileEnvironment(t, "definitions.php", `function divisor(){return 2;}`, nil)
	snapshot := NewSnapshot(Extract(defs, original.types.Index, phpversion.Max, nil, Options{}))
	consumer, file := closedFileEnvironment(t, "consumer.php", `function f($x){return $x/divisor();}`, snapshot)
	consumer.types.Index.Add(index.Extract(defs))
	extracted := Extract(file, consumer.types.Index, phpversion.Max, snapshot, Options{})
	if extracted.Summaries["f"].pure == nil || !slices.Contains(extracted.Consulted(), "divisor") {
		t.Fatal("integer dependency missing")
	}
	changed, updated := closedFileEnvironment(t, "definitions.php", `function divisor(){return 3;}`, nil)
	next := NewSnapshot(Extract(updated, changed.types.Index, phpversion.Max, nil, Options{}))
	if extracted.Current(snapshot, next) {
		t.Fatal("changed divisor did not invalidate program")
	}
}
