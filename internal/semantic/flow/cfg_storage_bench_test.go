package flow

import (
	"strings"
	"testing"
)

func BenchmarkControlFlowGraphStorage(b *testing.B) {
	for _, tc := range []struct{ name, source string }{
		{"linear", strings.Repeat(`$x=1;`, 64)},
		{"branches", strings.Repeat(`if($flag){$x=1;}else{$x=2;} while($flag){$x++;if($x){break;}}`, 16)},
		{"exceptions", strings.Repeat(`try{$x=1;}catch(RuntimeException $e){$x=2;}catch(LogicException $e){$x=3;}finally{$x++;}`, 16)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			e, file := environment(b, tc.source, nil)
			b.ReportAllocs()
			for b.Loop() {
				g := graph{complete: true, env: e}
				exit := g.add(nil)
				done := g.add(nil, exit)
				entry := g.list(file.Stmts, done, exits{ret: done, thrown: exit, brk: done, cont: done})
				if !g.complete || entry < 2 || len(g.blocks) < 3 {
					b.Fatal("incomplete graph")
				}
			}
		})
	}
}
