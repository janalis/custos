package splobjectstorageindirectarraymutation

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

func (probeRule) ID() string               { return "SplObjectStorageIndirectArrayMutation" }
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
	e, err := analysis.NewEngine([]analysis.Rule{probeRule{query, &answers}}, analysis.Config{Only: []string{"SplObjectStorageIndirectArrayMutation"}})
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

// storedArray holds while the storage slot read by probe() holds an array.
func TestStoredArray(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`$s=new SplObjectStorage();$o=new stdClass();$s[$o]=[];probe($s[$o]);$s[$o]=new stdClass();probe($s[$o]);`, "true,false"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			if got := probe(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) (string, bool) {
				slot, ok := semanticquery.CallArgument(c.Args, 0, "").(*syntax.ArrayDimFetch)
				return fmt.Sprint(ok && storedArray(ctx, slot, c)), probed(c) && ok
			}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
