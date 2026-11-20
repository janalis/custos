package pattern

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
)

func TestCompileErrors(t *testing.T) {
	no := false
	cases := []struct {
		def  Definition
		want string
	}{
		{Definition{Match: Patterns{"f()"}, Message: "m"}, `custom rule: id ""`},
		{Definition{ID: "lower", Match: Patterns{"f()"}, Message: "m"}, "must start with an uppercase letter"},
		{Definition{ID: "StaticClosureCanBeUsed", Match: Patterns{"f()"}, Message: "m"}, "names a built-in rule"},
		{Definition{ID: "R", Match: Patterns{"f()"}}, "message is required"},
		{Definition{ID: "R", Message: "m"}, "match is required"},
		{Definition{ID: "R", Match: Patterns{"f()"}, Message: "m", Severity: "fatal"}, "invalid severity"},
		{Definition{ID: "R", Match: Patterns{"f()"}, Message: "m", PHP: "9.9"}, "phpver"},
		{Definition{ID: "R", Match: Patterns{"f()"}, Message: "m", Paths: []string{"/abs"}}, "relative to the project root"},
		{Definition{ID: "R", Match: Patterns{"f()"}, Message: "m", Exclude: []string{"a/[b"}}, "syntax error"},
		{Definition{ID: "R", Match: Patterns{""}, Message: "m"}, "empty pattern"},
		{Definition{ID: "R", Match: Patterns{"$_; $S;"}, Message: "m"}, "needs a statement that is not a lone metavariable"},
		{Definition{ID: "R", Match: Patterns{";"}, Message: "m"}, "want an expression or statements"},
		{Definition{ID: "R", Match: Patterns{"f()"}, Message: "m", Inside: Patterns{"f(); g();"}}, `inside "f(); g();": want one expression or statement, got 2`},
		{Definition{ID: "R", Match: Patterns{"f()"}, Message: "m", NotInside: Patterns{"f("}}, `notInside "f("`},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Where: map[string]Where{"$X": {Not: Patterns{"a(); b();"}}}}, "got 2 statements"},
		{Definition{ID: "R", Match: Patterns{"f("}, Message: "m"}, "not valid PHP"},
		{Definition{ID: "R", Match: Patterns{"$X"}, Message: "m"}, "lone metavariable"},
		{Definition{ID: "R", Match: Patterns{"f($X)", "g()"}, Message: "use $X"}, `message uses X, which pattern "g()" does not bind`},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Where: map[string]Where{"x": {}}}, "is not a metavariable"},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Where: map[string]Where{"$Y": {}}}, "does not bind"},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Where: map[string]Where{"$X": {Type: ")"}}}, "invalid type"},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Where: map[string]Where{"$X": {Regex: "("}}}, "missing closing"},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Where: map[string]Where{"$X": {Not: Patterns{"g("}}}}, "not \"g(\""},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Fix: fix("g(")}, `fix "g("`},
		{Definition{ID: "R", Match: Patterns{"f($_)"}, Message: "m", Fix: fix("g($_)")}, "$_ binds nothing"},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Fix: fix("g($Y)")}, "fix uses Y"},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Fix: fix("return $X;")}, "statements cannot replace an expression"},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Fix: fix(`\log("value: $X")`)}, "inside a string"},
		{Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Fix: fix(`\log(<<<EOT
$X
EOT)`)}, "inside a string"},
		{Definition{ID: "R", Match: Patterns{"f()"}, Message: "m", Enabled: &no, Exclude: []string{"**/x"}}, ""},
	}
	for _, tc := range cases {
		_, err := Compile(tc.def)
		switch {
		case tc.want == "" && err != nil:
			t.Errorf("%+v: %v", tc.def, err)
		case tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)):
			t.Errorf("%+v: got %v, want %q", tc.def, err, tc.want)
		}
	}
}

func TestFacts(t *testing.T) {
	r, err := Compile(Definition{
		ID: "R", Match: Patterns{"f($X)"}, Message: "m", Severity: "info", PHP: "8.1", Fix: fix("g($X)"),
		Where: map[string]Where{"$X": {Type: `\Countable`}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(analysis.SemanticRule); !ok {
		t.Fatal("a type constraint makes the rule semantic")
	}
	r.(analysis.SemanticRule).Semantic()
	d := r.(analysis.Described)
	m := d.Facts()
	if m.ID != "R" || m.Group != Group || m.Severity != "info" || !m.EnabledByDefault || !m.HasFix {
		t.Fatalf("facts: %+v", m)
	}
	desc := d.Description()
	if desc.Summary != "m" || !desc.Fix || desc.PHPMin != "8.1" {
		t.Fatalf("description: %+v", desc)
	}
	r2, _ := Compile(Definition{ID: "R", Match: Patterns{"f($X)"}, Message: "m", Description: "long"})
	if d := r2.(analysis.Described).Description(); d.Summary != "long" || d.Fix || d.PHPMin != "" {
		t.Fatalf("description: %+v", d)
	}
	f1, f2 := r.(analysis.Fingerprinted).Fingerprint(), r2.(analysis.Fingerprinted).Fingerprint()
	if f1 == f2 || len(f1) != 64 {
		t.Fatalf("fingerprints %q %q", f1, f2)
	}
}

func TestSemanticNames(t *testing.T) {
	for _, tc := range []struct {
		match    string
		not      string
		semantic bool
	}{
		{`\dump()`, "", true},
		{`FOO`, "", true},
		{`$X == null`, "", false},
		{`$X === true`, "", false},
		{`$X->save()`, "", false},
		{`$F()`, "", false},
		{`$X->save()`, `\f()`, true},
		{`$X->a(); $X->b();`, "", false},
		{`$X->a(); \b();`, "", true},
	} {
		d := Definition{ID: "R", Match: Patterns{tc.match}, Message: "m"}
		if tc.not != "" {
			d.Match = Patterns{tc.match}
			d.Where = map[string]Where{"$X": {Not: Patterns{tc.not}}}
		}
		r, err := Compile(d)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := r.(analysis.FileRule); ok != strings.Contains(tc.match, ";") {
			t.Errorf("%s: file rule %t", tc.match, ok)
		}
		if sr, ok := r.(analysis.SemanticRule); ok {
			sr.Semantic()
		}
		if _, ok := r.(analysis.SemanticRule); ok != tc.semantic {
			t.Errorf("%s: semantic %t", tc.match, ok)
		}
	}
}

func TestPatternsJSON(t *testing.T) {
	var d Definition
	if err := json.Unmarshal([]byte(`{"match": "f()", "where": {"$X": {"not": ["a", "b"]}}}`), &d); err != nil || len(d.Match) != 1 || len(d.Where["$X"].Not) != 2 {
		t.Fatalf("%+v %v", d, err)
	}
	if err := json.Unmarshal([]byte(`{"match": 1}`), &d); err == nil || !strings.Contains(err.Error(), "list of patterns") {
		t.Fatalf("got %v", err)
	}
}

func TestSemanticCallees(t *testing.T) {
	for _, tc := range []struct {
		def      Definition
		semantic bool
	}{
		{Definition{Match: Patterns{`$F(...)`}}, false},
		{Definition{Match: Patterns{`$_(...)`}}, false},
		{Definition{Match: Patterns{`$F(...)`}, Where: map[string]Where{"$F": {Regex: `^App`}}}, true},
		{Definition{Match: Patterns{`$F($X)`}, Where: map[string]Where{"$F": {Not: Patterns{`$X`}}}}, true},
		{Definition{Match: Patterns{`$F($X)`}, Where: map[string]Where{"$X": {Regex: `a`}}}, false},
		{Definition{Match: Patterns{`$F($F)`}}, true},
		{Definition{Match: Patterns{`$F()`}, Inside: Patterns{`function f() { $F(); $_; }`}}, true},
	} {
		tc.def.ID, tc.def.Message = "R", "m"
		r, err := Compile(tc.def)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := r.(analysis.SemanticRule); ok != tc.semantic {
			t.Errorf("%v: semantic %t", tc.def, ok)
		}
	}
}
