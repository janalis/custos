package nonexhaustiveenummatch_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/nonexhaustiveenummatch"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\nenum Choice {case Left;case Right;} function label(Choice $c) {return match($c) {Choice::Left=>'L'};}\n", 1},
		{"negative0", "<?php\nenum E{case A;case B;}function f(E $e){return match($e){E::A=>1,E::B=>2};}\n", 0},
		{"negative1", "<?php\nenum E{case A;case B;}function f(E $e){return match($e){E::A=>1,default=>2};}\n", 0},
		{"negative2", "<?php\nmatch($x){1=>1};\n", 0},
		{"negative3", "<?php\nclass E{}function f(E $e){return match($e){1=>1};}\n", 0},
		{"negative4", "<?php\nenum E{case A;case B;}function f(E $e){return match($e){$x=>1};}\n", 0},
		{"negative5", "<?php\nenum E{case A;case B;}class C{const A=1;}function f(E $e){return match($e){C::A=>1};}\n", 0},
		{"known reachable case", "<?php enum E{case A;case B;} $e=E::A;echo match($e){E::A=>1};", 0},
		{"known missing case", "<?php enum E{case A;case B;} $e=E::B;echo match($e){E::A=>1};", 1},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\nenum Choice {case Left;case Right;} function label(Choice $c) {return match($c) {Choice::Left=>'L'};}"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
