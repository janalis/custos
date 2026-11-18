package callbackrequiredargumentmismatch

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "array_map(fn($value, $key) => $key . $value, $items);", Want: 1},
		{Src: "array_map(fn($value, $key = '') => $key . $value, $items);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
		{Src: "usort($xs, $unknown); array_map($unknown, $xs);", Want: 0},
		{Src: "array_map(fn($a, ...$rest) => $a, $xs);", Want: 0},
		{Src: "array_map(fn($a, $b) => $a, $xs, $ys);", Want: 0},
	})
}
