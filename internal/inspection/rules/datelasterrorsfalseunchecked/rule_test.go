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
		// The flow reads offset probes and follows a body that never ends.
		{"isset or", "<?php\n$e=DateTime::getLastErrors();echo !isset($e['warnings'])||$e['warning_count'];\n", 0},
		{"isset ternary", "<?php\n$e=DateTime::getLastErrors();echo isset($e['warnings'])?$e['warning_count']:0;\n", 0},
		{"isset elseif", "<?php\n$e=DateTime::getLastErrors();if($q){echo 1;}elseif(isset($e['warnings'])){echo $e['warning_count'];}\n", 0},
		{"isset branch", "<?php\n$e=DateTime::getLastErrors();if(isset($e['warnings'])){echo $e['warning_count'];}\n", 0},
		{"isset else", "<?php\n$e=DateTime::getLastErrors();if(!isset($e['warnings'])){echo 0;}else{echo $e['warning_count'];}\n", 0},
		{"isset after elseif", "<?php\n$e=DateTime::getLastErrors();if($q){echo 1;}elseif(!isset($e['warnings'])){echo 0;}else{echo $e['warning_count'];}\n", 0},
		{"isset early return", "<?php\nfunction f(){$e=DateTime::getLastErrors();if(!isset($e['warnings'])){return 0;}return $e['warning_count'];}\n", 0},
		{"loose exit through endless loop", "<?php\nfunction f(){$e=DateTime::getLastErrors();if($e==false){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"isset exit through endless loop", "<?php\nfunction f(){$e=DateTime::getLastErrors();if(!isset($e['warnings'])){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"and exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!($e&&$q)){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"or exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!$e||$q){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"strict exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!($e!==false)){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"loose yoda exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(false==$e){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"relational exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if($e>0){while(true){sleep(1);}}return $e['warning_count'];}\n", 1},
		{"empty exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(empty($e)){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"isset held exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(isset($e['warnings'])){while(true){sleep(1);}}return $e['warning_count'];}\n", 1},
		{"isset of another value exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!isset($q['warnings'])){while(true){sleep(1);}}return $e['warning_count'];}\n", 1},
		{"not array exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!is_array($e)){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"assignment exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!($e=DateTime::getLastErrors())){while(true){sleep(1);}}return $e['warning_count'];}\n", 0},
		{"another value exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!$q){while(true){sleep(1);}}return $e['warning_count'];}\n", 1},
		{"wrong conjunction exit through endless loop", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!$e&&$q){while(true){sleep(1);}}return $e['warning_count'];}\n", 1},
		// The statements of a `declare` block are followed in place.
		{"loose exit inside a declare block", "<?php\nfunction f(){$e=DateTime::getLastErrors();if($e==false){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"isset exit inside a declare block", "<?php\nfunction f(){$e=DateTime::getLastErrors();if(!isset($e['warnings'])){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"and exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!($e&&$q)){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"or exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!$e||$q){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"strict exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!($e!==false)){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"loose yoda exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(false==$e){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"relational exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if($e>0){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 1},
		{"empty exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(empty($e)){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"isset held exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(isset($e['warnings'])){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 1},
		{"isset of another value exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!isset($q['warnings'])){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 1},
		{"not array exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!is_array($e)){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"assignment exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!($e=DateTime::getLastErrors())){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 0},
		{"another value exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!$q){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 1},
		{"wrong conjunction exit inside a declare block", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!$e&&$q){declare(ticks=1){return 0;}}return $e['warning_count'];}\n", 1},
		{"declare block that falls through", "<?php\nfunction f($q){$e=DateTime::getLastErrors();if(!$e){declare(ticks=1){echo 1;}}return $e['warning_count'];}\n", 1},
		// A predicate around the assignment tests the value assigned.
		{"is_array of the assignment, early return", "<?php\nfunction f(){if(!is_array($e=DateTime::getLastErrors())){return 0;}return $e['warning_count'];}\n", 0},
		// A property is not a variable of the flow: the rule's own walk reads the
		// earlier exits that guard it.
		{"property unguarded", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();return $this->e['warning_count'];}}\n", 1},
		{"property strict early return", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if($this->e===false){return 0;}return $this->e['warning_count'];}}\n", 0},
		{"property either operand of or", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if($this->e===false||$q){return 0;}return $this->e['warning_count'];}}\n", 0},
		{"property negated strict test", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if(!($this->e!==false)){return 0;}return $this->e['warning_count'];}}\n", 0},
		{"property empty", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if(empty($this->e)){return 0;}return $this->e['warning_count'];}}\n", 0},
		{"property offset not set", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if(!isset($this->e['warnings'])){return 0;}return $this->e['warning_count'];}}\n", 0},
		{"property not an array", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if(!is_array($this->e)){return 0;}return $this->e['warning_count'];}}\n", 0},
		{"property falsy", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if(!$this->e){return 0;}return $this->e['warning_count'];}}\n", 0},
		{"property leaves on success", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if($this->e!==false){return 0;}return $this->e['warning_count'];}}\n", 1},
		{"property guard that falls through", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if($this->e===false){echo 0;}return $this->e['warning_count'];}}\n", 1},
		{"another property tested", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if($this->o===false){return 0;}return $this->e['warning_count'];}}\n", 1},
		{"same property of another object", "<?php\nclass C{private $e;private $o;function f($q){$this->e=DateTime::getLastErrors();if($q->e===false){return 0;}return $this->e['warning_count'];}}\n", 1},
		// An assignment tested in place guards the variable it stores, not another one.
		{"assignment of another variable exit", "<?php\nfunction f(){$e=DateTime::getLastErrors();if(!($m=DateTime::getLastErrors())){return 0;}return $e['warning_count'];}\n", 1},
		// A statement that leaves on the failing outcome is a fact of the flow.
		{"statement-level exit", "<?php\nfunction f(){$e=DateTime::getLastErrors();$e!==false or exit;return $e['warning_count'];}\n", 0},
		{"statement-level exit on another value", "<?php\nfunction f($q){$e=DateTime::getLastErrors();$q!==false or exit;return $e['warning_count'];}\n", 1},
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
