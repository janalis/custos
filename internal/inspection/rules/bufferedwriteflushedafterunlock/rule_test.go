package bufferedwriteflushedafterunlock

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"<?php\nflock($h, LOCK_EX); fwrite($h, $data); flock($h, LOCK_UN); fflush($h);\n", 1},
		{"<?php\nflock($h, LOCK_EX); fwrite($h, $data); fflush($h); flock($h, LOCK_UN);\n", 0},
		{"<?php other();", 0},
		{"<?php fflush($h);", 0},
		{"<?php fflush(open());", 0},
		{"<?php flock($h,LOCK_EX);fwrite($h,$s);flock($h,$flags);fflush($h);", 0},
		{"<?php flock($h,LOCK_SH);fwrite($h,$s);flock($h,LOCK_UN);fflush($h);", 0},
		{"<?php flock($h,LOCK_EX);fwrite($h,$s);flock($h,LOCK_UN);fclose($h);fflush($h);", 0},
		{"<?php flock($h,LOCK_EX);fwrite($h,$s);flock($h,LOCK_UN);$h=$other;fflush($h);", 0},
		{"<?php flock($h,LOCK_EX);fwrite($h,$s);flock($h,LOCK_UN);$x=&$h;fflush($h);", 0},
		{"<?php flock($h,LOCK_EX);fwrite($h,$s);flock($h,LOCK_UN);unset($h);fflush($h);", 0},
		{"<?php flock($h,LOCK_EX);fwrite($h,$s);flock($h,LOCK_UN);unknown($h);fflush($h);", 0},
		{"<?php flock($h,LOCK_EX);fwrite($h,$s);flock($h,LOCK_UN);$o->x();fflush($h);", 0},
		{"<?php flock($h,LOCK_EX);fwrite($h,$s);if($ok){flock($h,LOCK_UN);}fflush($h);", 0},
		{"<?php flock($other,LOCK_EX);fwrite($h,$s);flock($h,LOCK_UN);fflush($h);", 0},
		{"<?php function unreachable(){return;\nflock($h, LOCK_EX); fwrite($h, $data); flock($h, LOCK_UN); fflush($h);\n}", 0},
		{"<?php \nflock($h, LOCK_EX); fwrite($h, $data); flock($h, LOCK_UN); fflush($h);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"BufferedWriteFlushedAfterUnlock"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
