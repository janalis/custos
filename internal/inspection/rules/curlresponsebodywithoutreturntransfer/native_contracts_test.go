package curlresponsebodywithoutreturntransfer

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$ch = curl_init('https://example.invalid/'); $body = curl_exec($ch); json_decode($body, true);", Want: 1},
		{Src: "$ch = curl_init('https://example.invalid/'); curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); $body = curl_exec($ch); json_decode($body, true);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
