package filereadfailureunchecked

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$text = trim(file_get_contents($path));", Want: 1},
		{Src: "$text = file_get_contents($path); if ($text === false) { return; } trim($text);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
