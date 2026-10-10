package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type localContractExtraProbe struct{ localContractProbe }

func (p localContractExtraProbe) Check(ctx *analysis.Context, n syntax.Node) {
	c, ok := n.(*syntax.FuncCall)
	if !ok {
		return
	}
	name, ok := c.Name.(*syntax.Name)
	if !ok {
		return
	}
	if p.query == "plain-condition" {
		if name.Value == "readdir" {
			*p.results = append(*p.results, fmt.Sprint(NativeConditionUse(ctx, c, false) != nil))
		}
		return
	}
	if name.Value != "probe" {
		return
	}
	a := CallArgument(c.Args, 0, "")
	result := ""
	switch p.query {
	case "mode":
		m, k := NativeStreamMode(ctx, a, c)
		result = fmt.Sprintf("%s/%t", m, k)
	case "success":
		var call *syntax.FuncCall
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			if f, ok := n.(*syntax.FuncCall); ok && NativeBuiltin(ctx, f, "ftruncate") {
				call = f
			}
			return true
		})
		result = fmt.Sprint(NativeCallSucceeded(ctx, call, c))
	case "effects":
		result = fmt.Sprint(len(NativeMethodStateCalls(ctx, c, a, "setSize")))
	default:
		p.localContractProbe.Check(ctx, n)
		return
	}
	*p.results = append(*p.results, result)
}

func TestNativeLocalExtraProofs(t *testing.T) {
	for _, tc := range []struct{ query, source, want string }{
		{"plain-condition", `if(readdir($d)==false){}`, "false"},
		{"slots", `$a=new SplFixedArray(2);$a[0]=1;$a->setSize($unknown);probe($a);`, "map[]"},
		{"mode", `$s=fopen('x',$mode);if($s!==false){probe($s);}`, "/false"},
		{"mode", `$s=fopen('x','');if($s!==false){probe($s);}`, "/false"},
		{"success", `$s=fopen('x','w');if(!ftruncate($s,0)){return;}probe($s);`, "true"},
		{"success", `$s=fopen('x','w');if(ftruncate($s,0)>true){probe($s);}`, "false"},
		{"int", `probe(7);`, "7/true"},
		{"origin", `probe();`, "unknown"},
		{"origin", `$a=1;$a=probe($a);`, "1"},
		{"effects", `($a=new SplFixedArray(3))->{$method}(2);probe($a);`, "0"},
		{"effects", `$a=SplFixedArray::fromArray([]);$a->{'setSize'}(2);probe($a);`, "0"},
	} {
		t.Run(tc.query+tc.source, func(t *testing.T) {
			var got []string
			p := localContractExtraProbe{localContractProbe{tc.query, &got}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			e.Analyze(syntax.Parse("extra.php", []byte("<?php "+tc.source), syntax.Options{}))
			if strings.Join(got, ",") != tc.want {
				t.Fatalf("got %v want %s", got, tc.want)
			}
		})
	}
}

func TestNativeStateScopeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   bool
	}{
		{`if($x)unset($x);probe();`, false},
		{`function first(){unset($x);}function second(){probe();}`, false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			file := syntax.Parse("scope.php", []byte("<?php "+tc.source), syntax.Options{})
			var before *syntax.Unset
			var at *syntax.FuncCall
			syntax.InspectFile(file, func(n syntax.Node) bool {
				switch n := n.(type) {
				case *syntax.Unset:
					before = n
				case *syntax.FuncCall:
					at = n
				}
				return true
			})
			if before == nil || at == nil {
				t.Fatal("missing operations")
			}
			if got := NativeStateDominates(before, at); got != tc.want {
				t.Fatalf("dominates=%v", got)
			}
		})
	}
}
