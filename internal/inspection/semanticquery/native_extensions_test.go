package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"

	"github.com/janalis/custos/internal/php/syntax"
)

type extensionContractProbe struct {
	query   string
	results *[]string
}

func (extensionContractProbe) ID() string { return "AbstractClassInstantiation" }
func (extensionContractProbe) Semantic()  {}
func (extensionContractProbe) Flow()      {}
func (extensionContractProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall}
}

func (p extensionContractProbe) Check(ctx *analysis.Context, n syntax.Node) {
	if m, ok := n.(*syntax.MethodCall); ok {
		if LiteralMethodName(m) == "getimageblob" && p.query == "frames" {
			*p.results = append(*p.results, fmt.Sprint(appendedFrames(ctx, m.Var, m, 0)))
		}
		return
	}
	c := n.(*syntax.FuncCall)
	name, ok := c.Name.(*syntax.Name)
	if !ok || name.Value != "strlen" {
		return
	}
	e := CallArgument(c.Args, 0, "")
	result := ""
	switch p.query {
	case "methods":
		result = fmt.Sprintf("%d", len(ObjectMethodHistory(ctx, c, e)))
	case "frames":
		result = fmt.Sprint(appendedFrames(ctx, e, c, 0))
	case "depth":
		result = fmt.Sprint(appendedFrames(ctx, e, c, 9))
	case "unaliased":
		result = fmt.Sprint(UnaliasedLocal(ctx, e, c))
	case "pristine":
		result = fmt.Sprint(MembersUnwrittenBefore(ctx, e, c))
	case "success":
		for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
			if m, ok := record.Node.(*syntax.MethodCall); ok {
				result = fmt.Sprint(NativeCallSucceeded(ctx, m, c))
			}
		}
	}
	*p.results = append(*p.results, result)
}

func runExtensionContract(t testing.TB, query, src string) []string {
	t.Helper()
	out := []string{}
	e, err := analysis.NewEngine([]analysis.Rule{extensionContractProbe{query, &out}}, analysis.Config{Only: []string{"AbstractClassInstantiation"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("edge.php", []byte("<?php "+src), syntax.Options{})
	if len(f.Errors) > 0 {
		t.Fatal(f.Errors)
	}
	e.Analyze(f)
	return out
}

func TestExtensionContracts(t *testing.T) {
	var captures strings.Builder
	for i := range 4200 {
		if i > 0 {
			captures.WriteByte(',')
		}
		fmt.Fprintf(&captures, "$v%d", i)
	}

	for _, tc := range []struct{ query, src, want string }{
		{"methods", "$x=new XMLWriter();$x->openMemory();strlen($x);", "1"},
		{"methods", "$x=new XMLWriter();if($flag){$x->text('x');}strlen($x);", "0"},
		{"methods", "$x=new XMLWriter();" + strings.Repeat("strtoupper('x');", 260) + "strlen($x);", "0"},
		{"success", "$x=new XMLWriter();if($x->openUri('x')){strlen($x);}", "true"},
		{"success", "$x=new XMLWriter();if(!$x->openUri('x')){return;}strlen($x);", "true"},
		{"success", "$x=new XMLWriter();if($x->openUri('x')){}strlen($x);", "false"},
		{"success", "$x=new XMLWriter();if(!$x->openUri('x')){strlen($x);}", "false"},
		{"success", "$x=new XMLWriter();$x->openUri('x');if($flag){strlen($x);}", "false"},
		{"success", "$x=new XMLWriter();$ok=$x->openUri('x');if($ok){strlen($x);}", "true"},
		{"success", "$x=new XMLWriter();if($x->openUri('x')===false){return;}else{$y=1;}strlen($x);", "true"},
		{"success", "$x=new XMLWriter();$x->openUri('x')&&strlen($x);", "true"},
		{"success", "$x=new XMLWriter();if(!$x->openUri('x')){return;}back:strlen($x);goto back;", "false"},
		{"frames", "$x=new Imagick();if($x->newImage(2,2,'red')){if($x->newImage(2,2,'blue')){$x->getImageBlob();}}", "2"},
		{"frames", "$x=new Imagick();$x->newImage(2,2,'red');$x->getImageBlob();", "0"},
		{"frames", "$x=new Imagick();if($x->newImage(2,2,'red')){$x->getNumberImages();$x->getImageBlob();}", "1"},
		{"frames", "$x=new Imagick();if($x->newImage(2,2,'red')){$x->resizeImage(2,2,1,1);$x->getImageBlob();}", "0"},
		{"frames", "$x=new Imagick();$x->addImage($unknown);$x->getImageBlob();", "0"},
		{"frames", "$x=new Imagick();if($x->addImage($unknown)){$x->getImageBlob();}", "0"},
		{"frames", "$s=new Imagick();if($s->newImage(2,2,'red')){$x=new Imagick();if($x->addImage($s)){$x->getImageBlob();}}", "1"},
		{"frames", "$x=new Imagick('animation.gif');$x->getImageBlob();", "0"},
		{"frames", "$x=new Imagick();$alias=&$x;$x->getImageBlob();", "0"},
		{"depth", "strlen($unknown);", "0"},
		{"unaliased", "$x=new XMLWriter();$alias=&$x;strlen($x);", "false"},
		{"unaliased", "$x=new XMLWriter();$cb=function()use(&$x){$x=new XMLWriter();};strlen($x);", "false"},
		{"unaliased", "$x=new XMLWriter();$cb=function()use($x){$x->text('x');};strlen($x);", "false"},
		{"unaliased", "$x='text';$cb=function()use($x){};strlen($x);", "true"},
		{"unaliased", "$x='text';$cb=fn()=>strtoupper($x);strlen($x);", "true"},
		{"unaliased", "$x=new XMLWriter();strlen($x);", "true"},
		{"unaliased", "$x=new XMLWriter();strlen($x);$alias=&$x;", "true"},
		{"unaliased", "$x=new XMLWriter();$alias=&$other;strlen($x);", "true"},
		{"unaliased", "$x=new XMLWriter();$other=$x;$alias=&$other;strlen($x);", "false"},
		{"unaliased", "$x=new XMLWriter();$cb=function()use(&$other){};strlen($x);", "true"},
		{"unaliased", "$x=fopen('x','r');$cb=function()use($x){};strlen($x);", "false"},
		{"unaliased", "$x=new XMLWriter();$a=[" + strings.Repeat("1,", 4200) + "];strlen($x);", "false"},
		{"unaliased", "$x=new XMLWriter();$cb=function()use(" + captures.String() + "){};strlen($x);", "false"},
		{"unaliased", "$x=new XMLWriter();" + strings.Repeat("strtoupper('x');", 520) + "strlen($x);", "false"},
		{"pristine", "$x=new SimpleXMLElement('<r/>');strlen($x);", "true"},
		{"pristine", "$x=new SimpleXMLElement('<r/>');$x->p=1;strlen($x);", "false"},
		{"pristine", "$x=new SimpleXMLElement('<r/>');$x->child[0]=1;strlen($x);", "false"},
		{"pristine", "$x=new SimpleXMLElement('<r/>');$x[0][0]=1;strlen($x);", "false"},
		{"pristine", "$x=new SimpleXMLElement('<r/>');$x->p++;strlen($x);", "false"},
		{"pristine", "$x=new SimpleXMLElement('<r/>');unset($x['p']);strlen($x);", "false"},
		{"pristine", "function f(){$x=new SimpleXMLElement('<r/>');strlen($x);}", "true"},
		{"pristine", "function unrelated(){$x->p=1;}$x=new SimpleXMLElement('<r/>');strlen($x);", "true"},
		{"pristine", "$x=new SimpleXMLElement('<r/>');" + strings.Repeat("$a=1;", 200) + "strlen($x);", "false"},
	} {
		t.Run(tc.query+tc.src, func(t *testing.T) {
			got := runExtensionContract(t, tc.query, tc.src)
			if len(got) != 1 || got[0] != tc.want {
				t.Fatalf("got %v want %s", got, tc.want)
			}
		})
	}
}

func TestBCMathOperands(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"bcadd('1','2');", 2},
		{"bcsub('1','2');", 2},
		{"bcmul('1','2');", 2},
		{"bcdiv('1','2');", 2},
		{"bcmod('1','2');", 2},
		{"bccomp('1','2');", 2},
		{"bcpow('1',2);", 1},
		{"bcsqrt('1');", 1},
		{"bcpowmod('1','2','3');", 2},
		{"strlen('1');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			called := false
			p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
				called = true
				if got := len(BCMathOperands(ctx, c)); got != tc.want {
					t.Fatalf("got %d want %d", got, tc.want)
				}
			}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.source), syntax.Options{}))
			if !called {
				t.Fatal("missing call")
			}
		})
	}
}

func BenchmarkBCMathOperands(b *testing.B) {
	file := syntax.Parse("test.php", []byte("<?php bcadd('1','2'); bcsub('3','4'); bcpowmod('5','6','7');"), syntax.Options{})
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) { BCMathOperands(ctx, c) }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
