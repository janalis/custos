package index

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestDigestAndHolds(t *testing.T) {
	src := []byte(`<?php class A {}`)
	f := syntax.Parse("a.php", src, syntax.Options{Version: phpversion.PHP84})
	same := syntax.Parse("a.php", src, syntax.Options{Version: phpversion.PHP84})
	older := syntax.Parse("a.php", src, syntax.Options{Version: phpversion.PHP74})
	edited := syntax.Parse("a.php", []byte(`<?php class B {}`), syntax.Options{Version: phpversion.PHP84})
	if Digest(f) != Digest(same) || Digest(f) == Digest(older) || Digest(f) == Digest(edited) {
		t.Fatal("digest does not follow bytes and version")
	}
	ix := New(nil)
	ix.Add(Extract(f))
	if ix.Holds("a.php", Digest(f)) {
		t.Fatal("held without a recorded digest")
	}
	fs := Extract(f)
	fs.Digest = Digest(f)
	ix.Add(fs)
	if !ix.Holds("a.php", Digest(f)) || ix.Holds("a.php", Digest(edited)) || ix.Holds("b.php", Digest(f)) {
		t.Fatal("holds")
	}
}

func TestHoldsRequiresUniqueNames(t *testing.T) {
	held := func(src string) (*Index, uint64) {
		f := syntax.Parse("a.php", []byte(src), syntax.Options{})
		fs := Extract(f)
		fs.Digest = Digest(f)
		ix := New(nil)
		ix.Add(fs)
		return ix, fs.Digest
	}
	for _, tc := range []struct{ name, src, other string }{
		{"class", `<?php class A {}`, `<?php class A {}`},
		{"function", `<?php function f() {}`, `<?php function f() {}`},
		{"constant", `<?php const C = 1;`, `<?php const C = 2;`},
		{"alias", `<?php class_alias('A', 'B');`, `<?php class_alias('X', 'B');`},
	} {
		ix, digest := held(tc.src)
		if !ix.Holds("a.php", digest) {
			t.Fatalf("%s: a unique name is not held", tc.name)
		}
		ix.Add(Extract(syntax.Parse("b.php", []byte(tc.other), syntax.Options{})))
		if ix.Holds("a.php", digest) {
			t.Fatalf("%s: a name another file declares is held", tc.name)
		}
	}
}
