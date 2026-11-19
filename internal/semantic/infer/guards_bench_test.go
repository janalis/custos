package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/infer"
	"github.com/janalis/custos/internal/semantic/names"
	"github.com/janalis/custos/internal/semantic/stubs"
)

// BenchmarkPropertyReadGuards types every property read of methods whose
// statements mention a nullable property between reads: plain statements,
// writes into its offsets (assignments, increments, destructuring) and
// calls, the expression statements guards examines for each read; and the
// same writes into a property that cannot be null; and, in the proven
// environment, the writes into a property without a native type; and the
// plain statements in the native environment, which ignores PHPDoc.
func BenchmarkPropertyReadGuards(b *testing.B) {
	for _, scenario := range []struct {
		name, stmt     string
		proven, native bool
	}{
		{"plain", `$this->log($this->p, %[1]d);`, false, false},
		{"elemWrites", `$this->p['k%[1]d'] = %[1]d; $this->p['n']++; [$this->p['a'], $x] = [1, 2]; $this->log($this->p);`, false, false},
		// A property that cannot be null: no write into its offsets matters.
		{"nonNullable", `$this->q['k%[1]d'] = %[1]d; $this->q['n']++; $this->log($this->q);`, false, false},
		// Proven types: a property without a native type may hold false.
		{"untypedProven", `if ($this->u === false) { $this->u['k%[1]d'] = %[1]d; $this->log($this->u); }`, true, false},
		// Native types: the declaration of each read, without its PHPDoc.
		{"native", `$this->log($this->p, %[1]d);`, false, true},
	} {
		b.Run(scenario.name, func(b *testing.B) {
			var src strings.Builder
			src.WriteString("<?php class G {\n    private ?array $p = null;\n    private array $q = [];\n    /** @var array|false */\n    private $u = false;\n    private function log($v): void {}\n")
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
				if scenario.native {
					env = env.Native()
				}
				syntax.InspectFile(file, func(n syntax.Node) bool {
					if p, ok := n.(*syntax.PropertyFetch); ok && scenario.proven {
						env.ProvenTypeOf(p)
					} else if ok {
						env.TypeOf(p)
					}
					return true
				})
			}
		})
	}
}
