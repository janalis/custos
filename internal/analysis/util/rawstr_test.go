package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestQuotedStringRaw(t *testing.T) {
	cases := []struct {
		raw     string
		kind    syntax.LiteralKind
		content string
		quote   byte
		ok      bool
	}{
		{`'a\'b'`, syntax.LitString, `a\'b`, '\'', true},
		{`"x\n"`, syntax.LitString, `x\n`, '"', true},
		{`b'k'`, syntax.LitString, `k`, '\'', true},
		{`''`, syntax.LitString, ``, '\'', true},
		{"<<<'E'\nx\nE", syntax.LitString, "", 0, false},
		{`12`, syntax.LitInt, "", 0, false},
	}
	for _, c := range cases {
		got, q, ok := QuotedStringRaw(&syntax.Literal{LitKind: c.kind, Raw: c.raw})
		if got != c.content || q != c.quote || ok != c.ok {
			t.Errorf("%q: got (%q, %q, %v)", c.raw, got, q, ok)
		}
	}
	if _, _, ok := QuotedStringRaw(&syntax.Variable{Name: "a"}); ok {
		t.Error("variable accepted")
	}
}
