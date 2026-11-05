package flow

import (
	"strings"
	"testing"
)

func TestMappedMultipleArrays(t *testing.T) {
	for _, source := range []string{
		`$xs=array_map(fn($x,$flags)=>htmlspecialchars($x,$flags),[$_GET['x']],[ENT_QUOTES]);probe($xs[0]);`,
		`$xs=array_map(fn($x,$y)=>$y,['z'=>$_GET['x'],'a'=>$_GET['y']],['q'=>'clean']);probe($xs[0]);`,
		`$xs=array_map(fn($x,$y)=>$y,['z'=>$_GET['x'],'a'=>$_GET['y']],['q'=>'clean']);probe($xs[1]);`,
		`$xs=array_map(fn($x,$y)=>$x,['second'=>'clean','first'=>$_GET['x']],['x','y']);probe($xs[0]);`,
		`$xs=array_map(fn($x,$y)=>$y,[$_GET['x']],[]);probe($xs[0]);`,
		`$xs=array_map(fn($x,$y)=>$y,[],['safe']);probe($xs[0]);`,
		`$xs=array_map(fn($x,$y)=>$x,['a'=>'dirty','b'=>'safe','a'=>$_GET['x']],['x','y']);probe($xs[1]);`,
	} {
		e, file := summarized(t, source)
		x := argument(callsNamed(file, "probe")[0])
		if !e.Value(x).Complete || e.Tainted(x, HTML) {
			t.Fatalf("%s: %+v", source, e.Value(x))
		}
	}
}

func TestMappedNestedElements(t *testing.T) {
	for _, source := range []string{
		`$xs=array_map(fn($x)=>['clean'=>'safe','dirty'=>$x],[$_GET['x']]);$row=$xs[0];probe($row['clean']);`,
		`final class R{function row($x){return ['clean'=>'safe','dirty'=>$x];}}$worker=new R;$row=$worker->row($_GET['x']);probe($row['clean']);`,
		`function wrap($x){return ['clean'=>'safe','dirty'=>$x];}$xs=array_map('wrap',[$_GET['x']]);probe($xs[0]['clean']);`,
		`$xs=array_map(fn($x)=>['clean'=>'safe','dirty'=>$x],[$_GET['x']]);probe($xs[0]['clean']);`,
		`$xs=array_map(fn($x)=>$x,[['clean'=>'safe','dirty'=>$_GET['x']]]);probe($xs[0]['clean']);`,
	} {
		e, file := summarized(t, source)
		x := argument(callsNamed(file, "probe")[0])
		if !e.Value(x).Complete || e.Tainted(x, HTML) {
			t.Fatalf("%s: %+v", source, e.Value(x))
		}
	}
}

func TestMappedNestedUnknowns(t *testing.T) {
	for _, source := range []string{
		`function rows($flag){if($flag){return ['x'=>'safe','y'=>$_GET['x']];}return ['y'=>$_GET['x'],'x'=>'safe'];}$a=rows($flag);$xs=array_map(fn($x,$y)=>$x,$a,['a','b']);probe($xs[0]);`,
		`$xs=array_map(fn($x)=>['clean'=>'safe','dirty'=>$x],[$_GET['x']]);probe($xs[0]['dirty']);`,
		`$xs=array_map(fn($x)=>['clean'=>'safe','dirty'=>$x],[$_GET['x']]);$xs[0]['clean']=$_POST['x'];probe($xs[0]['clean']);`,
		`$xs=array_map(fn($x)=>['clean'=>'safe','dirty'=>$x],[$_GET['x']]);$alias=&$xs[0];opaque($alias);probe($xs[0]['clean']);`,
		`function caller($flag){$a=$flag?['x'=>'safe','y'=>$_GET['x']]:['y'=>$_GET['x'],'x'=>'safe'];$xs=array_map(fn($x,$y)=>$x,$a,['a','b']);probe($xs[0]);}`,
		`$a=['safe'];$a[$_GET['key']]=$_GET['value'];$xs=array_map(fn($x,$y)=>$x,$a,['a','b']);probe($xs[0]);`,
		`$xs=array_map(callback:fn($x,$y)=>$x,array:[$_GET['x']],arrays:['safe']);probe($xs[0]);`,
		`$xs=array_map(fn($x,$y)=>$x,[$_GET['x']],['a'],[],[],[],[],[],[],[]);probe($xs[0]);`,
	} {
		e, file := summarized(t, source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if v.Complete && !e.Tainted(x, HTML) {
			t.Fatalf("invented safe nested value: %s %+v", source, v)
		}
	}
}

func TestMappedRejectsUnorderedEntries(t *testing.T) {
	e, file := summarized(t, `array_map(fn($x,$y)=>$x,['a','b'],['c','d']);`)
	at := callsNamed(file, "array_map")[0]
	r := e.scope(nil)
	values := append([]Value(nil), r.calls[r.callIndex[at]].Arguments...)
	values[1].elements = values[1].elements.with("0", values[1].elements.items["0"])
	if _, known := e.mappedCallback(at, at.Args, values, &frame{}, r); known {
		t.Fatal("unordered entries established positional mapping")
	}
}

func TestNestedArrayGrowthStaysBounded(t *testing.T) {
	for _, source := range []string{
		`$a=['dirty'=>$_GET['x']];for($i=0;$i<1000;$i++){$a=[$a];}probe($a);`,
		`$a=['dirty'=>$_GET['x']];` + strings.Repeat(`$a=[$a];`, 64) + `probe($a);`,
	} {
		e, file := summarized(t, source)
		at := argument(callsNamed(file, "probe")[0])
		v := e.Value(at)
		if v.Complete && !tainted(v) {
			t.Fatal("depth fallback erased aggregate taint")
		}
		if v.elements != nil && v.elements.depth > MaxDepth {
			t.Fatal("unbounded local shape")
		}
	}
}

func TestNestedBranchingShapesStayBounded(t *testing.T) {
	var source strings.Builder
	source.WriteString(`$a=[$_GET['x']];$b=['safe'];`)
	for range 8 {
		source.WriteString(`$a=[` + strings.Repeat(`$a,`, 64) + `];$b=[` + strings.Repeat(`$b,`, 64) + `];`)
	}
	source.WriteString(`$joined=$flag?$a:$b;probe($joined);`)
	e, file := summarized(t, source.String())
	value := e.Value(argument(callsNamed(file, "probe")[0]))
	if value.Complete && !tainted(value) {
		t.Fatal("branching fallback lost aggregate provenance")
	}
	if value.elements != nil && value.elements.depth > MaxDepth {
		t.Fatal("branching shape exceeded structural budget")
	}
}

func TestNestedReturnShapesRequireClosedDispatch(t *testing.T) {
	for _, source := range []string{
		`class R{function row($x){return ['clean'=>'safe','dirty'=>$x];}}function caller(R $worker){$row=$worker->row($_GET['x']);probe($row['clean']);}`,
		`if($flag){function row($x){return ['clean'=>'safe','dirty'=>$x];}}$row=row($_GET['x']);probe($row['clean']);`,
	} {
		e, file := summarized(t, source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if v.Complete && !tainted(v) {
			t.Fatalf("unproven target supplied clean field: %s", source)
		}
	}
}
