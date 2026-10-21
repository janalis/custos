package curluploaddeclaredsizemismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The last-write option lookup is invalidated by anything that may change
// the handle in between.
func TestReviewCurlOptionLookupInvalidation(t *testing.T) {
	const open = "$h=fopen('data://text/plain,abcdef','r');"
	const tail = "curl_setopt($c,CURLOPT_INFILE,$h);curl_setopt($c,CURLOPT_INFILESIZE,99);"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"proven upload", open + "curl_setopt($c,CURLOPT_UPLOAD,true);" + tail, 1},
		{"conditional upload", open + "if($f){curl_setopt($c,CURLOPT_UPLOAD,true);}" + tail, 0},
		{"unknown call on handle", open + "curl_setopt($c,CURLOPT_UPLOAD,true);configure($c);" + tail, 0},
		{"other builtin call on handle", open + "curl_setopt($c,CURLOPT_UPLOAD,true);curl_reset($c);" + tail, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlUploadDeclaredSizeMismatch"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("audit.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
