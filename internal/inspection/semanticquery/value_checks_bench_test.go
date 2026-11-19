package semanticquery

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

const valueChecksBenchSrc = `<?php
declare(strict_types=1);
namespace App;
final class Point {}
/**
 * @param int $id
 * @param string $names
 * @param int[] $counts
 */
function probe($id, $names, $counts, Point $point, int $n) {
    $id->run();
    foreach ($names as $name) {
        echo $name;
    }
    return $counts * $n + $point;
}`

// benchValueChecks runs check on valueChecksBenchSrc with its member
// access, foreach and arithmetic nodes.
func benchValueChecks(b *testing.B, check func(*analysis.Context, *syntax.MethodCall, *syntax.Foreach, []*syntax.Binary)) {
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, _ *syntax.FuncCall) {
		if ran {
			return
		}
		ran = true
		var access *syntax.MethodCall
		var loop *syntax.Foreach
		var ops []*syntax.Binary
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.MethodCall:
				access = n
			case *syntax.Foreach:
				loop = n
			case *syntax.Binary:
				ops = append(ops, n)
			}
			return true
		})
		check(ctx, access, loop, ops)
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("bench.php", []byte(valueChecksBenchSrc+"\nprobe(1, '', [], new Point(), 1);"), syntax.Options{}))
	if !ran {
		b.Fatal("probe not reached")
	}
}

// BenchmarkValueChecks judges one member access, one foreach and two
// arithmetic operations in the proven view, as the default-level rules do,
// and in the documented view.
func BenchmarkValueChecks(b *testing.B) {
	for _, documented := range []bool{false, true} {
		name := "proven"
		if documented {
			name = "documented"
		}
		b.Run(name, func(b *testing.B) {
			benchValueChecks(b, func(ctx *analysis.Context, access *syntax.MethodCall, loop *syntax.Foreach, ops []*syntax.Binary) {
				_, _, okAccess := NonObjectReceiver(ctx, access, documented)
				_, okLoop := NonIterableSubject(ctx, loop, documented)
				if okAccess != documented || okLoop != documented || len(ops) != 2 {
					b.Fatalf("access %v, loop %v, %d operations", okAccess, okLoop, len(ops))
				}
				b.ReportAllocs()
				for b.Loop() {
					NonObjectReceiver(ctx, access, documented)
					NonIterableSubject(ctx, loop, documented)
					for _, op := range ops {
						invalidOperands(ctx, op, documented)
					}
				}
			})
		})
	}
}
