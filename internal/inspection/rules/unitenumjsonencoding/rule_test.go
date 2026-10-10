package unitenumjsonencoding_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/unitenumjsonencoding"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\nenum Mode {case Quick;} echo json_encode(Mode::Quick);\n", 1},
		{"negative0", "<?php\njson_encode($x);\n", 0},
		{"negative1", "<?php\nenum E:string{case A='a';}json_encode(E::A);\n", 0},
		{"negative2", "<?php\nenum E implements JsonSerializable{case A;function jsonSerialize():mixed{return 'a';}}json_encode(E::A);\n", 0},
		{"negative3", "<?php\njson_encode('x');\n", 0},
		{"negative4", "<?php\nenum E{case A;}json_encode(E::A->name);\n", 0},
		{"negative5", "<?php\nstrlen('x');\n", 0},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\nenum Mode {case Quick;} echo json_encode(Mode::Quick);"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
