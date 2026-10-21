package gdpaletteindexdecodedasrgb

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type probeRule struct {
	query   func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool)
	answers *[]string
}

func (probeRule) ID() string               { return "GdPaletteIndexDecodedAsRgb" }
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
	e, err := analysis.NewEngine([]analysis.Rule{probeRule{query, &answers}}, analysis.Config{Only: []string{"GdPaletteIndexDecodedAsRgb"}})
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

// probed reports whether c calls probe().
func probed(c *syntax.FuncCall) bool {
	name, ok := c.Name.(*syntax.Name)
	return ok && name.Value == "probe"
}

// localOnly is the one value a variable holds on every path, or nothing.
func TestLocalOnly(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`function f($a){$u=scandir($a); probe($u);}`, "scandir($a)"},
		{`function f($a, $c){if($c){$u=scandir($a);}else{$u=glob($a);} probe($u);}`, ""},
		{`function f($a, $c){if($c){$u=scandir($a);}else{$u=[];} probe($u);}`, ""},
		{`function f($a, $c){$u=[]; if($c){$u=scandir($a);} probe($u);}`, ""},
		{`function f($a, $c, $u){if($c){$u=scandir($a);} probe($u);}`, ""},
		{`function f($a, $c){$u=scandir($a); while($c--){probe($u); $u=[];}}`, ""},
		{`function f($a, $c){while($c--){$u=scandir($a); probe($u);}}`, "scandir($a)"},
		{`function f($a){$u=scandir($a); $v=$u; probe($v);}`, "scandir($a)"},
		{`function f($a){probe(scandir($a));}`, "scandir($a)"},
		{`function f($a){probe($a);}`, ""},
		{`function f($a){$u=scandir($a); probe($u); yield 1;}`, "scandir($a)"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			if got := probe(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool) {
				if v := localOnly(ctx, semanticquery.CallArgument(c.Args, 0, "")); v != nil {
					return ctx.Text(v), probed(c)
				}
				return "", probed(c)
			}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
