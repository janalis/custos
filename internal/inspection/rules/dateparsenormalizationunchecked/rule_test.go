package dateparsenormalizationunchecked_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/dateparsenormalizationunchecked"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$day=DateTimeImmutable::createFromFormat('!Y-m-d','2025-02-31');\n", 1},
		{"negative0", "<?php\nDateTimeImmutable::createFromFormat('!Y-m-d','2025-02-28');\n", 0},
		{"negative1", "<?php\nDateTimeImmutable::createFromFormat($f,$d);\n", 0},
		{"negative2", "<?php\nDateTimeImmutable::createFromFormat('!Y-m-d','nonsense');\n", 0},
		{"negative3", "<?php\nDateTimeImmutable::createFromFormat('!Y-m-d','2025-02-31');$errors=DateTimeImmutable::getLastErrors();if($errors){reject();}\n", 0},
		{"procedural diagnostics check", "<?php\nfunction parse(){$d=DateTimeImmutable::createFromFormat('Y-m-d','2025-02-31');$e=date_get_last_errors();if($d===false||($e&&($e['warning_count']||$e['error_count']))){return null;}return $d;}\n", 0},
		{"procedural diagnostics before parse", "<?php\n$e=date_get_last_errors();$d=DateTimeImmutable::createFromFormat('Y-m-d','2025-02-31');\n", 1},
		{"procedural diagnostics in other scope", "<?php\nfunction a(){return DateTimeImmutable::createFromFormat('Y-m-d','2025-02-31');}\nfunction b(){return date_get_last_errors();}\n", 1},
		{"shadowed procedural name", "<?php\nnamespace App;\nfunction date_get_last_errors(){return [];}\nfunction parse(){$d=\\DateTimeImmutable::createFromFormat('Y-m-d','2025-02-31');$e=date_get_last_errors();return $e?null:$d;}\n", 1},
		{"negative4", "<?php\nUnknown::createFromFormat('Y-m-d','2025-02-31');\n", 0},
		{"negative5", "<?php\nDateTimeImmutable::createFromFormat('!Y-m-d',$unknown);\n", 0},
		{"invalid non-calendar input", "<?php DateTimeImmutable::createFromFormat('Y-m-d','xxxx-xx-xx');", 0},
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
