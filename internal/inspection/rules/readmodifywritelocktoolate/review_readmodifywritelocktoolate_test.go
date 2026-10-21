package readmodifywritelocktoolate

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewReadModifyWriteLockTooLateSpanningLock(t *testing.T) {
	const rmw = "$n=(int)file_get_contents($p);file_put_contents($p,$n+1,LOCK_EX);"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"no lock", "function f($p){" + rmw + "}", 1},
		{"top level", "$p='/tmp/c';" + rmw, 1},
		{"top level lock", "$p='/tmp/c';$h=fopen('/tmp/l','c');flock($h,LOCK_EX);" + rmw, 0},
		{"exclusive lock", "function f($p,$h){flock($h,LOCK_EX);" + rmw + "flock($h,LOCK_UN);}", 0},
		{"non-blocking lock", "function f($p,$h){if(!flock($h,LOCK_EX|LOCK_NB)){return;}" + rmw + "}", 0},
		{"integer operation", "function f($p,$h){flock($h,2);" + rmw + "}", 0},
		{"unknown operation", "function f($p,$h,$op){flock($h,$op);" + rmw + "}", 0},
		{"member handle", "class C{function f($p){flock($this->h,LOCK_EX);" + rmw + "}}", 0},
		{"shared lock", "function f($p,$h){flock($h,LOCK_SH);" + rmw + "}", 1},
		{"unlock only", "function f($p,$h){flock($h,LOCK_UN);" + rmw + "}", 1},
		{"missing handle", "function f($p){flock();" + rmw + "}", 1},
		{"other builtin", "function f($p,$h){rewind($h);" + rmw + "}", 1},
		{"lock after read", "function f($p,$h){$n=(int)file_get_contents($p);flock($h,LOCK_EX);file_put_contents($p,$n+1,LOCK_EX);}", 1},
		{"lock after write", "function f($p,$h){" + rmw + "flock($h,LOCK_EX);}", 1},
		{"released before read", "function f($p,$h){flock($h,LOCK_EX);flock($h,LOCK_UN);" + rmw + "}", 1},
		{"closed before read", "function f($p,$h){flock($h,LOCK_EX);fclose($h);" + rmw + "}", 1},
		{"downgraded before read", "function f($p,$h){flock($h,LOCK_EX);flock($h,LOCK_SH);" + rmw + "}", 1},
		{"released before write", "function f($p,$h){flock($h,LOCK_EX);$n=(int)file_get_contents($p);flock($h,LOCK_UN);file_put_contents($p,$n+1,LOCK_EX);}", 1},
		{"other handle released", "function f($p,$h,$g){flock($h,LOCK_EX);flock($g,LOCK_UN);fclose($g);" + rmw + "}", 0},
		{"one of two released", "function f($p,$h,$g){flock($h,LOCK_EX);flock($g,LOCK_EX);fclose($g);" + rmw + "}", 0},
		{"release in leaving branch", "function f($p,$h){if(!flock($h,LOCK_EX)){fclose($h);return;}" + rmw + "}", 0},
		{"braceless failure close", "function f($p,$h){if(!flock($h,LOCK_EX))fclose($h);" + rmw + "}", 0},
		{"release in finally", "function f($p,$h){flock($h,LOCK_EX);try{" + rmw + "}finally{flock($h,LOCK_UN);fclose($h);}}", 0},
		{"earlier locked section", "function f($p,$h,$c){if($c){flock($h,LOCK_EX);work();flock($h,LOCK_UN);}" + rmw + "}", 1},
		{"tested release", "function f($p,$h){flock($h,LOCK_EX);if(!flock($h,LOCK_UN)){return;}" + rmw + "}", 1},
		{"nested scope lock", "function f($p,$h){$g=function()use($h){flock($h,LOCK_EX);};" + rmw + "}", 1},
		{"inline read", "function f($p,$h){flock($h,LOCK_EX);file_put_contents($p,file_get_contents($p).'x',LOCK_EX);}", 0},
		{"two reads", "function f($p,$h){$a=file_get_contents($p);flock($h,LOCK_EX);$b=file_get_contents($p);file_put_contents($p,$b.$a,LOCK_EX);}", 1},
		{"two reads reversed", "function f($p,$h){$a=file_get_contents($p);flock($h,LOCK_EX);$b=file_get_contents($p);file_put_contents($p,$a.$b,LOCK_EX);}", 1},
		{"lock method", "function f($p,$m){$m->lock();" + rmw + "$m->unlock();}", 0},
		{"static lock", "function f($p){Mutex::acquireLock('counter');" + rmw + "}", 0},
		{"lock function", "function f($p){take_lock();" + rmw + "}", 0},
		{"unlock method", "function f($p,$m){$m->unlock();" + rmw + "}", 1},
		{"path method", "function f($p,$m){$m->enter($p);" + rmw + "}", 0},
		{"literal path method", "function f($m){$m->enter('/tmp/c');$n=(int)file_get_contents('/tmp/c');file_put_contents('/tmp/c',$n+1,LOCK_EX);}", 0},
		{"literal path function", "function f(){enter(1, '/tmp/c');$n=(int)file_get_contents('/tmp/c');file_put_contents('/tmp/c',$n+1,LOCK_EX);}", 0},
		{"unrelated calls", "function f($p,$m,$q){$m->log($q);Audit::note($q);note($q);$m->touch(...);" + rmw + "}", 1},
		{"lock call after read", "function f($p,$m){$n=(int)file_get_contents($p);$m->lock();file_put_contents($p,$n+1,LOCK_EX);}", 1},
		{"dynamic callee", "function f($p,$m,$fn){$m->$fn();$fn();" + rmw + "}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := New()
			engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{probe.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			got := engine.Analyze(Parse(t, "<?php "+tc.source))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func Parse(t *testing.T, src string) *syntax.File {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.Max})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %q: %v", src, f.Errors)
	}
	return f
}
