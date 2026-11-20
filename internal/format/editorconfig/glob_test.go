package editorconfig

import (
	"strings"
	"testing"
	"time"

	"github.com/janalis/custos/internal/testing/testbudget"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		glob, rel string
		want      bool
	}{
		// Literals and base-name matching at any depth.
		{"a.php", "a.php", true},
		{"a.php", "src/deep/a.php", true},
		{"a.php", "b.php", false},
		{"a.php", "xa.php", false},
		{"a.php", "a.php/x", false},
		// *, **, ?
		{"*.php", "a.php", true},
		{"*.php", "src/a.php", true},
		{"*.php", "a.phpx", false},
		{"*", "x", true},
		{"src/*.php", "src/a.php", true},
		{"src/*.php", "src/b/a.php", false},
		{"src/**.php", "src/b/a.php", true},
		{"**.php", "a/b/c.php", true},
		{"src/**", "src/a/b", true},
		{"src/**", "lib/a", false},
		{"?.php", "a.php", true},
		{"?.php", "é.php", true},
		{"?.php", "ab.php", false},
		{"src?a.php", "src/a.php", false},
		{"*é?", "aéb", true},
		{"*é??", "aéb", false},
		// /**/ matches zero or more directories.
		{"a/**/b", "a/b", true},
		{"a/**/b", "a/x/b", true},
		{"a/**/b", "a/x/y/b", true},
		{"a/**/b", "ab", false},
		{"a/**/b", "x/a/b", false},
		{"**/b.php", "b.php", true},
		{"**/b.php", "x/y/b.php", true},
		// Anchoring: a '/' anchors the glob to the file's directory.
		{"src/a.php", "src/a.php", true},
		{"src/a.php", "x/src/a.php", false},
		{"/a.php", "a.php", true},
		{"/a.php", "x/a.php", false},
		{"/src/a.php", "src/a.php", true},
		{"src/", "src", false},
		// Character classes.
		{"[abc].php", "b.php", true},
		{"[abc].php", "d.php", false},
		{"[a-z].php", "q.php", true},
		{"[a-z].php", "Q.php", false},
		{"[!a-z].php", "Q.php", true},
		{"[!a-z].php", "q.php", false},
		{"[!a]", "/", false},
		{"x[!a]y", "x/y", false},
		{"[a-]", "-", true},
		{"[a\\-z]", "-", true},
		{"[a\\-z]", "b", false},
		{"[\\]]", "]", true},
		{"[!]", "x", true},
		{"[abc", "[abc", true},
		{"[abc", "a", false},
		{"ab[e/]cd.i", "ab[e/]cd.i", true},
		{"ab[e/]cd.i", "abecd.i", false},
		// Braces.
		{"*.{php,phtml}", "a.php", true},
		{"*.{php,phtml}", "a.phtml", true},
		{"*.{php,phtml}", "a.js", false},
		{"{a,{b,c}d}.x", "cd.x", true},
		{"{a,{b,c}d}.x", "a.x", true},
		{"{a,{b,c}d}.x", "c.x", false},
		{"a{,b}.x", "a.x", true},
		{"a{,b}.x", "ab.x", true},
		{"{single}.b", "{single}.b", true},
		{"{single}.b", "single.b", false},
		{"{}.b", "{}.b", true},
		{"{a,b", "{a,b", true},
		{"{a\\,b,c}", "a,b", true},
		{"{a\\,b,c}", "a", false},
		{"{a\\}b,c}", "a}b", true},
		{"{src,lib}/*.php", "lib/a.php", true},
		{"{src,lib}/*.php", "x/lib/a.php", false},
		// Integer ranges.
		{"f{1..3}.php", "f2.php", true},
		{"f{1..3}.php", "f4.php", false},
		{"f{3..1}.php", "f1.php", true},
		{"f{-3..3}.php", "f-2.php", true},
		{"f{-3..3}.php", "f-4.php", false},
		{"f{-3..3}.php", "f+3.php", true},
		{"f{-3..3}.php", "f003.php", true},
		{"f{-10..-5}", "f-7", true},
		{"f{1..30}x", "f12x", true},
		{"f{1..30}x", "f123x", false},
		{"f{1..3}", "fx", false},
		{"f{1..3}", "f", false},
		{"f{1..3}", "f-", false},
		{"f{0..9}", "f99999999999999999999", false},
		{"f{a..b}", "fa", false},
		{"f{a..b}", "f{a..b}", true},
		// Escapes.
		{"\\*.php", "*.php", true},
		{"\\*.php", "a.php", false},
		{"a\\?", "a?", true},
		{"a\\", "a\\", true},
		{"a\\/b", "a/b", true},
		{"a\\/b", "x/a/b", false},
	}
	for _, tt := range tests {
		if got := Match(tt.glob, tt.rel); got != tt.want {
			t.Errorf("Match(%q, %q) = %v, want %v", tt.glob, tt.rel, got, tt.want)
		}
	}
	// A second call is served from the cache.
	if !Match("*.php", "a.php") {
		t.Error("cached glob no longer matches")
	}
}

func BenchmarkMatch(b *testing.B) {
	for b.Loop() {
		Match("{src,lib}/**/*.{php,phtml}", "src/a/b/c/d/e/file.php")
	}
}

// TestMatchBudget: a glob that would backtrack for an exponential time
// fails once it exhausts its budget, quickly.
func TestMatchBudget(t *testing.T) {
	glob := strings.Repeat("*a", 30) + "b"
	name := strings.Repeat("a", 60)
	done := make(chan bool, 1)
	go func() { done <- Match(glob, name) }()
	select {
	case got := <-done:
		if got {
			t.Error("pathological glob matched")
		}
	case <-time.After(testbudget.Of(5 * time.Second)):
		t.Fatal("pathological glob still running")
	}
}
