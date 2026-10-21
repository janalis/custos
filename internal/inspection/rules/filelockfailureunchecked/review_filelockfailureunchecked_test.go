package filelockfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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
		// A later case label can be entered below the test.
		{"case below the test", "function f($h,$mode){$ok=flock($h,LOCK_EX);switch($mode){case 1: if(!$ok){return;} case 2: fwrite($h,'x');}}", 1},
		{"stored into a property", "function f($h,$box){$ok=flock($h,LOCK_EX);$box->ok=$ok;if(!$box->ok){return;}fwrite($h,'x');}", 1},
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

func TestAssertionLockAttemptSharesScopeBudget(t *testing.T) {
	probeNative(t, `function f($h){assert(flock($h,LOCK_EX));probe($h);}`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		if name, ok := at.Name.(*syntax.Name); !ok || name.Value != "probe" {
			return
		}
		handle := semanticquery.CallArgument(at.Args, 0, "")
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

func Parse(t *testing.T, src string) *syntax.File {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.Max})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %q: %v", src, f.Errors)
	}
	return f
}

func probeNative(t *testing.T, src string, check func(*analysis.Context, *syntax.FuncCall)) {
	t.Helper()
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
			check(ctx, c)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{}))
}

type nativeProbe struct {
	check func(*analysis.Context, *syntax.FuncCall)
}

func (nativeProbe) ID() string { return "ArrayFlipInvalidValueType" }

func (nativeProbe) Semantic() {}

func (nativeProbe) Flow() {}

func (nativeProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (p nativeProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n.(*syntax.FuncCall)) }
