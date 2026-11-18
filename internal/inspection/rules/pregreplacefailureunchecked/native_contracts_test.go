package pregreplacefailureunchecked

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$text = strtoupper(preg_replace('/z/u', '', $input));", Want: 1},
		{Src: "$text = preg_replace('/z/u', '', $input); if ($text === null) { return; } strtoupper($text);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
