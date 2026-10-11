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
		{"membership double-quoted literal", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],[\"a\"],true)){E::from($_GET['x']);}", 0},
		{"membership double-quoted case", "<?php enum E:string{case A=\"a\";}if(in_array($_GET['x'],['a'],true)){E::from($_GET['x']);}", 0},
		{"membership escaped literal", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],[\"\\x61\"],true)){E::from($_GET['x']);}", 0},
		{"membership int literal", "<?php enum E:int{case A=1;}if(in_array($_GET['x'],[1],true)){E::from($_GET['x']);}", 0},
		{"membership int literal mismatch", "<?php enum E:int{case A=1;}if(in_array($_GET['x'],[2],true)){E::from($_GET['x']);}", 1},
		{"membership string for int case", "<?php enum E:int{case A=1;}if(in_array($_GET['x'],['1'],true)){E::from($_GET['x']);}", 1},
		{"membership ignores plain constants", "<?php enum E:string{case A='a';const B='b';}if(in_array($_GET['x'],['b'],true)){E::from($_GET['x']);}", 1},
		{"membership last conjunct", "<?php enum E:string{case A='a';}if(isset($_GET['x'])&&in_array($_GET['x'],['a'],true)){E::from($_GET['x']);}", 0},
		{"membership last of three conjuncts", "<?php enum E:string{case A='a';}if($on&&(isset($_GET['x'])&&(in_array($_GET['x'],['a'],true)))){E::from($_GET['x']);}", 0},
		{"membership keyword conjunction", "<?php enum E:string{case A='a';}if(isset($_GET['x']) and in_array($_GET['x'],['a'],true)){E::from($_GET['x']);}", 0},
		{"membership followed by evaluation", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],['a'],true)&&unknown()){E::from($_GET['x']);}", 1},
		{"membership in disjunction", "<?php enum E:string{case A='a';}if($on||in_array($_GET['x'],['a'],true)){E::from($_GET['x']);}", 1},
		{"membership guard around closure", "<?php enum E:string{case A='a';}if(in_array($_GET['x'],['a'],true)){$f=function(){return E::from($_GET['x']);};}", 1},
		{"from in catch block", "<?php enum E:string{case A='a';}try{work();}catch(ValueError $e){E::from($_GET['x']);}", 1},
		{"from in finally block", "<?php enum E:string{case A='a';}try{work();}catch(ValueError $e){}finally{E::from($_GET['x']);}", 1},
		{"from in catch block of protected try", "<?php enum E:string{case A='a';}try{try{work();}catch(ValueError $e){E::from($_GET['x']);}}catch(Throwable $e){}", 0},
		{"closure created in try body", "<?php enum E:string{case A='a';}try{$f=function(){return E::from($_GET['x']);};}catch(ValueError $e){}", 1},
		{"arrow function created in try body", "<?php enum E:string{case A='a';}try{$f=fn()=>E::from($_GET['x']);}catch(ValueError $e){}", 1},
		{"try inside closure", "<?php enum E:string{case A='a';}$f=function(){try{return E::from($_GET['x']);}catch(ValueError $e){return null;}};", 0},
		{"unrelated catch type", "<?php enum E:string{case A='a';}try{E::from($_GET['x']);}catch(RuntimeException $e){}", 1},
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
