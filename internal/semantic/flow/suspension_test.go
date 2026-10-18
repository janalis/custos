package flow

import "testing"

func TestSuspensionKeepsLocals(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		nonNull   bool
	}{
		{"generator", `function g(){$x=unknown();if($x===null){return;}yield 1;probe($x);}`, true},
		{"delegating generator", `function g($i){$x=unknown();if($x===null){return;}yield from $i;probe($x);}`, true},
		{"by-reference function", `function &g(){$x=unknown();if($x===null){return;}yield $x;probe($x);}`, false},
		{"by-reference method", `class A{function &g(){$x=unknown();if($x===null){return;}yield 1=>$x;probe($x);}}`, false},
		{"by-reference closure", `$g=function &(){$x=unknown();if($x===null){return;}yield $x;probe($x);};`, false},
		{"by-reference element", `function &g(){$x=[unknown()];if($x[0]===null){return;}yield $x[0];probe($x);}`, false},
		{"yielded by value", `function g(){$x=unknown();if($x===null){return;}yield $x;probe($x);}`, true},
		{"file level", `$x=unknown();if($x===null){return;}yield $x;probe($x);`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			if got := e.Excludes(argument(callsNamed(f, "probe")[0]), "null"); got != tc.nonNull {
				t.Fatalf("non-null=%v", got)
			}
		})
	}
}

func TestSuspensionEndsRequestState(t *testing.T) {
	e, f := environment(t, `function g(){ob_start();probe(1);} function h(){ob_start();yield 1;probe(2);}`, nil)
	calls := callsNamed(f, "probe")
	if _, ok := e.GlobalStateBefore(calls[0], "output-buffer"); !ok {
		t.Fatal("buffer without a suspension")
	}
	if _, ok := e.GlobalStateBefore(calls[1], "output-buffer"); ok {
		t.Fatal("the consumer may end the buffer")
	}
}

func TestJumpsOutOfTryStatements(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		nonNull   bool
	}{
		{"label in a nested function", `function f(){$x=unknown();if($x===null){return;}$g=function(){done: return;};try{goto done;}finally{$x=null;}done: probe($x);}`, false},
		// A jump to no label leaves the scope unfollowed.
		{"missing label", `function f($c){$x=unknown();if($x===null){return;}try{if($c){goto nowhere;}}finally{echo 1;}probe($x);}`, false},
		{"file level", `$x=unknown();if($x===null){return;}try{goto done;}finally{$x=null;}done: probe($x);`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			if got := e.Excludes(argument(callsNamed(f, "probe")[0]), "null"); got != tc.nonNull {
				t.Fatalf("non-null=%v", got)
			}
		})
	}
}
