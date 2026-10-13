package pregreplacefailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow reads these guards itself, also on a result computed from a
// property or a project call: the value is incomplete, the checks are not.
func TestFlowGuards(t *testing.T) {
	const produce = "$n=preg_replace($t,$r,load($b));"
	run(t, []guardCase{
		{"unguarded", produce + "strlen($n);", 1},
		{"branch body", produce + "if($n!==null){strlen($n);}", 0},
		{"elseif own condition", produce + "if($x){echo 1;}elseif(is_string($n)){strlen($n);}", 0},
		{"ternary else", produce + "return $n===null?0:strlen($n);", 0},
		{"ternary wrong arm", produce + "return $n===null?strlen($n):0;", 1},
		{"and operand", produce + "return !is_null($n)&&strlen($n);", 0},
		{"isset", produce + "if(isset($n)){strlen($n);}", 0},
		{"truthiness", produce + "if($n){strlen($n);}", 0},
		{"property operand unguarded", "$n=preg_replace($t,$r,$this->v);strlen($n);", 1},
		{"property operand, statement-level exit", "$n=preg_replace($t,$r,$this->v);$n!==null or exit;strlen($n);", 0},
		{"project call operand, strict return", produce + "if($n===null){return;}strlen($n);", 0},
	})
}

// The flow follows a guard body that never ends and the statements of a
// `declare` block, so these rows are decided before the rule's own walk.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=preg_replace($t,$r,$b);"
	for _, exit := range []struct{ name, body string }{
		{"endless loop", "{while(true){}}"},
		{"declare block", "{declare(ticks=1){return;}}"},
	} {
		leave := exit.body
		t.Run(exit.name, func(t *testing.T) {
			run(t, []guardCase{
				{"strict null", produce + "if($n===null)" + leave + "strlen($n);", 0},
				{"null on the left", produce + "if(null===$n)" + leave + "strlen($n);", 0},
				{"negated strict test", produce + "if(!($n!==null))" + leave + "strlen($n);", 0},
				{"not set", produce + "if(!isset($n))" + leave + "strlen($n);", 0},
				{"is_null", produce + "if(is_null($n))" + leave + "strlen($n);", 0},
				{"not a string", produce + "if(!is_string($n))" + leave + "strlen($n);", 0},
				{"either operand of or", produce + "if($x||$n===null)" + leave + "strlen($n);", 0},
				{"failed conjunction", produce + "if(!($n!==null&&$x))" + leave + "strlen($n);", 0},
				{"assignment tested in place", "if(null===$n=preg_replace($t,$r,$b))" + leave + "strlen($n);", 0},
				{"from an enclosing list", produce + "if($n===null)" + leave + "if($x){strlen($n);}", 0},
				{"leaves on success", produce + "if($n!==null)" + leave + "strlen($n);", 1},
				{"both operands of and", produce + "if($x&&$n===null)" + leave + "strlen($n);", 1},
				{"other variable", produce + "if($x===null)" + leave + "strlen($n);", 1},
				{"reassigned after the guard", produce + "if($n===null)" + leave + "$n=preg_replace($t,$r,$n);strlen($n);", 1},
			})
		})
	}
	run(t, []guardCase{
		{"loop with a way out", produce + "if($n===null){while(true){break;}}strlen($n);", 1},
		{"declare block that falls through", produce + "if($n===null){declare(ticks=1){echo 1;}}strlen($n);", 1},
		// A predicate around the assignment tests the value assigned.
		{"not a string, assigned in place", "if(!is_string($n=preg_replace($t,$r,$b))){return;}strlen($n);", 0},
		{"is_null, assigned in place", "if(is_null($n=preg_replace($t,$r,$b))){return;}strlen($n);", 0},
		{"is_null of another assignment", "$n=preg_replace($t,$r,$b);if(is_null($m=preg_replace($t,$r,$b))){return;}strlen($n);", 1},
	})
}

type guardCase struct {
	name, source string
	want         int
}

func run(t *testing.T, cases []guardCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PregReplaceFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f($t,$r,$b,$x){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

// A property is not a variable of the flow: the rule's own walk reads the
// strict tests of an earlier exit that guard a result held in one.
func TestPropertyGuardWalk(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PregReplaceFailureUnchecked"}})
	if err != nil {
		t.Fatal(err)
	}
	const produce = "<?php class C{private $n;private $o;function f($t,$r,$b,$x){$this->n=preg_replace($t,$r,$b);"
	for _, tc := range []guardCase{
		{"unguarded", "strlen($this->n);", 1},
		{"strict null", "if($this->n===null){return;}strlen($this->n);", 0},
		{"enclosing branch", "if($this->n!==null){strlen($this->n);}", 0},
		{"enclosing branch, wrong sense", "if($this->n===null){strlen($this->n);}", 1},
		{"else of a null test", "if($this->n===null){echo 1;}else{strlen($this->n);}", 0},
		{"elseif own condition", "if($x){echo 1;}elseif($this->n!==null){strlen($this->n);}", 0},
		// Only the first condition of a chain is read for its later arms.
		{"else after a failed elseif", "if($x){echo 1;}elseif($this->n===null){echo 2;}else{strlen($this->n);}", 1},
		{"elseif of an unrelated chain", "if($x){echo 1;}elseif($b){strlen($this->n);}", 1},
		{"ternary arm", "return $this->n!==null?strlen($this->n):0;", 0},
		{"ternary wrong arm", "return $this->n===null?strlen($this->n):0;", 1},
		{"and operand", "return $this->n!==null&&strlen($this->n);", 0},
		{"or operand", "return $this->n===null||strlen($this->n);", 0},
		{"or operand, wrong sense", "return $this->n!==null||strlen($this->n);", 1},
		{"null on the left", "if(null===$this->n){return;}strlen($this->n);", 0},
		{"negated strict test", "if(!($this->n!==null)){return;}strlen($this->n);", 0},
		{"not set", "if(!isset($this->n)){return;}strlen($this->n);", 0},
		{"not set among several", "if(!isset($x,$this->n)){return;}strlen($this->n);", 0},
		{"is_null", "if(is_null($this->n)){return;}strlen($this->n);", 0},
		{"not a string", "if(!is_string($this->n)){return;}strlen($this->n);", 0},
		{"either operand of or", "if($x||$this->n===null){return;}strlen($this->n);", 0},
		{"failed conjunction", "if(!($this->n!==null&&$x)){return;}strlen($this->n);", 0},
		{"from an enclosing list", "if($this->n===null){return;}if($x){strlen($this->n);}", 0},
		{"leaves on success", "if($this->n!==null){return;}strlen($this->n);", 1},
		{"both operands of and", "if($x&&$this->n===null){return;}strlen($this->n);", 1},
		{"guard that falls through", "if($this->n===null){echo 1;}strlen($this->n);", 1},
		{"another property", "if($this->o===null){return;}strlen($this->n);", 1},
		{"same property of another object", "if($x->n===null){return;}strlen($this->n);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.Parse("guard.php", []byte(produce+tc.source+"}}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
