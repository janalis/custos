package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewFileLockFailureUncheckedResultTests(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"not a call", "foreach ([] as $entry) {}", 0},
		{"reader", "function f($h){flock($h,LOCK_EX);fread($h,1);}", 0},
		{"unchecked", "function f($h){flock($h,LOCK_EX);fwrite($h,'x');}", 1},
		{"stored guard", "function f($h){$ok=flock($h,LOCK_EX);if(!$ok){return;}ftruncate($h,0);}", 0},
		{"reassigned guard", "function f($h){$ok=flock($h,LOCK_EX);$ok=true;if(!$ok){return;}ftruncate($h,0);}", 1},
		{"terminating alternative", "function f($h){flock($h,LOCK_EX) or exit(1);fwrite($h,'x');}", 0},
		{"continuing alternative", "function f($h){flock($h,LOCK_EX) or print('x');fwrite($h,'x');}", 1},
		{"retry", "function f($h){do{$ok=flock($h,LOCK_EX);}while($ok!==true);fwrite($h,'x');}", 0},
		{"compared ternary", "function f($h){return flock($h,LOCK_EX)===true?fwrite($h,'x'):0;}", 0},
		{"abandoned retry", "function f($h,$n){do{$ok=flock($h,LOCK_EX);if(!--$n){break;}}while($ok!==true);fwrite($h,'x');}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := nativeStreamProbe{"FileLockFailureUnchecked", CheckFileLockFailureUnchecked}
			engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{probe.id}})
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

// The flow engine currently hides these shapes from the check, so the guard
// walk is exercised on the syntax tree directly.
func TestReviewFileLockFailureUncheckedGuardWalk(t *testing.T) {
	variables := func(f *syntax.File, name string) []*syntax.Variable {
		var found []*syntax.Variable
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok && v.Name == name {
				found = append(found, v)
			}
			return true
		})
		return found
	}
	held := nativeLockTruth{whenFalse: true}

	f := Parse(t, "<?php $ok = 1; $box->ok = $ok; $sum += $ok;")
	for _, use := range variables(f, "ok") {
		if stored, _, guarded := nativeLockTruthGuards(use, held, f.Stmts[2]); stored != nil || guarded {
			t.Fatalf("assignment at %d must not carry the lock result", use.Span().Start)
		}
	}

	f = Parse(t, "<?php switch ($mode) { case 1: if (!$ok) { return; } case 2: fwrite($h, 'x'); }")
	write := f.Stmts[0].(*syntax.Switch).Cases[1].Stmts[0]
	if _, _, guarded := nativeLockTruthGuards(variables(f, "ok")[0], nativeLockTruth{whenTrue: true}, write); guarded {
		t.Fatal("a later case label can be entered below the test")
	}
}
