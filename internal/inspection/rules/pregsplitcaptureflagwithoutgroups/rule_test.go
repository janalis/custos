package pregsplitcaptureflagwithoutgroups_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/pregsplitcaptureflagwithoutgroups"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$parts=preg_split('/,/',$text,-1,PREG_SPLIT_DELIM_CAPTURE);\n", 1},
		{"negative0", "<?php\npreg_split('/(,)/',$s,-1,PREG_SPLIT_DELIM_CAPTURE);\n", 0},
		{"negative1", "<?php\npreg_split('/,/',$s);\n", 0},
		{"negative2", "<?php\npreg_split('/,/',$s,-1,0);\n", 0},
		{"negative3", "<?php\npreg_split($pattern,$s,-1,PREG_SPLIT_DELIM_CAPTURE);\n", 0},
		{"negative4", "<?php\npreg_split('/(?|,)/',$s,-1,PREG_SPLIT_DELIM_CAPTURE);\n", 0},
		{"negative5", "<?php\npreg_split('/,/',$s,-1,$flags);\n", 0},
		{"negative6", "<?php\ntrim($s);\n", 0},
		{"negative7", "<?php\npreg_split('/,/',$s,-1,PREG_SPLIT_NO_EMPTY);\n", 0},
		{"extra0", "<?php\npreg_split('/,/',$s,-1,PREG_SPLIT_DELIM_CAPTURE|PREG_SPLIT_NO_EMPTY);\n", 1},
		{"extra1", "<?php\npreg_split('/,/',$s,-1,PREG_SPLIT_NO_EMPTY|PREG_SPLIT_DELIM_CAPTURE);\n", 1},
		{"extra2", "<?php\npreg_split('/,/',$s,-1,PREG_SPLIT_NO_EMPTY|PREG_SPLIT_DELIM_CAPTURE|PREG_SPLIT_OFFSET_CAPTURE);\n", 1},
		{"extra3", "<?php\npreg_split('/,/',$s,-1,PREG_SPLIT_DELIM_CAPTURE|0);\n", 1},
		{"extra4", "<?php\npreg_split('/,/',$s,-1,PREG_SPLIT_DELIM_CAPTURE /*why*/);\n", 1},
		{"extra5", "<?php\npreg_split('/,/',$s,-1,(PREG_SPLIT_DELIM_CAPTURE /*why*/));\n", 1},
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

func TestHashCommentPreservesFixBoundary(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
	if err != nil {
		t.Fatal(err)
	}
	findings := e.Analyze(syntax.Parse("test.php", []byte("<?php preg_split('/,+/','a,b',-1,PREG_SPLIT_DELIM_CAPTURE | # keep this explanation\n PREG_SPLIT_NO_EMPTY);"), syntax.Options{}))
	if len(findings) != 1 || len(findings[0].Fixes) != 0 {
		t.Fatalf("comment must survive automatic repair: %+v", findings)
	}
}
