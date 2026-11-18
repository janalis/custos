package curltransportsuccessashttpsuccess

import (
	"testing"

	"custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "function bad($url) { $ch = curl_init('https://example.invalid/api'); if (curl_exec($ch) !== false) { return true; } return false; }", Want: 1},
		{Src: "function good($url) { $ch = curl_init('https://example.invalid/api'); curl_setopt($ch, CURLOPT_FAILONERROR, true); return curl_exec($ch) !== false; }", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
