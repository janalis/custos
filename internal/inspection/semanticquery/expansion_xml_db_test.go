package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestExpansionXMLChildren(t *testing.T) {
	for _, tc := range []struct {
		xml  string
		want bool
	}{{"<one><two/></one>", true}, {"<one>text</one>", false}, {"<one/>", false}, {"<one><two></one>", false}, {strings.Repeat("x", 65537), false}} {
		if got := expansionXMLChildren(tc.xml); got != tc.want {
			t.Errorf("%q got %v", tc.xml, got)
		}
	}
}

func TestExpansionDOMOwnership(t *testing.T) {
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
			got := ExpansionDOMOwner(ctx, CallArgument(c.Args, 0, "")) != nil
			if got != tc.want {
				t.Errorf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestExpansionSQLSkeleton(t *testing.T) {
	for _, tc := range []struct {
		src   string
		parts []string
		want  bool
	}{
		{`probe("SELECT '$v'");`, []string{"SELECT '", "'"}, true},
		{`probe('SELECT ' . $v . ' WHERE x=3');`, []string{"SELECT ", " WHERE x=3"}, true},
		{`probe('SELECT 3');`, nil, false},
		{"probe(`echo $v`);", nil, false},
		{`probe($v);`, []string{"", ""}, true},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			p, _, ok := ExpansionSQLParts(ctx, CallArgument(c.Args, 0, ""))
			if ok != tc.want || tc.want && strings.Join(p, "|") != strings.Join(tc.parts, "|") {
				t.Errorf("%s got %v %v", tc.src, p, ok)
			}
		})
	}
}

func BenchmarkExpansionXMLChildren(b *testing.B) {
	s := `<catalog><item><name>Amber</name></item></catalog>`
	b.ReportAllocs()
	for b.Loop() {
		expansionXMLChildren(s)
	}
}

func TestExpansionDOMChildrenProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');probe($d->documentElement);`, true},
		{`$d=new DOMDocument();probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$d->loadXML('<x><y/></x>');$d->load('other.xml');probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$d->loadXML($xml);probe($d->documentElement);`, false},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createElement('y'));probe($n);`, true},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createTextNode('y'));probe($n);`, false},
		{`$d=new DOMDocument();$n=$d->createElement('x');$child=$d->createElement('y');$n->appendChild($child);$n->removeChild($child);probe($n);`, false},
		{`$d=new DOMDocument();probe($d->other);`, false},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->hasChildNodes();probe($n);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := ExpansionDOMElementChildren(ctx, CallArgument(c.Args, 0, ""), c); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}

func TestExpansionDOMBeforeWriteProof(t *testing.T) {
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
			if got := ExpansionDOMBeforeWrite(ctx, CallArgument(c.Args, 0, ""), c); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}

func TestExpansionObjectIdentity(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$d=new DOMDocument();probe($d,$d);`, true},
		{`$d=new DOMDocument();$n=$d->createElement('x');probe($n,$n);`, true},
		{`probe($a,$b);`, false},
		{`$n=7;probe($n,$n);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := ExpansionSameObject(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}

func TestExpansionGlobalConstant(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`probe(SQLITE3_NULL);`, true}, {`probe(5);`, false}, {`namespace Custom;const SQLITE3_NULL=4;probe(SQLITE3_NULL);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := ExpansionGlobalConstant(ctx, CallArgument(c.Args, 0, ""), "SQLITE3_NULL"); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}

func TestExpansionDOMDepthBudget(t *testing.T) {
	expr := `$d->createElement("x")`
	for range 18 {
		expr += `->cloneNode(true)`
	}
	probeNative(t, `$d=new DOMDocument();probe(`+expr+`);`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		if got := ExpansionDOMOwner(ctx, CallArgument(c.Args, 0, "")); got != nil {
			t.Fatal("depth budget")
		}
	})
}

func TestExpansionDOMConstructedWriteProof(t *testing.T) {
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
			if got := ExpansionDOMCreatedBeforeWrite(ctx, CallArgument(c.Args, 0, ""), c); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}
