package interfaceinstantiation

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"InterfaceInstantiation"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"\ninterface Repository {} new Repository();\n", 1},
		{"\ninterface Repository {} class DiskRepository implements Repository {} new DiskRepository();\n", 0},
		{"interface X{} $name='X';new $name();", 1},
		{"new Missing();", 0},
		{"$name=$unknown;new $name();", 0},
		{"trait T{}", 0},
		{"if(false){\ninterface Repository {} new Repository();\n}", 0},
		{"\ninterface Repository {} new Repository();\n $broken = ;", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
