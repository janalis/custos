package infer_test

import "testing"

// TestDestructuringReplacesValue: destructuring by value assigns every
// variable of the pattern (null where the key is missing), so the value a
// variable held before no longer reaches past it.
func TestDestructuringReplacesValue(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"short list", `<?php function f(string $h) { $p = 1.5; [$h, $p] = explode('/', $h); t('v', $p); }`, "string"},
		{"list()", `<?php function f(string $h) { $p = 1.5; list(, $p) = explode('/', $h); t('v', $p); }`, "string"},
		{"in a branch", `<?php function f(string $h) { $p = 1.5; if ($h !== '') { [$h, $p] = explode('/', $h); t('v', $p); } }`, "string"},
		{"nested", `<?php function f(string $h) { $p = 1.5; [[$p]] = [explode('/', $h)]; t('v', $p); }`, "string"},
		{"keyed", `<?php function f(string $h) { $p = 1.5; ['k' => $p] = ['k' => 'x']; t('v', $p); }`, "string"},
		{"after the branch", `<?php function f(string $h) { $p = 1.5; if ($h !== '') { [$h, $p] = explode('/', $h); } t('v', $p); }`, "float|string"},
		// A reference binding is no assignment of the others' values.
		{"by reference", `<?php function f(string $h) { $l = explode('/', $h); $q = 1.5; [$q, &$r] = $l; t('v', $q); }`, "float|string"},
		{"nested by reference", `<?php function f(string $h) { $l = [explode('/', $h)]; $q = 1.5; [[$q, &$r]] = $l; t('v', $q); }`, "float|string"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			native, proven := provenProbe(t, c.src)
			if native != c.want || proven != c.want {
				t.Errorf("native %s, proven %s; want %s", native, proven, c.want)
			}
		})
	}
}
