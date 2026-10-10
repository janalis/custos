package datelasterrorsfalseunchecked_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/datelasterrorsfalseunchecked"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$errors=DateTimeImmutable::getLastErrors(); echo $errors['warning_count'];\n", 1},
		{"negative0", "<?php\n$e=DateTimeImmutable::getLastErrors();if($e){echo $e['warning_count'];}\n", 0},
		{"negative1", "<?php\n$e=DateTimeImmutable::getLastErrors();if($e!==false){echo $e['warning_count'];}\n", 0},
		{"negative2", "<?php\n$e=DateTimeImmutable::getLastErrors()?:[];echo $e['warning_count'];\n", 0},
		{"negative3", "<?php\necho $x['warning_count'];\n", 0},
		{"negative4", "<?php\n$e=DateTimeImmutable::getLastErrors();if($x!==false){echo 1;}\n", 0},
		{"negative5", "<?php\n$e=DateTimeImmutable::getLastErrors();if(false!==$e){echo $e['warning_count'];}\n", 0},
		{"negative6", "<?php\nfunction f(){$e=DateTimeImmutable::getLastErrors();if($e===false){return;}echo $e['warning_count'];}\n", 0},
		{"extra0", "<?php\n$e=DateTimeImmutable::getLastErrors();if($q){echo $e['warning_count'];}\n", 1},
		{"extra1", "<?php\n$e=DateTimeImmutable::getLastErrors();if($q!==false){echo $e['warning_count'];}\n", 1},
		{"extra2", "<?php\n$e=DateTimeImmutable::getLastErrors();if($q!==$r){echo $e['warning_count'];}\n", 1},
		{"extra3", "<?php\n$e=DateTimeImmutable::getLastErrors();if($e!==false){echo $e['warning_count'];}\n", 0},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\n$errors=DateTimeImmutable::getLastErrors(); echo $errors['warning_count'];"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
