package sortresultusedasarray

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$sorted = sort($items); foreach ($sorted as $item) {}", Want: 1},
		{Src: "$ok = sort($items); if ($ok) { foreach ($items as $item) {} }", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
