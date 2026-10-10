package flow

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestFixedScalarTextCannotInjectSyntax(t *testing.T) {
	for _, expression := range []string{
		`$_GET['x'] === 'yes'`, `$_GET['x'] != 'yes'`, `$_GET['x'] <=> 2`,
		`$_GET['x'] && true`, `$_GET['x'] xor false`, `!$_GET['x']`,
		`(int)$_GET['x']`, `(float)$_GET['x']`, `(bool)$_GET['x']`,
		`intval($_GET['x'])`, `floatval($_GET['x'])`, `isset($_GET['x'])`,
		`empty($_GET['x'])`, `$_GET['x'] instanceof stdClass`,
	} {
		t.Run(expression, func(t *testing.T) {
			e, f := environment(t, "probe("+expression+");", nil)
			x := argument(callsNamed(f, "probe")[0])
			for _, context := range []Context{HTML, SQL, Shell, Header} {
				if e.Tainted(x, context) {
					t.Fatalf("fixed scalar text retained injectable context %d", context)
				}
			}
			if !e.Tainted(x, Path) {
				t.Fatal("scalar conversion discarded input-dependent path choice")
			}
		})
	}
}

func TestDigestHexAndBinaryOutputRemainDistinct(t *testing.T) {
	for _, tc := range []struct {
		expression string
		tainted    bool
	}{
		{`md5($_GET['x'])`, false},
		{`sha1($_GET['x'], false)`, false},
		{`hash('sha256', $_GET['x'], 0)`, false},
		{`hash_hmac('sha256', $_GET['x'], 'key')`, false},
		{`md5($_GET['x'], true)`, true},
		{`md5($_GET['x'], 'false')`, true},
		{`sha1($_GET['x'], $_GET['raw'])`, true},
		{`hash('sha256', $_GET['x'], binary: true)`, true},
		{`md5($_GET['x'], 1)`, true},
		{`md5($_GET['x'], unknown: false)`, true},
	} {
		t.Run(tc.expression, func(t *testing.T) {
			e, f := environment(t, "probe("+tc.expression+");", nil)
			x := argument(callsNamed(f, "probe")[0])
			if got := e.Tainted(x, HTML); got != tc.tainted {
				t.Fatalf("HTML taint %v, want %v: %+v", got, tc.tainted, e.Value(x))
			}
		})
	}
	e, f := environment(t, `namespace App; function md5($x){return $x;} probe(md5($_GET['x']));`, nil)
	snapshot := NewSnapshot(Extract(f, e.types.Index, phpversion.Max, nil))
	e.snapshot = snapshot
	if !e.Tainted(argument(callsNamed(f, "probe")[0]), HTML) {
		t.Fatal("application function inherited native digest safety")
	}
}

func TestInvalidWrapperBindingDoesNotExpandExternalSink(t *testing.T) {
	e, f := environment(t, `function sink($required){system($_GET['command']);}`, nil)
	snapshot := NewSnapshot(Extract(f, e.types.Index, phpversion.Max, nil))
	local, g := environment(t, `sink();sink(unknown: 1);sink(1);`, snapshot)
	local.types.Index = e.types.Index
	calls := callsNamed(g, "sink")
	for _, call := range calls[:2] {
		if sinks := local.Sinks(call); len(sinks) != 0 {
			t.Fatalf("invalid call expanded unreachable sink: %+v", sinks)
		}
	}
	if sinks := local.Sinks(calls[2]); len(sinks) != 1 || !sinks[0].Argument.Complete {
		t.Fatalf("valid external sink lost: %+v", sinks)
	}
}

func BenchmarkScalarTextFlow(b *testing.B) {
	e, f := environment(b, `probe(($_GET['x'] === 'yes') . md5($_GET['x']));`, nil)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		local := New(f, e.types, nil)
		if local.Tainted(x, HTML) {
			b.Fatal("fixed scalar text lost its output contract")
		}
	}
}

func TestChangedWrapperSignatureCannotRetainParameterProof(t *testing.T) {
	e, f := environment(t, `function sink($parameter){system($parameter);}`, nil)
	snapshot := NewSnapshot(Extract(f, e.types.Index, phpversion.Max, nil))
	local, g := environment(t, `function sink(){} sink();`, snapshot)
	if sinks := local.Sinks(callsNamed(g, "sink")[0]); len(sinks) != 1 || sinks[0].Argument.Complete {
		t.Fatalf("changed signature retained old parameter proof: %+v", sinks)
	}
}
