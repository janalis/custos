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
