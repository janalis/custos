package infer_test

import (
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/infer"
	"github.com/janalis/custos/internal/semantic/names"
	"github.com/janalis/custos/internal/semantic/stubs"
)

// TestMemberTypeCacheTracksIndex covers the cached types of declarations:
// a PHPDoc type narrower than the declared one counts once the index shows
// it is (here, once the documented class implements the declared
// interface), and the cache follows a replaced index.
func TestMemberTypeCacheTracksIndex(t *testing.T) {
	parse := func(path, src string) *syntax.File {
		f := syntax.Parse(path, []byte(src), syntax.Options{Version: phpversion.PHP84})
		if len(f.Errors) > 0 {
			t.Fatalf("parse %s: %v", path, f.Errors)
		}
		return f
	}
	f := parse("main.php", `<?php function f(Q $q) { $q->a; $q->b; $q->c; }`)
	var reads []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if p, ok := n.(*syntax.PropertyFetch); ok {
			reads = append(reads, p)
		}
		return true
	})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(parse("q.php", `<?php interface I {} class Q { /** @var M */ public I $a; /** @var M */ public I $b; /** @var M */ public I $c; }`)))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
	if got := env.TypeOf(reads[0]).String(); got != `\I` {
		t.Fatalf("before M: %s", got)
	}
	ix.Add(index.Extract(parse("m.php", `<?php class M implements I {}`)))
	if got := env.TypeOf(reads[1]).String(); got != `\M` {
		t.Fatalf("after M: %s", got)
	}
	env.Index = index.New(stubs.Index())
	env.Index.Add(index.Extract(parse("q2.php", `<?php interface I {} class Q { /** @var M */ public I $a; /** @var M */ public I $b; /** @var M */ public I $c; }`)))
	if got := env.TypeOf(reads[2]).String(); got != `\I` {
		t.Fatalf("replaced index: %s", got)
	}
}

// TestNativeAndDocTypesStayDistinct reads the same declarations, a property
// and a method return each documented narrower than declared, alternately
// through an Env and its native twin: each view keeps its own cached type,
// whichever view parses the declaration first.
func TestNativeAndDocTypesStayDistinct(t *testing.T) {
	src := `<?php interface I {} class M implements I {}
class Q { /** @var M */ public I $a; /** @return M */ public function r(): I { return new M; } }
function f(Q $q) { $q->a; $q->r(); $q->a; $q->r(); }`
	f := syntax.Parse("main.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	var reads []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.PropertyFetch:
			reads = append(reads, x)
		case *syntax.MethodCall:
			reads = append(reads, x)
		}
		return true
	})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
	native := env.Native()
	// Native first for the first two reads, PHPDoc-aware first after.
	for i, r := range reads {
		first, second, want := native, env, [2]string{`\I`, `\M`}
		if i >= 2 {
			first, second, want = env, native, [2]string{`\M`, `\I`}
		}
		if got := first.TypeOf(r).String(); got != want[0] {
			t.Fatalf("read %d, first view: %s, want %s", i, got, want[0])
		}
		if got := second.TypeOf(r).String(); got != want[1] {
			t.Fatalf("read %d, second view: %s, want %s", i, got, want[1])
		}
	}
}
