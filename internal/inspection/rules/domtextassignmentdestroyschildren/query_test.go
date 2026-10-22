package domtextassignmentdestroyschildren

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

func TestDOMConstructedWriteProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createElement('y'));probe($n);`, true},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createTextNode('y'));probe($n);`, false},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createElement('y'));if($flag){}probe($n);`, false},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createElement('y'));mutate($n);probe($n);`, false},
		{`probe(3);`, false},
		{strings.Repeat(`$a=1;`, 257) + `probe($n);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := createdBeforeWrite(ctx, semanticquery.CallArgument(c.Args, 0, ""), c); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}

func probeNative(t *testing.T, src string, check func(*analysis.Context, *syntax.FuncCall)) {
	t.Helper()
	p := funcCallProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
			check(ctx, c)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{}))
}

type funcCallProbe struct {
	check func(*analysis.Context, *syntax.FuncCall)
}

func (funcCallProbe) ID() string { return "DomTextAssignmentDestroysChildren" }

func (funcCallProbe) Semantic() {}

func (funcCallProbe) Flow() {}

func (funcCallProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (p funcCallProbe) Check(ctx *analysis.Context, n syntax.Node) {
	p.check(ctx, n.(*syntax.FuncCall))
}

func TestDOMBeforeWriteProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');probe($d->documentElement);`, true},
		{`$d=new DOMDocument();$d->loadXML('<x/>');probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');if($flag){}probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');mutate($d);probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');$other=3;probe($d->documentElement);`, true},
		{`probe(3);`, false},
		{`$d=new DOMDocument();probe($d->other);`, false},
		{`probe((new DOMDocument())->documentElement);`, false},
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');$d=$unknown;probe($d->documentElement);`, false},
		{strings.Repeat(`$a=1;`, 257) + `probe($d->documentElement);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := documentElementHasChildren(ctx, semanticquery.CallArgument(c.Args, 0, ""), c); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}
