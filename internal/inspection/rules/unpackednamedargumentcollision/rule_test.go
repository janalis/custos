package unpackednamedargumentcollision_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/unpackednamedargumentcollision"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\nfunction send($id) {} send(...['id'=>7],id:8);\n", 1},
		{"negative0", "<?php\nfunction f($x,$y){} f(...['x'=>1],y:2);\n", 0},
		{"negative1", "<?php\nfunction f($x){}f(...$args,x:2);\n", 0},
		{"negative2", "<?php\nfunction f($x){}f(x:1);\n", 0},
		{"negative3", "<?php\nfunction f($x){}f(...);\n", 0},
		{"negative4", "<?php\nfunction f($x){}f(...[...$args],x:2);\n", 0},
		{"negative5", "<?php\nclass C{static function f($x){}}C::f(...['x'=>1]);\n", 0},
		{"negative6", "<?php\nunknown(...['x'=>1],x:2);\n", 0},
		{"negative7", "<?php\nfunction f(){}f(...[1]);\n", 0},
		{"negative8", "<?php\nfunction f(){}f(1);\n", 0},
		{"negative9", "<?php\nfunction f($x){}f(...[1]);\n", 0},
		{"negative10", "<?php\nfunction f($x){}f(1);\n", 0},
		{"negative11", "<?php\nclass C{function f($x){}}(new C())->f(...['x'=>1]);\n", 0},
		{"extra0", "<?php\nfunction f($x){}f(...['x'=>1],...['x'=>2]);\n", 1},
		{"overwritten named key", "<?php function f($x){}f(...['x'=>1,'x'=>2]);", 0},
		{"overwritten integer key", "<?php function f($x,$y){}f(...[0=>1,0=>2],y:3);", 0},
		{"numeric string key", "<?php function f($x,$y){}f(...['0'=>1,0=>2],y:3);", 0},
		{"implicit index after explicit", "<?php function f($x,$y){}f(...[2=>1,2],y:3);", 1},
		{"unknown key", "<?php function f($x){}f(...[$key=>1],x:2);", 0},
		{"version-dependent negative indices", "<?php function f($x,$y,$z){}f(...[-2=>1,2,-1=>3],z:4);", 0},
		{"positional spread after named spread", "<?php function f($a,$b=0){}f(...['a'=>1],...[2]);", 0},
		{"positional spread after later named slot", "<?php function f($a=0,$b=0,$c=0){}f(...['b'=>1],...[2]);", 0},
		{"positional spread past named last slot", "<?php function f($a=0,$b=0,$c=0){}f(...['c'=>1],...[2,3]);", 0},
		{"positional spread fills following slots", "<?php function f($a=0,$b=0,$c=0){}f(...['a'=>1],...[2,3]);", 0},
		{"named after shifted positional spread", "<?php function f($a=0,$b=0,$c=0){}f(...['b'=>1],...[7],c:3);", 1},
		{"named rebinding earlier named spread", "<?php function f($a=0,$b=0,$c=0){}f(...['a'=>1],...[7],a:3);", 1},
		{"free slot after shifted positional spread", "<?php function f($a=0,$b=0,$c=0){}f(...['b'=>1],...[7],a:3);", 0},
		{"collected name keeps counter", "<?php function f($a=0,...$r){}f(...['q'=>1],...[2],a:3);", 1},
		{"positional after named in one array", "<?php function f($a=0,$b=0){}f(...['a'=>1,5]);", 0},
		{"variadic slot shares no name", "<?php function f(...$r){}f(...[1],r:2);", 0},
		{"explicit variadic slot shares no name", "<?php function f(...$r){}f(1,r:2);", 0},
		{"collected name bound twice", "<?php function f(...$r){}f(...['r'=>1],r:2);", 1},
		{"maximum integer array key", "<?php function f($x){}f(...[9223372036854775807=>1],x:2);", 0},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\nfunction send($id) {} send(...['id'=>7],id:8);"), syntax.Options{}))
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
