package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestParseLiteralXML(t *testing.T) {
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
		children, text, plain, known := ParseLiteralXML(tc.src)
		if known != tc.known || plain != tc.plain || text != tc.text || len(children) != tc.children {
			t.Fatalf("%q got %v %q %v %v", tc.src, children, text, plain, known)
		}
	}
}

func TestXMLChildren(t *testing.T) {
	for _, tc := range []struct {
		xml  string
		want bool
	}{{"<one><two/></one>", true}, {"<one>text</one>", false}, {"<one/>", false}, {"<one><two></one>", false}, {strings.Repeat("x", 65537), false}} {
		if got := XMLChildren(tc.xml); got != tc.want {
			t.Errorf("%q got %v", tc.xml, got)
		}
	}
}

func TestDOMOwnerDocument(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$d=new DOMDocument(); probe($d);`, true},
		{`$d=new DOMDocument(); probe($d->documentElement);`, true},
		{`$d=new DOMDocument(); probe($d->other);`, false},
		{`$d=new DOMDocument(); probe($d->createElement('x'));`, true},
		{`$d=new DOMDocument(); probe($d->createTextNode('x'));`, true},
		{`$d=new DOMDocument(); probe($d->createElementNS('urn:test','x'));`, true},
		{`$d=new DOMDocument(); probe($d->importNode($node));`, true},
		{`$d=new DOMDocument(); probe($d->adoptNode($node));`, true},
		{`$d=new DOMDocument(); $n=$d->createElement('x'); probe($n->cloneNode(true));`, true},
		{`probe($unknown);`, false},
		{`probe(new stdClass());`, false},
		{`$d=new DOMDocument(); probe($d->getElementsByTagName('x'));`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			got := DOMOwnerDocument(ctx, CallArgument(c.Args, 0, "")) != nil
			if got != tc.want {
				t.Errorf("%s got %v", tc.src, got)
			}
		})
	}
}

func BenchmarkXMLChildren(b *testing.B) {
	s := `<catalog><item><name>Amber</name></item></catalog>`
	b.ReportAllocs()
	for b.Loop() {
		XMLChildren(s)
	}
}

func TestDOMElementHasChildren(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');probe($d->documentElement);`, true},
		{`$d=new DOMDocument();probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');if($flag){$d->loadXML('<a><b/></a>');}probe($d->documentElement);`, true},
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');if($flag){$d->loadXML('<a/>');}probe($d->documentElement);`, false},
		{`$d=new DOMDocument();if($flag){$d->load('other.xml');}$d->loadXML('<x><y/></x>');probe($d->documentElement);`, true},
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');$d->load('other.xml');probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$d->loadXML($xml);probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createElement('y'));probe($n);`, true},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createTextNode('y'));probe($n);`, false},
		{`$d=new DOMDocument();$n=$d->createElement('x');$child=$d->createElement('y');$n->appendChild($child);$n->removeChild($child);probe($n);`, false},
		{`$d=new DOMDocument();probe($d->other);`, false},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->hasChildNodes();probe($n);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := DOMElementHasChildren(ctx, CallArgument(c.Args, 0, ""), c); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}

func TestDOMOwnerDepthBudget(t *testing.T) {
	expr := `$d->createElement("x")`
	for range 18 {
		expr += `->cloneNode(true)`
	}
	probeNative(t, `$d=new DOMDocument();probe(`+expr+`);`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		if got := DOMOwnerDocument(ctx, CallArgument(c.Args, 0, "")); got != nil {
			t.Fatal("depth budget")
		}
	})
}
