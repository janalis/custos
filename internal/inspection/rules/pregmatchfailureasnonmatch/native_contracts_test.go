package pregmatchfailureasnonmatch

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "if (!preg_match($pattern, $text)) { return false; }", Want: 1},
		{Src: "if (preg_match($pattern, $text) === false) { throw new RuntimeException(); }", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
