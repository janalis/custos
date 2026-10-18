package syntax

import "testing"

func TestSuperglobal(t *testing.T) {
	for name, want := range map[string]bool{
		"GLOBALS": true, "_SERVER": true, "_GET": true, "_POST": true, "_FILES": true,
		"_COOKIE": true, "_SESSION": true, "_REQUEST": true, "_ENV": true,
		"_get": false, "globals": false, "this": false, "argv": false, "HTTP_RAW_POST_DATA": false, "": false,
	} {
		if got := Superglobal(name); got != want {
			t.Errorf("Superglobal(%q) = %v, want %v", name, got, want)
		}
	}
}
