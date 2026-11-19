package shortstreamreadunchecked

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$header = fread($stream, 4); unpack('Nsize', $header);", Want: 1},
		{Src: "$header = fread($stream, 4); if (strlen($header) !== 4) { return; } unpack('Nsize', $header);", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
