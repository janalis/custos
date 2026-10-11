package bufferedwriteflushedafterunlock

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const locked = "flock($h,LOCK_EX);fwrite($h,$s);"
	const tail = "flock($h,LOCK_UN);fflush($h);"
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"closure capture called", "<?php $f=function() use($h){fflush($h);};" + locked + "$f();" + tail, 0},
		{"closure capture by reference", "<?php $f=function() use(&$h){fflush($h);};" + locked + "$f();" + tail, 0},
		{"closure capture among others", "<?php $f=function() use($a,$h){fflush($h);};" + locked + "$f();" + tail, 0},
		{"arrow capture called", "<?php $f=fn()=>fflush($h);" + locked + "$f();" + tail, 0},
		{"nested arrow capture", "<?php $f=fn()=>fn()=>fflush($h);" + locked + "$f()();" + tail, 0},
		{"capture inside function", "<?php function w($h,$s){$f=function() use($h){fflush($h);};" + locked + "$f();" + tail + "}", 0},
		{"second argument", "<?php " + locked + "done(1,$h);" + tail, 0},
		{"parenthesized argument", "<?php " + locked + "done(1,($h));" + tail, 0},
		{"named argument", "<?php " + locked + "done(target: $h);" + tail, 0},
		{"static call argument", "<?php " + locked + "Sink::done($h);" + tail, 0},
		{"constructor argument", "<?php " + locked + "new Sink($h);" + tail, 0},
		{"variable variable", "<?php flock($$a,LOCK_EX);fwrite($$a,$s);flock($$a,LOCK_UN);fflush($$b);", 0},
		{"unrelated closure", "<?php $f=function() use($o){fflush($o);};" + locked + "$f();" + tail, 1},
		{"closure without capture", "<?php $f=function($h){fflush($h);};" + locked + tail, 1},
		{"unrelated arrow", "<?php $f=fn()=>strlen($s);" + locked + tail, 1},
		{"named function with same name", "<?php function g($h){fflush($h);}" + locked + tail, 1},
		{"handed over after flush", "<?php " + locked + tail + "done(1,$h);", 1},
		{"parenthesized stream", "<?php flock(($h),LOCK_EX);fwrite(($h),$s);flock(($h),LOCK_UN);fflush(($h));", 1},
		{"other variable argument", "<?php " + locked + "done(1,$o);" + tail, 1},
		{"read outside calls", "<?php " + locked + "$ok=$h!==false;" + tail, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
