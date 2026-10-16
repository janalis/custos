package lsp

import (
	"io"
	"path/filepath"
	"testing"

	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
)

func TestUnsavedSinkProjectionsAndDispatchRefresh(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		root := t.TempDir()
		s := whiteBox(t, root, io.Discard, untrustedshellcommand.New())
		caller, callee := "a.php", "z.php"
		if reverse {
			caller, callee = callee, caller
		}
		outer := filepath.Join(root, caller)
		inner := filepath.Join(root, callee)
		s.docs["file://"+outer] = &document{path: outer, text: []byte(`<?php function invoke($rows){inner($rows);} function dispatch(C $o,$text){$o->send($text);}`)}
		uri := "file://" + inner
		for i, tc := range []struct {
			inner, main string
			want        int
		}{
			{`function inner($rows){system(implode(' ',$rows));} final class C {function send($x){system($x);}}`, `$rows=array_map('intval',$_GET['rows']);invoke($rows);`, 0},
			{`function inner($rows){system(implode(' ',array_keys($rows)));} final class C {function send($x){system($x);}}`, `$rows=array_map('intval',$_GET['rows']);invoke($rows);`, 1},
			{`function inner($rows){system(implode(' ',$rows));} final class C {function send($x){system($x);}}`, `function useIt(C $o){dispatch($o,$_GET['command']);}`, 1},
			{`function inner($rows){system(implode(' ',$rows));} class C {function send($x){system($x);}}`, `function useIt(C $o){dispatch($o,$_GET['command']);}`, 0},
			{`function inner($rows){system(implode(' ',$rows));} class C {final function send($x){system($x);}}`, `function useIt(C $o){dispatch($o,$_GET['command']);}`, 1},
		} {
			s.docs[uri] = &document{path: inner, text: []byte("<?php " + tc.inner)}
			if i == 0 {
				if !s.beginIndexing() {
					t.Fatal("index not requested")
				}
				s.buildIndex()
			} else if !s.reindexDoc(uri) {
				t.Fatal("unsaved dependency not reindexed")
			}
			f := syntax.Parse("main.php", []byte("<?php "+tc.main), syntax.Options{Version: s.cfg.PHP})
			if got := len(s.engine.Analyze(f)); got != tc.want {
				t.Fatalf("reverse=%v source=%s findings=%d want=%d", reverse, tc.inner, got, tc.want)
			}
		}
	}
}
