package flow

import (
	"fmt"
	"testing"
)

func BenchmarkFrameBranchWrite(b *testing.B) {
	for _, width := range []int{1, 8, 32, 128} {
		b.Run(fmt.Sprint(width), func(b *testing.B) {
			f := frame{}
			f.resetVars()
			for i := range width {
				f.setVar(fmt.Sprint(i), Value{Complete: true, Identity: uint32(i + 1)})
			}
			changed := Value{Complete: true, Identity: 1000}
			b.ReportAllocs()
			for b.Loop() {
				branch := clone(&f)
				branch.setVar("0", changed)
				if f.vars["0"].Identity != 1 || branch.vars["0"].Identity != 1000 {
					b.Fatal("branch changed original")
				}
			}
		})
	}
}
