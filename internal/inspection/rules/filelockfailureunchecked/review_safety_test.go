package filelockfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewLockResultTests pins which tests of a flock result establish the
// acquisition before a write, and which leave the write reported.
func TestReviewLockResultTests(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"stored early return", "function f($h){$ok=flock($h,LOCK_EX);if(!$ok){return;}fwrite($h,'x');}", 0},
		{"stored positive branch", "function f($h){$ok=flock($h,LOCK_EX);if($ok){fwrite($h,'x');}}", 0},
		{"stored loose comparison", "function f($h){$ok=flock($h,LOCK_EX);if($ok!=true){return;}fwrite($h,'x');}", 0},
		{"stored qualified constant", "function f($h){$ok=flock($h,LOCK_EX);if($ok===\\FALSE){return;}fwrite($h,'x');}", 0},
		{"or die", "function f($h){flock($h,LOCK_EX) or die('x');fwrite($h,'x');}", 0},
		{"or exit in parentheses", "function f($h){(flock($h,LOCK_EX)) || (exit(1));fwrite($h,'x');}", 0},
		{"strict false", "function f($h){if(flock($h,LOCK_EX)===false){throw new Exception('x');}fwrite($h,'x');}", 0},
		{"conjunction", "function f($h){return flock($h,LOCK_EX)&&fwrite($h,'x');}", 0},
		{"keyword conjunction", "function f($h){$ok=flock($h,LOCK_EX);$ok and fwrite($h,'x');}", 0},
		{"do while retry", "function f($h){do{$ok=flock($h,LOCK_EX);}while(!$ok);fwrite($h,'x');}", 0},
		{"else terminates", "function f($h){$ok=flock($h,LOCK_EX);if($ok){$n=1;}else{return;}fwrite($h,'x');}", 0},
		{"file level", "$h=fopen('f','cb');$ok=flock($h,LOCK_EX);if(!$ok){exit(1);}fwrite($h,'x');", 0},
		{"braceless nested test", "function f($h,$v){$ok=flock($h,LOCK_EX);if($v)if(!$ok)return;fwrite($h,'x');}", 1},
		{"unchecked", "function f($h){flock($h,LOCK_EX);fwrite($h,'x');}", 1},
		{"stored and ignored", "function f($h){$ok=flock($h,LOCK_EX);fwrite($h,'x');return $ok;}", 1},
		{"assertion only", "function f($h){assert(flock($h,LOCK_EX));fwrite($h,'x');}", 1},
		{"else continues", "function f($h){$ok=flock($h,LOCK_EX);if($ok){$n=1;}else{$n=2;}fwrite($h,'x');}", 1},
		{"else terminates behind elseif", "function f($h,$v){$ok=flock($h,LOCK_EX);if($ok){$n=1;}elseif($v){$n=2;}else{return;}fwrite($h,'x');}", 1},
		{"write on failure branch", "function f($h){if(flock($h,LOCK_EX)){fwrite($h,'x');}else{fwrite($h,'y');}}", 1},
		{"elseif failure body", "function f($h,$v){if($v){return;}elseif(flock($h,LOCK_EX)===false){fwrite($h,'x');}}", 1},
		{"do while positive", "function f($h){do{$ok=flock($h,LOCK_EX);}while($ok);fwrite($h,'x');}", 1},
		{"bitwise negation", "function f($h){$n=~flock($h,LOCK_EX);if(!$n){return;}fwrite($h,'x');}", 1},
		{"cast result", "function f($h){$n=(int)flock($h,LOCK_EX);if(!$n){return;}fwrite($h,'x');}", 1},
		{"stored in ternary branch", "function f($h,$v){$ok=flock($h,LOCK_EX);$r=$v?$ok:false;fwrite($h,'x');return $r;}", 1},
		{"bitwise test", "function f($h){$ok=flock($h,LOCK_EX);if($ok&2){fwrite($h,'x');}}", 1},
		{"null comparison", "function f($h){$ok=flock($h,LOCK_EX);if($ok===null){return;}fwrite($h,'x');}", 1},
		{"write inside retry body", "function f($h){do{$ok=flock($h,LOCK_EX);fwrite($h,'x');}while(!$ok);}", 1},
		{"numeric comparison", "function f($h){$ok=flock($h,LOCK_EX);if($ok===0){return;}fwrite($h,'x');}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FileLockFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
