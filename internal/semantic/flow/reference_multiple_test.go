package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestDistinctReferenceOutputs(t *testing.T) {
	for _, tc := range []struct {
		body  string
		known bool
	}{
		{`$a=$_GET['x'];$b='safe';fill($a,$b);`, true},
		{`$a=$_GET['x'];$b=&$a;fill($a,$b);`, false},
		{`$a=$_GET['x'];fill($a,$a);`, false},
		{`$a='x';$b='y';fill(...[$a,$b]);`, false},
		{`$a='x';$b='y';fill($a,unknown:$b);`, false},
		{`global $a;$b='safe';fill($a,$b);`, false},
		{`static $a;$b='safe';fill($a,$b);`, false},
		{`$a=[];$b='safe';fill($a[0],$b);`, false},
		{`$a=$_GET['x'];$cb=function()use(&$a){};store($cb);$b='safe';fill($a,$b);`, false},
	} {
		e, file := summarized(t, `function fill(&$a,&$b){$a='safe';$b='safe';}function caller(){`+tc.body+`}`)
		call := callsNamed(file, "fill")[0]
		for i, node := range call.Args.Args {
			value, known := e.ReferenceResult(call, node.(*syntax.Arg))
			if known != tc.known || known && tainted(value) {
				t.Fatalf("%s output %d: known=%v value=%+v", tc.body, i, known, value)
			}
		}
	}
}

func TestReferenceProofJoinsLoopAliases(t *testing.T) {
	e, file := summarized(t, `function fill(&$a,&$b){$a='safe';$b='safe';}function caller($flag){$a='x';$b='y';while($flag){fill($a,$b);$b=&$a;}}`)
	at := callsNamed(file, "fill")[0]
	if _, known := e.ReferenceResult(at, at.Args.Args[0].(*syntax.Arg)); known {
		t.Fatal("one distinct iteration authorized aliased iterations")
	}
}

func TestDistinctReferenceOutputsKeepInputProvenance(t *testing.T) {
	e, file := summarized(t, `function fill(&$a,&$b,$input){$a='safe';$b=$input;}function caller(){$a=$_GET['x'];$b='safe';fill(b:$b,input:$_POST['x'],a:$a);}`)
	at := callsNamed(file, "fill")[0]
	for _, i := range []int{0, 2} {
		value, known := e.ReferenceResult(at, at.Args.Args[i].(*syntax.Arg))
		if !known || tainted(value) != (i == 0) {
			t.Fatalf("named output %d: known=%v value=%+v", i, known, value)
		}
	}
}
