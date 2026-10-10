package uninitializedtypedpropertyread_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/uninitializedtypedpropertyread"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\nclass Meter {public int $value;} echo (new Meter())->value;\n", 1},
		{"negative0", "<?php\nclass P{public int $p=1;} echo (new P())->p;\n", 0},
		{"negative1", "<?php\nclass P{public int $p;function __construct(){$this->p=1;}} echo (new P())->p;\n", 0},
		{"negative2", "<?php\nclass P{public int $p;} isset((new P())->p);\n", 0},
		{"negative3", "<?php\nclass P{public int $p;} empty((new P())->p);\n", 0},
		{"negative4", "<?php\nclass P{public int $p;} (new P())->p=1;\n", 0},
		{"negative5", "<?php\necho $unknown->p;\n", 0},
		{"negative6", "<?php\nclass P{public $p;}echo (new P())->p;\n", 0},
		{"negative7", "<?php\nclass P{public int $p;}echo (new P())->{$name};\n", 0},
		{"negative8", "<?php\nclass P{public int $p;}$p=new P();$p->p=1;echo $p->p;\n", 0},
		{"negative9", "<?php\nclass P{public int $p;}$p=new P();echo $p->p;\n", 0},
		{"coalescing safe read", "<?php class P{public int $p;}echo (new P)->p ?? 1;", 0},
		{"coalescing fallback read", "<?php class P{public int $p;}echo null ?? (new P)->p;", 1},
		{"unknown inherited constructor", "<?php class P extends Unknown{public int $p;}echo (new P)->p;", 0},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\nclass Meter {public int $value;} echo (new Meter())->value;"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
