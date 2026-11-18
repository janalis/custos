package iteratormaterializationkeycollision

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "function rows($flag){if($flag){yield 'k'=>1;}else{yield 'k'=>2;}} iterator_to_array(rows($flag));", Want: 0},

		{Src: "function rows() { yield 'k' => 2; yield 'k' => 5; } $a = iterator_to_array(rows());", Want: 1},
		{Src: "function rows() { yield 'k' => 2; yield 'k' => 5; } $a = iterator_to_array(rows(), false);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
