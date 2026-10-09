package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func TestBuildIndexSurvivesExtractPanic(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.php")
	bad := filepath.Join(dir, "bad.php")
	if err := os.WriteFile(good, []byte("<?php class Good {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("<?php class Bad {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := extract
	t.Cleanup(func() { extract = old })
	extract = func(path string, src []byte, opt syntax.Options) *index.FileSymbols {
		if strings.HasSuffix(path, "bad.php") {
			panic("simulated extraction crash")
		}
		return old(path, src, opt)
	}
	ix := BuildIndex([]string{good, bad}, syntax.Options{})
	if ix.Class("Good", 0) == nil {
		t.Error("symbols of the healthy file are missing")
	}
	if ix.Class("Bad", 0) != nil {
		t.Error("the crashing file should contribute no symbols")
	}
}
