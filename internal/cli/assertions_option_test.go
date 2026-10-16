package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIAssertionConfiguration(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.php"), []byte(`<?php function f(string $p){$v=file_get_contents($p);assert($v!==false);trim($v);}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []bool{false, true} {
		if err := os.WriteFile(filepath.Join(root, "custos.json"), []byte(fmt.Sprintf(`{"flow":{"assertions":%t}}`, enabled)), 0o644); err != nil {
			t.Fatal(err)
		}
		code, out, stderr := runCLI(t, "analyse", "--format=json", "--fail-on=never", "--rule", "FileReadFailureUnchecked", root)
		if code != 0 {
			t.Fatalf("exit=%d %s", code, stderr)
		}
		reported := strings.Contains(out, `"rule": "FileReadFailureUnchecked"`)
		if reported == enabled {
			t.Fatalf("enabled=%v output=%s", enabled, out)
		}
	}
}
