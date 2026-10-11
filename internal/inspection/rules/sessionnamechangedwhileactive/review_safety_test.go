package sessionnamechangedwhileactive

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestNullArgumentGetter(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		php       phpversion.Version
		want      int
	}{
		{"no argument reads the name", "session_start(); session_name();", phpversion.PHP80, 0},
		{"null reads the name", "session_start(); session_name(null);", phpversion.PHP80, 0},
		{"named null reads the name", "session_start(); session_name(name: null);", phpversion.PHP85, 0},
		{"qualified uppercase null", `\session_start(); \session_name(\NULL);`, phpversion.PHP80, 0},
		{"variable proven null", "session_start(); $n = null; session_name($n);", phpversion.PHP80, 0},
		{"null is coerced before 8.0", "session_start(); session_name(null);", phpversion.PHP74, 1},
		{"empty string is a setter", `session_start(); session_name("");`, phpversion.PHP80, 1},
		{"unknown value is a setter", "session_start(); session_name($name);", phpversion.PHP80, 1},
		{"variable proven string", `session_start(); $n = "x"; session_name($n);`, phpversion.PHP80, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, Only: []string{"SessionNameChangedWhileActive"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
