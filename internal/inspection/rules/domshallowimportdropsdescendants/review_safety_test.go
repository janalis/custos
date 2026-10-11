package domshallowimportdropsdescendants

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const (
		loaded = "$s=new DOMDocument();$s->loadXML('<list><entry/></list>');$d=new DOMDocument();"
		built  = "$s=new DOMDocument();$d=new DOMDocument();$e=$s->createElement('a');$b=$s->createElement('b');$e->appendChild($b);"
	)
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"loaded root", loaded + "$d->importNode($s->documentElement);", 1, loaded + "$d->importNode($s->documentElement, true);"},
		{"loaded root in function", "function f(){" + loaded + "return $d->importNode($s->documentElement);}", 1, "function f(){" + loaded + "return $d->importNode($s->documentElement, true);}"},
		{"node fetched then filled", "$x=new DOMDocument();$d=new DOMDocument();$e=$x->firstChild;$e->appendChild($x->createElement('b'));$d->importNode($e);", 0, ""},
		{"node fetched then reached again", "$x=new DOMDocument();$d=new DOMDocument();$e=$x->firstChild;$e->appendChild($x->createElement('b'));$x->firstChild->textContent='';$d->importNode($e);", 0, ""},
		{"root fetched then filled", loaded + "$e=$s->documentElement;$e->appendChild($s->createElement('b'));$d->importNode($e);", 0, ""},
		{"root fetched", loaded + "$e=$s->documentElement;$d->importNode($e);", 0, ""},
		{"document only read", loaded + "$x=$s->createElement('x');echo $s->saveXML();$d->importNode($s->documentElement);", 1, loaded + "$x=$s->createElement('x');echo $s->saveXML();$d->importNode($s->documentElement, true);"},
		{"document reloaded first", "$s=new DOMDocument();$s->load('a.xml');$s->loadXML('<l><e/></l>');$d=new DOMDocument();$d->importNode($s->documentElement);", 1, "$s=new DOMDocument();$s->load('a.xml');$s->loadXML('<l><e/></l>');$d=new DOMDocument();$d->importNode($s->documentElement, true);"},
		{"use after the import", loaded + "$d->importNode($s->documentElement);strip($s);", 1, loaded + "$d->importNode($s->documentElement, true);strip($s);"},
		{"built element", built + "$d->importNode($e);", 1, built + "$d->importNode($e, true);"},
		{"built element decorated", built + "$e->appendChild($s->createElement('c'));$e->setAttribute('k','v');$b->setAttribute('k','v');$b=null;$d->importNode($e);", 1, built + "$e->appendChild($s->createElement('c'));$e->setAttribute('k','v');$b->setAttribute('k','v');$b=null;$d->importNode($e, true);"},
		{"built element attached", built + "$s->appendChild($e);$d->importNode($e);", 1, built + "$s->appendChild($e);$d->importNode($e, true);"},
		{"built element imported twice", built + "$d->importNode($e, true);$d->importNode($e);", 1, built + "$d->importNode($e, true);$d->importNode($e, true);"},
		{"arrow function elsewhere", built + "$g=fn($e)=>1;function h($e){prune($e);}$k=function($e){prune($e);};$d->importNode($e);", 0, ""},
		{"nested scopes ignored", built + "function h($q){$q=$q->e;prune($q);}$k=function($q){prune($q);};$d->importNode($e);", 1, built + "function h($q){$q=$q->e;prune($q);}$k=function($q){prune($q);};$d->importNode($e, true);"},

		{"element passed to a call", built + "prune($e);$d->importNode($e);", 0, ""},
		{"element passed to a method", built + "$o->prune($e);$d->importNode($e);", 0, ""},
		{"element passed to a constructor", built + "new Pruner($e);$d->importNode($e);", 0, ""},
		{"element passed as second argument", built + "$s->insertBefore($x, $e);$d->importNode($e);", 0, ""},
		{"element attached to an element", built + "$p=$s->createElement('p');$p->appendChild($e);$d->importNode($e);", 0, ""},
		{"element aliased", built + "$t=$e;$t->textContent='';$d->importNode($e);", 0, ""},
		{"element bound by reference", built + "$t=&$e;$d->importNode($e);", 0, ""},
		{"element from a variable", "$s=new DOMDocument();$d=new DOMDocument();$o=$s->createElement('a');$e=$o;$e->appendChild($s->createElement('b'));$o->textContent='';$d->importNode($e);", 0, ""},
		{"element property written", built + "$e->textContent='';$d->importNode($e);", 0, ""},
		{"element child read", built + "$e->firstChild->remove();$d->importNode($e);", 0, ""},
		{"element emptied", built + "$e->replaceChildren();$d->importNode($e);", 0, ""},
		{"element child removed on a branch", built + "if($c){$e->removeChild($b);}$d->importNode($e);", 0, ""},
		{"element as method name", built + "$o->$e();$d->importNode($e);", 0, ""},
		{"element as property name", built + "$o->$e;$d->importNode($e);", 0, ""},
		{"element captured", built + "$g=function() use($e){$e->textContent='';};$g();$d->importNode($e);", 0, ""},
		{"element changed later in a loop", built + "while($c){$d->importNode($e);$e->textContent='';}", 0, ""},
		{"element changed later in other loops", built + "foreach($l as $i){for(;;){do{$d->importNode($e);}while($c);}}$e->textContent='';", 0, ""},
		{"child detaches itself", built + "$b->remove();$d->importNode($e);", 0, ""},
		{"child moved elsewhere", built + "$o->appendChild($b);$d->importNode($e);", 0, ""},
		{"child passed to a call", built + "detach($b);$d->importNode($e);", 0, ""},
		{"child text written", built + "$e->appendChild($s->createElement('c'));$b->textContent='t';$d->importNode($e);", 0, ""},
		{"child aliased", built + "$t=$b;$t->remove();$d->importNode($e);", 0, ""},
		{"child appended to", built + "$b.='x';$d->importNode($e);", 0, ""},
		{"parenthesised root", loaded + "$d->importNode(($s->documentElement));", 0, ""},
		{"parenthesised document", loaded + "$d->importNode(($s)->documentElement);", 0, ""},
		{"parenthesised element", built + "$d->importNode(($e));", 0, ""},
		{"child parent read", built + "$b->parentNode->removeChild($b);$d->importNode($e);", 0, ""},
		{"child is the element", "$s=new DOMDocument();$d=new DOMDocument();$e=$s->createElement('a');$e->appendChild($s->createElement('b'));$e->appendChild($e);$d->importNode($e);", 0, ""},
		{"root children removed", loaded + "$s->documentElement->removeChild($s->documentElement->firstChild);$d->importNode($s->documentElement);", 0, ""},
		{"root text written", loaded + "$s->documentElement->textContent='';$d->importNode($s->documentElement);", 0, ""},
		{"root replaced", loaded + "$s->replaceChild($s->createElement('x'), $s->documentElement);$d->importNode($s->documentElement);", 0, ""},
		{"root bound elsewhere", loaded + "$r=$s->documentElement;$r->textContent='';$d->importNode($s->documentElement);", 0, ""},
		{"root bound by reference", loaded + "$e=$s->documentElement;$r=&$s->documentElement;$d->importNode($e);", 0, ""},
		{"other document property", loaded + "$s->firstChild->textContent='';$d->importNode($s->documentElement);", 0, ""},
		{"dynamic document property", loaded + "$s->$p->textContent='';$d->importNode($s->documentElement);", 0, ""},
		{"document reloaded on a branch", loaded + "if($c){$s->loadXML('<a/>');}$d->importNode($s->documentElement);", 0, ""},
		{"document reloaded as html", loaded + "$s->loadHTML('<p>x</p>');$s->documentElement->replaceChildren();$d->importNode($s->documentElement);", 0, ""},
		{"document passed to a call", loaded + "strip($s);$d->importNode($s->documentElement);", 0, ""},
		{"document aliased", loaded + "$t=$s;$t->loadXML('<z/>');$d->importNode($s->documentElement);", 0, ""},
		{"document replaced by a call", loaded + "$e=$s->documentElement;$s=other();$d->importNode($e);", 0, ""},
		{"document is also the element", loaded + "$s=$s->documentElement;$d->importNode($s);", 0, ""},
		{"document behind a property", "$this->s=new DOMDocument();$this->s->loadXML('<list><entry/></list>');$d=new DOMDocument();$d->importNode($this->s->documentElement);", 0, ""},
		{"root behind a property", "$o->s=new DOMDocument();$o->s->loadXML('<list><entry/></list>');$e=$o->s->documentElement;$e->appendChild($o->s->createElement('b'));$d=new DOMDocument();$d->importNode($e);", 0, ""},
		{"created inline", "$s=new DOMDocument();$d=new DOMDocument();$d->importNode($s->createElement('a'));", 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte("<?php " + tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != "<?php "+want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
