package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"

	"custos/internal/php/syntax"
)

type expansionDExtensionsProbe struct {
	query   string
	results *[]string
}

func (expansionDExtensionsProbe) ID() string { return "AbstractClassInstantiation" }
func (expansionDExtensionsProbe) Semantic()  {}
func (expansionDExtensionsProbe) Flow()      {}
func (expansionDExtensionsProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall}
}

func (p expansionDExtensionsProbe) Check(ctx *analysis.Context, n syntax.Node) {
	if m, ok := n.(*syntax.MethodCall); ok {
		if ExpansionDMethodName(m) == "getimageblob" && p.query == "frames" {
			*p.results = append(*p.results, fmt.Sprint(expansionDFrames(ctx, m.Var, m, 0)))
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
		result = fmt.Sprintf("%d", len(ExpansionDMethods(ctx, c, e)))
	case "frames":
		result = fmt.Sprint(expansionDFrames(ctx, e, c, 0))
	case "depth":
		result = fmt.Sprint(expansionDFrames(ctx, e, c, 9))
	case "unaliased":
		result = fmt.Sprint(ExpansionDUnaliased(ctx, e, c))
	case "pristine":
		result = fmt.Sprint(ExpansionDPristine(ctx, e, c))
	case "success":
		for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
			if m, ok := record.Node.(*syntax.MethodCall); ok {
				result = fmt.Sprint(ExpansionDMethodSucceeded(ctx, m, c))
			}
		}
	}
	*p.results = append(*p.results, result)
}

func expansionDExtensionsRun(t testing.TB, query, src string) []string {
	t.Helper()
	out := []string{}
	e, err := analysis.NewEngine([]analysis.Rule{expansionDExtensionsProbe{query, &out}}, analysis.Config{Only: []string{"AbstractClassInstantiation"}})
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

func TestExpansionDExtensionsContracts(t *testing.T) {
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
			got := expansionDExtensionsRun(t, tc.query, tc.src)
			if len(got) != 1 || got[0] != tc.want {
				t.Fatalf("got %v want %s", got, tc.want)
			}
		})
	}
}

func TestExpansionDLiteralXML(t *testing.T) {
	for _, tc := range []struct {
		src          string
		known, plain bool
		text         string
		children     int
	}{
		{"<r/>", true, true, "", 0},
		{"<r>false</r>", true, true, "false", 0},
		{"<r><c>text</c></r>", true, false, "", 1},
		{"<r a='x'/>", true, false, "", 0},
		{"<r xmlns='urn:x'/>", false, false, "", 0},
		{"<r xmlns:x='urn:x'/>", false, false, "", 0},
		{"<!DOCTYPE r><r/>", false, false, "", 0},
		{"<r>", false, false, "", 0},
		{"", false, true, "", 0},
		{strings.Repeat("x", 65537), false, false, "", 0},
	} {
		children, text, plain, known := ExpansionDXML(tc.src)
		if known != tc.known || plain != tc.plain || text != tc.text || len(children) != tc.children {
			t.Fatalf("%q got %v %q %v %v", tc.src, children, text, plain, known)
		}
	}
}

func TestExpansionDMethodName(t *testing.T) {
	if ExpansionDMethodName(&syntax.MethodCall{Name: &syntax.Variable{Name: "dynamic"}}) != "" {
		t.Fatal("dynamic method")
	}
	if ExpansionDMethodName(&syntax.MethodCall{Name: &syntax.Identifier{Value: "NEWImage"}}) != "newimage" {
		t.Fatal("literal method")
	}
}

func BenchmarkExpansionDPristine(b *testing.B) {
	out := []string{}
	e, err := analysis.NewEngine([]analysis.Rule{expansionDExtensionsProbe{"pristine", &out}}, analysis.Config{Only: []string{"AbstractClassInstantiation"}})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte("<?php $x=new SimpleXMLElement('<r/>');strlen($x);"), syntax.Options{})
	e.Analyze(f)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out = out[:0]
		e.Analyze(f)
	}
}

func BenchmarkExpansionDUnaliasedAllocations(b *testing.B) {
	var source strings.Builder
	source.WriteString("<?php ")
	for i := range 100 {
		fmt.Fprintf(&source, "$v%d=new stdClass();strlen($v%d);", i, i)
	}
	out := []string{}
	e, err := analysis.NewEngine([]analysis.Rule{expansionDExtensionsProbe{"unaliased", &out}}, analysis.Config{Only: []string{"AbstractClassInstantiation"}})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte(source.String()), syntax.Options{})
	e.Analyze(f)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out = out[:0]
		e.Analyze(f)
	}
}
