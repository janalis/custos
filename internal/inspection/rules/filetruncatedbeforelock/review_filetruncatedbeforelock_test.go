package filetruncatedbeforelock

import (
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewFileTruncatedBeforeLock(t *testing.T) {
	const open = "function f($p,$op){$h=fopen($p,'w');"
	// Every case writes through the handle (E4) unless it is raw.
	for _, tc := range []struct {
		name, src string
		want      int
		raw       bool
	}{
		{"exclusive", open + "flock($h,LOCK_EX);}", 1, false},
		{"qualified exclusive", open + "\\flock($h,\\LOCK_EX);}", 1, false},
		{"exclusive non-blocking", open + "flock($h,LOCK_EX|LOCK_NB);}", 1, false},
		{"parenthesized non-blocking", open + "flock($h,(LOCK_NB|LOCK_EX));}", 1, false},
		{"integer exclusive", open + "flock($h,2);}", 1, false},
		{"integer exclusive non-blocking", open + "flock($h,6);}", 1, false},
		{"named operation", open + "flock(operation:LOCK_EX,stream:$h);}", 1, false},
		{"local operation", open + "$mode=LOCK_EX|LOCK_NB;flock($h,$mode);}", 1, false},
		{"exclusive after unlock", open + "flock($h,LOCK_UN);flock($h,LOCK_EX);}", 1, false},
		{"unlock only", open + "fwrite($h,'x');flock($h,LOCK_UN);}", 0, false},
		{"unlock non-blocking", open + "flock($h,LOCK_UN|LOCK_NB);}", 0, false},
		{"shared only", open + "flock($h,LOCK_SH);}", 0, false},
		{"shared non-blocking", open + "flock($h,LOCK_SH|LOCK_NB);}", 0, false},
		{"integer shared", open + "flock($h,1);}", 0, false},
		{"integer unlock", open + "flock($h,3);}", 0, false},
		{"dynamic operation", open + "flock($h,$op);}", 0, false},
		{"partly dynamic operation", open + "flock($h,LOCK_EX|$op);}", 0, false},
		{"other operator", open + "flock($h,LOCK_EX&$op);}", 0, false},
		{"missing operation", open + "flock($h);}", 0, false},
		{"other handle", open + "$g=fopen($p,'c');flock($g,LOCK_EX);}", 0, false},
		{"lock file only", open + "flock($h,LOCK_EX);fclose($h);}", 0, true},
		{"read through it", open + "flock($h,LOCK_EX);echo fgets($h);}", 1, true},
		{"written by name", open + "flock($h,LOCK_EX);fwrite(data:'x',stream:$h);}", 1, true},
		{"other stream written", open + "flock($h,LOCK_EX);fwrite($g,'x');fwrite(...$all);}", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := tc.src
			if !tc.raw {
				src = src[:len(src)-1] + "fwrite($h,'x');}"
			}
			got := analyzeReview(t, src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func analyzeReview(t *testing.T, src string) []diagnostic.Finding {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FileTruncatedBeforeLock"}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{}))
}
