package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

// Constructors that only store their parameters are indexed as such.
func TestStoresParamsIndex(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php
class A { public function __construct($db, protected int $n) { $this->db = $db; } }
class B { public function __construct($db) { $this->db = strtolower($db); } }
class C { public function __construct($db) { $this->db = $db; foo(); } }
class D { public function __construct() {} public function m() { return random_bytes(8); } }
class E { public function __construct($x) { $this->a[] = $x; } }
class F { public function __construct($x) { $y->a = $x; } }
class G { public function __construct($x) { $this->a = $other; } }
class H { public function __construct($x) { $this->a = &$x; } }
abstract class I { abstract public function __construct($x); }
class J { public function __construct($x) { $this?->a = $x; } }
function iv() { return openssl_random_pseudo_bytes(16); }
function weak() { return 'x'; }
function viaMethod($o) { return $o->random_bytes(8); }
function viaStatic() { return R::mcrypt_create_iv(8); }
function qualified() { return \random_bytes(8); }
`), syntax.Options{Version: phpversion.PHP84})
	fs := index.Extract(f)
	want := map[string]bool{"A": true, "B": false, "C": false, "D": true, "E": false, "F": false, "G": false, "H": false, "I": false, "J": false}
	for _, c := range fs.Classes {
		if got := c.Methods["__construct"].StoresParams; got != want[c.FQN] {
			t.Errorf("%s: StoresParams %v", c.FQN, got)
		}
	}
	if s := fs.Classes[0].Methods["__construct"].Stores; len(s) != 2 || s[0] != "n" || s[1] != "db" {
		t.Errorf("A stores %v", s)
	}
	if !fs.Classes[3].Methods["m"].CSPRNG {
		t.Error("D::m CSPRNG")
	}
	csprng := map[string]bool{"iv": true, "weak": false, "viaMethod": true, "viaStatic": true, "qualified": true}
	for _, fn := range fs.Functions {
		if fn.CSPRNG != csprng[fn.FQN] {
			t.Errorf("%s: CSPRNG %v", fn.FQN, fn.CSPRNG)
		}
	}
}

// AncestorsComplete reports a hierarchy cut by MaxAncestors.
func TestAncestorsComplete(t *testing.T) {
	var b strings.Builder
	b.WriteString("<?php\nclass Big implements ")
	for i := 0; i < index.MaxAncestors+10; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "I%d", i)
	}
	b.WriteString(" {}\nclass Small {}\n")
	for i := 0; i < index.MaxAncestors+10; i++ {
		fmt.Fprintf(&b, "interface I%d {}\n", i)
	}
	f := syntax.Parse("t.php", []byte(b.String()), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	if ix.AncestorsComplete("Big", 0) || !ix.AncestorsComplete("Small", 0) || !ix.AncestorsComplete("Missing", 0) {
		t.Error("AncestorsComplete")
	}
}

// Properties extensions create but the stubs omit are declared.
func TestStubMissingProps(t *testing.T) {
	ix := stubs.Index()
	if ix.FindProperty("OAuthProvider", "nonce", phpversion.PHP84) == nil || ix.FindProperty("OAuthProvider", "bogus", phpversion.PHP84) != nil {
		t.Error("OAuthProvider properties")
	}
}
