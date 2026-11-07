package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReturnedClosureShapeProvenance(t *testing.T) {
	for _, tc := range []struct {
		source string
		taint  bool
	}{
		{`function factory(){return fn($x)=>['clean'=>'safe','dirty'=>$x];}$cb=factory();$row=$cb($_GET['x']);probe($row['clean']);`, false},
		{`function factory(){return fn($x)=>['clean'=>'safe','dirty'=>$x];}$cb=factory();$row=$cb($_GET['x']);probe($row['dirty']);`, true},
		{`function factory(string $x){$held=['clean'=>'safe','dirty'=>$x];return fn()=>$held;}$cb=factory($_GET['x']);$row=$cb();probe($row['clean']);`, false},
		{`function factory(string $x){$held=['clean'=>'safe','dirty'=>$x];return fn()=>$held;}$cb=factory($_GET['x']);$row=$cb();probe($row['dirty']);`, true},
		{`function factory(){return fn($x='safe')=>['value'=>$x];}$cb=factory();$row=$cb();probe($row['value']);`, false},
		{`function factory(){return fn($x)=>['clean'=>'safe','dirty'=>$x];}$a=factory();$b=factory();$one=$a($_GET['x']);$two=$b('safe');probe($two['dirty']);`, false},
	} {
		e, file := summarized(t, tc.source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if !v.Complete || e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, v)
		}
	}
}

func TestCrossFileReturnedClosureShapes(t *testing.T) {
	definitions := `<?php function factory(){return fn($x)=>['clean'=>'safe','dirty'=>$x];}`
	source := `<?php function caller(){$cb=factory();$row=$cb($_GET['x']);probe($row['clean']);}`
	ix := stubsIndexForResource(t, definitions, source)
	defs := syntax.Parse("definitions.php", []byte(definitions), syntax.Options{Version: phpversion.Max})
	snapshot := NewSnapshot(Extract(defs, ix, phpversion.Max, nil, Options{}))
	caller := syntax.Parse("caller.php", []byte(source), syntax.Options{Version: phpversion.Max})
	e, out := Summarize(caller, resourceTypes(caller, ix), snapshot, Options{})
	at := argument(callsNamed(caller, "probe")[0])
	if !e.Value(at).Complete || e.Tainted(at, HTML) {
		t.Fatal("cross-file closure field lost precision")
	}
	assertCompactElements(t, snapshot.files["definitions.php"].Summaries["factory"].returned.result.elements)
	replacement := syntax.Parse("definitions.php", []byte(`<?php function factory(){return fn($x)=>['clean'=>$x,'dirty'=>$x];}`), syntax.Options{Version: phpversion.Max})
	next := snapshot.WithFile(Extract(replacement, ix, phpversion.Max, nil, Options{}))
	if out.Current(snapshot, next) {
		t.Fatal("returned shape edit did not invalidate caller")
	}
	revised, _ := Summarize(caller, resourceTypes(caller, ix), next, Options{})
	if !revised.Tainted(at, HTML) {
		t.Fatal("updated returned field did not acquire taint")
	}
}
