package regexbyteoffsetascharacteroffset_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/regexbyteoffsetascharacteroffset"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\npreg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE); echo mb_substr('ñx',$m[0][1],1);\n", 1},
		{"negative0", "<?php\npreg_match('/x/','ax',$m,PREG_OFFSET_CAPTURE);mb_substr('ax',$m[0][1],1);\n", 0},
		{"negative1", "<?php\npreg_match('/x/u','ñx',$m);mb_substr('ñx',$m[0][1],1);\n", 0},
		{"negative2", "<?php\npreg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);substr('ñx',$m[0][1],1);\n", 0},
		{"negative3", "<?php\npreg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);mb_substr('other',$m[0][1],1);\n", 0},
		{"negative4", "<?php\nmb_substr('ñx',0,1);\n", 0},
		{"negative5", "<?php\nmb_substr('ñx',$m[0][0],1);\n", 0},
		{"negative6", "<?php\nmb_substr('ñx',$m[1],1);\n", 0},
		{"negative7", "<?php\n$m=other();mb_substr('ñx',$m[0][1],1);\n", 0},
		{"negative8", "<?php\nother();mb_substr('ñx',$m[0][1],1);\n", 0},
		{"negative9", "<?php\npreg_match('/x/u','ñx',$other,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1);\n", 0},
		{"negative10", "<?php\nmb_substr('ñx',items()[0][1],1);\n", 0},
		{"negative11", "<?php\npreg_match('/x/u','öx',$m,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1);\n", 0},
		{"no prior capture", "<?php mb_substr('ñx',$m[0][1],1);", 0},
		{"control flow capture", "<?php if($q){preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);}mb_substr('ñx',$m[0][1],1);", 0},
		{"byte encoding consumer", "<?php preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1,'8bit');", 0},
		{"unknown encoding consumer", "<?php preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1,$encoding);", 0},
		{"explicit utf8 consumer", "<?php preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1,'UTF-8');", 1},
		{"utf8 alias consumer", "<?php preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1,'UTF8');", 1},
		{"grapheme consumer", "<?php preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);grapheme_substr('ñx',$m[0][1],1);", 1},
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
