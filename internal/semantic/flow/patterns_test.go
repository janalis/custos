package flow

import "testing"

func TestEmptyless(t *testing.T) {
	for _, tc := range []struct {
		pattern string
		want    bool
	}{
		{`/class\s+(\w+)/`, true},
		{`#\d+#u`, true},
		{`(abc)i`, true},
		{`[a]`, true},
		{`{a}D`, true},
		{`<a>msUS`, true},
		{`/a*/`, false},
		{`/a/x`, false},
		{`/(?=a)b/`, false},
		{``, false},
		{`abc`, false},
		{`9a9`, false},
		{`\a\`, false},
		{` a `, false},
		{`/a`, false},
	} {
		if got := emptyless(tc.pattern); got != tc.want {
			t.Errorf("emptyless(%q) = %v, want %v", tc.pattern, got, tc.want)
		}
	}
}
