package aeaddecryptiontaglengthunchecked

import (
	"fmt"
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

func (probeRule) ID() string               { return "AeadDecryptionTagLengthUnchecked" }
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
	e, err := analysis.NewEngine([]analysis.Rule{probeRule{query, &answers}}, analysis.Config{Only: []string{"AeadDecryptionTagLengthUnchecked"}})
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

// lengthProven holds where the flow or a guard proves a 4-byte string.
func TestLengthProven(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"if(strlen($s)===4){probe($s);}", "true"},
		{"if(strlen($s)===8){probe($s);}", "false"},
		{"if(strlen($other)===4){probe($s);}", "false"},
		{"if(count($s)===4){probe($s);}", "false"},
		{"if(strlen($s)!==4){return;} probe($s);", "true"},
		{"if(strlen($s)!==4){echo 'short';} probe($s);", "false"},
		{"probe($s);", "false"},
		{`function f($s){probe();}`, "false"},
		{`function f($s){if(strlen($s) !== 4){return;} echo 1; probe($s);}`, "true"},
		{`final class H { public $s; } function f(H $o){if(strlen($o->s) === 4){probe($o->s);}}`, "true"},
		{`final class H { public $s; } function f(H $o){if(4 === strlen($o->s)){probe($o->s);}}`, "true"},
		{`final class H { public $s; } function f(H $o){if(strlen($o->s) === 5){probe($o->s);}}`, "false"},
		{`final class H { public $s; } function f(H $o){return strlen($o->s) === 4 ? probe($o->s) : 0;}`, "true"},
		{`final class H { public $s; } function f(H $o){if(count($o->s) === 4){probe($o->s);}}`, "false"},
		{`final class H { public $s; } function f(H $o){if(strlen($o->s) == 4){probe($o->s);}}`, "true"},
		{`final class H { public $s; } function f(H $o, $c){if($c){return 0;}elseif(strlen($o->s) === 4){probe($o->s);}}`, "true"},
		{`final class H { public $s; } function f(H $o){if(strlen($o->s) !== 4){return 0;}else{probe($o->s);}}`, "true"},
		{`final class H { public $s; } function f(H $o, $c){if($c){return 0;}elseif(strlen($o->s) !== 4){return 1;}else{probe($o->s);}}`, "true"},
		{`final class H { public $s; } function f(H $o){return strlen($o->s) === 4 && probe($o->s);}`, "true"},
		{`final class H { public $s; } function f(H $o){return strlen($o->s) !== 4 || probe($o->s);}`, "true"},
		{`function f($o){if(strlen($o->s()) === 4){probe($o->s());}}`, "false"},
		{`function f($s, $a){if(strlen($s) === 4){extract($a); probe($s);}}`, "false"},
		{"$v=fread($h,4);if(strlen($v)===4){$v=fread($h,4);probe($v);}", "false"},
		{"if($v){probe($v);}", "false"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			if got := probe(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool) {
				return fmt.Sprint(lengthProven(ctx, semanticquery.CallArgument(c.Args, 0, ""), 4)), probed(c)
			}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
