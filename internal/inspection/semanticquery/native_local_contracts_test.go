package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type localContractProbe struct {
	query   string
	results *[]string
}

func (localContractProbe) ID() string { return "AbstractClassInstantiation" }
func (localContractProbe) Semantic()  {}
func (localContractProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KAttribute}
}

func (p localContractProbe) Check(ctx *analysis.Context, n syntax.Node) {
	if a, ok := n.(*syntax.Attribute); ok {
		if p.query == "attributes" {
			_, flags, known := NativeAttributeFlags(ctx, a)
			*p.results = append(*p.results, fmt.Sprintf("%d/%t", flags, known))
		}
		return
	}
	c := n.(*syntax.FuncCall)
	name, ok := c.Name.(*syntax.Name)
	if !ok {
		return
	}
	if p.query == "conditions" {
		if name.Value != "readdir" {
			return
		}
		*p.results = append(*p.results, fmt.Sprint(NativeConditionUse(ctx, c, true) != nil))
		return
	}
	if name.Value != "probe" {
		return
	}
	a := CallArgument(c.Args, 0, "")
	b := CallArgument(c.Args, 1, "")
	result := ""
	switch p.query {
	case "construction":
		result = fmt.Sprint(NativeConstruction(ctx, a, "SplFixedArray") != nil)
	case "origin":
		v := NativeLocalValue(ctx, a)
		if v == nil {
			result = "unknown"
		} else {
			result = ctx.Text(v)
		}
	case "nil-same":
		result = fmt.Sprint(NativeSameValue(ctx, nil, a))
	case "dominates":
		result = fmt.Sprint(NativeStateDominates(c, a), NativeStateDominates(a, c))
	case "same":
		result = fmt.Sprint(NativeSameValue(ctx, a, b))
	case "size":
		size, k := NativeFixedArraySize(ctx, a, c)
		result = fmt.Sprintf("%d/%t", size, k)
	case "slots":
		result = fmt.Sprint(NativeFixedArraySlots(ctx, a, c))
	case "stored":
		slot, ok := a.(*syntax.ArrayDimFetch)
		if ok {
			result = fmt.Sprint(NativeStoredArray(ctx, slot, c))
		}
	case "int":
		v, k := NativeContractInt(ctx, a)
		result = fmt.Sprintf("%d/%t", v, k)
	case "reflection":
		params, ret, k := NativeReflectionTarget(ctx, a)
		result = fmt.Sprintf("%d/%s/%t", len(params), ret, k)
	case "sentinel":
		result = fmt.Sprint(NativeLocalSentinelGuard(ctx, a, "null"))
	case "csv":
		args, pos, name, k := NativeCSVSignature(ctx, a)
		_ = args
		result = fmt.Sprintf("%d/%s/%t", pos, name, k)
	}
	*p.results = append(*p.results, result)
}

func TestNativeLocalQueries(t *testing.T) {
	for _, tc := range []struct{ query, src, want, php string }{
		{"nil-same", `probe(1);`, "false", ""},
		{"dominates", `probe(1);`, "false false", ""},
		{"origin", `$a=new stdClass();$a->unknown();probe($a);`, "unknown", ""},
		{"origin", `$a=new stdClass();$alias=$a;$alias[0]=1;probe($alias);`, "new stdClass()", ""},
		{"origin", `$a=new stdClass();$a+=1;probe($a);`, "unknown", ""},
		{"size", `$a=SplFixedArray::fromArray([4=>'x']);$a->setSize(2);probe($a);`, "2/true", ""},
		{"slots", `$a=SplFixedArray::fromArray([4=>'x']);$a->setSize(2);probe($a);`, "map[]", ""},
		{"slots", `function f(){ $a=new SplFixedArray(2);$a[1]=1;unset($a[1]);probe($a);}`, "map[]", ""},
		{"int", `$flags=Attribute::TARGET_CLASS|Attribute::TARGET_METHOD;probe($flags);`, "5/true", ""},
		{"conditions", `do{echo readdir($d);}while(true);if($x){echo readdir($d);}if(readdir($d)<true){} $x=readdir($d);`, "false,false,false,false", ""},
		{"construction", `probe(new SplFixedArray(2));probe(new $unknown());probe(new Other());`, "true,false,false", ""},
		{"origin", `$a=new SplFixedArray(2);$a[0]=1;probe($a);probe($unknown);`, "new SplFixedArray(2),unknown", ""},
		{"same", `$a=new stdClass();$b=$a;probe($a,$b);probe($a,new stdClass());`, "true,false", ""},
		{"size", `probe(new SplFixedArray());probe(new SplFixedArray(-1));probe(SplFixedArray::fromArray($unknown));probe(SplFixedArray::fromArray(['s'=>'x']));probe(SplFixedArray::fromArray([-1=>'x']));probe(SplFixedArray::fromArray(['x'],$unknown));`, "0/true,0/false,0/false,0/false,0/false,0/false", ""},
		{"size", `$a=new SplFixedArray(1);$a->setSize($unknown);probe($a);$b=new SplFixedArray(1);$b->setSize(-1);probe($b);`, "0/false,0/false", ""},
		{"size", `$a=SplFixedArray::fromArray([3=>'x'],false);probe($a);`, "1/true", ""},
		{"slots", `$a=new SplFixedArray(5);$a[4]=null;unset($a[4]);probe($a);`, "map[]", ""},
		{"slots", `$a=new SplFixedArray(5);if($unknown){$a[4]=1;}probe($a);`, "map[]", ""},
		{"stored", `$s=new SplObjectStorage();$o=new stdClass();$s[$o]=[];probe($s[$o]);$s[$o]=new stdClass();probe($s[$o]);`, "true,false", ""},
		{"origin", `$a=new stdClass();unknown($a);probe($a);`, "unknown", ""},
		{"origin", `$a=new stdClass();if($unknown){$a=new SplFixedArray();}probe($a);`, "unknown", ""},
		{"origin", `$a=new stdClass();$b=&$a;probe($b);`, "unknown", ""},
		{"reflection", `probe(new ReflectionFunction($unknown));probe(new ReflectionFunction('absent'));probe(new ReflectionMethod('A',$unknown));probe(new ReflectionMethod($unknown,'f'));probe(new ReflectionMethod('Absent','f'));`, "0//false,0//false,0//false,0//false,0//false", ""},
		{"reflection", `class A{function f($x):int{return 1;}}probe(new ReflectionMethod(new A(),'f'));`, "1/int/true", ""},
		{"int", `probe(Attribute::TARGET_ALL);probe(Attribute::IS_REPEATABLE);probe(Attribute::TARGET_CONSTANT);probe(Attribute::TARGET_CLASS|Attribute::TARGET_METHOD);probe($unknown);probe(Missing::C);probe($unknown::C);`, "63/true,64/true,0/false,5/true,0/false,0/false,0/false", "8.4"},
		{"int", `probe(Attribute::TARGET_ALL);probe(Attribute::IS_REPEATABLE);`, "127/true,128/true", "8.5"},
		{"attributes", `#[Custom]#[Attribute]class Mark{}#[Mark]class A{}`, "0/false,0/false,127/true", "8.5"},
		{"sentinel", `$a=static function(){};$b=$a->bindTo(new stdClass());if($b===null){return;}probe($b);`, "true", ""},
		{"sentinel", `$a=static function(){};$b=$a->bindTo(new stdClass());if($b){probe($b);}`, "false", ""},
		{"sentinel", `$a=static function(){};$b=$a->bindTo(new stdClass());if($other!==null){probe($b);}`, "false", ""},
		{"csv", `probe(1);probe(str_getcsv('a'));probe($unknown->fgetcsv());`, "0//false,1/separator/true,0//false", ""},
		{"conditions", `$x=readdir($d);if(readdir($d)&&$x){} if((readdir($d))){} if($x=readdir($d)){} if(-readdir($d)){} if(1+readdir($d)){} if(readdir($d)==$unknown){} if(false==readdir($d)){} while(readdir($d)){} do{}while(readdir($d));`, "false,false,true,true,false,false,false,true,true,true", ""},
	} {
		t.Run(tc.query+tc.src+tc.php, func(t *testing.T) {
			var got []string
			r := localContractProbe{tc.query, &got}
			cfg := analysis.Config{Only: []string{r.ID()}}
			if tc.php != "" {
				cfg.PHP = phpversion.MustParse(tc.php)
			}
			e, err := analysis.NewEngine([]analysis.Rule{r}, cfg)
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("local.php", []byte("<?php "+tc.src), syntax.Options{})
			e.Analyze(f)
			if strings.Join(got, ",") != tc.want {
				t.Fatalf("got %v want %s", got, tc.want)
			}
		})
	}
}

func TestNativeLocalEventBudget(t *testing.T) {
	var src strings.Builder
	src.WriteString("<?php $a=new SplFixedArray(1);")
	for range 513 {
		src.WriteString("$other=1;")
	}
	src.WriteString("$a[0]=1;probe($a);")
	var got []string
	r := localContractProbe{"origin", &got}
	e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{Only: []string{r.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("budget.php", []byte(src.String()), syntax.Options{}))
	if strings.Join(got, ",") != "unknown" {
		t.Fatalf("budget: %v", got)
	}
}

func BenchmarkNativeLocalContracts(b *testing.B) {
	src := []byte(`<?php $a=new SplFixedArray(10);$a[9]=1;probe($a);`)
	f := syntax.Parse("bench.php", src, syntax.Options{})
	var results []string
	r := localContractProbe{"construction", &results}
	e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{Only: []string{r.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(src)))
	b.ResetTimer()
	for range b.N {
		results = results[:0]
		e.Analyze(f)
	}
}
