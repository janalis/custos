package enumfromuncheckedexternalvalue_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/enumfromuncheckedexternalvalue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\nenum Phase:string {case Ready='ready';} $p=Phase::from($_GET['phase']);\n", 1},
		{"negative0", "<?php\nenum E:string{case A='a';} E::tryFrom($_GET['x']);\n", 0},
		{"negative1", "<?php\nenum E:string{case A='a';} E::from('a');\n", 0},
		{"negative2", "<?php\nenum E:string{case A='a';} try{E::from($_GET['x']);}catch(ValueError $e){}\n", 0},
		{"negative3", "<?php\nenum E{case A;} E::from($_GET['x']);\n", 0},
		{"negative4", "<?php\nUnknown::from($_GET['x']);\n", 0},
		{"negative5", "<?php\nclass E{static function from($x){}} E::from($_GET['x']);\n", 0},
		{"negative6", "<?php\nenum E:string{case A='a';} E::from($x);\n", 0},
		{"negative7", "<?php\nenum E:string{case A='a';}E::from($value['x']);\n", 0},
		{"negative8", "<?php\nenum E:string{case A='a';}E::from(getInput()['x']);\n", 0},
		{"strict membership guard", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],['a'],true)){E::from($_GET['x']);}", 0},
		{"membership not strict", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],['a'])){E::from($_GET['x']);}", 1},
		{"membership includes invalid value", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],['a','bad'],true)){E::from($_GET['x']);}", 1},
		{"membership empty set", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],[],true)){E::from($_GET['x']);}", 1},
		{"membership unknown set", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],$allowed,true)){E::from($_GET['x']);}", 1},
		{"membership other input", "<?php enum E:string{case A='a';}if(in_array($_GET['other'],['a'],true)){E::from($_GET['x']);}", 1},
		{"membership intervening invalidation", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],['a'],true)){unknown();E::from($_GET['x']);}", 1},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\nenum Phase:string {case Ready='ready';} $p=Phase::from($_GET['phase']);"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
