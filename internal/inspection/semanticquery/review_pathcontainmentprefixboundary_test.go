package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/php/version"
)

func TestReviewPathContainmentPrefixBoundary(t *testing.T) {
	const get = "$r=realpath($p);"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"str_starts_with", get + "return str_starts_with($r,'/srv/files');", 1},
		{"str_starts_with named", get + "return str_starts_with(needle:'/srv/files',haystack:$r);", 1},
		{"str_starts_with boundary", get + "return str_starts_with($r,'/srv/files/');", 0},
		{"str_starts_with windows boundary", get + "return str_starts_with($r,'C:\\\\srv\\\\');", 0},
		{"str_starts_with empty prefix", get + "return str_starts_with($r,'');", 0},
		{"str_starts_with dynamic prefix", get + "return str_starts_with($r,$q);", 0},
		{"str_starts_with raw path", "return str_starts_with($p,'/srv/files');", 0},
		{"str_starts_with other origin", "$r=trim($p);return str_starts_with($r,'/srv/files');", 0},
		{"unrelated call", get + "return str_ends_with($r,'/srv/files');", 0},

		{"strpos identical", get + "return strpos($r,'/srv/files')===0;", 1},
		{"strpos yoda", get + "return 0===strpos($r,'/srv/files');", 1},
		{"strpos not identical", get + "if(strpos($r,'/srv/files')!==0){return false;}", 1},
		{"strpos parenthesised", get + "return ((strpos($r,'/srv/files')))===0;", 1},
		{"strpos zero offset", get + "return strpos($r,'/srv/files',0)===0;", 1},
		{"strpos prefix variable", "$b='/srv/files';" + get + "return strpos($r,$b)===0;", 1},
		{"strpos loose", get + "return strpos($r,'/srv/files')==0;", 0},
		{"strpos other offset", get + "return strpos($r,'/srv/files',1)===0;", 0},
		{"strpos dynamic offset", get + "return strpos($r,'/srv/files',$q)===0;", 0},
		{"strpos other position", get + "return strpos($r,'/srv/files')===1;", 0},
		{"strpos compared with false", get + "return strpos($r,'/srv/files')!==false;", 0},
		{"strpos ordering", get + "return strpos($r,'/srv/files')<1;", 0},
		{"strpos bare", get + "return strpos($r,'/srv/files');", 0},
		{"strpos boundary", get + "return strpos($r,'/srv/files/')===0;", 0},
		{"strpos raw path", "return strpos($p,'/srv/files')===0;", 0},
		{"strpos incomplete", get + "return strpos($r,'/srv/files')===", 0},

		{"strncmp identical", get + "return strncmp($r,'/srv/files',10)===0;", 1},
		{"strncmp loose", get + "return strncmp($r,'/srv/files',10)==0;", 1},
		{"strncmp not equal", get + "return strncmp($r,'/srv/files',10)!=0;", 1},
		{"strncmp swapped", get + "return 0===strncmp('/srv/files',$r,10);", 1},
		{"strncmp strlen", get + "return strncmp($r,'/srv/files',strlen('/srv/files'))===0;", 1},
		{"strncmp strlen variable", "$b='/srv/files';" + get + "return strncmp($r,$b,strlen($b))===0;", 1},
		{"strncmp shorter length", get + "return strncmp($r,'/srv/files',4)===0;", 0},
		{"strncmp other strlen", get + "return strncmp($r,'/srv/files',strlen('/srv'))===0;", 0},
		{"strncmp other measure", get + "return strncmp($r,'/srv/files',mb_strlen('/srv/files'))===0;", 0},
		{"strncmp dynamic length", get + "return strncmp($r,'/srv/files',$q)===0;", 0},
		{"strncmp missing length", get + "return strncmp($r,'/srv/files')===0;", 0},
		{"strncmp ordering result", get + "return strncmp($r,'/srv/files',10)===1;", 0},
		{"strncmp negated", get + "return !strncmp($r,'/srv/files',10);", 0},
		{"strncmp boundary", get + "return strncmp($r,'/srv/files/',11)===0;", 0},
		{"strncmp raw path", "return strncmp($p,'/srv/files',10)===0;", 0},

		{"substr identical", get + "return substr($r,0,10)==='/srv/files';", 1},
		{"substr yoda loose", get + "return '/srv/files'==substr($r,0,10);", 1},
		{"substr not identical", get + "return substr($r,0,10)!=='/srv/files';", 1},
		{"substr strlen", "$b='/srv/files';" + get + "return substr($r,0,strlen($b))===$b;", 1},
		{"substr whole string", get + "return substr($r,0)==='/srv/files';", 0},
		{"substr other offset", get + "return substr($r,1,10)==='/srv/files';", 0},
		{"substr other length", get + "return substr($r,0,9)==='/srv/files';", 0},
		{"substr dynamic operand", get + "return substr($r,0,10)===$q;", 0},
		{"substr bare", get + "return substr($r,0,10);", 0},
		{"substr concatenated", get + "return substr($r,0,10).'/srv/files';", 0},
		{"substr boundary", get + "return substr($r,0,11)==='/srv/files/';", 0},
		{"substr raw path", "return substr($p,0,10)==='/srv/files';", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := nativeCheckProbe{"PathContainmentPrefixBoundary", CheckPathContainmentPrefixBoundary}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php function f($p,$q){"+tc.src+"}"), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

// The comparison forms keep the rule effective before str_starts_with exists,
// and highlight the whole comparison rather than the inner call.
func TestReviewPathContainmentPrefixBoundaryOldPHPAndRange(t *testing.T) {
	const src = "<?php function f($p){$r=realpath($p);return strpos($r,'/srv/files')===0||str_starts_with($r,'/srv/files');}"
	p := nativeCheckProbe{"PathContainmentPrefixBoundary", CheckPathContainmentPrefixBoundary}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}, PHP: version.MustParse("5.3")})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("test.php", []byte(src), syntax.Options{}))
	if len(got) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(got), got)
	}
	if text := src[got[0].Span.Start:got[0].Span.End]; text != "strpos($r,'/srv/files')===0" {
		t.Fatalf("unexpected range %q", text)
	}
}
