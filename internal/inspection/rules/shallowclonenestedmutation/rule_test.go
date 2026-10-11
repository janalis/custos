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
		{"declared open type", "<?php class C{public $v;}class B{public $child;}class S extends B{function __clone(){$this->child=clone $this->child;}}function f(B $a){$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"declared final type", "<?php class C{public $v;}final class B{public $child;}function f(B $a){$a->child=new C;$b=clone $a;$b->child->v=1;}", 1},
		{"constructed open type", "<?php class C{public $v;}class B{public $child;}function f(){$a=new B;$a->child=new C;$b=clone $a;$b->child->v=1;}", 1},
		{"late static instance", "<?php class C{public $v;}class B{public $child;static function f(){$a=new static;$a->child=new C;$b=clone $a;$b->child->v=1;}}", 0},
		{"dynamic class instance", "<?php class C{public $v;}class B{public $child;}function f(){$n='B';$a=new $n;$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"literal stores between", "<?php class C{public $v;}class B{public $child;public $id;}function f(){$a=new B;$a->id=7;$a->child=new C;$b=clone $a;$b->child->v=1;}", 1},
		{"computed store between", "<?php class C{public $v;}class B{public $child;public $id;}function f(){$a=new B;$a->id=g();$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"foreign store between", "<?php class C{public $v;}class B{public $child;}function f($o){$a=new B;$o->id=7;$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"other variable between", "<?php class C{public $v;}class B{public $child;}function f(){$a=new B;$x=7;$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"call between", "<?php class C{public $v;}class B{public $child;}function f(){$a=new B;g($a);$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"branch between", "<?php class C{public $v;}class B{public $child;}function f($q){$a=new B;if($q){$a=h();}$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"reference binding", "<?php class C{public $v;}class B{public $child;}function f(B $o){$a=&$o;$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"coalescing creation", "<?php class C{public $v;}class B{public $child;}function f(B $a){$a??=new B;$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"factory result", "<?php class C{public $v;}class B{public $child;}function h():B{return new B;}function f(){$a=h();$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"property receiver", "<?php class C{public $v;}class B{public $child;}class H{public B $o;function f(){$this->o=new B;$this->o->child=new C;$b=clone $this->o;$b->child->v=1;}}", 0},
		{"statement before store", "<?php class C{public $v;}class B{public $child;}function f(B $a){echo 1;$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
		{"undeclared class", "<?php class C{public $v;}function f(){$a=new Missing;$a->child=new C;$b=clone $a;$b->child->v=1;}", 0},
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
