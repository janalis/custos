package testbudget

import (
	"testing"
	"time"
)

func TestScale(t *testing.T) {
	if f := scale(time.Millisecond, 2*time.Millisecond); f != 1 {
		t.Errorf("faster machine: factor %v, want 1", f)
	}
	if f := scale(6*time.Millisecond, 2*time.Millisecond); f != 3 {
		t.Errorf("3x slower: factor %v, want 3", f)
	}
}

func TestOf(t *testing.T) {
	if d := Of(time.Second); d < time.Second {
		t.Errorf("Of shrank the budget: %v", d)
	}
	t.Logf("calibration factor %.2f", factor)
}
