package datelasterrorsfalseunchecked_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/datelasterrorsfalseunchecked"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$errors=DateTimeImmutable::getLastErrors(); echo $errors['warning_count'];\n", 1},
		{"negative0", "<?php\n$e=DateTimeImmutable::getLastErrors();if($e){echo $e['warning_count'];}\n", 0},
		{"negative1", "<?php\n$e=DateTimeImmutable::getLastErrors();if($e!==false){echo $e['warning_count'];}\n", 0},
		{"negative2", "<?php\n$e=DateTimeImmutable::getLastErrors()?:[];echo $e['warning_count'];\n", 0},
		{"negative3", "<?php\necho $x['warning_count'];\n", 0},
		{"negative4", "<?php\n$e=DateTimeImmutable::getLastErrors();if($x!==false){echo 1;}\n", 0},
		{"negative5", "<?php\n$e=DateTimeImmutable::getLastErrors();if(false!==$e){echo $e['warning_count'];}\n", 0},
		{"negative6", "<?php\nfunction f(){$e=DateTimeImmutable::getLastErrors();if($e===false){return;}echo $e['warning_count'];}\n", 0},
		{"extra0", "<?php\n$e=DateTimeImmutable::getLastErrors();if($q){echo $e['warning_count'];}\n", 1},
		{"extra1", "<?php\n$e=DateTimeImmutable::getLastErrors();if($q!==false){echo $e['warning_count'];}\n", 1},
		{"extra2", "<?php\n$e=DateTimeImmutable::getLastErrors();if($q!==$r){echo $e['warning_count'];}\n", 1},
		{"coalesce", "<?php\n$e=DateTime::getLastErrors();echo $e['warning_count']??0;\n", 0},
		{"isset", "<?php\n$e=DateTime::getLastErrors();echo isset($e['warnings'])?1:0;\n", 0},
		{"and", "<?php\n$e=DateTime::getLastErrors();echo $e&&$e['warning_count'];\n", 0},
		{"strict or", "<?php\n$e=DateTime::getLastErrors();echo $e===false||$e['warning_count']===0;\n", 0},
		{"ternary", "<?php\n$e=DateTime::getLastErrors();echo $e?$e['warning_count']:0;\n", 0},
		{"early return", "<?php\nfunction f(){$e=DateTime::getLastErrors();if(!$e){return 0;}return $e['warning_count'];}\n", 0},
		{"is_array", "<?php\n$e=DateTime::getLastErrors();if(is_array($e)){echo $e['warning_count'];}\n", 0},
		// An offset probe and an exit through an endless loop are proven by the rule itself.
		{"isset or", "<?php\n$e=DateTime::getLastErrors();echo !isset($e['warnings'])||$e['warning_count'];\n", 0},
		{"isset ternary", "<?php\n$e=DateTime::getLastErrors();echo isset($e['warnings'])?$e['warning_count']:0;\n", 0},
		{"isset elseif", "<?php\n$e=DateTime::getLastErrors();if($q){echo 1;}elseif(isset($e['warnings'])){echo $e['warning_count'];}\n", 0},
		{"isset branch", "<?php\n$e=DateTime::getLastErrors();if(isset($e['warnings'])){echo $e['warning_count'];}\n", 0},
		{"isset else", "<?php\n$e=DateTime::getLastErrors();if(!isset($e['warnings'])){echo 0;}else{echo $e['warning_count'];}\n", 0},
		{"isset after elseif", "<?php\n$e=DateTime::getLastErrors();if($q){echo 1;}elseif(!isset($e['warnings'])){echo 0;}else{echo $e['warning_count'];}\n", 0},
		{"isset early return", "<?php\nfunction f(){$e=DateTime::getLastErrors();if(!isset($e['warnings'])){return 0;}return $e['warning_count'];}\n", 0},
		{"loose exit through endless loop", "<?php\nfunction f(){$e=DateTime::getLastErrors();if($e==false){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"isset failed branch", "<?php\n$e=DateTime::getLastErrors();if(!isset($e['warnings'])){echo $e['warning_count'];}\n", 1},
		{"reassigned after guard", "<?php\n$e=DateTime::getLastErrors();if($e){$e=DateTime::getLastErrors();echo $e['warning_count'];}\n", 1},
		{"wrong ternary arm", "<?php\n$e=DateTime::getLastErrors();echo $e?0:$e['warning_count'];\n", 1},
		{"index of a probe", "<?php\n$e=DateTime::getLastErrors();echo $e[$e['warning_count']]??0;\n", 1},
		{"extra3", "<?php\n$e=DateTimeImmutable::getLastErrors();if($e!==false){echo $e['warning_count'];}\n", 0},
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
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\n$errors=DateTimeImmutable::getLastErrors(); echo $errors['warning_count'];"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
