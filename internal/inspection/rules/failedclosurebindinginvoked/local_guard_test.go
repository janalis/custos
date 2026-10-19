package failedclosurebindinginvoked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow analysis answers every guard of a plain variable in a scope it
// follows. Type narrowing decides for a binding result kept in a property or
// static property, and in a scope the flow does not follow (`goto`,
// generator).
func TestLocalGuards(t *testing.T) {
	const bind = "(static function(){})->bindTo(new stdClass())"
	property := func(body string) string {
		return "class A{public $b; function f($flag){$this->b=" + bind + ";" + body + "}}"
	}
	static := func(body string) string {
		return "class A{static $b; function f($flag){self::$b=" + bind + ";" + body + "}}"
	}
	unfollowed := func(body string) string {
		return "function f($flag){$b=" + bind + ";" + body + " goto done; done: ;}"
	}
	generator := func(body string) string {
		return "function f($flag){$b=" + bind + ";" + body + " yield 0;}"
	}
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"property unguarded", property("return ($this->b)();"), 1},
		{"property if not null", property("if($this->b!==null){return ($this->b)();}return 0;"), 0},
		{"property if null literal first", property("if(null!=$this->b){return ($this->b)();}return 0;"), 0},
		{"property if null branch", property("if($this->b===null){return ($this->b)();}return 0;"), 1},
		{"property unrelated if", property("if($flag){return ($this->b)();}return 0;"), 1},
		{"property ternary then", property("return $this->b?($this->b)():0;"), 0},
		{"property ternary else", property("return $this->b===null?0:($this->b)();"), 0},
		{"property ternary wrong arm", property("return $this->b?0:($this->b)();"), 1},
		{"property and", property("return $this->b!==null&&($this->b)();"), 0},
		{"property or", property("return !$this->b or ($this->b)();"), 0},
		{"property wrong operand", property("return $this->b||($this->b)();"), 1},
		{"property early return", property("if($this->b===null){return 0;}echo 'bound';return ($this->b)();"), 0},
		{"property early return in enclosing list", property("if(!$this->b){return 0;}if($flag){return ($this->b)();}return 0;"), 0},
		{"property non-terminating guard", property("if(!$this->b){echo 'failed';}return ($this->b)();"), 1},
		{"property guard with else", property("if(!$this->b){return 0;}else{echo 'bound';}return ($this->b)();"), 0},
		{"property else branch", property("if($this->b===null){return 0;}else{return ($this->b)();}"), 0},
		{"property elseif body", property("if($flag){return 0;}elseif($this->b){return ($this->b)();}return 0;"), 0},
		{"property after failed branch", property("if(!$this->b){return 0;}elseif($flag){return ($this->b)();}return 0;"), 0},
		{"property in later condition", property("if(!$this->b){return 0;}elseif(($this->b)()){return 1;}return 0;"), 0},
		{"property in first condition", property("if(($this->b)()){return 1;}elseif($flag){return 2;}return 0;"), 1},
		{"property else after unrelated condition", property("if($flag){return 0;}else{return ($this->b)();}"), 1},
		{"property conjunction", property("if($flag&&$this->b!=null){return ($this->b)();}return 0;"), 0},
		{"property conjunction when false", property("if($flag&&$this->b!=null){return 0;}return ($this->b)();"), 1},
		{"property disjunction", property("if($flag||$this->b==null){return 0;}return ($this->b)();"), 0},
		{"property disjunction when true", property("if($flag||$this->b!=null){return ($this->b)();}return 0;"), 1},
		{"property instanceof", property("if($this->b instanceof Closure){return ($this->b)();}return 0;"), 0},
		{"property not instanceof", property("if(!($this->b instanceof Closure)){return ($this->b)();}return 0;"), 1},
		{"property isset", property("return isset($this->b)?($this->b)():0;"), 0},
		{"property isset of another value", property("return isset($flag)?($this->b)():0;"), 1},
		{"property empty", property("if(empty($this->b)){return 0;}return ($this->b)();"), 0},
		{"property empty branch", property("if(empty($this->b)){return ($this->b)();}return 0;"), 1},
		{"property is_null", property("if(is_null($this->b)){return 0;}return ($this->b)();"), 0},
		{"property is_callable", property("return is_callable($this->b)?($this->b)():0;"), 0},
		{"property is_object", property("if(!is_object($this->b)){return 0;}return ($this->b)();"), 0},
		{"property is_object failed", property("return is_object($this->b)?0:($this->b)();"), 1},
		{"property is_string", property("return is_string($this->b)?0:($this->b)();"), 1},
		{"property compared with false", property("if($this->b!==false){return ($this->b)();}return 0;"), 1},
		{"other property tested", "class A{public $b; public $c; function f(){$this->b=" + bind + ";if($this->c!==null){return ($this->b)();}return 0;}}", 1},
		{"static property unguarded", static("return (self::$b)();"), 1},
		{"static property guarded", static("if(self::$b){return (self::$b)();}return 0;"), 0},
		{"static property early return", static("if(self::$b===null){return 0;}return (self::$b)();"), 0},
		{"goto scope unguarded", unfollowed("$b();"), 1},
		{"goto scope guarded", unfollowed("if($b!==null){$b();}"), 0},
		{"goto scope early return", unfollowed("if(!$b){return;}$b();"), 0},
		{"generator unguarded", generator("$b();"), 1},
		{"generator guarded", generator("if($b!==null){$b();}"), 0},
		// Forms the flow does not read and the walk does not accept stay reported.
		{"predicate compared with literal", "function f(){$b=" + bind + ";if(is_callable($b)===true){return $b();}return 0;}", 0},
		// An arm of `match (true)` runs when its condition holds, and after
		// every earlier condition failed.
		{"match arm", "function f(){$b=" + bind + ";return match(true){$b!==null=>$b(),default=>0};}", 0},
		{"match arm of the failure", "function f(){$b=" + bind + ";return match(true){$b===null=>$b(),default=>0};}", 1},
		{"match default after the failure", "function f(){$b=" + bind + ";return match(true){$b===null=>0,default=>$b()};}", 0},
		{"match arm of another test", "function f($flag){$b=" + bind + ";return match(true){$flag!==null=>$b(),default=>0};}", 1},
		// The flow proves a result assigned inside the condition that tests it.
		{"assigned in condition", "function f(){$c=static function(){};if($b=$c->bindTo(new stdClass())){return $b();}return 0;}", 0},
		{"assigned in failed condition", "function f(){$c=static function(){};if(!($b=$c->bindTo(new stdClass()))){return $b();}return 0;}", 1},
		{"flag holds the test", "function f(){$c=static function(){};$b=$c->bindTo(new stdClass());$ok=$b!==null;if(!$ok){return 0;}return $b();}", 0},
		{"flag unused", "function f(){$c=static function(){};$b=$c->bindTo(new stdClass());$ok=$b!==null;return $b();}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FailedClosureBindingInvoked"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
