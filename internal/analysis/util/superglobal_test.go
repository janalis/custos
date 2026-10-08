package util

import "testing"

func TestIsSuperglobal(t *testing.T) {
	for name, want := range map[string]bool{"_SESSION": true, "GLOBALS": true, "_ENV": true, "session": false, "_session": false, "this": false} {
		if IsSuperglobal(name) != want {
			t.Errorf("IsSuperglobal(%q) = %v", name, !want)
		}
	}
}
