package splfixedarrayshrinkdiscardsvalues

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

func (probeRule) ID() string               { return "SplFixedArrayShrinkDiscardsValues" }
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
	e, err := analysis.NewEngine([]analysis.Rule{probeRule{query, &answers}}, analysis.Config{Only: []string{"SplFixedArrayShrinkDiscardsValues"}})
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

// fixedArraySlots lists the slots known to hold a value before probe().
func TestFixedArraySlots(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`$a=SplFixedArray::fromArray([4=>'x']);$a->setSize(2);probe($a);`, "map[]"},
		{`function f(){ $a=new SplFixedArray(2);$a[1]=1;unset($a[1]);probe($a);}`, "map[]"},
		{`$a=new SplFixedArray(5);$a[4]=null;unset($a[4]);probe($a);`, "map[]"},
		{`$a=new SplFixedArray(5);if($unknown){$a[4]=1;}probe($a);`, "map[]"},
		{`$a=SplFixedArray::fromArray([1,2],false);probe($a);`, "map[0:true 1:true]"},
		{`$a=SplFixedArray::fromArray([1,2],$keep);probe($a);`, "map[]"},
		{`$a=SplFixedArray::fromArray([1,2],...$options);probe($a);`, "map[]"},
		{`$a=new SplFixedArray(2);$a[0]=1;$a->setSize($unknown);probe($a);`, "map[]"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			if got := probe(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool) {
				return fmt.Sprint(fixedArraySlots(ctx, semanticquery.CallArgument(c.Args, 0, ""), c)), probed(c)
			}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// A call argument neither dominates its call nor the other way round.
func TestStateDominatesArgument(t *testing.T) {
	got := probe(t, "probe(1);", func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool) {
		a := semanticquery.CallArgument(c.Args, 0, "")
		return fmt.Sprint(stateDominates(c, a), stateDominates(a, c)), probed(c)
	})
	if got != "false false" {
		t.Fatalf("got %q", got)
	}
}
