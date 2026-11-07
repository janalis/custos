package flow

import "testing"

func TestFiniteMappingBindings(t *testing.T) {
	for _, tc := range []struct {
		source string
		taint  bool
	}{
		{`$args=['htmlspecialchars',[$_GET['x']]];$xs=array_map(...$args);probe($xs[0]);`, false},
		{`function esc($v,$ignored){return htmlspecialchars($v);}$xs=array_map(...['esc',[$_GET['x']],['safe']]);probe($xs[0]);`, false},
		{`$xs=array_map(...['array'=>[$_GET['x']],'callback'=>'htmlspecialchars']);probe($xs[0]);`, false},
		{`$xs=array_map(null,['key'=>$_GET['x']]);probe($xs['key']);`, true},
		{`$xs=array_map(null,['key'=>'safe'],['other'=>$_GET['x'],'extra'=>'safe']);probe($xs[0][0]);`, false},
		{`$xs=array_map(null,['key'=>'safe'],['other'=>$_GET['x'],'extra'=>'safe']);probe($xs[0][1]);`, true},
		{`$xs=array_map(null,['key'=>'safe'],['other'=>$_GET['x'],'extra'=>'safe']);probe($xs[1][0]);`, false},
		{`$xs=array_map(...['callback'=>null,'array'=>['key'=>'safe']]);probe($xs['key']);`, false},
	} {
		e, file := summarized(t, tc.source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if !v.Complete || e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, v)
		}
	}
}

func TestArrayShapesDoNotSurviveScalarCoercions(t *testing.T) {
	for _, source := range []string{
		`$a=(string)[$_GET['x']];array_map('header',$a);`,
		`$a=(bool)[$_GET['x']];array_filter($a,'header');`,
		`$a=[$_GET['x']]==[$_GET['x']];array_filter($a,'header');`,
		`$a=(object)[$_GET['x']];array_map('header',$a);`,
		`function bad($x):bool{return [$x];}$a=bad($_GET['x']);array_filter($a,'header');`,
		`function make(){return fn($x):bool=>[$x];}$f=make();$a=$f($_GET['x']);array_map('header',$a);`,
	} {
		e, file := summarized(t, source)
		for _, name := range []string{"array_map", "array_filter"} {
			for _, at := range callsNamed(file, name) {
				if len(e.Sinks(at)) != 0 {
					t.Fatalf("scalar retained array callback proof: %s", source)
				}
			}
		}
	}
}
