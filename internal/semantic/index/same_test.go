package index

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestSameDeclarations(t *testing.T) {
	extract := func(src string) *FileSymbols {
		return Extract(syntax.Parse("a.php", []byte(src), syntax.Options{Version: phpversion.PHP84}))
	}
	base := extract(`<?php
namespace App;
const LIMIT = 3;
class A { public int $p = 1; /** @return int[] */ public function f(array $x): array { return $x; } }
function g(string $s): int { return 1; }
`)
	for _, tc := range []struct {
		name, src string
		same      bool
	}{
		{"moved and body edited", `<?php


namespace App;
const LIMIT = 3;
class A { public int $p = 1; /** @return int[] */ public function f(array $x): array { $y = $x; return $y; } }
function g(string $s): int { echo $s; return 1; }
`, true},
		{"return type", `<?php
namespace App;
const LIMIT = 3;
class A { public int $p = 1; /** @return string[] */ public function f(array $x): array { return $x; } }
function g(string $s): int { return 1; }
`, false},
		{"new method", `<?php
namespace App;
const LIMIT = 3;
class A { public int $p = 1; /** @return int[] */ public function f(array $x): array { return $x; } public function h() {} }
function g(string $s): int { return 1; }
`, false},
		{"constant value", `<?php
namespace App;
const LIMIT = 4;
class A { public int $p = 1; /** @return int[] */ public function f(array $x): array { return $x; } }
function g(string $s): int { return 1; }
`, false},
	} {
		other := extract(tc.src)
		if got := SameDeclarations(base, other); got != tc.same {
			t.Errorf("%s: same = %v, want %v", tc.name, got, tc.same)
		}
		// The fingerprint agrees, whatever order maps list their keys in.
		if got := Fingerprint(base) == Fingerprint(other); got != tc.same {
			t.Errorf("%s: same fingerprint = %v, want %v", tc.name, got, tc.same)
		}
	}
	if !SameDeclarations(nil, nil) || SameDeclarations(base, nil) {
		t.Error("nil symbols")
	}
	var empty, none FileSymbols
	none.Globals = []string{}
	if SameDeclarations(&empty, &none) || Fingerprint(&empty) == Fingerprint(&none) {
		t.Error("nil and empty slice compared equal")
	}
	if Fingerprint(nil) == Fingerprint(&empty) {
		t.Error("nil symbols fingerprinted as empty ones")
	}
	// Maps, nil or empty, in key order; fallbacks keyed by number.
	withMap := func(m map[int]string) *FileSymbols { return &FileSymbols{ClassAliasFallbacks: m} }
	if Fingerprint(withMap(nil)) == Fingerprint(withMap(map[int]string{})) ||
		Fingerprint(withMap(map[int]string{1: "a", -2: "b"})) != Fingerprint(withMap(map[int]string{-2: "b", 1: "a"})) ||
		Fingerprint(withMap(map[int]string{1: "a"})) == Fingerprint(withMap(map[int]string{1: "b"})) {
		t.Error("maps")
	}
	if Fingerprint(&FileSymbols{Path: "a", Digest: 1}) != Fingerprint(&FileSymbols{Path: "a", Digest: 2}) {
		t.Error("digest fingerprinted")
	}
}

func TestSameDeclarationsIgnoresDigest(t *testing.T) {
	a, b := &FileSymbols{Path: "a.php", Digest: 1}, &FileSymbols{Path: "a.php", Digest: 2}
	if !SameDeclarations(a, b) || a.Digest != 1 || b.Digest != 2 {
		t.Fatal("digest compared or cleared")
	}
}
