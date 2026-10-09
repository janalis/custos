package index

import (
	"testing"

	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestAnonymousClassIndex(t *testing.T) {
	const src = `<?php namespace App;
 interface Contract { public function inherited(): string; }
 trait Members { public int $traitProp = 1; public function traitMethod(): int { return 1; } public const TRAIT_VALUE = 1; }
 class Base { public string $baseProp = ''; public function inherited(): string { return ''; } }
 $one = new class(1) extends Base implements Contract {
 use Members;
 public const VALUE = 2;
 public function __construct(public int $promoted) {}
 public function own(): self { return $this; }
 public function nested() { return new class { public string $nested = ''; }; }
 };
 $two = new class {};
 `
	f := syntax.Parse("classes.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	if len(f.Errors) > 0 {
		t.Fatal(f.Errors)
	}
	fs := Extract(f)
	var anonymous []*Class
	for _, c := range fs.Classes {
		if c.Anonymous {
			anonymous = append(anonymous, c)
		}
	}
	if len(anonymous) != 3 {
		t.Fatalf("anonymous classes: %d", len(anonymous))
	}
	ix := New(nil)
	ix.Add(fs)
	c := anonymous[0]
	if !names.IsAnonymousClassName(c.FQN) || c.Parent != `App\Base` || len(c.Interfaces) != 1 || c.Interfaces[0] != `App\Contract` {
		t.Fatalf("bad identity/hierarchy: %+v", c)
	}
	for _, property := range []string{"promoted", "traitProp", "baseProp"} {
		if ix.FindProperty(c.FQN, property, phpver.PHP84) == nil {
			t.Error("missing property", property)
		}
	}
	for _, method := range []string{"own", "traitMethod", "inherited"} {
		if ix.FindMethod(c.FQN, method, phpver.PHP84) == nil {
			t.Error("missing method", method)
		}
	}
	for _, constant := range []string{"VALUE", "TRAIT_VALUE"} {
		if ix.FindConst(c.FQN, constant, phpver.PHP84) == nil {
			t.Error("missing constant", constant)
		}
	}
	if anonymous[1].Props["nested"] == nil || c.FQN == anonymous[2].FQN {
		t.Fatal("nested or separate identity")
	}
	other := Extract(syntax.Parse("other.php", []byte(src), syntax.Options{Version: phpver.PHP84}))
	ix.Add(other)
	if ix.ClassCount(c.FQN) != 1 {
		t.Fatal("different files collide")
	}
	ix.Add(Extract(syntax.Parse("classes.php", []byte(`<?php $replacement = new class { public int $replacement; };`), syntax.Options{Version: phpver.PHP84})))
	if ix.Class(c.FQN, phpver.PHP84) != nil {
		t.Fatal("old declaration survived replacement")
	}
	ix.Remove("classes.php")
	for _, cls := range ix.files["other.php"].Classes {
		if cls.Anonymous && ix.Class(cls.FQN, phpver.PHP84) == nil {
			t.Fatal("removal affected another file")
		}
	}
	ix.Remove("other.php")
	for _, cls := range other.Classes {
		if ix.Class(cls.FQN, phpver.PHP84) != nil {
			t.Fatal("classes survived removal")
		}
	}
}
