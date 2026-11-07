//go:build unix

package main

import "testing"

type noUsage struct{}

func (noUsage) SysUsage() any { return nil }

func TestRusageMaxUnknown(t *testing.T) {
	if got := rusageMax(noUsage{}); got != 0 {
		t.Errorf("rusageMax = %d", got)
	}
}
