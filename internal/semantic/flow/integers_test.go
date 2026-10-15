package flow

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"unsafe"

	"custos/internal/php/syntax"
)

func TestIntegerRanges(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		lo, hi    int64
		known     bool
	}{
		{"literal", `probe(-3);`, -3, -3, true},
		{"union", `if($c){$n=-2;}else{$n=-7;}probe($n);`, -7, -2, true},
		{"crossing", `if($c){$n=-2;}else{$n=7;}probe($n);`, -2, 7, true},
		{"unknown path", `if($c){$n=-2;}probe($n);`, 0, 0, false},
		{"copy", `if($c){$n=-2;}else{$n=-7;}$m=$n;probe($m);`, -7, -2, true},
		{"addition", `$n=-3;$n+=1;probe($n);`, -2, -2, true},
		{"subtraction", `$n=3;$n-=8;probe($n);`, -5, -5, true},
		{"unary", `$n=3;probe(-$n);`, -3, -3, true},
		{"poststep", `$n=3;probe($n--);`, 3, 3, true},
		{"prestep", `$n=3;probe(--$n);`, 2, 2, true},
		{"multiply", `$n=3;probe($n*2);`, 0, 0, false},
		{"overflow", `$n=9223372036854775807;$n++;probe($n);`, 0, 0, false},
		{"underflow", `$n=-9223372036854775807;$n-=2;probe($n);`, 0, 0, false},
		{"reference", `$n=-2;unknown($n);probe($n);`, 0, 0, false},
		{"do assigns", `$n=5;do{$n=-1;}while($c);probe($n);`, -1, -1, true},
		{"bounded loop", `for($n=-4;$n<0;$n++){probe($n);}`, -4, -1, true},
		{"bounded decrement", `for($n=4;$n>0;$n--){probe($n);}`, 1, 4, true},
		{"unbounded overflow", `$n=-1;while($c){probe($n);$n--;}`, 0, 0, false},
		{"narrow", `if($c){$n=-7;}else{$n=3;}if($n<0){probe($n);}`, -7, -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, `function probe($v) {}`+tc.src, nil)
			lo, hi, known := e.IntRange(argument(callsNamed(f, "probe")[0]))
			if known != tc.known || known && (lo != tc.lo || hi != tc.hi) {
				t.Fatalf("range=[%d,%d],%v want=[%d,%d],%v", lo, hi, known, tc.lo, tc.hi, tc.known)
			}
		})
	}
	t.Logf("Value size=%d", unsafe.Sizeof(Value{}))
	if _, _, known := (&Env{}).IntRange(nil); known {
		t.Fatal("nil range")
	}
	if _, ok := addInt(math.MaxInt64, 1); ok {
		t.Fatal("overflow")
	}
	if _, ok := subInt(math.MinInt64, 1); ok {
		t.Fatal("underflow")
	}
}

func BenchmarkIntegerRanges(b *testing.B) {
	for _, src := range []string{`$n=-2;probe($n);`, `if($c){$n=-2;}else{$n=-8;}probe($n);`, `for($n=-200;$n<0;$n++){probe($n);}`} {
		b.Run(fmt.Sprint(len(src)), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				e, f := environment(b, `function probe($v) {}`+src, nil)
				e.IntRange(argument(callsNamed(f, "probe")[0]))
			}
		})
	}
}

func TestIntegerEntryPrefix(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int64
		known     bool
	}{
		{"do first", `$n=-1;do{probe($n);$n=5;}while($c);`, -1, true},
		{"for first", `for($n=-4;$n<0;$n++){probe($n);}`, -4, true},
		{"zero turns", `for($n=4;$n<0;$n++){probe($n);}`, 0, false},
		{"while false", `$n=-1;while(false){probe($n);}`, 0, false},
		{"continue", `$n=-1;do{continue;probe($n);}while(false);`, 0, false},
		{"goto", `$n=-1;goto done;do{probe($n);}while(false);done:;`, 0, false},
		{"nested", `$n=-1;do{do{probe($n);$n=5;}while(false);}while(false);`, -1, true},
		{"unknown branch", `$n=-1;if($c){$n=5;}do{probe($n);}while(false);`, 0, false},
		{"prior call", `$n=-1;other();do{probe($n);}while(false);`, 0, false},
		{"prior reference", `$n=-1;$m=&$n;do{probe($n);}while(false);`, 0, false},
		{"overflow", `$n=9223372036854775807;$n++;probe($n);`, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, `function probe($v){} function f(){`+tc.src+`}`, nil)
			v, ok := e.IntEntry(argument(callsNamed(f, "probe")[0]))
			if ok != tc.known || ok && v != tc.want {
				t.Fatalf("entry=%d,%v", v, ok)
			}
		})
	}
}

// BenchmarkValueCopy deliberately references no range fields, allowing the
// same harness to measure the preceding stage's Value layout and copy cost.
func BenchmarkValueCopy(b *testing.B) {
	src := make([]Value, 256)
	dst := make([]Value, len(src))
	for i := range src {
		src[i].Identity = uint32(i)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(src)) * int64(unsafe.Sizeof(Value{})))
	for b.Loop() {
		copy(dst, src)
		copy(src, dst)
	}
	if dst[255].Identity != 255 {
		b.Fatal("copy failed")
	}
}

func TestIntegerDerivedValues(t *testing.T) {
	for _, src := range []string{
		`$n=-2;$s="$n";probe($s);`,
		`$n=-2;probe(isset($n));`,
		`$n=-2;probe(empty($n));`,
		`$n=-2;probe($n[0]);`,
		`$n=-2;probe((string)$n);`,
		`$n=-2;probe(strlen($n));`,
		`$n=-2;probe($n instanceof C);`,
		`$n=-2;probe([$n]);`,
		`function f(&$n,&$m){$n=-2;$m=5;do{probe($n);$n=6;}while($c);}`,
	} {
		e, f := environment(t, `function probe($v){}`+src, nil)
		x := argument(callsNamed(f, "probe")[0])
		if lo, hi, known := e.IntRange(x); known {
			t.Errorf("%s: stale range [%d,%d]", src, lo, hi)
		}
	}
	e, f := environment(t, `const LIMIT=-6;class C{const LIMIT=-7;}if($c){$n=LIMIT;}else{$n=C::LIMIT;}probe($n);`, nil)
	if lo, hi, ok := e.IntRange(argument(callsNamed(f, "probe")[0])); !ok || lo != -7 || hi != -6 {
		t.Fatalf("constant range=[%d,%d],%v", lo, hi, ok)
	}
}

func TestIntegerEntryArguments(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int64
	}{
		{`$n=-2;do{probe($n=5,$n);$n=6;}while($c);`, 5},
		{`$n=-2;do{probe($n,$n=5);$n=6;}while($c);`, -2},
	} {
		e, f := environment(t, `function probe($a,$b){} function f(){`+tc.src+`}`, nil)
		call := callsNamed(f, "probe")[0]
		at := call.Args.Args[1].(*syntax.Arg).Value
		if _, ok := at.(*syntax.Assign); ok {
			at = call.Args.Args[0].(*syntax.Arg).Value
		}
		value, ok := e.IntEntry(at)
		if !ok || value != tc.want {
			t.Fatalf("%s entry=%d,%v", tc.src, value, ok)
		}
	}
}

func TestIntegerEntryBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$n=-2;if(true){probe($n);}`, true},
		{`$n=-2;if($n>-3){probe($n);}`, true},
		{`$n=-2;if($n<=-1){probe($n);}`, true},
		{`$n=-2;if($n>=-3){probe($n);}`, true},
		{`$n=-2;$s=$_GET;probe($n);`, false},
		{`$n=-2;$s=4/2;probe($n);`, false},
		{`$n=-2;$s=function(){};probe($n);`, false},
		{`$n=-2;while($n<0){$n--;}probe($n);`, false},
		{`$n=-2;do{echo probe($n);$n=5;}while($c);`, true},
		{`$n=-2;do{echo "prefix",probe($n);$n=5;}while($c);`, false},
		{`$n=-2;do{return probe($n);}while($c);`, true},
		{`$n=-2;probe($_GET, $n);`, false},
		{`$n=-2;probe(${'n'}, $n);`, false},
	} {
		e, f := environment(t, `function probe($a,$b=null){} function f(){`+tc.src+`}`, nil)
		args := callsNamed(f, "probe")[0].Args.Args
		x := args[len(args)-1].(*syntax.Arg).Value
		n, known := e.IntEntry(x)
		if known != tc.want || known && n != -2 {
			t.Errorf("%s: entry=%d,%v", tc.src, n, known)
		}
	}
	e, f := environment(t, `function f(&$n){$n=-2;probe($n);}`, nil)
	if _, known := e.IntEntry(argument(callsNamed(f, "probe")[0])); known {
		t.Fatal("reference parameter entry must remain unknown")
	}
	e, f = environment(t, `$n=-2;if($n>9223372036854775807){probe($n);}`, nil)
	e.IntRange(argument(callsNamed(f, "probe")[0]))
	_, f = environment(t, `probe(`+strings.Repeat("- (", 70)+`2`+strings.Repeat(")", 70)+`);`, nil)
	if _, known := literalInteger(argument(callsNamed(f, "probe")[0])); known {
		t.Fatal("deep literal sign chain must stop at the depth budget")
	}
}

func TestImmutableIntegerRangeReuse(t *testing.T) {
	e, f := environment(t, `if($pick){$n=-2;}else{$n=-7;}if($n<0){probe(+$n);}`, nil)
	x := argument(callsNamed(f, "probe")[0]).(*syntax.Unary)
	held := e.Value(x.Expr)
	got := e.Value(x)
	if held.interval == nil || got.interval != held.interval {
		t.Fatal("unary plus must retain the immutable interval")
	}
	before := *held.interval
	state := frame{vars: map[string]Value{"n": held}}
	var condition *syntax.Binary
	for parent := x.Parent(); parent != nil; parent = parent.Parent() {
		if guard, ok := parent.(*syntax.If); ok {
			condition = guard.Cond.(*syntax.Binary)
			break
		}
	}
	e.refineInts(condition, &state, true)
	if state.vars["n"].interval != held.interval || *held.interval != before {
		t.Fatal("unchanged refinement must retain immutable bounds")
	}
	literal := Value{Complete: true, Expr: condition.Right}
	if joinInts(literal, literal) != nil {
		t.Fatal("joining identical literal expressions must not allocate a range")
	}
}
