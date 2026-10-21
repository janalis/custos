package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeUnbracedStatementDoesNotDominate(t *testing.T) {
	probeNative(t, "if ($flag) earlier(); probe();", func(ctx *analysis.Context, at *syntax.FuncCall) {
		var earlier *syntax.FuncCall
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			if call, ok := n.(*syntax.FuncCall); ok {
				if name, ok := call.Name.(*syntax.Name); ok && name.Value == "earlier" {
					earlier = call
				}
			}
			return true
		})
		if earlier == nil || NativeDominates(earlier, at) {
			t.Fatal("conditional unbraced statement established unconditional dominance")
		}
	})
}

func TestNativePriorResolvedCalls(t *testing.T) {
	for _, tc := range []struct {
		source string
		count  int
		known  bool
	}{
		{"probe($unknown);", 0, false},
		{"$h=tmpfile();flock($h,LOCK_EX);probe($h);", 1, true},
		{"$h=fopen('path','rb'); flock($h,LOCK_EX); probe($h);", 1, true},
		{"$h=fopen('path','rb'); if($flag) flock($h,LOCK_EX); probe($h);", 0, false},
		{"$h=fopen('path','rb'); $other=fopen('other','rb'); flock($other,LOCK_EX); probe($h);", 0, true},
		{"$h=fopen('path','rb'); probe($h); flock($h,LOCK_EX);", 0, true},
		{"namespace App; function flock($handle,$mode){} $h=\\fopen('path','rb'); flock($h,LOCK_EX); probe($h);", 0, false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			probeNative(t, tc.source, func(ctx *analysis.Context, at *syntax.FuncCall) {
				calls, known := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, "value"), "flock").Certain()
				if len(calls) != tc.count || known != tc.known {
					t.Fatalf("prior resolved calls: %d/%v, want %d/%v", len(calls), known, tc.count, tc.known)
				}
			})
		})
	}
}

func TestNativeFinallyConditionalAndShadowedCleanup(t *testing.T) {
	for _, source := range []string{
		"$h=fopen('path','rb'); try{probe($h);}finally{if($flag) fclose($h);}",
		"namespace App; function fclose($handle){} $h=\\fopen('path','rb'); try{probe($h);}finally{fclose($h);}",
		"$h=fopen('path','rb'); try{probe($h);}finally{$h=fopen('other','rb'); fclose($h);}",
	} {
		t.Run(source, func(t *testing.T) {
			probeNative(t, source, func(ctx *analysis.Context, at *syntax.FuncCall) {
				if CalledInFinally(ctx, at, CallArgument(at.Args, 0, "value"), "fclose") {
					t.Fatal("cleanup without a proven unconditional matching contract")
				}
			})
		})
	}
}
