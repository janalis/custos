package usortdiscardsrequiredkeys

import (
	"testing"

	"github.com/janalis/custos/internal/testing/ruletest"
)

func TestNativeContracts(t *testing.T) {
	ruletest.Count(t, New(), message, []ruletest.Case{
		{Src: "$users = [31 => 'Mia', 52 => 'Noor']; usort($users, fn($a, $b) => $a <=> $b); echo $users[31];", Want: 1},
		{Src: "$users = ['Mia', 'Noor']; usort($users, fn($a, $b) => $a <=> $b); echo $users[0];", Want: 0},
		{Src: "namespace Custom; function unrelated($x) {} unrelated(1);", Want: 0},
	})
}
