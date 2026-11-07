package flow

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkPureBindingStorage(b *testing.B) {
	for _, count := range []int{1, 8, 9, 64} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			var params, args strings.Builder
			for i := range count {
				if i != 0 {
					params.WriteByte(',')
					args.WriteByte(',')
				}
				fmt.Fprintf(&params, "$p%d", i)
				fmt.Fprintf(&args, "%d", i)
			}
			source := fmt.Sprintf(`function last(%s){return $p%d;}probe(last(%s));`, params.String(), count-1, args.String())
			b.ReportAllocs()
			for b.Loop() {
				e, f := summarized(b, source)
				e.Value(argument(callsNamed(f, "probe")[0]))
			}
		})
	}
}
