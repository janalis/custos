package arrowfunctioncapturedcounter

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$n = 0; $next = fn() => ++$n; echo $next(), $next();", Want: 1},
		{Src: "$n = 0; $next = fn() => ++$n; echo $next();", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
