package infer_test

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func TestTRulesDivisionIntOrFloat(t *testing.T) {
	src := `<?php
function f(int $a, int $b, float $c) {
    t('div', $a / $b);
    t('mul', $a * $b);
    t('fdiv', $c / $a);
    $q = $a / 2;
    t('var', $q);
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
	for _, c := range []struct {
		div  bool
		want map[string]string
	}{
		{false, map[string]string{"div": "int", "mul": "int", "fdiv": "float", "var": "int"}},
		{true, map[string]string{"div": "float|int", "mul": "int", "fdiv": "float", "var": "float|int"}},
	} {
		tr := infer.NewTRules(env)
		tr.DivisionIntOrFloat = c.div
		got := map[string]string{}
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if call, ok := n.(*syntax.FuncCall); ok {
				if nm, ok := call.Name.(*syntax.Name); ok && nm.Value == "t" {
					label := strings.Trim(string(f.Src[call.Args.Args[0].Span().Start:call.Args.Args[0].Span().End]), "'")
					got[label] = strings.Join(tr.TypeOf(call.Args.Args[1].(*syntax.Arg).Value).Atoms(), "|")
				}
			}
			return true
		})
		for k, w := range c.want {
			if got[k] != w {
				t.Errorf("division=%v %s: got %q want %q", c.div, k, got[k], w)
			}
		}
	}
}
