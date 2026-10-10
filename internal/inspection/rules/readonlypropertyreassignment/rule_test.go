package readonlypropertyreassignment_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/readonlypropertyreassignment"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\nclass Item { public readonly int $id; public function __construct() {$this->id=1; $this->id=2;} }\n", 1},
		{"negative0", "<?php\nclass P {public int $p;function f(){$this->p=1;$this->p=2;}}\n", 0},
		{"negative1", "<?php\nclass P{public readonly int $p;function f(){$this->p=1;}}\n", 0},
		{"negative2", "<?php\nclass P{public readonly int $p;function f(){unknown();$this->p=2;}}\n", 0},
		{"negative3", "<?php\nclass P{public readonly int $p;function __clone(){$this->p=1;$this->p=2;}}\n", 0},
		{"negative4", "<?php\n$a=1;\n", 0},
		{"negative5", "<?php\n$x->p=1;\n", 0},
		{"negative6", "<?php\nclass P{public readonly int $p;function f(){if($q){$this->p=1;}$this->p=2;}}\n", 0},
		{"negative7", "<?php\nclass P{public readonly int $p;function f(){$this->{$name}=1;}}\n", 0},
		{"negative8", "<?php\nclass P{public readonly int $p;function f(){$this->p=&$x;$this->p=2;}}\n", 0},
		{"negative9", "<?php\nclass P{public readonly int $p;function f(){for($this->p=1;;){}}}\n", 0},
		{"negative10", "<?php\nclass P{public readonly int $p;function f(){if($q)$this->p=1;}}\n", 0},
		{"receiver replaced", "<?php class P {public readonly int $p; static function run() {$o=new P; $o->p=1; $o=new P; $o->p=2;}}", 0},
		{"unknown RHS invalidation", "<?php class P{public readonly int $p;function f(){$this->p=resetProperty($this);$this->p=2;}}", 0},
		{"non-variable receiver", "<?php class P{public readonly int $p;function f(){make()->p=1;make()->p=2;}}", 0},
		{"intervening output", "<?php class P{public readonly int $p;function f(){$this->p=1;echo 'ok';$this->p=2;}}", 0},
		{"different adjacent assignment", "<?php class P{public readonly int $p;function f(){$x=1;$this->p=2;}}", 0},
		{"repeat parameter write", "<?php class P{public readonly int $p;function f($n){$this->p=$n;$this->p=2;}}", 1},
		{"computed receiver", "<?php class P{public readonly int $p;static function make():self{return new self;}function f(){self::make()->p=1;self::make()->p=2;}}", 0},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\nclass Item { public readonly int $id; public function __construct() {$this->id=1; $this->id=2;} }"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
