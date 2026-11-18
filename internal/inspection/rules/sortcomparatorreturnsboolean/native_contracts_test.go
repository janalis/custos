package sortcomparatorreturnsboolean

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "usort($xs, function($a,$b){ if($a){return $a > $b;} });", Want: 0},
		{Src: "usort($xs, fn($a, $b) => $a > $b);", Want: 1},
		{Src: "usort($xs, fn($a, $b) => $a <=> $b);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
		{Src: "usort($xs, $unknown); array_map($unknown, $xs);", Want: 0},
		{Src: "usort($xs, fn($a, $b) => $a >= $b);", Want: 1},
		{Src: "usort($xs, function ($a, $b) { return; });", Want: 0},
	})
}
