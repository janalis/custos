package untrustedldapfilter

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestAuditRegressions exercises independent valid programs and comment-preservation boundaries.
func TestAuditRegressions(t *testing.T) {
	for _, tc := range []struct {
		src         string
		want, fixes int
	}{
		{"$_GET[\"q\"]=ldap_escape($_GET[\"q\"],\"\",LDAP_ESCAPE_FILTER);ldap_search($ldap,$base,$_GET[\"q\"]);", 0, -1},
		// Interpolated request values reach the filter like concatenated ones.
		{"ldap_search($l,$b,\"(uid={$_GET['u']})\");", 1, 0},
		{"ldap_search($l,$b,\"(uid=$_GET[u])\");", 1, 0},
		{"ldap_search($l,$b,\"(uid=${_GET['u']})\");", 1, 0},
		{"ldap_search($l,$b,<<<F\n(uid={$_POST['u']})\nF\n);", 1, 0},
		{"$f=\"(uid={$_GET['u']})\";ldap_search($l,$b,$f);", 1, 0},
		{"ldap_search($l,$b,\"(uid={$_GET[$k]})\");", 0, -1},
		{"ldap_search($l,$b,`id {$_GET['u']}`);", 0, -1},
		// Only filter-context escaping protects the filter.
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',LDAP_ESCAPE_DN).')');", 1, 0},
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',2).')');", 1, 0},
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',LDAP_ESCAPE_DN|4).')');", 1, 0},
		{"ldap_search($l,$b,\"(uid=\".ldap_escape(\"{$_GET['u']}\",'',LDAP_ESCAPE_DN).\")\");", 1, 0},
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u']).')');", 0, -1},
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',0).')');", 0, -1},
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',LDAP_ESCAPE_FILTER|LDAP_ESCAPE_DN).')');", 0, -1},
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',3).')');", 0, -1},
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',$flags).')');", 0, -1},
		{"ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',LDAP_ESCAPE_DN|$flags).')');", 0, -1},
		{"ldap_search($l,$b,'(uid='.ldap_escape('x','',LDAP_ESCAPE_DN).')');", 0, -1},
		{"ldap_search($l,$b,'(uid='.ldap_escape(...$args).')');", 0, -1},
		{"ldap_search($l,$b,'(uid='.trim($_GET['u']).')');", 0, -1},
		{"namespace App;function ldap_escape($v,$i,$f){return '';}ldap_search($l,$b,'(uid='.ldap_escape($_GET['u'],'',LDAP_ESCAPE_DN).')');", 0, -1},
		// Server-generated entries are not client text; other keys still are.
		{"ldap_search($l,$b,'(t<='.$_SERVER['REQUEST_TIME'].')');", 0, -1},
		{"ldap_search($l,$b,'(t<='.$_SERVER['REQUEST_TIME_FLOAT'].')');", 0, -1},
		{"ldap_search($l,$b,\"(p={$_SERVER['SERVER_PORT']})\");", 0, -1},
		{"ldap_search($l,$b,\"(p=$_SERVER[argc])\");", 0, -1},
		{"ldap_search($l,$b,'(f='.$_SERVER['SCRIPT_FILENAME'].')');", 0, -1},
		{"ldap_search($l,$b,'(h='.$_SERVER['HTTP_HOST'].')');", 1, 0},
		{"ldap_search($l,$b,'(h='.$_SERVER['SERVER_NAME'].')');", 1, 0},
		{"ldap_search($l,$b,'(h='.$_SERVER['REMOTE_USER'].')');", 1, 0},
		{"ldap_search($l,$b,'(h='.$_SERVER['CONTENT_TYPE'].')');", 1, 0},
		{"ldap_search($l,$b,'(t='.$_GET['REQUEST_TIME'].')');", 1, 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("audit.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("bad regression source: %+v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d findings want %d: %+v", len(got), tc.want, got)
			}
			if tc.fixes >= 0 {
				count := 0
				for _, d := range got {
					count += len(d.Fixes)
				}
				if count != tc.fixes {
					t.Fatalf("got %d fixes want %d", count, tc.fixes)
				}
			}
		})
	}
}
