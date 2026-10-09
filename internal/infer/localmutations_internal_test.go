package infer

import (
	"strings"
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestTRulesLocalMutationModes(t *testing.T) {
	source := `<?php
function modes() {
    $unset = 's'; unset($unset); t('unset', $unset);
    $increment = null; ++$increment; t('increment', $increment);
}
`
	checkLabels(t, trulesLabels(t, source, phpver.PHP84, nil), map[string]string{"unset": "", "increment": ""})
	checkLabels(t, trulesLabels(t, source, phpver.PHP84, func(r *TRules) { r.SpecOnly = true }), map[string]string{"unset": "string", "increment": "null"})
	checkLabels(t, trulesLabels(t, source, phpver.PHP84, func(r *TRules) { r.Env = r.Env.Native() }), map[string]string{"unset": "", "increment": ""})
}

func TestArrayElementMutationModes(t *testing.T) {
	source := `<?php
function modes() {
    $a = ['x' => null]; ++$a['x']; t('increment', $a['x']);
    $b = ['x' => null]; --$b['x']; t('decrement', $b['x']);
}
`
	checkLabels(t, trulesLabels(t, source, phpver.PHP84, nil), map[string]string{"increment": "int|null", "decrement": "null"})
	checkLabels(t, trulesLabels(t, source, phpver.PHP84, func(r *TRules) { r.SpecOnly = true }), map[string]string{"increment": "int|null", "decrement": "null"})
	got := trulesLabels(t, source, phpver.PHP84, func(r *TRules) {
		r.Env = r.Env.Native()
		syntax.InspectFile(r.Env.File, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				value := c.Args.Args[1].(*syntax.Arg).Value
				label := c.Args.Args[0].Span()
				want := "int|null"
				if strings.Trim(string(r.Env.File.Src[label.Start:label.End]), "'") == "decrement" {
					want = "null"
				}
				if got := r.Env.TypeOf(value).String(); got != want {
					t.Errorf("native element: got %s want %s", got, want)
				}
			}
			return true
		})
	})
	checkLabels(t, got, map[string]string{"increment": "int|null", "decrement": "null"})
}
