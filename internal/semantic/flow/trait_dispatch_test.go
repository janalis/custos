package flow

import "testing"

// `$this->m()` in a trait method reaches the trait's own or used non-abstract
// declaration unless the using class overrides it: the call is resolved for
// its may-return, never as a complete value.
func TestTraitThisDispatch(t *testing.T) {
	const probe = `function probe($x) {} `
	for _, tc := range []struct {
		src               string
		resolved, tainted bool
	}{
		{`trait T { function a($x) { return $x; } function f() { probe($this->a($_GET['x'])); } }`, true, true},
		{`trait U { function a($x) { return $x; } } trait T { use U; function f() { probe($this->a($_GET['x'])); } }`, true, true},
		{`trait T { function a($x) { return htmlspecialchars($x); } function f() { probe($this->a($_GET['x'])); } }`, true, false},
		{`trait T { abstract function a($x); function f() { probe($this->a($_GET['x'])); } }`, false, false},
		{`trait T { function f() { probe($this->b($_GET['x'])); } } class C { use T; function b($x) { return $x; } }`, false, false},
		{`trait T { function a($x) { return $x; } function f() { probe((function () { return $this->a($_GET['x']); })()); } }`, false, false},
		{`trait T { function a($x) { return $x; } static function f() { probe($this->a($_GET['x'])); } }`, false, false},
		{`trait T { function a($x) { return $x; } function f() { probe($this?->a($_GET['x'])); } }`, false, false},
		{`trait T { function a($x) { return $x; } function f() { $o = $this; probe($o->a($_GET['x'])); } }`, false, false},
		{`trait T { function __call($n, $a) { return $a[0]; } function f() { probe($this->a($_GET['x'])); } }`, false, false},
		{`class C { function a($x) { return $x; } function f() { probe($this->a($_GET['x'])); } }`, true, true},
	} {
		e, f := closedEnvironment(t, probe+tc.src)
		x := argument(callsNamed(f, "probe")[0])
		v := e.Value(x)
		if v.Complete && tc.resolved {
			t.Errorf("%s: overridable trait target gave a complete value", tc.src)
		}
		if resolved := len(e.Gaps()) == 0; resolved != tc.resolved || e.Tainted(x, HTML) != tc.tainted {
			t.Errorf("%s: resolved=%v tainted=%v gaps=%v", tc.src, resolved, e.Tainted(x, HTML), e.Gaps())
		}
	}
}
