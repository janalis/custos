package sortcomparatorreturnsfloat

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "usort($xs, fn(float $a, float $b) => $a - $b);", Want: 1},
		{Src: "usort($xs, fn(int $a, int $b) => $a - $b);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
		{Src: "usort($xs, $unknown); array_map($unknown, $xs);", Want: 0},
	})
}
