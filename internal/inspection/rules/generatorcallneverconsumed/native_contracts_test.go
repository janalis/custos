package generatorcallneverconsumed

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "function numbers(){yield 1;} numbers();", Want: 0},

		{Src: "function deliver() { yield sendPacket(); } deliver();", Want: 1},
		{Src: "function deliver() { yield sendPacket(); } foreach (deliver() as $v) {}", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
