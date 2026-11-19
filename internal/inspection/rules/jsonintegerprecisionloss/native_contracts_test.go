package jsonintegerprecisionloss

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$data = json_decode('{\"id\":9223372036854775808}', true);", Want: 1},
		{Src: "$data = json_decode('{\"id\":9223372036854775808}', true, 512, JSON_BIGINT_AS_STRING);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
