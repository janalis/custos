package unknownnamedargument_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/unknownnamedargument"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\nfunction pause($timeout) {} pause(timeot:3);\n", 1},
		{"negative0", "<?php\nfunction f($x){}f(x:1);\n", 0},
		{"negative1", "<?php\nfunction f(...$args){}f(anything:1);\n", 0},
		{"negative2", "<?php\nunknown(x:1);\n", 0},
		{"negative3", "<?php\nclass C{function f($x){}} (new C())->f(x:1);\n", 0},
		{"negative4", "<?php\nclass C{static function f($x){}} C::f(x:1);\n", 0},
		{"negative5", "<?php\nstrlen(string:'x');\n", 0},
		{"negative6", "<?php\nstrlen(...);\n", 0},
		{"negative7", "<?php\nfunction f(){}f();\n", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rule.New().(analysis.SemanticRule).Semantic()
			e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("test.php", []byte(tc.source), syntax.Options{})
			findings := e.Analyze(f)
			if len(findings) != tc.count {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.count, findings)
			}
		})
	}
}

func TestLegacyVersion(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{PHP: phpversion.PHP53, Only: []string{rule.New().ID()}})
	if err != nil {
		t.Fatal(err)
	}
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\nfunction pause($timeout) {} pause(timeot:3);"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}

func TestIncompleteArguments(t *testing.T) {
	file := syntax.Parse("test.php", []byte("<?php function f($x){}f();"), syntax.Options{})
	syntax.InspectFile(file, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			c.Args = nil
		}
		return true
	})
	e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
	if err != nil {
		t.Fatal(err)
	}
	if findings := e.Analyze(file); len(findings) != 0 {
		t.Fatal(findings)
	}
}
