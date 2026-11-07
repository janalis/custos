package flow

import (
	"math"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

var scanCases = []struct {
	name, source string
	want         constantScalar
}{
	{"unique", `function f($xs){return count(array_unique($xs));}probe(f(['a','b','a']));`, constantScalar{k: 'i', i: 2}},
	{"unique_string_forms", `function f($xs){return array_keys(array_unique($xs))[1];}probe(f([3=>1,5=>'1',7=>'01']));`, constantScalar{k: 'i', i: 7}},
	{"unique_named", `function f($xs){return array_unique(array:$xs)['b'];}probe(f(['a'=>'x','b'=>'y','c'=>'x']));`, constantScalar{k: 's', s: "y"}},
	{"unique_empty", `function f($xs){return count(array_unique($xs));}probe(f([]));`, constantScalar{k: 'i'}},
	{"search", `function f($xs){return array_search('b',$xs,true);}probe(f(['x'=>'a','y'=>'b','z'=>'b']));`, constantScalar{k: 's', s: "y"}},
	{"search_int_key", `function f($xs){return array_search(strict:true,haystack:$xs,needle:1);}probe(f(['1',1]));`, constantScalar{k: 'i', i: 1}},
	{"search_miss", `function f($xs){return array_search('q',$xs,true);}probe(f(['a']));`, constantScalar{k: 'b'}},
	{"max_array", `function f($xs){return max($xs);}probe(f([3,9,2]));`, constantScalar{k: 'i', i: 9}},
	{"min_array", `function f($xs){return min($xs);}probe(f([5,-2,7]));`, constantScalar{k: 'i', i: -2}},
	{"min_args", `function f($a,$b){return min($a,$b,8);}probe(f(3,2));`, constantScalar{k: 'i', i: 2}},
	{"max_args", `function f($a){return max($a,-3);}probe(f(1));`, constantScalar{k: 'i', i: 1}},
	{"product", `function f($xs){return array_product($xs);}probe(f([2,3,4]));`, constantScalar{k: 'i', i: 24}},
	{"product_empty", `function f($xs){return array_product($xs);}probe(f([]));`, constantScalar{k: 'i', i: 1}},
	{"product_zero", `function f($xs){return array_product($xs);}probe(f([9223372036854775807,0]));`, constantScalar{k: 'i'}},
	{"product_negative", `function f($xs){return array_product($xs);}probe(f([-1,9223372036854775807]));`, constantScalar{k: 'i', i: -9223372036854775807}},
	{"strrev", `function f($s){return strrev($s);}probe(f('abc'));`, constantScalar{k: 's', s: "cba"}},
	{"substr_count", `function f($s){return substr_count($s,'aa');}probe(f('aaaa'));`, constantScalar{k: 'i', i: 2}},
}

func TestPureScans(t *testing.T) {
	for _, tc := range scanCases {
		t.Run(tc.name, func(t *testing.T) {
			e, file := summarized(t, tc.source)
			contract, _ := e.contract("f")
			if contract.pure == nil {
				t.Fatal("missing pure contract")
			}
			for _, env := range []*Env{e, NewLocal(file, e.types, e.snapshot, Options{}, nil)} {
				got, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0])))
				if !known || got != tc.want {
					t.Fatalf("got=%+v known=%v want=%+v", got, known, tc.want)
				}
			}
		})
	}
}

func TestPureScanRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($xs){return count(array_unique($xs,SORT_REGULAR));}probe(f([1]));`,
		`function f($xs){return count(array_unique($xs));}probe(f([1.5]));`,
		`function f($xs){return count(array_unique($xs));}probe(f([true]));`,
		`function f($xs){return count(array_unique($xs));}probe(f(opaque()));`,
		`function f($xs){return array_search('a',$xs);}probe(f(['a']));`,
		`function f($xs){return array_search('a',$xs,false);}probe(f(['a']));`,
		`function f($xs){return array_search('a',$xs,1);}probe(f(['a']));`,
		`function f($xs){return array_search('a',$xs,true);}probe(f([opaque()]));`,
		`function f($xs){return array_search('a',$xs,true);}probe(f(opaque()));`,
		`function f($xs,$v){return array_search('a',$xs,$v['missing']);}probe(f(['a'],[]));`,
		`function f($xs,$v){return array_search('a',$v['missing'],true);}probe(f(['a'],[]));`,
		`function f($xs){return max($xs);}probe(f([]));`,
		`function f($xs){return max($xs);}probe(f([1,'2']));`,
		`function f($xs){return max($xs);}probe(f([1,2.5]));`,
		`function f($xs){return max($xs);}probe(f(opaque()));`,
		`function f($a){return max($a,'2');}probe(f(1));`,
		`function f($a,$v){return max($a,$v['missing']);}probe(f(1,[]));`,
		`function f($a){return max(...$a);}probe(f([1,2]));`,
		`function f($a){return max(value:$a);}probe(f([1,2]));`,
		`function f($a){return max(opaque());}probe(f([1,2]));`,
		`function f($a){return max();}probe(f([1,2]));`,
		`function f($a){return max($a,opaque());}probe(f(1));`,
		`function f($xs){return array_product($xs);}probe(f([9223372036854775807,2]));`,
		`function f($xs){return array_product($xs);}probe(f([-1,-9223372036854775807-1]));`,
		`function f($xs){return array_product($xs);}probe(f([-9223372036854775807-1,-1]));`,
		`function f($xs){return array_product($xs);}probe(f(['2']));`,
		`function f($xs){return array_product($xs);}probe(f([new stdClass]));`,
		`function f($s){return substr_count($s,'');}probe(f('aaaa'));`,
		`function f($s){return substr_count($s,'a',1);}probe(f('aaaa'));`,
		`function f($s){return substr_count($s,1);}probe(f('aaaa'));`,
		`function f($s,$v){return substr_count($s,$v['missing']);}probe(f('aaaa',[]));`,
		`function f($s){return strrev($s);}probe(f(1));`,
		`function f($s){return substr_count($s);}probe(f('a'));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s invented %+v", source, got)
		}
	}
}

func TestPureScanRuntimeBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	integer := func(i int64) Value { return scalarValue(constantScalar{k: 'i', i: i}) }
	text := func(s string) Value { return scalarValue(constantScalar{k: 's', s: s}) }
	deep := constructionRecord(text("x"))
	deep.elements.depth = MaxDepth - 1
	unique := &pureExpr{kind: pureUnique, left: &pureSlots[0]}
	product := &pureExpr{kind: pureProduct, left: &pureSlots[0]}
	search := &pureExpr{kind: pureSearch, left: &pureSlots[0], right: &pureSlots[1], other: &pureSlots[2]}
	count := &pureExpr{kind: pureTextScan, literal: &constantScalar{k: 's', s: "substr_count"}, left: &pureSlots[0], right: &pureSlots[1]}
	maxArray := &pureExpr{kind: pureExtreme, literal: &constantScalar{k: 's', s: "max"}, left: &pureSlots[0]}
	yes := scalarValue(constantScalar{k: 'b', b: true})
	a, b := integer(2), integer(3)
	ta, tb := text("aa"), text("a")
	for i := 0; i <= MaxSources; i++ {
		source := Source{Kind: "input", Parameter: i}
		if i < MaxSources/2 {
			a.Sources, ta.Sources = append(a.Sources, source), append(ta.Sources, source)
		} else {
			b.Sources, tb.Sources = append(b.Sources, source), append(tb.Sources, source)
		}
	}
	many, manyText := integer(2), text("a")
	for i := 0; i <= MaxSources; i++ {
		many.Sources = append(many.Sources, Source{Kind: "input", Parameter: i})
	}
	manyText.Sources = many.Sources
	for i, tc := range []struct {
		expr  *pureExpr
		slots []Value
		r     scopeResult
	}{
		{unique, []Value{constructionRecord(text("a"))}, scopeResult{copies: MaxCopied}},
		{unique, []Value{constructionRecord(deep)}, scopeResult{}},
		{unique, []Value{constructionRecord(manyText)}, scopeResult{}},
		{product, []Value{constructionRecord(a, b)}, scopeResult{}},
		{product, []Value{constructionRecord(many)}, scopeResult{}},
		{search, []Value{text("a"), constructionRecord(text("a")), yes}, scopeResult{copies: MaxCopied}},
		{search, []Value{text("a"), constructionRecord(Value{Complete: true, plain: true, deep: true}), yes}, scopeResult{}},
		{maxArray, []Value{constructionRecord(integer(1))}, scopeResult{copies: MaxCopied}},
		{count, []Value{ta, tb}, scopeResult{}},
		{count, []Value{text(strings.Repeat("a", 64)), text("a")}, scopeResult{operations: MaxTransfers - 8}},
	} {
		if _, ok := e.evalPure(tc.expr, tc.slots, &tc.r, MaxDepth); ok {
			t.Errorf("case%d accepted unsafe input/budget", i)
		}
	}
	for _, tc := range []struct {
		a, b    int64
		product int64
		fits    bool
	}{
		{0, math.MinInt64, 0, true},
		{-1, 5, -5, true},
		{5, -1, -5, true},
		{-1, math.MinInt64, 0, false},
		{math.MinInt64, -1, 0, false},
		{math.MaxInt64, 2, 0, false},
		{-3, 4, -12, true},
	} {
		got, fits := multiplyChecked(tc.a, tc.b)
		if fits != tc.fits || fits && got != tc.product {
			t.Errorf("%d*%d = %d,%v", tc.a, tc.b, got, fits)
		}
	}
}

func TestPureScanCompilerBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	c := pureCompiler{env: e}
	if c.query(&syntax.FuncCall{}, "array_unique", nil) != nil || c.extreme(&syntax.FuncCall{}, "max") != nil {
		t.Fatal("invalid native metadata accepted")
	}
	for _, name := range []string{"strrev", "max"} {
		e, file := environment(t, `function f($xs){return `+name+`($xs);}`, nil)
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
			if name == "max" && c.extreme(call, name) != nil || name != "max" && c.query(call, name, params) != nil {
				t.Fatalf("%s: expanded bound bypassed", name)
			}
		}
	}
}

func BenchmarkPureScans(b *testing.B) {
	for _, tc := range scanCases {
		b.Run(tc.name, func(b *testing.B) {
			e, file := summarized(b, tc.source)
			call := argument(callsNamed(file, "probe")[0])
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				NewLocal(file, e.types, e.snapshot, Options{}, nil).Value(call)
			}
		})
	}
}
