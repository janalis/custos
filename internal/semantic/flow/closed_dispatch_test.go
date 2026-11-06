package flow

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func closedEnvironment(t testing.TB, src string) (*Env, *syntax.File) {
	t.Helper()
	e, f := environment(t, src, nil)
	file := Extract(f, e.types.Index, phpversion.Max, nil, Options{})
	return New(f, e.types, NewSnapshot(file), Options{}), f
}

func TestClosedDispatchReturnsAndSinks(t *testing.T) {
	for _, tc := range []struct {
		bodyA, bodyB string
		tainted      bool
	}{
		{`return $text;`, `return $text;`, true},
		{`return htmlspecialchars($text);`, `return htmlspecialchars($text);`, false},
		{`return htmlspecialchars($text);`, `return $text;`, true},
	} {
		e, f := closedEnvironment(t, `final class A { function render($text) { `+tc.bodyA+` } }
final class B { function render($text) { `+tc.bodyB+` } }
function f(A|B $object) { probe($object->render($_GET['x'])); }`)
		x := argument(callsNamed(f, "probe")[0])
		if !e.Value(x).Complete || e.Tainted(x, HTML) != tc.tainted {
			t.Fatalf("%s / %s: %+v", tc.bodyA, tc.bodyB, e.Value(x))
		}
	}
	e, f := closedEnvironment(t, `final class A { function send($text) { header($text); } }
final class B { function send($other, $text = '') { header($text); } }
function f(A|B $object) { $object->send(text: $_GET['x'], other: 'safe'); }`)
	var call *syntax.MethodCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.MethodCall); ok {
			call = c
		}
		return true
	})
	// One target cannot bind the extra named argument: no partial sink proof.
	if len(e.Sinks(call)) != 0 {
		t.Fatal("partial named binding")
	}
	e, f = closedEnvironment(t, `final class A { function send($text) { header($text); } }
final class B { function send($text) { header($text); } }
function f(A|B $object) { $object->send(text: $_GET['x']); }`)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.MethodCall); ok {
			call = c
		}
		return true
	})
	if sinks := e.Sinks(call); len(sinks) != 2 || !tainted(sinks[0].Argument) || !tainted(sinks[1].Argument) {
		t.Fatalf("sinks: %+v", sinks)
	}
}

func TestClosedDispatchStaysConservative(t *testing.T) {
	for _, src := range []string{
		`final class A {function render($x){return $x;}} final class B {function render($x){return $x;}} class Caller extends MissingParent {function f(A|B $o){probe($o->render($_GET['x']));}}`,
		`final class A {function render($x){return $x;}} final class B {function render($x){return $x;}} function f(A|B $o,$method){probe($o->$method($_GET['x']));}`,
		`class A { function render($x) { return $x; } } final class B { function render($x) { return $x; } } function f(A|B $o) { probe($o->render($_GET['x'])); }`,
		`final class A { function render($x) { return $x; } } final class B {} function f(A|B $o) { probe($o->render($_GET['x'])); }`,
		`final class A { private function render($x) { return $x; } } final class B { function render($x) { return $x; } } function f(A|B $o) { probe($o->render($_GET['x'])); }`,
		`final class A { function render($x) { return $x; } } final class B { function render($x) { yield $x; } } function f(A|B $o) { probe($o->render($_GET['x'])); }`,
		`final class A { function render($x) { return $x; } } final class B { function render($x) { return unknown($x); } } function f(A|B $o) { probe($o->render($_GET['x'])); }`,
		`final class A { function render($x) { return $x; } } final class B { function render($x) { return $x; } } /** @param A|B $o */ function f($o) { probe($o->render($_GET['x'])); }`,
	} {
		e, f := closedEnvironment(t, src)
		x := argument(callsNamed(f, "probe")[0])
		if e.Value(x).Complete {
			t.Fatalf("open or incomplete call accepted: %s", src)
		}
	}
}

func BenchmarkClosedDispatch(b *testing.B) {
	e, f := closedEnvironment(b, `final class A { function render($x) { return $x; } } final class B { function render($x) { return $x; } } function f(A|B $o) { probe($o->render($_GET['x'])); }`)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	for b.Loop() {
		New(f, e.types, e.snapshot, Options{}).Value(x)
	}
}

func TestClosedDispatchCrossFileDependencies(t *testing.T) {
	definitions := `final class A { function render($x){return $x;} } final class B { function render($x){return $x;} }`
	original, defs := closedFileEnvironment(t, "definitions.php", definitions, nil)
	file := Extract(defs, original.types.Index, phpversion.Max, nil, Options{})
	snapshot := NewSnapshot(file)
	consumer, source := closedFileEnvironment(t, "caller.php", `function wrap(A|B $o,$text){return $o->render($text);}`, snapshot)
	consumer.types.Index.Add(index.Extract(defs))
	extracted := Extract(source, consumer.types.Index, phpversion.Max, snapshot, Options{})
	for _, key := range []string{"a::render", "b::render", sinkClassKey("A"), sinkClassKey("B")} {
		if !slices.Contains(extracted.Consulted(), key) {
			t.Fatalf("missing dependency %q: %v", key, extracted.Consulted())
		}
	}
	changed, updated := closedFileEnvironment(t, "definitions.php", strings.Replace(definitions, "return $x;", "return htmlspecialchars($x);", 1), nil)
	next := NewSnapshot(Extract(updated, changed.types.Index, phpversion.Max, nil, Options{}))
	if extracted.Current(snapshot, next) {
		t.Fatal("target replacement did not invalidate caller")
	}
}

func TestClosedDispatchReferenceEscape(t *testing.T) {
	e, f := closedEnvironment(t, `final class A {function render($x){return $x;}} final class B {function render($x){return $x;}}
function f(A|B $o){$alias=&$o; mutate($alias); probe($o->render($_GET['x']));}`)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("escaped receiver established a closed target")
	}
}

func TestClosedDispatchInheritedAndUnreachable(t *testing.T) {
	e, f := closedEnvironment(t, `class Base {function render($x){return $x;}}
final class A extends Base {} final class B extends Base {}
function f(A|B $o){probe($o->render($_GET['x']));}`)
	if x := argument(callsNamed(f, "probe")[0]); !e.Value(x).Complete || !e.Tainted(x, HTML) {
		t.Fatalf("inherited: %+v", e.Value(x))
	}
	e, f = closedEnvironment(t, `final class A {function render($x){header($x);}} final class B {function render($x){header($x);}}
function make():A|B {} function f(){return;make()->render('x');}`)
	var call *syntax.MethodCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.MethodCall); ok {
			call = c
		}
		return true
	})
	if len(e.Sinks(call)) != 0 {
		t.Fatal("unreachable sink")
	}
}

func TestClosedDispatchIncompleteMetadata(t *testing.T) {
	const source = `final class A {function render($x){return $x;}} final class B {function render($x){return $x;}} function f(A|B $o){probe($o->render($_GET['x']));}`
	for _, corrupt := range []func(*File){
		func(file *File) { delete(file.Summaries, sinkClassKey("A")) },
		func(file *File) {
			summary := file.Summaries[sinkClassKey("A")]
			summary.dispatch.Methods["render"] = sinkMethod{Count: 2}
		},
	} {
		e, f := environment(t, source, nil)
		file := Extract(f, e.types.Index, phpversion.Max, nil, Options{})
		corrupt(file)
		local := New(f, e.types, NewSnapshot(file), Options{})
		if local.Value(argument(callsNamed(f, "probe")[0])).Complete {
			t.Fatal("incomplete or ambiguous metadata accepted")
		}
	}
}

func TestClosedDispatchTraitAndDepth(t *testing.T) {
	e, f := closedEnvironment(t, `trait Body {function render($x){return $x;}} final class A {use Body;} final class B {function render($x){return $x;}} function f(A|B $o){probe($o->render($_GET['x']));}`)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("trait target accepted without a fixed declaring chain")
	}
	var code strings.Builder
	code.WriteString(`class Root {function render($x){return $x;}}`)
	parent := "Root"
	for i := range MaxDepth*4 + 1 {
		name := fmt.Sprintf("C%d", i)
		fmt.Fprintf(&code, "class %s extends %s {}", name, parent)
		parent = name
	}
	fmt.Fprintf(&code, `final class A extends %s {} final class B extends %s {} function f(A|B $o){probe($o->render($_GET['x']));}`, parent, parent)
	e, f = closedEnvironment(t, code.String())
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("over-budget hierarchy accepted")
	}
}

func TestClosedDispatchRejectsPartialBindings(t *testing.T) {
	for _, src := range []string{
		`final class A {function render(&$x){return $x;}} final class B {function render(&$x){return $x;}} function f(A|B $o,$x){probe($o->render($x));}`,
		`final class A {function render($x){return $x;}} final class B {function render($y){return $y;}} function f(A|B $o){probe($o->render(x:$_GET['x']));}`,
		`final class A {function render($x){return $x;}} final class B {function render($x){return $x;}} function f(A|B|null $o){probe($o?->render($_GET['x']));}`,
	} {
		e, f := closedEnvironment(t, src)
		if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
			t.Fatal("partial binding or target proof")
		}
	}
	var code strings.Builder
	var names []string
	for i := range maxClosedTargets + 1 {
		name := fmt.Sprintf("T%d", i)
		names = append(names, name)
		fmt.Fprintf(&code, `final class %s {function render($x){return $x;}}`, name)
	}
	fmt.Fprintf(&code, `function f(%s $o){probe($o->render($_GET['x']));}`, strings.Join(names, "|"))
	e, f := closedEnvironment(t, code.String())
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("dispatch fanout budget ignored")
	}
}

func TestClosedDispatchInferredObjectUnion(t *testing.T) {
	e, f := closedEnvironment(t, `final class A {function render($x){return $x;}} final class B {function render($x){return $x;}}
function f($flag){$o=$flag ? new A : new B; probe($o->render($_GET['x']));}`)
	if x := argument(callsNamed(f, "probe")[0]); !e.Value(x).Complete || !e.Tainted(x, HTML) {
		t.Fatalf("inferred native union: %+v", e.Value(x))
	}
}

func TestClosedDispatchLoopAndInferredSinks(t *testing.T) {
	e, f := closedEnvironment(t, `final class A {function render($x){return $x;}} final class B {function render($x){return $x;}}
function f($flag){$o=$flag ? new A : new B;while($flag){probe($o->render($_GET['x']));$o=$flag ? new A : new B;}}`)
	if x := argument(callsNamed(f, "probe")[0]); !e.Value(x).Complete || !e.Tainted(x, HTML) {
		t.Fatalf("loop union: %+v", e.Value(x))
	}
	e, f = closedEnvironment(t, `final class A {function send($x){header($x);}} final class B {function send($x){header($x);}}
function f($flag){$o=$flag ? new A : new B;$o->send($_GET['x']);}`)
	var call *syntax.MethodCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if method, ok := n.(*syntax.MethodCall); ok {
			call = method
		}
		return true
	})
	if sinks := e.Sinks(call); len(sinks) != 2 || !tainted(sinks[0].Argument) || !tainted(sinks[1].Argument) {
		t.Fatalf("inferred sinks: %+v", sinks)
	}
}

func TestClosedDispatchTracksCallerAncestry(t *testing.T) {
	const definitions = `class Root {protected function send($x){header($x);}}
final class A extends Root {} final class B extends Root {}
class Other {} class ParentCaller extends Root {}`
	original, defs := closedFileEnvironment(t, "definitions.php", definitions, nil)
	old := NewSnapshot(Extract(defs, original.types.Index, phpversion.Max, nil, Options{}))
	consumer, source := closedFileEnvironment(t, "caller.php", `class Caller extends ParentCaller {function wrap(A|B $o,$text){$o->send($text);}}`, old)
	consumer.types.Index.Add(index.Extract(defs))
	file := Extract(source, consumer.types.Index, phpversion.Max, old, Options{})
	if len(file.Summaries["caller::wrap"].Sinks) != 1 {
		t.Fatal("protected target was not composed")
	}
	if !slices.Contains(file.Consulted(), sinkClassKey("ParentCaller")) {
		t.Fatal("caller ancestry not tracked")
	}
	changed, updated := closedFileEnvironment(t, "definitions.php", strings.Replace(definitions, "ParentCaller extends Root", "ParentCaller extends Other", 1), nil)
	next := NewSnapshot(Extract(updated, changed.types.Index, phpversion.Max, nil, Options{}))
	if file.Current(old, next) {
		t.Fatal("caller visibility change did not invalidate summary")
	}
}

func closedFileEnvironment(t testing.TB, path, src string, snapshot *Snapshot) (*Env, *syntax.File) {
	t.Helper()
	file := syntax.Parse(path, []byte("<?php "+src), syntax.Options{Version: phpversion.Max})
	if len(file.Errors) > 0 {
		t.Fatal(file.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(file))
	return New(file, infer.NewEnv(file, names.New(file), ix, phpversion.Max, infer.Options{}), snapshot, Options{}), file
}

func TestClosedDispatchFinalMethods(t *testing.T) {
	for _, classes := range []string{
		`class A{final function render($x){return $x;}} class B{final function render($x){return $x;}}`,
		`class Base{final function render($x){return $x;}} class A extends Base{} class B extends Base{}`,
		`class A{final function render($x){return $x;}} final class B{function render($x){return $x;}}`,
	} {
		e, f := closedEnvironment(t, classes+`function f(A|B $o){probe($o->render($_GET['x']));}`)
		x := argument(callsNamed(f, "probe")[0])
		if !e.Value(x).Complete || !e.Tainted(x, HTML) {
			t.Fatalf("%s: %+v", classes, e.Value(x))
		}
	}
}

func BenchmarkFinalMethodDispatch(b *testing.B) {
	e, f := closedEnvironment(b, `class A{final function render($x){return $x;}} class B{final function render($x){return $x;}} function f(A|B $o){probe($o->render($_GET['x']));}`)
	x := argument(callsNamed(f, "probe")[0])
	e.Value(x)
	b.ReportAllocs()
	for b.Loop() {
		e.Value(x)
	}
}

func TestClosedDispatchRejectsUncertainClassDeclarations(t *testing.T) {
	for _, classes := range []string{
		`interface A{function render($x);} interface B{function render($x);}`,
		`final class A{function render($x){return $x;}} final class A{function render($x){return 'safe';}} final class B{function render($x){return $x;}}`,
	} {
		e, f := closedEnvironment(t, classes+`function f(A|B $o){probe($o->render($_GET['x']));}`)
		if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
			t.Fatalf("uncertain class declaration accepted: %s", classes)
		}
	}
}

// A summary describes the body that runs only for a non-overridable target.
// Otherwise the result keeps the sources the declared body may yield but is
// not a complete value.
func TestOpenDispatchReturnsStayIncomplete(t *testing.T) {
	for _, tc := range []struct {
		src               string
		complete, tainted bool
	}{
		{`interface R { function get(); } function f(R $r) { probe($r->get()); }`, false, false},
		{`abstract class R { abstract function get(); } function f(R $r) { probe($r->get()); }`, false, false},
		{`class R { function get() { return 'safe'; } } class S extends R { function get() { return $_GET['x']; } } function f(R $r) { probe($r->get()); }`, false, false},
		{`class R { function get() { return $_GET['x']; } } function f(R $r) { probe($r->get()); }`, false, true},
		{`class R { function get() { return 1; } function f() { probe($this->get()); } }`, false, false},
		{`class R { static function get() { return 1; } function f() { probe(static::get()); } }`, false, false},
		{`/** @param R $r */ function f($r) { probe($r->get()); } final class R { function get() { return 1; } }`, false, false},
		{`final class R { function get() { return 'safe'; } } function f(R $r) { probe($r->get()); }`, true, false},
		{`class R { final function get() { return 1; } } function f(R $r) { probe($r->get()); }`, true, false},
		{`class R { private function get() { return 1; } function f() { probe($this->get()); } }`, true, false},
		{`function get() { return 1; } probe(get());`, true, false},
	} {
		e, f := closedEnvironment(t, tc.src)
		x := argument(callsNamed(f, "probe")[0])
		if v := e.Value(x); v.Complete != tc.complete || e.Tainted(x, HTML) != tc.tainted {
			t.Errorf("%s: complete=%v tainted=%v", tc.src, v.Complete, e.Tainted(x, HTML))
		}
	}
}
