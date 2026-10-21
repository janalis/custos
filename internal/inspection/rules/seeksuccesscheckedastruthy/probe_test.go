package seeksuccesscheckedastruthy

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type probeRule struct {
	query   func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool)
	answers *[]string
}

func (probeRule) ID() string               { return "SeekSuccessCheckedAsTruthy" }
func (probeRule) Semantic()                {}
func (probeRule) Flow()                    {}
func (probeRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (p probeRule) Check(ctx *analysis.Context, n syntax.Node) {
	if answer, ok := p.query(ctx, n.(*syntax.FuncCall)); ok {
		*p.answers = append(*p.answers, answer)
	}
}

// probe analyzes "<?php "+src and joins with commas what query answers for
// the calls it takes.
func probe(t *testing.T, src string, query func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool)) string {
	t.Helper()
	var answers []string
	e, err := analysis.NewEngine([]analysis.Rule{probeRule{query, &answers}}, analysis.Config{Only: []string{"SeekSuccessCheckedAsTruthy"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("probe.php", []byte("<?php "+src), syntax.Options{})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %s: %v", src, f.Errors)
	}
	e.Analyze(f)
	return strings.Join(answers, ",")
}

// conditionUse finds the condition a readdir() result is tested in.
func TestConditionUse(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`do{echo readdir($d);}while(true);if($x){echo readdir($d);}if(readdir($d)<true){} $x=readdir($d);`, "false,false,false,false"},
		{`$x=readdir($d);if(readdir($d)&&$x){} if((readdir($d))){} if($x=readdir($d)){} if(-readdir($d)){} if(1+readdir($d)){} if(readdir($d)==$unknown){} if(false==readdir($d)){} while(readdir($d)){} do{}while(readdir($d));`, "false,false,true,true,false,false,false,true,true,true"},
		{`if($x){}elseif(readdir($d)){}`, "true"},
		{`if($x):elseif(!readdir($d)):endif;`, "true"},
		{`if($x){}else if(readdir($d)){}`, "true"},
		{`for(;readdir($d);){}`, "true"},
		{`for(;$x,($e=readdir($d));){}`, "true"},
		{`for(;readdir($d),$x;){}`, "false"},
		{`for(readdir($d);;readdir($d)){}`, "false,false"},
		{`if(readdir(...)){}`, "false"},
		{`if($x+=readdir($d)){}`, "false"},
		{`if($x??=readdir($d)){}`, "false"},
		{`if($x.=readdir($d)){}`, "false"},
		{`if($x[readdir($d)]=1){}`, "false"},
		{`if($x=readdir($d)){}`, "true"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			if got := probe(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool) {
				return fmt.Sprint(conditionUse(ctx, c, true) != nil), readdirCall(c)
			}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// Without boolean comparisons only the bare result is a condition.
func TestConditionUseWithoutComparisons(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`if(readdir($d)==false){}`, "false"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			if got := probe(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool) {
				return fmt.Sprint(conditionUse(ctx, c, false) != nil), readdirCall(c)
			}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func readdirCall(c *syntax.FuncCall) bool {
	name, ok := c.Name.(*syntax.Name)
	return ok && name.Value == "readdir"
}
