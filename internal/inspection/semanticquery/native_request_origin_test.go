package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeRequestSourceBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`probe($_POST["x"]);`, true},
		{`probe("text"[0]);`, false},
		{`function read(){probe($_POST["x"]);}`, true},
		{`$f=function(){$_POST["x"]="derived";};probe($_POST["x"]);`, true},
		{`$_POST["x"]++;probe($_POST["x"]);`, false},
		{`unset($_POST["x"]);probe($_POST["x"]);`, false},
		{`$alias=&$_POST;probe($_POST["x"]);`, false},
		{`strlen("a");probe($_POST["x"]);`, true},
		{`sanitizeRequest();probe($_POST["x"]);`, false},
		{`$s->sanitize();probe($_POST["x"]);`, false},
		{`Sanitizer::sanitize();probe($_POST["x"]);`, false},
		{`new Sanitizer();probe($_POST["x"]);`, false},
		{`$GLOBALS["_POST"]=["x"=>"derived"];probe($_POST["x"]);`, false},
		{strings.Repeat(`$x=1;`, 1400) + `probe($_POST["x"]);`, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				source, ok := CallArgument(c.Args, 0, "value").(*syntax.ArrayDimFetch)
				if !ok {
					t.Fatal("missing source access")
				}
				if got := NativeRequestUnwritten(ctx, source); got != tc.want {
					t.Fatalf("got %v want %v", got, tc.want)
				}
				if got := NativeRequestUnwritten(ctx, source); got != tc.want {
					t.Fatal("memo changed source proof")
				}
			})
		})
	}
}
