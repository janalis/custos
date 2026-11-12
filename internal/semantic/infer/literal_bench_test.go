package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// BenchmarkTypeOfLiterals types every expression of a file dense in
// literals: integers, strings (quoted, escaped, heredoc), negated values and
// literal arrays read through variables.
func BenchmarkTypeOfLiterals(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("<?php\n")
	for i := range 200 {
		fmt.Fprintf(&sb, `function f%d(int $n) {
    $a = %d; $b = -%d; $c = 'key%d'; $d = "line\n%d";
    $rows = ['id' => %d, 'name' => 'n%d', 'tags' => [1, 2, 3, 'x', 'y']];
    $list = [0, 1, 2, -1, 0x1F, 'a', 'bb', "c\t"];
    $s = <<<EOT
    text %d
    EOT;
    if ($n > 10) { $a = 20; }
    return [$a, $b, $c, $d, $rows['id'], $list, $s, strlen('abc') + 1];
}
`, i, i, i, i, i, i, i, i)
	}
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpversion.PHP84}
	b.ReportAllocs()
	for b.Loop() {
		f := syntax.Parse("t.php", src, opt)
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if x, ok := n.(syntax.Expr); ok {
				_ = env.TypeOf(x)
			}
			return true
		})
	}
}
