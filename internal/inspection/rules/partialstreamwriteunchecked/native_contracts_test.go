package partialstreamwriteunchecked

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "function bad($socket, $payload) { fwrite($socket, $payload); return true; }", Want: 1},
		{Src: "function good($socket, $payload) { return fwrite($socket, $payload) === strlen($payload); }", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
