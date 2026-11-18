package exhaustedgeneratorreused

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "function values() { yield 1; } $g = values(); iterator_to_array($g); foreach ($g as $v) {}", Want: 1},
		{Src: "function values() { yield 1; } $g = values(); iterator_to_array($g); foreach (values() as $v) {}", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
