package flow

import "testing"

func TestCallbackCompletionProbes(t *testing.T) {
	for _, source := range []string{
		`$xs=array_map(...['htmlspecialchars',[$_GET['x']]]);probe($xs[0]);`,
		`$xs=array_map(null,['safe'],[$_GET['x']]);probe($xs[0][0]);`,
		`function factory(){return fn($x)=>['clean'=>'safe','dirty'=>$x];}$cb=factory();$row=$cb($_GET['x']);probe($row['clean']);`,
	} {
		e, file := summarized(t, source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if !v.Complete || e.Tainted(x, HTML) {
			t.Fatalf("%s: %+v", source, v)
		}
	}
}

func TestFiniteFilterCallbackSink(t *testing.T) {
	e, file := summarized(t, `function wrap($x){array_filter([$x],fn($v)=>header($v));}wrap($_GET['x']);`)
	if sinks := e.Sinks(callsNamed(file, "wrap")[0]); len(sinks) != 1 || !tainted(sinks[0].Argument) {
		t.Fatalf("sinks=%+v", sinks)
	}
}

func TestResultGuardedWrapperClose(t *testing.T) {
	e, file := summarized(t, `function shut($h){return fclose($h);}$h=tmpfile();if(shut($h)){probe($h);}`)
	at := callsNamed(file, "probe")[0]
	if _, closed := e.StateBefore(at, argument(at), "fclose"); !closed {
		t.Fatal("returned close result lost its success proof")
	}
}
