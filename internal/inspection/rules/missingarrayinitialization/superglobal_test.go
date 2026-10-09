package missingarrayinitialization

import (
	"testing"
)

func TestIsSuperglobal(t *testing.T) {
	for name, want := range map[string]bool{"_SESSION": true, "GLOBALS": true, "_ENV": true, "session": false, "_session": false, "this": false} {
		if isSuperglobal(name) != want {
			t.Errorf("isSuperglobal(%q) = %v", name, !want)
		}
	}
}
