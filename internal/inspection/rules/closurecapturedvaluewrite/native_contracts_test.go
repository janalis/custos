package closurecapturedvaluewrite

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$n = 0; $f = function () use ($n) { ++$n; }; $f(); echo $n;", Want: 1},
		{Src: "$n = 0; $f = function () use (&$n) { ++$n; }; $f(); echo $n;", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
