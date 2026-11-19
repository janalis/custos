package config

import (
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
)

func TestTypeCheckSetting(t *testing.T) {
	root := t.TempDir()
	for in, want := range map[string]analysis.TypeCheck{
		"": analysis.TypeCheckOff, "off": analysis.TypeCheckOff, "proven": analysis.TypeCheckProven,
		"documented": analysis.TypeCheckDocumented, "DOCUMENTED": analysis.TypeCheckDocumented,
	} {
		c, err := Resolve(root, File{TypeCheck: in})
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if c.TypeCheck != want || c.Analysis().TypeCheck != want {
			t.Errorf("%q: TypeCheck = %v, want %v", in, c.TypeCheck, want)
		}
	}
	if _, err := Resolve(root, File{TypeCheck: "strict"}); err == nil || !strings.Contains(err.Error(), "typeCheck") {
		t.Errorf("invalid level: err = %v", err)
	}
}

func TestTypeCheckOverlay(t *testing.T) {
	base, err := Resolve(t.TempDir(), File{TypeCheck: "proven"})
	if err != nil {
		t.Fatal(err)
	}
	for in, want := range map[string]analysis.TypeCheck{
		"": analysis.TypeCheckProven, "off": analysis.TypeCheckOff, "documented": analysis.TypeCheckDocumented,
	} {
		c, err := Overlay(base, File{TypeCheck: in})
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if c.TypeCheck != want {
			t.Errorf("client %q: TypeCheck = %v, want %v", in, c.TypeCheck, want)
		}
	}
}
