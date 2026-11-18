package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestDirectCondition(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`if(probe()) {}`, true},
		{`while(probe()) {}`, true},
		{`do{}while(probe());`, true},
		{`if(!probe()) {}`, true},
		{`if($r=probe()) {}`, true},
		{`$r=probe();`, false},
		{`if(probe()===0) {}`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := DirectCondition(c); got != tc.want {
				t.Fatalf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestLiteralMethodName(t *testing.T) {
	if LiteralMethodName(&syntax.MethodCall{Name: &syntax.Variable{Name: "dynamic"}}) != "" {
		t.Fatal("dynamic method")
	}
	if LiteralMethodName(&syntax.MethodCall{Name: &syntax.Identifier{Value: "NEWImage"}}) != "newimage" {
		t.Fatal("literal method")
	}
}
