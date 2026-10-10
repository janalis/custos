package shallowclonenestedmutation_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/shallowclonenestedmutation"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$a=new stdClass(); $a->child=new stdClass(); $b=clone $a; $b->child->label='changed';\n", 1},
		{"negative0", "<?php\n$a->p=1;\n", 0},
		{"negative1", "<?php\n$a->child=2;$b=clone $a;$b->child->p=1;\n", 0},
		{"negative2", "<?php\n$a->child=new stdClass();$b=$a;$b->child->p=1;\n", 0},
		{"negative3", "<?php\n$a->child=new stdClass();$b=clone $a;unknown();$b->child->p=1;\n", 0},
		{"negative4", "<?php\nclass A{function __clone(){}}$a=new A();$a->child=new stdClass();$b=clone $a;$b->child->p=1;\n", 0},
		{"negative5", "<?php\n$b->child->p=1;\n", 0},
		{"negative6", "<?php\n$x=1;$a->child=new stdClass();$z=clone $a;$b->child->p=1;\n", 0},
		{"negative7", "<?php\n$a->child=new stdClass();$b=clone $a;echo 1;$b->child->p=1;\n", 0},
		{"negative8", "<?php\n$a->different=new stdClass();$b=clone $a;$b->child->p=1;\n", 0},
		{"negative9", "<?php\nif($q){echo 1;}$b=clone $a;$b->child->p=1;\n", 0},
		{"negative10", "<?php\necho 1;$b=clone $a;$b->child->p=1;\n", 0},
		{"negative11", "<?php\n$a=new stdClass();$b=clone $a;$b->child->p=1;\n", 0},
		{"negative12", "<?php\n$a->child=1;$b=clone $a;$b->child->p=1;\n", 0},
		{"negative13", "<?php\n$a=new stdClass();if($q){echo 1;}$b=clone $a;$b->child->p=1;\n", 0},
		{"negative14", "<?php\n $a=new stdClass();echo 1;$b=clone $a;$b->child->p=1;\n", 0},
		{"negative15", "<?php\n$a=new stdClass();$a->different=new stdClass();$b=clone $a;$b->child->p=1;\n", 0},
		{"negative16", "<?php\n$a=new stdClass();$a->child=1;$b=clone $a;$b->child->p=1;\n", 0},
		{"unrelated expression", "<?php $a=new stdClass();new stdClass();$b=clone $a;$b->child->p=1;", 0},
		{"magic storage", "<?php class P{public function __set($n,$v){} public function __get($n){return new stdClass;}}$o=new P;$o->child=new stdClass;$c=clone $o;$c->child->p=1;", 0},
		{"clone overrides nested state", "<?php class P{public object $child;}$o=new P;$o->child=new stdClass;$c=clone($o,['child'=>new stdClass]);$c->child->p=1;", 0},
		{"unresolved inherited clone", "<?php class P extends Unknown{public object $child;}$o=new P;$o->child=new stdClass;$c=clone $o;$c->child->p=1;", 0},
		{"dynamic child property", "<?php class P{public object $child;}$o=new P;$o->{$name}=new stdClass;$c=clone $o;$c->{$name}->p=1;", 0},
		{"hooked child property", "<?php class P{public object $child{get{return new stdClass;}set{}}}$o=new P;$o->child=new stdClass;$c=clone $o;$c->child->p=1;", 0},
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
