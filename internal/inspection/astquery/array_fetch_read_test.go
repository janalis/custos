package astquery

import (
	"slices"
	"testing"

	"custos/internal/php/syntax"
)

func TestArrayFetchRequiresRead(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want []bool
	}{
		{"echo $p['x'];", []bool{true}},
		{"$p['x']=1;", []bool{false}},
		{"$p['x']+=1;", []bool{true}},
		{"$p['x']??=1;", []bool{false}},
		{"$p['x']['y']=1;", []bool{false, false}},
		{"isset($p['x']);", []bool{false}},
		{"empty($p['x']);", []bool{false}},
		{"unset($p['x']);", []bool{false}},
		{"echo ($p['x']) ?? 'fallback';", []bool{false}},
		{"echo $a ?? $p['x'];", []bool{true}},
		{"empty($p['x'] + 1);", []bool{true}},
		{"isset($p[$p['x']]);", []bool{false, true}},
		{"foreach ($items as $p['x']) {}", []bool{false}},
		{"foreach ($p['x'] as $item) {}", []bool{true}},
		{"[$p['x']] = [1];", []bool{false}},
		{"list($p['x']) = [1];", []bool{false}},
		{"[[$p['x']]] = [[1]];", []bool{false}},
		{"foreach ($items as [$p['x']]) {}", []bool{false}},
		{"echo [$p['x']];", []bool{true}},
		{"[$p['x'] => $target] = [1];", []bool{true}},
		{"$target = [$p['x']];", []bool{true}},
		{"echo [$p['x']] ?? [];", []bool{true}},
		{"foreach ([$p['x']] as $item) {}", []bool{true}},
		{"[[$p['x']] => $target] = [1];", []bool{true}},
	} {
		file := syntax.Parse("case.php", []byte("<?php "+tc.src), syntax.Options{})
		if len(file.Errors) != 0 {
			t.Fatal(file.Errors)
		}
		var got []bool
		syntax.InspectFile(file, func(n syntax.Node) bool {
			if fetch, ok := n.(*syntax.ArrayDimFetch); ok {
				got = append(got, ArrayFetchRequiresRead(fetch))
			}
			return true
		})
		if !slices.Equal(got, tc.want) {
			t.Fatalf("%s: got %v want %v", tc.src, got, tc.want)
		}
	}
}

func BenchmarkArrayFetchRequiresRead(b *testing.B) {
	file := syntax.Parse("bench.php", []byte("<?php echo ($p['key']['value']) ?? 'fallback';"), syntax.Options{})
	var fetch *syntax.ArrayDimFetch
	syntax.InspectFile(file, func(n syntax.Node) bool {
		if n, ok := n.(*syntax.ArrayDimFetch); ok && fetch == nil {
			fetch = n
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		ArrayFetchRequiresRead(fetch)
	}
}
