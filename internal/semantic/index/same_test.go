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
		if got := SameDeclarations(base, extract(tc.src)); got != tc.same {
			t.Errorf("%s: same = %v, want %v", tc.name, got, tc.same)
		}
	}
	if !SameDeclarations(nil, nil) || SameDeclarations(base, nil) {
		t.Error("nil symbols")
	}
	var empty, none FileSymbols
	none.Globals = []string{}
	if SameDeclarations(&empty, &none) {
		t.Error("nil and empty slice compared equal")
	}
}
