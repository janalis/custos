package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
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
