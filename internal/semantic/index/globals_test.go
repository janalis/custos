package index

import (
	"maps"
	"slices"
	"testing"
)

func TestBoundGlobals(t *testing.T) {
	fs := extract(t, "a.php", `<?php
global $top;
$GLOBALS['file'] = 1;
function f() { global $a, $b, $$dynamic; $GLOBALS['c'] = 1; $GLOBALS["d"] .= 2; }
class K { function m() { global $a; $x = $GLOBALS; } }
$g = function () { $GLOBALS[$key] = 1; };
`)
	if got, want := fs.Globals, []string{"a", "b", "c", "d", ""}; !slices.Equal(got, want) {
		t.Fatalf("globals=%q, want %q", got, want)
	}
	if none := extract(t, "b.php", `<?php function f() { return 1; }`); none.Globals != nil {
		t.Fatalf("globals=%q", none.Globals)
	}
}

func TestGlobalNamesAcrossLayers(t *testing.T) {
	base := New(nil)
	base.Add(extract(t, "a.php", `<?php function f() { global $a; }`))
	overlay := New(base)
	overlay.Add(extract(t, "b.php", `<?php function g() { global $b; }`))
	names := func(ix *Index) []string { return slices.Sorted(maps.Keys(ix.GlobalNames())) }
	if got := names(overlay); !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("names=%q", got)
	}
	if got := names(overlay); !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("cached names=%q", got)
	}
	base.Add(extract(t, "c.php", `<?php function h() { global $c; }`))
	overlay.Remove("b.php")
	if got := names(overlay); !slices.Equal(got, []string{"a", "c"}) {
		t.Fatalf("names after changes=%q", got)
	}
}
