package infer

import (
	"testing"

	"custos/internal/phpver"
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
