package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestConstantTruth(t *testing.T) {
	for _, tc := range []struct {
		expr         string
		truth, known bool
	}{
		{`true`, true, true},
		{`false`, false, true},
		{`null`, false, true},
		{`0`, false, true},
		{`-3`, true, true},
		{`0.0`, false, true},
		{`-1.5`, true, true},
		{`""`, false, true},
		{`"0"`, false, true},
		{`"\x30"`, false, true},
		{`"\u0030"`, false, false},
		{`"\a"`, false, false},
		{`"\b"`, false, false},
		{`"\f"`, true, true},
		{`"\u{30}"`, false, false},
		{`'\u0030'`, true, true},
		{`"\a" === "\x07"`, false, false},
		{`"\\u0030"`, true, true},
		{`'00'`, true, true},
		{`'text'`, true, true},
		{`!0`, true, true},
		{`!true`, false, true},
		{`+3`, true, true},
		{`+1.5`, true, true},
		{`+'1'`, false, false},
		{`2+3`, true, true},
		{`2-2`, false, true},
		{`2*0`, false, true},
		{`3*-2`, true, true},
		{`9223372036854775807+1`, false, false},
		{`-9223372036854775807-2`, false, false},
		{`9223372036854775807*2`, false, false},
		{`1e999`, false, false},
		{`true&&false`, false, true},
		{`false||true`, true, true},
		{`true xor true`, false, true},
		{`false and false`, false, true},
		{`true or false`, true, true},
		{`2===2`, true, true},
		{`2!==2`, false, true},
		{`2===2.0`, false, true},
		{`"01"==="1"`, false, true},
		{`"01"=="1"`, false, false},
		{`null==null`, true, true},
		{`2==2`, true, true},
		{`2!=3`, true, true},
		{`1==true`, false, false},
		{`2<3`, true, true},
		{`2>3`, false, true},
		{`2<=2`, true, true},
		{`2>=2`, true, true},
		{`1.5<2.0`, true, true},
		{`3.0>2.0`, true, true},
		{`2.0<=2.0`, true, true},
		{`"a"."b"`, true, true},
		{`""."0"`, false, true},
		{`'1' . 2`, false, false},
		{`1/0`, false, false},
		{`2/1`, false, false},
		{`3%2`, false, false},
		{`true < false`, false, false},
		{`SYMBOL`, false, false},
		{`f()`, false, false},
		{`$x`, false, false},
		{`$x=1`, false, false},
		{`false&&f()`, false, false},
		{`(new C)`, false, false},
		{`~1`, false, false},
	} {
		t.Run(tc.expr, func(t *testing.T) {
			_, f := environment(t, `probe(`+tc.expr+`);`, nil)
			truth, known := constantTruth(argument(callsNamed(f, "probe")[0]))
			if known != tc.known || known && truth != tc.truth {
				t.Fatalf("truth=%v known=%v", truth, known)
			}
		})
	}
	if _, known := constantTruth(nil); known {
		t.Fatal("nil must be unknown")
	}
	_, f := environment(t, `probe(`+strings.Repeat("! (", 70)+`true`+strings.Repeat(")", 70)+`);`, nil)
	if _, known := constantTruth(argument(callsNamed(f, "probe")[0])); known {
		t.Fatal("deep evaluator must stop")
	}
}

func TestConstantCFGSuccessors(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		lo, hi    int64
		known     bool
	}{
		{"if false", `$n=-2;if(false){$n=5;}probe($n);`, -2, -2, true},
		{"if true", `$n=5;if(!0){$n=-2;}else{$n=7;}probe($n);`, -2, -2, true},
		{"elseif", `$n=5;if(false){$n=7;}elseif(2<3){$n=-2;}else{$n=8;}probe($n);`, -2, -2, true},
		{"while false", `$n=-2;while(0){$n=5;}probe($n);`, -2, -2, true},
		{"do false", `$n=5;do{$n=-2;}while(false);probe($n);`, -2, -2, true},
		{"for last false", `$n=-2;for(;f(),false;){$n=5;}probe($n);`, -2, -2, true},
		{"ternary true", `$n=true?-2:unknown();probe($n);`, -2, -2, true},
		{"ternary false", `$n=false?unknown():-2;probe($n);`, -2, -2, true},
		{"short true", `$n=-2?:unknown();probe($n);`, -2, -2, true},
		{"short false", `$n=0?:-2;probe($n);`, -2, -2, true},
		{"dead mutation", `$n=-2;true?1:($n=5);probe($n);`, -2, -2, true},
		{"unknown", `$n=SYMBOL?-2:5;probe($n);`, -2, 5, true},
		{"PHP escape unknown", `$n="\u0030"?-2:5;probe($n);`, -2, 5, true},
		{"goto enters", `$n=5;goto inside;if(false){inside:$n=-2;}probe($n);`, -2, -2, true},
		{"switch subject unknown", `switch($x){case 1:$n=-2;break;default:$n=5;}probe($n);`, -2, 5, true},
		{"switch false", `switch(false){case true:$n=5;break;case false:$n=-2;break;}probe($n);`, -2, -2, true},
		{"dead exit", `$n=-2;if(false){exit;}probe($n);`, -2, -2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, `function probe($v){}`+tc.src, nil)
			lo, hi, known := e.IntRange(argument(callsNamed(f, "probe")[0]))
			if known != tc.known || known && (lo != tc.lo || hi != tc.hi) {
				t.Fatalf("range=[%d,%d],%v", lo, hi, known)
			}
		})
	}
}

func TestSyntacticGeneratorProducers(t *testing.T) {
	for _, src := range []string{
		`$g=function()use(&$u){$u=parse_url($x);if(false){yield 1;}};$g();probe($u);`,
		`$g=function()use(&$u){$u=parse_url($x);true?1:yield 1;};$g();probe($u);`,
	} {
		e, f := summarized(t, src)
		origins, _ := e.Reaching(argument(callsNamed(f, "probe")[0]))
		if len(origins) > 0 {
			t.Fatalf("unadvanced generator exported producers: %v", origins)
		}
	}
	e, f := environment(t, `function outer(){function nested(){if(false){yield 1;}}return 1;}function gen(){if(false){yield from [];}}`, nil)
	outer := f.Stmts[0].(*syntax.Function)
	gen := f.Stmts[1].(*syntax.Function)
	if e.generator(outer) || !e.generator(gen) {
		t.Fatal("syntactic generator scope mismatch")
	}
	if e.scope(gen).generator != true {
		t.Fatal("dead-yield scope lost generator identity")
	}
}

func BenchmarkConstantBranches(b *testing.B) {
	var src strings.Builder
	src.WriteString(`function f(){$n=-2;`)
	for range 200 {
		src.WriteString(`if(false){$n=5;}else{$n=-2;} `)
	}
	src.WriteString(`probe($n);}`)
	e, f := environment(b, src.String(), nil)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, _, known := New(f, e.types, nil).IntRange(x); !known {
			b.Fatal("range lost")
		}
	}
}

func constantTree(depth int) string {
	if depth == 0 {
		return "1"
	}
	child := constantTree(depth - 1)
	return "(" + child + "+" + child + ")"
}

func TestConstantTotalBudget(t *testing.T) {
	_, f := environment(t, `probe(`+constantTree(8)+`);`, nil)
	if _, known := constantTruth(argument(callsNamed(f, "probe")[0])); known {
		t.Fatal("wide expression must stop at the shared node budget")
	}
}

func BenchmarkConstantEvaluatorBudget(b *testing.B) {
	_, f := environment(b, `probe(`+constantTree(10)+`);`, nil)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, known := constantTruth(x); known {
			b.Fatal("budget exceeded without abstaining")
		}
	}
}

func TestChosenTernaryValue(t *testing.T) {
	for _, src := range []string{`$v=true?parse_url($x):false;probe($v);`, `$v=false?false:parse_url($x);probe($v);`} {
		e, f := environment(t, src, nil)
		value, known := e.Resolve(argument(callsNamed(f, "probe")[0]))
		call, ok := value.(*syntax.FuncCall)
		if !known || !ok || call.Name.(*syntax.Name).Value != "parse_url" {
			t.Fatalf("selected producer=%T,%v", value, known)
		}
	}
	for _, src := range []string{`$v=true?['key'=>'safe','other'=>$_GET['x']]:['key'=>$_GET['x']];probe($v['key']);`, `$v=false?['key'=>$_GET['x']]:['key'=>'safe','other'=>$_GET['x']];probe($v['key']);`} {
		e, f := environment(t, src, nil)
		value := e.Value(argument(callsNamed(f, "probe")[0]))
		if !value.Complete || tainted(value) {
			t.Fatal("selected array element lost its own metadata")
		}
	}
}

func TestConstantLiteralByteBudget(t *testing.T) {
	for _, raw := range []string{`"` + strings.Repeat("x", 4096) + `"`, strings.Repeat("1", 4097), `1.` + strings.Repeat("1", 4096)} {
		_, f := environment(t, `probe(`+raw+`);`, nil)
		if _, known := constantTruth(argument(callsNamed(f, "probe")[0])); known {
			t.Fatal("oversized raw literal must remain unknown")
		}
	}
}
