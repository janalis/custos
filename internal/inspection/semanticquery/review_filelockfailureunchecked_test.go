package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
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
		// A comparison with `true` is read from the statement alone.
		{"terminating alternative of a comparison", "function f($h){flock($h,LOCK_EX)===true or exit(1);fwrite($h,'x');}", 0},
		{"terminating alternative of a comparison at file level", "$h=fopen('f','c');flock($h,LOCK_EX)===true or exit(1);fwrite($h,'x');", 0},
		{"stored result compared at file level", "$h=fopen('f','c');$ok=flock($h,LOCK_EX);if($ok===true){fwrite($h,'x');}", 0},
		{"continuing alternative of a comparison", "function f($h){flock($h,LOCK_EX)===true or print('x');fwrite($h,'x');}", 1},
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

func TestAssertionLockAttemptSharesScopeBudget(t *testing.T) {
	probeNative(t, `function f($h){assert(flock($h,LOCK_EX));probe($h);}`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		if name, ok := at.Name.(*syntax.Name); !ok || name.Value != "probe" {
			return
		}
		handle := CallArgument(at.Args, 0, "")
		if _, known := assertionLockAttempt(ctx, at, handle); !known {
			t.Fatal("missing initial lock witness")
		}
		scopes := ctx.Memo("file-lock.assertion-attempts", nil).(map[syntax.Node]*assertionLockIndex)
		ix := scopes[syntax.EnclosingVariableScope(at)]
		for _, used := range []int{flow.MaxTransfers, flow.MaxTransfers - 1} {
			ix.work = used
			if _, known := assertionLockAttempt(ctx, at, handle); known {
				t.Fatal("exhausted work budget retained witness")
			}
		}
	})
}
