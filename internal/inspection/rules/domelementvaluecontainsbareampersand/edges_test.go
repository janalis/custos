package domelementvaluecontainsbareampersand

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{`function loadXML($doc) {} $d=new DOMDocument();loadXML($d);$d->createElement('x','&author;');`, 1},

		{`class CustomDocument extends DOMDocument {function loadXML($source,$options=0):bool{return true;}} $d=new CustomDocument();$d->loadXML('<x/>');$d->createElement('x','&author;');`, 1},
		{`$d=new DOMDocument();$d->loadXML('<!SOMETHING text><x/>');$d->createElement('x','&author;');`, 1},

		{`$d=new DOMDocument();$d->loadXML('<!DOCTYPE x [<!ENTITY author "Bex">]><x/>');$d->createElement('x','&author;');`, 0},
		{`$d=new DOMDocument();$d->loadXML('<x/>');$d->createElement('x','&author;');`, 1},
		{`$d=new DOMDocument();$d->loadXML($xml);$d->createElement('x','&author;');`, 1},
		{`$d=new DOMDocument();$d->loadXML('<!DOCTYPE x SYSTEM "x.dtd"><x/>');$d->createElement('x','&author;');`, 1},
		{`$d=new DOMDocument();$d->loadXML('<!DOCTYPE x [<!ELEMENT x EMPTY>]><x/>');$d->createElement('x','&author;');`, 1},
		{`$d=new DOMDocument();$d->loadXML('bad');$d->createElement('x','&author;');`, 1},
		{`$d=new DOMDocument();$d->loadXML('<?xml version="1.0"?><!DOCTYPE x [<!ENTITY other "Other">]><x/>');$d->createElement('x','&author;');`, 1},

		{"$d=new DOMDocument();$d->createElement('x','&amp; &#65; &#x41;');", 0},
		{"$d=new DOMDocument();$d->createElement('x','&unknown;');", 1},
		{"$d=new DOMDocument();$d->createElement('x','&#0;');", 1},
		{"$d=new DOMDocument();$d->createElement('x',$unknown);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DomElementValueContainsBareAmpersand"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edges.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
