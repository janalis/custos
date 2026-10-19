package intlformattingfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow reads these guards itself, also on a result computed from a
// property or a project call: the value is incomplete, the checks are not.
func TestFlowGuards(t *testing.T) {
	const produce = "$n=$fm->format(load($b));"
	run(t, []guardCase{
		{"unguarded", produce + "strlen($n);", 1},
		{"ternary then", produce + "return $n?strlen($n):0;", 0},
		{"ternary else", produce + "return !$n?0:strlen($n);", 0},
		{"ternary wrong arm", produce + "return $n?0:strlen($n);", 1},
		{"and operand", produce + "return $n&&strlen($n);", 0},
		{"else of failure test", produce + "if(!$n){echo 1;}else{strlen($n);}", 0},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif(!$n){echo 2;}else{strlen($n);}", 0},
		{"empty exit", produce + "if(empty($n)){return;}strlen($n);", 0},
		{"is_bool exit", produce + "if(is_bool($n)){return;}strlen($n);", 0},
		{"is_scalar", produce + "if(is_scalar($n)){strlen($n);}", 1},
		{"assignment tested in place", "if($n=$fm->format(load($x))){strlen($n);}", 0},
		{"property operand unguarded", "$n=$fm->format($this->v);strlen($n);", 1},
		{"property operand, statement-level exit", "$n=$fm->format($this->v);$n!==false or exit;strlen($n);", 0},
		{"project call operand, strict return", produce + "if($n===false){return;}strlen($n);", 0},
	})
}

// A condition whose two operands each rule out the failure is not a fact of
// the flow, so type narrowing decides every guard below.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=$fm->format($b);"
	run(t, []guardCase{
		{"unguarded", produce + "strlen($n);", 1},
		{"branch body", produce + "if(is_string($n)||$n!==false){strlen($n);}", 0},
		{"branch body, one operand only", produce + "if(is_string($n)||$x){strlen($n);}", 1},
		{"ternary then", produce + "return ($n||$n!==false)?strlen($n):0;", 0},
		{"ternary else", produce + "return (!$n&&$n===false)?0:strlen($n);", 0},
		{"ternary wrong arm", produce + "return (!$n&&$n===false)?strlen($n):0;", 1},
		{"and operand", produce + "return ($n||is_string($n))&&strlen($n);", 0},
		{"or operand", produce + "return ($n||is_string($n))||strlen($n);", 1},
		{"earlier exit", produce + "if(!$n&&$n===false){return;}strlen($n);", 0},
		{"earlier branch that continues", produce + "if(!$n&&$n===false){echo 1;}strlen($n);", 1},
		{"else of failure test", produce + "if(!$n&&$n===false){echo 1;}else{strlen($n);}", 0},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif(!$n&&$n===false){echo 2;}else{strlen($n);}", 0},
		{"empty exit", produce + "if(empty($n)&&$n===false){return;}strlen($n);", 0},
		{"is_bool exit", produce + "if(is_bool($n)&&!$n){return;}strlen($n);", 0},
		{"assignment tested in place", "if(($n=$fm->format($b))||is_string($n)){strlen($n);}", 0},
		{"other variable", produce + "if(is_string($x)||$x!==false){strlen($n);}", 1},
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
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"IntlFormattingFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f(NumberFormatter $fm,$b,$x){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
