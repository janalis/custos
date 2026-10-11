package reflectioncompositetypeassumednamed

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/testing/conformance"
)

func TestContracts(t *testing.T) {
	r := New()
	if semantic, ok := r.(analysis.SemanticRule); ok {
		semantic.Semantic()
	}
	if flow, ok := r.(analysis.FlowRule); ok {
		flow.Flow()
	}
	fixtures, err := filepath.Glob("../../../../testdata/rules/ReflectionCompositeTypeAssumedNamed/*.php")
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("fixtures: %v", err)
	}
	for _, path := range fixtures {
		if strings.HasSuffix(path, ".fixed.php") {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			src, want, err := conformance.ParseMarkup(raw)
			if err != nil {
				t.Fatal(err)
			}
			cfg := analysis.Config{Only: []string{"ReflectionCompositeTypeAssumedNamed"}}
			var side struct {
				PHP     string         `json:"php"`
				Options map[string]any `json:"options"`
			}
			settings, err := os.ReadFile(strings.TrimSuffix(path, ".php") + ".json")
			if err == nil {
				if err = json.Unmarshal(settings, &side); err != nil {
					t.Fatal(err)
				}
				if side.PHP != "" {
					cfg.PHP = phpversion.MustParse(side.PHP)
				}
				cfg.Rules = map[string]analysis.RuleConfig{"ReflectionCompositeTypeAssumedNamed": {Options: side.Options}}
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
			e, err := analysis.NewEngine([]analysis.Rule{r}, cfg)
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse(path, src, syntax.Options{})
			got := e.Analyze(f)
			if len(got) != len(want) {
				t.Fatalf("got %d want %d: %+v", len(got), len(want), got)
			}
			var edits []diagnostic.TextEdit
			for i, g := range got {
				w := want[i]
				if int(g.Span.Start) != w.Start || int(g.Span.End) != w.End || g.Message != w.Message || g.Severity != w.Severity {
					t.Fatalf("got %+v want %+v", g, w)
				}
				if len(g.Fixes) > 0 {
					edits = append(edits, g.Fixes[0].Edits()...)
				}
			}
			expected, err := os.ReadFile(strings.TrimSuffix(path, ".php") + ".fixed.php")
			if os.IsNotExist(err) {
				if len(edits) > 0 {
					t.Fatal("fix without expected output")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			fixed, _ := fixing.Apply(src, edits)
			if !bytes.Equal(fixed, expected) {
				t.Fatalf("fixed:\n%s\nwant:\n%s", fixed, expected)
			}
			parsed := syntax.Parse(path, fixed, syntax.Options{})
			if len(parsed.Errors) > 0 {
				t.Fatalf("invalid fix: %+v", parsed.Errors)
			}
			if again := e.Analyze(parsed); len(again) > 0 {
				t.Fatalf("fix not stable: %+v", again)
			}
		})
	}
}

func TestUnknownAndShadowed(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ReflectionCompositeTypeAssumedNamed"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("edge.php", []byte(`<?php namespace Local; class SplQueue { function dequeue() {} } class SplFixedArray { function __construct($n) {} } class WeakMap {} class ReflectionFunction { function __construct($s) {} function invokeArgs($a) {} } class Attribute {} class Override {} function fseek($s,$n){} function readdir($s){} function fputcsv($s,$r){} function unknown($v) {} function test($x,$fp) { $x->dequeue(); $x->extract(); $x->top(); $x->getName(); $x->invokeArgs([]); $x->setSize(0); $x->fgetcsv(); $x->fputcsv([]); $x->setCsvControl(); $x[9]=1; $x[$fp]=[]; $x[$fp][0]++; foreach($x as $v){} if(fseek($fp,0)){} while(readdir($fp)){} fputcsv($fp,[]); echo unknown($x); } new SplQueue(); new SplFixedArray(1); new ReflectionFunction('test'); #[Attribute] class Annotated{} #[Override] function method() {}`), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("unknown/shadowed: %+v", got)
	}
}

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function f():int|string{}$r=new ReflectionFunction('f');$r->getReturnType()->getName();", 1},
		{"class A{public int|string $x;}$r=new ReflectionProperty('A','x');$r->getType()->getName();", 1},
		{"function f(int $x){}$r=new ReflectionFunction('f');$r->getParameters()[0]->getType()->getName();", 0},
		{"function f(int|string $x){}$r=new ReflectionFunction('f');$r->getParameters()[5]->getType()->getName();", 0},
		{"$unknown->unknown()->getName();", 0},
		{"function f(int|string $x){}$r=new ReflectionFunction('f');$t=$r->getParameters()[0]->getType();if($t instanceof ReflectionNamedType){$t->getName();}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ReflectionCompositeTypeAssumedNamed"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edge.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func TestVersionAndOtherEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
		php  string
	}{
		{"function f(int|string $x){}$r=new ReflectionFunction('f');$r->getParameters()[0]->getType()->getName();", 0, "7.4"},
		{"$x->$method()->getName();", 0, "8.4"},
		{"$p=$unknown[0];$p->getType()->getName();", 0, "8.4"},
		{"$p=$unknown->items()[0];$p->getType()->getName();", 0, "8.4"},
		{"$p=new ReflectionProperty($unknown,'x');$p->getType()->getName();", 0, "8.4"},
	} {
		t.Run(tc.src+tc.php, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.MustParse(tc.php), Only: []string{"ReflectionCompositeTypeAssumedNamed"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edge.php", []byte("<?php "+tc.src), syntax.Options{})
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func TestReviewSafety(t *testing.T) {
	const fn = "function f(int|string $x):int|string{}function g(A&B $x){}function n(?int $a,int|null $b,null|string $c,int|string|null $d,(A&B)|null $e):null|false{}"
	param := func(f, i string) string {
		return "$r=new ReflectionFunction('" + f + "');$t=$r->getParameters()[" + i + "]->getType();"
	}
	u := param("f", "0")
	for _, tc := range []struct {
		name string
		src  string
		want int
		php  string
	}{
		{"nullable shorthand", param("n", "0") + "$t->getName();", 0, "8.4"},
		{"trailing null", param("n", "1") + "$t->getName();", 0, "8.4"},
		{"leading null", param("n", "2") + "$t->getName();", 0, "8.4"},
		{"nullable union", param("n", "3") + "$t->getName();", 1, "8.4"},
		{"nullable intersection group", param("n", "4") + "$t->getName();", 1, "8.4"},
		{"null or false return", "$r=new ReflectionFunction('n');$r->getReturnType()->getName();", 0, "8.4"},
		{"intersection", param("g", "0") + "$t->getName();", 1, "8.4"},
		{"intersection before 8.1", param("g", "0") + "$t->getName();", 0, "8.0"},
		{"negated guard leaves", u + "if(!$t instanceof ReflectionNamedType){return;}$t->getName();", 0, "8.4"},
		{"negated guard throws", u + "if(!($t instanceof \\ReflectionNamedType)){throw new Exception();}$t->getName();", 0, "8.4"},
		{"negated guard falls through", u + "if(!$t instanceof ReflectionNamedType){echo 1;}$t->getName();", 1, "8.4"},
		{"guard with elseif", u + "if(!$t instanceof ReflectionNamedType){return;}elseif($x){}$t->getName();", 0, "8.4"},
		{"union leaves before elseif", u + "if($t instanceof ReflectionUnionType){return;}elseif($x){return;}$t->getName();", 0, "8.4"},
		{"else leaves with elseif", u + "if($t instanceof ReflectionNamedType){echo 1;}elseif($x){}else{return;}$t->getName();", 1, "8.4"},
		{"elseif guard leaves", u + "if($x){}elseif(!$t instanceof ReflectionNamedType){return;}$t->getName();", 1, "8.4"},
		{"else leaves", u + "if($t instanceof ReflectionNamedType){echo 1;}else{return;}$t->getName();", 0, "8.4"},
		{"else leaves wrong test", u + "if($t instanceof ReflectionUnionType){echo 1;}else{return;}$t->getName();", 1, "8.4"},
		{"guard after call", u + "$t->getName();if(!$t instanceof ReflectionNamedType){return;}", 1, "8.4"},
		{"guard on other value", u + "if(!$o instanceof ReflectionNamedType){return;}$t->getName();", 1, "8.4"},
		{"reassigned after guard", u + "if(!$t instanceof ReflectionNamedType){return;}$t=$r->getParameters()[0]->getType();$t->getName();", 1, "8.4"},
		{"dynamic class", u + "if(!$t instanceof $c){return;}$t->getName();", 1, "8.4"},
		{"union leaves", u + "if($t instanceof ReflectionUnionType){return;}$t->getName();", 0, "8.4"},
		{"intersection leaves on union", u + "if($t instanceof ReflectionIntersectionType){return;}$t->getName();", 1, "8.4"},
		{"intersection leaves", param("g", "0") + "if($t instanceof ReflectionIntersectionType){return;}$t->getName();", 0, "8.4"},
		{"union leaves on intersection", param("g", "0") + "if($t instanceof ReflectionUnionType){return;}$t->getName();", 1, "8.4"},
		{"either composite leaves", u + "if($t instanceof ReflectionUnionType||$t instanceof ReflectionIntersectionType){return;}$t->getName();", 0, "8.4"},
		{"both required to leave", u + "if($t instanceof ReflectionUnionType&&$x){return;}$t->getName();", 1, "8.4"},
		{"unrelated comparison", u + "if($t===null){return;}$t->getName();", 1, "8.4"},
		{"unrelated negation", u + "if(-$x){return;}$t->getName();", 1, "8.4"},
		{"assert", u + "assert($t instanceof ReflectionNamedType);$t->getName();", 0, "8.4"},
		{"assert unrelated", u + "assert($x);other($t instanceof ReflectionNamedType);$t->getName();", 1, "8.4"},
		{"guard in enclosing block", u + "if(!$t instanceof ReflectionNamedType){return;}if($x){foreach($y as $z){$t->getName();}}", 0, "8.4"},
		{"guard in switch case", u + "switch($x){case 1:if(!$t instanceof ReflectionNamedType){break;}$t->getName();}", 0, "8.4"},
		{"braced namespace", "namespace N{function f(int|string $x){}$r=new \\ReflectionFunction('N\\f');$t=$r->getParameters()[0]->getType();if(!$t instanceof \\ReflectionNamedType){return;}$t->getName();}", 0, "8.4"},
		{"if body", u + "if($t instanceof ReflectionNamedType){$t->getName();}", 0, "8.4"},
		{"if condition", u + "if($t->getName()){}", 1, "8.4"},
		{"else of named", u + "if($t instanceof ReflectionNamedType){}else{$t->getName();}", 1, "8.4"},
		{"else of negated named", u + "if(!$t instanceof ReflectionNamedType){}else{$t->getName();}", 0, "8.4"},
		{"elseif body", u + "if($x){}elseif($t instanceof ReflectionNamedType){$t->getName();}", 0, "8.4"},
		{"elseif condition", u + "if($x){}elseif($t->getName()){}", 1, "8.4"},
		{"elseif after union", u + "if($t instanceof ReflectionUnionType){}elseif($x){$t->getName();}", 0, "8.4"},
		{"else after elseif union", u + "if($x){}elseif($t instanceof ReflectionUnionType){}else{$t->getName();}", 0, "8.4"},
		{"elseif before union", u + "if($x){}elseif($y){$t->getName();}elseif($t instanceof ReflectionUnionType){}", 1, "8.4"},
		{"ternary then", u + "echo $t instanceof ReflectionNamedType?$t->getName():'';", 0, "8.4"},
		{"ternary else", u + "echo $t instanceof ReflectionNamedType?'':$t->getName();", 1, "8.4"},
		{"ternary else of union", u + "echo $t instanceof ReflectionUnionType?'':$t->getName();", 0, "8.4"},
		{"ternary condition", u + "echo $t->getName()?1:2;", 1, "8.4"},
		{"short ternary", u + "echo $t instanceof ReflectionUnionType?:$t->getName();", 0, "8.4"},
		{"and operand", u + "echo $t instanceof ReflectionNamedType&&$t->getName();", 0, "8.4"},
		{"and keyword", u + "$t instanceof ReflectionNamedType and print $t->getName();", 0, "8.4"},
		{"or operand", u + "echo $t instanceof ReflectionUnionType||$t->getName();", 0, "8.4"},
		{"or keyword", u + "$t instanceof ReflectionUnionType or print $t->getName();", 0, "8.4"},
		{"or operand of named", u + "echo $t instanceof ReflectionNamedType||$t->getName();", 1, "8.4"},
		{"and left operand", u + "echo $t->getName()&&$t instanceof ReflectionNamedType;", 1, "8.4"},
		{"other operator", u + "echo $t instanceof ReflectionNamedType.$t->getName();", 1, "8.4"},
		{"match arm", u + "echo match(true){$t instanceof ReflectionNamedType=>$t->getName(),default=>''};", 0, "8.4"},
		{"match default", u + "echo match(true){$t instanceof ReflectionNamedType=>'',default=>$t->getName()};", 1, "8.4"},
		{"match mixed conditions", u + "echo match(true){$t instanceof ReflectionNamedType,$x=>$t->getName(),default=>''};", 1, "8.4"},
		{"match other subject", u + "echo match($x){$t instanceof ReflectionNamedType=>$t->getName(),default=>''};", 1, "8.4"},
		{"match false subject", u + "echo match(false){$t instanceof ReflectionNamedType=>$t->getName(),default=>''};", 1, "8.4"},
		{"match arm condition", u + "echo match(true){$t->getName()=>1,default=>2};", 1, "8.4"},
		{"guard outside closure", u + "if(!$t instanceof ReflectionNamedType){return;}$f=function(){$r=new ReflectionFunction('f');$t=$r->getParameters()[0]->getType();return $t->getName();};$g=fn()=>$t->getName();", 1, "8.4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.MustParse(tc.php), Only: []string{"ReflectionCompositeTypeAssumedNamed"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("review.php", []byte("<?php "+fn+tc.src), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
