package duplicatematchcondition_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/duplicatematchcondition"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$v=match($input) {4=>'first',4=>'second',default=>'other'};\n", 1},
		{"negative0", "<?php\nmatch($x){1=>1,'1'=>2,default=>3};\n", 0},
		{"negative1", "<?php\nmatch($x){$a=>1,$b=>2};\n", 0},
		{"negative2", "<?php\nmatch($x){f()=>1,f()=>2};\n", 0},
		{"negative3", "<?php\nmatch($x){1=>1,2=>2};\n", 0},
		{"extra0", "<?php\nmatch($x){true=>1,true=>2};\n", 1},
		{"extra1", "<?php\nmatch($x){false=>1,false=>2};\n", 1},
		{"extra2", "<?php\nmatch($x){null=>1,null=>2};\n", 1},
		{"extra3", "<?php\nenum E{case A;}match($x){E::A=>1,E::A=>2};\n", 1},
		{"extra4", "<?php\nmatch($x){Missing::X=>1,Missing::X=>2};\n", 0},
		{"duplicate final arm", "<?php match($x){4=>1,4=>2};", 1},
		{"duplicate with comment", "<?php match($x){4=>1,4=>/*retain*/2};", 1},
		{"duplicate among conditions", "<?php match($x){4=>1,4,5=>2};", 1},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\n$v=match($input) {4=>'first',4=>'second',default=>'other'};"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}

func TestHashCommentPreservesFixBoundary(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
	if err != nil {
		t.Fatal(err)
	}
	findings := e.Analyze(syntax.Parse("test.php", []byte("<?php $x=match($v){1=>'one',1 # keep this explanation\n=>'two'};"), syntax.Options{}))
	if len(findings) != 1 || len(findings[0].Fixes) != 0 {
		t.Fatalf("comment must survive automatic repair: %+v", findings)
	}
}
