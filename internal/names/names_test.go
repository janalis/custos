package names

import (
	"strings"
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestResolve(t *testing.T) {
	src := `<?php
namespace App\Service;
use Psr\Log\LoggerInterface as Logger, Foo\Bar;
use function Util\helper;
use const Util\LIMIT;
use Vendor\{Pkg\One, function two, const THREE};
// marker
`
	f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	r := New(f)
	at := uint32(strings.Index(src, "// marker"))
	cases := []struct{ got, want string }{
		{r.Class("Logger", at), `Psr\Log\LoggerInterface`},
		{r.Class("Bar\\Baz", at), `Foo\Bar\Baz`},
		{r.Class("\\Exception", at), "Exception"},
		{r.Class("Local", at), `App\Service\Local`},
		{r.Class("namespace\\X", at), `App\Service\X`},
		{r.Class("self", at), "self"},
		{r.Class("One", at), `Vendor\Pkg\One`},
	}
	for i, c := range cases {
		if c.got != c.want {
			t.Errorf("#%d: got %s want %s", i, c.got, c.want)
		}
	}
	if fqn, fb := r.Function("helper", at); fqn != `Util\helper` || fb != "" {
		t.Errorf("helper: %s %s", fqn, fb)
	}
	if fqn, fb := r.Function("strlen", at); fqn != `App\Service\strlen` || fb != "strlen" {
		t.Errorf("strlen: %s %s", fqn, fb)
	}
	if fqn, _ := r.Function("two", at); fqn != `Vendor\two` {
		t.Errorf("two: %s", fqn)
	}
	if fqn, _ := r.Const("LIMIT", at); fqn != `Util\LIMIT` {
		t.Errorf("LIMIT: %s", fqn)
	}
	if fqn, _ := r.Const("THREE", at); fqn != `Vendor\THREE` {
		t.Errorf("THREE: %s", fqn)
	}
	if !r.IsGlobalFunction("STRLEN", at, "strlen") || r.IsGlobalFunction("helper", at, "helper") {
		t.Error("IsGlobalFunction")
	}
}

func TestBracedNamespaces(t *testing.T) {
	src := `<?php
namespace A { use X\Y; function f() { /*a*/ } }
namespace B { /*b*/ }
namespace { /*g*/ }
`
	f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	r := New(f)
	for marker, want := range map[string]string{"/*a*/": "A", "/*b*/": "B", "/*g*/": ""} {
		if got := r.Namespace(uint32(strings.Index(src, marker))); got != want {
			t.Errorf("%s: got %q want %q", marker, got, want)
		}
	}
	if got := r.Class("Y", uint32(strings.Index(src, "/*a*/"))); got != `X\Y` {
		t.Errorf("Y in A: %s", got)
	}
	if got := r.Class("Y", uint32(strings.Index(src, "/*b*/"))); got != `B\Y` {
		t.Errorf("Y in B: %s", got)
	}
}

func TestConstQualifiedAndFallback(t *testing.T) {
	src := `<?php
namespace App;
use Vendor\Pkg;
// marker
`
	f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	r := New(f)
	at := uint32(strings.Index(src, "// marker"))
	cases := []struct{ written, fqn, fallback string }{
		{`\PHP_EOL`, "PHP_EOL", ""},
		{`namespace\LIMIT`, `App\LIMIT`, ""},
		{`Pkg\LIMIT`, `Vendor\Pkg\LIMIT`, ""},  // first segment is an imported namespace
		{`Other\LIMIT`, `App\Other\LIMIT`, ""}, // qualified, not imported: relative to the namespace
		{`PHP_EOL`, `App\PHP_EOL`, "PHP_EOL"},  // unqualified in a namespace: global fallback
	}
	for _, c := range cases {
		if fqn, fb := r.Const(c.written, at); fqn != c.fqn || fb != c.fallback {
			t.Errorf("Const(%q) = %q, %q; want %q, %q", c.written, fqn, fb, c.fqn, c.fallback)
		}
	}
	g := New(syntax.Parse("g.php", []byte("<?php echo X;"), syntax.Options{Version: phpver.PHP84}))
	if fqn, fb := g.Const("X", 6); fqn != "X" || fb != "" {
		t.Errorf("global Const: %q %q", fqn, fb)
	}
}

func TestFunctionAndDeclarations(t *testing.T) {
	gsrc := "<?php\nuse function Lib\\g;\nclass G {}\n"
	gf := syntax.Parse("g.php", []byte(gsrc), syntax.Options{Version: phpver.PHP84})
	g := New(gf)
	if fqn, _ := g.Function("g", 1); fqn != `Lib\g` {
		t.Errorf("global use function: %q", fqn)
	}
	if fqn, fb := g.Function("strlen", 1); fqn != "strlen" || fb != "" {
		t.Errorf("global strlen: %q %q", fqn, fb)
	}
	src := `<?php
namespace A;
use Lib\Tools;
interface I extends \Countable {}
class C extends Base {}
class D {}
$o = new class {};
// marker
namespace B;
function h() {}
`
	f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	r := New(f)
	at := uint32(strings.Index(src, "// marker"))
	if r.Namespace(at) != "A" || r.Namespace(uint32(strings.Index(src, "function h"))) != "B" {
		t.Error("unbraced namespaces must extend to the next one")
	}
	fcases := []struct{ written, fqn string }{
		{`\strlen`, "strlen"},
		{`namespace\f`, `A\f`},
		{`Tools\f`, `Lib\Tools\f`},
		{`Other\f`, `A\Other\f`},
	}
	for _, c := range fcases {
		if fqn, fb := r.Function(c.written, at); fqn != c.fqn || fb != "" {
			t.Errorf("Function(%q) = %q, %q", c.written, fqn, fb)
		}
	}
	var decls []*syntax.ClassLike
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ClassLike); ok {
			decls = append(decls, c)
		}
		return true
	})
	if len(decls) != 4 {
		t.Fatalf("got %d class-likes", len(decls))
	}
	if r.DeclFQN(decls[0]) != `A\I` || r.ParentFQN(decls[0]) != "" {
		t.Error("interface")
	}
	if r.DeclFQN(decls[1]) != `A\C` || r.ParentFQN(decls[1]) != `A\Base` {
		t.Error("class C")
	}
	if r.ParentFQN(decls[2]) != "" || r.DeclFQN(decls[3]) != "" || r.DeclFQN(nil) != "" || r.ParentFQN(nil) != "" {
		t.Error("no parent / anonymous / nil")
	}
	var gc *syntax.ClassLike
	syntax.InspectFile(gf, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ClassLike); ok {
			gc = c
		}
		return true
	})
	if g.DeclFQN(gc) != "G" {
		t.Error("global class")
	}
	if !g.IsGlobalFunction(`STRLEN`, 1, "strlen") || g.Class(`namespace\X`, 1) != "X" {
		t.Error("global namespace")
	}
	if !IsBuiltinType("Int") || IsBuiltinType("Foo") {
		t.Error("IsBuiltinType")
	}
}
