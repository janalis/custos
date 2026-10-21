package flow

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"custos/internal/php/syntax"
)

func projectionEnvironment(t testing.TB, src string) (*Env, *syntax.File, *File) {
	t.Helper()
	e, f := environment(t, src, nil)
	snapshot := NewSnapshot()
	var own *File
	for range 5 {
		_, own = Summarize(f, e.types, snapshot, Options{})
		next := NewSnapshot(own)
		if snapshot.Equal(next) {
			break
		}
		snapshot = next
	}
	return New(f, e.types, snapshot, Options{}), f, own
}

func TestProjectedWrapperDependencies(t *testing.T) {
	for _, tc := range []struct {
		name, body, call string
		safe             bool
	}{
		{"values", `function sink($a){system(implode(' ',$a));}`, `sink($a);`, true},
		{"search key in array", `function sink($a){system(implode(' ',[array_search(1,$a)]));}`, `sink($a);`, false},
		{"random key in array", `function sink($a){system(implode(' ',[array_rand($a)]));}`, `sink($a);`, false},
		{"trimmed value", `function sink($a){system(implode(' ',[trim($a[0])]));}`, `sink($a);`, true},
		{"single key in array", `function sink($a){system(implode(' ',[key($a)]));}`, `sink($a);`, false},
		{"flipped keys", `function sink($a){system(implode(' ',array_flip($a)));}`, `sink($a);`, false},
		{"keys through values", `function sink($a){system(implode(' ',array_values(array_keys($a))));}`, `sink($a);`, false},
		{"numeric keys", `function sink($a){system(implode(' ',array_map('intval',array_keys($a))));}`, `sink($a);`, true},
		{"cast key", `function sink(array $a){foreach($a as $key=>$x){system((int)$key);}}`, `sink($a);`, true},
		{"foreach key", `function sink(array $a){foreach($a as $key=>$x){system($key);}}`, `sink($a);`, false},
		{"keys", `function sink($a){system(implode(' ',array_keys($a)));}`, `sink($a);`, false},
		{"whole nested in values", `function encode($a){return json_encode($a);} function sink($a){system(implode(' ',[encode($a)]));}`, `sink($a);`, false},
		{"whole", `function sink($a){system(json_encode($a));}`, `sink($a);`, false},
		{"mixed", `function sink($a){system(implode(' ',$a).json_encode($a));}`, `sink($a);`, false},
		{"nested", `function inner($a){system(implode(' ',$a));}function sink($a){inner($a);}`, `sink($a);`, true},
		{"return", `function values($a){return implode(' ',$a);}function sink($a){system(values($a));}`, `sink($a);`, true},
		{"reference", `function values(&$out,$a){$out=implode(' ',$a);}function sink($a){$out="";values($out,$a);system($out);}`, `sink($a);`, true},
		{"foreach", `function sink(array $a){foreach($a as $x){system($x);}}`, `sink($a);`, true},
		{"offset", `function sink($a){system($a[0]);}`, `sink($a);`, true},
		{"named", `function sink($a){system(implode(' ',$a));}`, `sink(a:$a);`, true},
		{"rewritten", `function sink($a){system(implode(' ',$a));}`, `$a['x']=$_GET['x'];sink($a);`, false},
		{"path", `function sink($a){file_get_contents(implode(' ',$a));}`, `sink($a);`, false},
		{"url", `function sink($a){curl_init(implode(' ',$a));}`, `sink($a);`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f, own := projectionEnvironment(t, tc.body+`$a=array_map('intval',$_GET['items']);`+tc.call)
			sinks := e.Sinks(callsNamed(f, "sink")[0])
			if len(sinks) != 1 {
				t.Fatalf("sinks=%+v", sinks)
			}
			got := sinks[0].Argument.Safe&sinks[0].Context != 0
			if !sinks[0].Argument.Complete || len(sinks[0].Argument.Sources) == 0 || got != tc.safe {
				t.Fatalf("safe=%v want %v: %+v summaries=%+v", got, tc.safe, sinks[0], own.Summaries)
			}
		})
	}
}

func TestProjectionUnionWholeDominates(t *testing.T) {
	s := Source{Kind: "parameter", Parameter: 0}
	values := s
	values.ValuesOnly = true
	for _, pair := range [][2][]Source{{{s}, {values}}, {{values}, {s}}, {{values}, {values}}} {
		got, whole := union(pair[0], pair[1])
		if !whole || len(got) != 1 || got[0].ValuesOnly != (pair[0][0].ValuesOnly && pair[1][0].ValuesOnly) {
			t.Fatalf("union=%+v", got)
		}
	}
	t.Logf("Source=%d bytes", unsafe.Sizeof(Source{}))
}

func BenchmarkProjectedWrapper(b *testing.B) {
	e, f, _ := projectionEnvironment(b, `function sink($a){system(implode(' ',$a));}$a=array_map('intval',$_GET['items']);sink($a);`)
	call := callsNamed(f, "sink")[0]
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Sinks(call)
	}
}

func TestExactMethodSinkWrappers(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      bool
	}{
		{"final class", `final class C {function send($x){system($x);}} function run(C $o){$o->send($_GET['x']);}`, true},
		{"final method", `class C {final function send($x){system($x);}} function run(C $o){$o->send($_GET['x']);}`, true},
		{"open method", `class C {function send($x){system($x);}} function run(C $o){$o->send($_GET['x']);}`, false},
		{"inherited final", `class B {final function send($x){system($x);}} class C extends B {} function run(C $o){$o->send($_GET['x']);}`, true},
		{"final receiver", `class B {function send($x){system($x);}} final class C extends B {} function run(C $o){$o->send($_GET['x']);}`, true},
		{"protected inherited visible", `class A{final protected function send($x){system($x);}} class B extends A{} class C extends B{function run(A $o){$o->send($_GET['x']);}}`, true},
		{"private visible", `class C {private function send($x){system($x);} function run(){ $this->send($_GET['x']);}}`, true},
		{"private inaccessible", `class C {private function send($x){system($x);}} function run(C $o){$o->send($_GET['x']);}`, false},
		{"static final", `class C {final static function send($x){system($x);}} C::send($_GET['x']);`, true},
		{"static open", `class C {static function send($x){system($x);}} C::send($_GET['x']);`, false},
		{"trait", `trait T {final function send($x){system($x);}} final class C {use T;} function run(C $o){$o->send($_GET['x']);}`, false},
		{"conditional", `if($c){final class C {function send($x){system($x);}}} function run(C $o){$o->send($_GET['x']);}`, false},
		{"duplicate lexical scope", `final class A {function send($x){system($x);}} class C {function run(A $o){$o->send($_GET['x']);}} class C {}`, false},
		{"duplicate method", `final class C {function send($x){system($x);}function send($x){system($x);}} function run(C $o){$o->send($_GET['x']);}`, false},
		{"duplicate", `final class C {function send($x){system($x);}} final class C {function send($x){}} function run(C $o){$o->send($_GET['x']);}`, false},
		{"dynamic static class", `function run($class){$class::send($_GET['x']);}`, false},
		{"missing inherited method", `class B{} class C extends B{} function run(C $o){$o->missing($_GET['x']);}`, false},
		{"dynamic", `final class C {function send($x){system($x);}} function run(C $o,$m){$o->$m($_GET['x']);}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f, _ := projectionEnvironment(t, tc.src)
			var sinks []Sink
			syntax.InspectFile(f, func(n syntax.Node) bool {
				switch n.(type) {
				case *syntax.MethodCall, *syntax.StaticCall:
					sinks = append(sinks, e.Sinks(n)...)
				}
				return true
			})
			if (len(sinks) > 0) != tc.want {
				t.Fatalf("sinks=%+v want=%v", sinks, tc.want)
			}
		})
	}
}

func TestSinkDispatchMetadataInvalidatesSnapshots(t *testing.T) {
	for _, tc := range []struct{ name, old, new string }{
		{"receiver finality", `class B{function send($x){system($x);}} final class C extends B{}`, `class B{function send($x){system($x);}} class C extends B{}`},
		{"method finality", `class C{final function send($x){system($x);}}`, `class C{function send($x){system($x);}}`},
		{"parameter renamed", `final class C{function send($x){system($x);}}`, `final class C{function send($value){system($value);}}`},
		{"parameter default", `final class C{function send($x='first'){system($x);}}`, `final class C{function send($x='other'){system($x);}}`},
		{"parameter type", `final class C{function send(string $x){system($x);}}`, `final class C{function send($x){system($x);}}`},
		{"visibility", `class C{final public function send($x){system($x);}}`, `class C{final private function send($x){system($x);}}`},
		{"method added", `final class C{}`, `final class C{function send($x){system($x);}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldEnv, _, oldFile := projectionEnvironment(t, tc.old)
			_, _, newFile := projectionEnvironment(t, tc.new)
			before, after := NewSnapshot(oldFile), NewSnapshot(newFile)
			caller, file := environment(t, `function invoke(C $o,$x){$o->send($x);}`, before)
			caller.types.Index = oldEnv.types.Index
			_, own := Summarize(file, caller.types, before, Options{})
			if !own.Current(before, before) || own.Current(before, after) || before.Equal(after) || before.Holds(newFile) {
				t.Fatal("dispatch metadata change retained stale caller contract")
			}
			if before.WithFile(newFile).Holds(oldFile) || !before.WithFile(newFile).Equal(after) {
				t.Fatal("overlay dispatch metadata differs from fresh snapshot")
			}
		})
	}
}

func TestSinkMethodDepthBudgetAndConstructor(t *testing.T) {
	var src strings.Builder
	src.WriteString(`class C0 {final function send($x){system($x);}}`)
	for i := 1; i <= MaxDepth*4; i++ {
		fmt.Fprintf(&src, "class C%d extends C%d{}", i, i-1)
	}
	fmt.Fprintf(&src, `function run(C%d $o){$o->send($_GET['x']);} new C0;`, MaxDepth*4)
	e, f, _ := projectionEnvironment(t, src.String())
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.MethodCall, *syntax.New:
			if got := e.Sinks(n); len(got) != 0 {
				t.Fatalf("unsupported sink target: %+v", got)
			}
		}
		return true
	})
}

func TestUnknownWrapperTransformCannotProveProjectedSafety(t *testing.T) {
	e, f, _ := projectionEnvironment(t, `function sink($a){system(implode(' ',[unknown($a)]));}$a=array_map('intval',$_GET['rows']);sink($a);`)
	sinks := e.Sinks(callsNamed(f, "sink")[0])
	if len(sinks) != 1 || sinks[0].Argument.Complete {
		t.Fatalf("unknown transform established complete proof: %+v", sinks)
	}
}

func TestProtectedSinkAccessibilityDependency(t *testing.T) {
	base := `class A {final protected function send($x){system($x);}} class Z {}`
	caller := `class C extends B {function run(A $o,$x){$o->send($x);}}`
	_, _, old := projectionEnvironment(t, base+`class B extends A {}`+caller)
	_, _, next := projectionEnvironment(t, base+`class B extends Z {}`+caller)
	before, after := NewSnapshot(old), NewSnapshot(next)
	if old.Current(before, after) {
		t.Fatal("changed lexical ancestry retained protected access contract")
	}
}

func TestNamedMethodBindingInvalidatesCallerSummary(t *testing.T) {
	oldEnv, _, old := projectionEnvironment(t, `final class C{function send($value){system($value);}}`)
	_, _, next := projectionEnvironment(t, `final class C{function send($other){system($other);}}`)
	before, after := NewSnapshot(old), NewSnapshot(next)
	caller, file := environment(t, `function invoke(C $o,$x){$o->send(value:$x);}`, before)
	caller.types.Index = oldEnv.types.Index
	_, own := Summarize(file, caller.types, before, Options{})
	if len(own.Summaries["invoke"].Sinks) != 1 || own.Current(before, after) {
		t.Fatal("renamed method parameter retained named binding")
	}
}
