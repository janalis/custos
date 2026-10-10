package emptysplcollectionextraction

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
	fixtures, err := filepath.Glob("../../../../testdata/rules/EmptySplCollectionExtraction/*.php")
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
			cfg := analysis.Config{Only: []string{"EmptySplCollectionExtraction"}}
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
				cfg.Rules = map[string]analysis.RuleConfig{"EmptySplCollectionExtraction": {Options: side.Options}}
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
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"EmptySplCollectionExtraction"}})
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
		{"$q=new SplQueue();$q->enqueue('x');$q->dequeue();$q->dequeue();", 1},
		{"$q=new SplQueue();$q->dequeue();$q->dequeue();", 1},
		{"$q=new SplStack();$q->push('x');$q->pop();$q->bottom();", 1},
		{"class Q extends SplQueue{}$q=new Q();$q->dequeue();", 0},
		{"function f(SplQueue $q){$q->dequeue();}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"EmptySplCollectionExtraction"}})
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
		{"$q=new SplQueue();$q->$method();", 0, "8.4"},
	} {
		t.Run(tc.src+tc.php, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.MustParse(tc.php), Only: []string{"EmptySplCollectionExtraction"}})
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

func TestEmptyCollectionGuards(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{`$q=new SplQueue();if(!$q->isEmpty()){$q->dequeue();}`, 0},
		{`$q=new SplQueue();if($q->isEmpty()){return;}$q->dequeue();`, 0},
		{`$q=new SplQueue();if($q->isEmpty()){}else{$q->dequeue();}`, 0},
		{`$q=new SplQueue();if($q->isEmpty()){$q->dequeue();}`, 1},
		{`$q=new SplQueue();if($q->count()>0){$q->dequeue();}`, 0},
		{`$q=new SplQueue();if($q->count()>=1){$q->dequeue();}`, 0},
		{`$q=new SplQueue();if($q->count()!==0){$q->dequeue();}`, 0},
		{`$q=new SplQueue();if($unknown>0){$q->dequeue();}`, 1},
		{`$q=new SplQueue();if($q->count()>2){$q->dequeue();}`, 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"EmptySplCollectionExtraction"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("guard.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Fatalf("%s: %+v want %d", tc.src, got, tc.want)
		}
	}
}

func TestCountGuardEdges(t *testing.T) {
	for _, src := range []string{
		`$q=new SplQueue();if($q->count()==0){return;}$q->dequeue();`,
		`$q=new SplQueue();if($q->count()>0){}else{echo 1;}`,
		`$q=new SplQueue();if($q->count()>$unknown){$q->dequeue();}`,
	} {
		e, _ := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"EmptySplCollectionExtraction"}})
		got := e.Analyze(syntax.Parse("guard.php", []byte("<?php "+src), syntax.Options{}))
		if strings.Contains(src, "$unknown") {
			if len(got) != 1 {
				t.Fatalf("unknown bound: %+v", got)
			}
		} else if len(got) != 0 {
			t.Fatalf("guard: %+v", got)
		}
	}
}

func TestUnrelatedBooleanGuard(t *testing.T) {
	src := `<?php $q=new SplQueue();if($unknown){$q->dequeue();}`
	e, _ := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"EmptySplCollectionExtraction"}})
	if got := e.Analyze(syntax.Parse("guard.php", []byte(src), syntax.Options{})); len(got) != 1 {
		t.Fatalf("unrelated guard: %+v", got)
	}
}
