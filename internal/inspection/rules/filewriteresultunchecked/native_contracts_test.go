package filewriteresultunchecked

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "function bad($path, $payload) { file_put_contents($path, $payload); return true; }", Want: 1},
		{Src: "function good($path, $payload) { return file_put_contents($path, $payload) === strlen($payload); }", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
