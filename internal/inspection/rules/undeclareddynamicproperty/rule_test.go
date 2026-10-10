package undeclareddynamicproperty_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/undeclareddynamicproperty"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\nclass Person {} $p=new Person(); $p->label='Mia';\n", 1},
		{"negative0", "<?php\n$x->p=1;\n", 0},
		{"negative1", "<?php\nclass P {public $p;} (new P())->p=1;\n", 0},
		{"negative2", "<?php\nclass P {function __set($n,$v){}} (new P())->p=1;\n", 0},
		{"negative3", "<?php\n#[AllowDynamicProperties] class P{} (new P())->p=1;\n", 0},
		{"negative4", "<?php\n(new stdClass())->p=1;\n", 0},
		{"negative5", "<?php\nclass P extends Missing{} (new P())->p=1;\n", 0},
		{"negative6", "<?php\n$a=1;\n", 0},
		{"negative7", "<?php\nclass P{} (new P())->{$name}=1;\n", 0},
		{"negative8", "<?php\n(new Missing())->p=1;\n", 0},
		{"negative9", "<?php\nclass P{}(new P())->{$name}=1;\n", 0},
		{"negative10", "<?php\nclass P extends stdClass{}(new P())->p=1;\n", 0},
		{"nonfinal nominal parameter", "<?php class Base{}class Child extends Base{public $p;}function put(Base $o){$o->p=1;}", 0},
		{"nominal guarded subclass property", "<?php class Base{}class Child extends Base{public $p;}function put(Base $o){if(!property_exists($o,'p')){throw new Exception;}$o->p=1;}", 0},
		{"final parameter missing property", "<?php final class P{} function put(P $o){$o->p=1;}", 1},
		{"final existing guard throw", "<?php final class P{} function put(P $o){if(!property_exists($o,'p')){throw new Exception;}$o->p=1;}", 0},
		{"final existing guard return", "<?php final class P{} function put(P $o){if(!property_exists($o,'p')){return;}$o->p=1;}", 0},
		{"nonfinal unknown local origin", "<?php class P{}$o=getObject();$o->p=1;", 0},
		{"late static allocation", "<?php class P{static function put(){$o=new static;$o->p=1;}}", 0},
		{"nonallocation typed local", "<?php class P{}function make():P{return new P;} $o=make();$o->p=1;", 0},
		{"intervening echo", "<?php class P{}$o=new P;echo 'go';$o->p=1;", 0},
		{"intervening different allocation", "<?php class P{}$o=new P;$other=new P;$o->p=1;", 0},
		{"nonfinal nominal first stmt", "<?php class P{}function put(P $o){$o->p=1;}", 0},
		{"sideeffectful guarded write", "<?php final class P{}function put(P $o){if(!property_exists($o,'p')){return;}$o->p=clear($o);}", 1},
		{"guard has else branch", "<?php final class P{}function put(P $o){if(!property_exists($o,'p')){return;}else{echo 'go';}$o->p=1;}", 1},
		{"guard not negated", "<?php final class P{}function put(P $o){if(property_exists($o,'p')){return;}$o->p=1;}", 1},
		{"guard different receiver", "<?php final class P{}function put(P $o,P $other){if(!property_exists($other,'p')){return;}$o->p=1;}", 1},
		{"guard different property", "<?php final class P{}function put(P $o){if(!property_exists($o,'q')){return;}$o->p=1;}", 1},
		{"guard nonterminal body", "<?php final class P{}function put(P $o){if(!property_exists($o,'p')){echo 'go';}$o->p=1;}", 1},
		{"guard nonsingle body", "<?php final class P{}function put(P $o){if(!property_exists($o,'p')){echo 'go';return;}$o->p=1;}", 1},
		{"guard nonscalar expression body", "<?php final class P{}function put(P $o){if(!property_exists($o,'p')){1+2;}$o->p=1;}", 1},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\nclass Person {} $p=new Person(); $p->label='Mia';"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
