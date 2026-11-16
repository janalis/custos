package flow

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkExtractScopes summarizes a file of callables that branch, loop
// and call each other and project methods: the cross-file flow phase of a
// cold run, which follows every scope once per file.
func BenchmarkExtractScopes(b *testing.B) {
	var source strings.Builder
	source.WriteString(`class Repo { function find($id) { return $this->load($id); } function load($id) { return $id; } }
function pass($x) { return $x; }
`)
	for i := range 32 {
		fmt.Fprintf(&source, `function caller%d(Repo $r, $x, $y) {
	$a = pass($x);
	if ($a) { $b = $r->find($a); } else { $b = pass($y); }
	foreach ([$a, $b] as $k => $v) { if ($v) { $a .= $k; continue; } $b = $r->load($v); }
	try { $c = strlen($a) + count([$b]); } catch (Exception $e) { $c = 0; }
	return [$a, $b, $c];
}
`, i)
	}
	e, file := environment(b, source.String(), nil)
	b.ReportAllocs()
	for b.Loop() {
		Extract(file, e.types.Index, e.types.PHP, nil, Options{})
	}
}
