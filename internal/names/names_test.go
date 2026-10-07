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
