package validatedintegerzerorejected

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

func (probeRule) ID() string               { return "ValidatedIntegerZeroRejected" }
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
	e, err := analysis.NewEngine([]analysis.Rule{probeRule{query, &answers}}, analysis.Config{Only: []string{"ValidatedIntegerZeroRejected"}})
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

// integerAllowsZero reads the filter options passed to probe().
func TestIntegerAllowsZero(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"probe();", "true"},
		{"probe(0);", "true"},
		{"probe(FILTER_NULL_ON_FAILURE);", "false"},
		{"probe(['flags'=>FILTER_NULL_ON_FAILURE]);", "false"},
		{"probe(['options'=>['min_range'=>1]]);", "false"},
		{"probe(['options'=>['max_range'=>-1]]);", "false"},
		{"probe(['options'=>['min_range'=>0,'max_range'=>5]]);", "true"},
		{"probe(['options'=>$unknown]);", "false"},
		{"probe(['flags'=>$unknown]);", "false"},
		{"probe(['options'=>['min_range'=>$unknown]]);", "false"},
		{"probe($unknown);", "false"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			if got := probe(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool) {
				return fmt.Sprint(integerAllowsZero(ctx, semanticquery.CallArgument(c.Args, 0, "value"))), probed(c)
			}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
