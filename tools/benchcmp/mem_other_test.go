//go:build unix && !linux

package main

import "testing"

func TestMachine(t *testing.T) {
	if m := machine(); m.CPU == "unknown" || m.MemoryGB == 0 || m.Cores == 0 {
		t.Errorf("machine = %+v", m)
	}
	withPath(t, t.TempDir(), true)
	if m := machine(); m.CPU != "unknown" || m.MemoryGB != 0 {
		t.Errorf("machine without sysctl = %+v", m)
	}
}
