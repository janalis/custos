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

// TestStringOffsetWrite: a write into a string offset changes the value
// (and may lengthen it) without assigning the local, so the literal it held
// is dropped while its type stays a string.
func TestStringOffsetWrite(t *testing.T) {
	cases := []struct {
		name, body, want string // want: type, then its string values
	}{
		{"replaced character", `$s = 'abc'; $s[0] = 'x'; t('v', $s);`, `string`},
		{"appended character", `$s = 'abc'; $s[3] = 'd'; t('v', $s);`, `string`},
		{"in a branch", `$s = 'abc'; if (g()) { $s[1] = 'x'; } t('v', $s);`, `string`},
		// A write into null makes an array, which this path does not model.
		{"nullable", `$s = g() ? 'abc' : null; $s[0] = 'x'; t('v', $s);`, `null|string`},
		// The literal stands when no write reaches.
		{"no write", `$s = 'abc'; t('v', $s);`, `string abc`},
		{"write before the assignment", `$s = 'q'; $s[0] = 'x'; $s = 'abc'; t('v', $s);`, `string abc`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := `<?php function g(): bool { return true; } function t($label, $value) {} function f() { ` + c.body + ` }`
			f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if call, ok := n.(*syntax.FuncCall); ok && call.Args != nil && len(call.Args.Args) == 2 {
					x := call.Args.Args[1].(*syntax.Arg).Value
					for _, e := range []*infer.Env{env, env.Proven()} {
						ty := e.TypeOf(x)
						got := ty.String()
						if vals, ok := ty.StringValueList(); ok {
							got += " " + strings.Join(vals, ",")
						}
						if got != c.want {
							t.Errorf("got %q, want %q", got, c.want)
						}
					}
				}
				return true
			})
		})
	}
}
