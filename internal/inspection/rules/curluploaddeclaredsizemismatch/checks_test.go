package curluploaddeclaredsizemismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlUploadDeclaredSizeMismatch"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"\n$h = fopen('data://text/plain,abcdef', 'r'); curl_setopt($ch, CURLOPT_UPLOAD, true); curl_setopt($ch, CURLOPT_INFILE, $h); curl_setopt($ch, CURLOPT_INFILESIZE, 99);\n", 1},
		{"\n$h = fopen('data://text/plain,abcdef', 'r'); curl_setopt($ch, CURLOPT_UPLOAD, true); curl_setopt($ch, CURLOPT_INFILE, $h); curl_setopt($ch, CURLOPT_INFILESIZE, 6);\n", 0},
		{"$h=fopen('data://text/plain,abcdef','r');curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,-1);", 0},
		{"$h=fopen('data://text/plain,abcdef','r');curl_setopt($c,CURLOPT_UPLOAD,false);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0},
		{"$h=fopen('/tmp/x','r');curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0},
		{"$h=fopen('data://text/plain,abcdef','w');curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0},
		{"$h=fopen('data://text/plain,abcdef','r');fread($h,1);curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0},
		{"curl_setopt($c,CURLOPT_INFILESIZE,99);", 0},
		{"curl_setopt($c,CURLOPT_INFILESIZE,$size);", 0},
		{"if(false){\n$h = fopen('data://text/plain,abcdef', 'r'); curl_setopt($ch, CURLOPT_UPLOAD, true); curl_setopt($ch, CURLOPT_INFILE, $h); curl_setopt($ch, CURLOPT_INFILESIZE, 99);\n}", 0},
		{"curl_setopt($c,CURLOPT_UPLOAD,true);curl_setopt($c,CURLOPT_INFILESIZE,99);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
