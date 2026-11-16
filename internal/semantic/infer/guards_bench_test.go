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

// BenchmarkPropertyReadGuards types every property read of methods whose
// statements mention a nullable property between reads: plain statements,
// writes into its offsets (assignments, increments, destructuring) and
// calls, the expression statements guards examines for each read.
func BenchmarkPropertyReadGuards(b *testing.B) {
	for _, scenario := range []struct{ name, stmt string }{
		{"plain", `$this->log($this->p, %[1]d);`},
		{"elemWrites", `$this->p['k%[1]d'] = %[1]d; $this->p['n']++; [$this->p['a'], $x] = [1, 2]; $this->log($this->p);`},
	} {
		b.Run(scenario.name, func(b *testing.B) {
			var src strings.Builder
			src.WriteString("<?php class G {\n    private ?array $p = null;\n    private function log($v): void {}\n")
			for m := range 20 {
				fmt.Fprintf(&src, "    function m%d(): void {\n", m)
				for i := range 30 {
					fmt.Fprintf(&src, "        "+scenario.stmt+"\n", i)
				}
				src.WriteString("    }\n")
			}
			src.WriteString("}\n")
			file := syntax.Parse("bench.php", []byte(src.String()), syntax.Options{Version: phpversion.PHP84})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(file))
			resolver := names.New(file)
			b.ReportAllocs()
			for b.Loop() {
				env := infer.NewEnv(file, resolver, ix, phpversion.PHP84, infer.Options{})
				syntax.InspectFile(file, func(n syntax.Node) bool {
					if p, ok := n.(*syntax.PropertyFetch); ok {
						env.TypeOf(p)
					}
					return true
				})
			}
		})
	}
}
