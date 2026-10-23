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

// TestCLIFlowOrderConfiguration: flow.order in custos.json, as --flow-order,
// saves the flow order in the cache.
func TestCLIFlowOrderConfiguration(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.php"), []byte(`<?php function f(string $p){$v=file_get_contents($p);trim($v);}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		config string
		flag   bool
		want   bool
	}{{`{}`, false, false}, {`{"flow":{"order":true}}`, false, true}, {`{"flow":{"order":false}}`, true, true}} {
		if err := os.WriteFile(filepath.Join(root, "custos.json"), []byte(tc.config), 0o644); err != nil {
			t.Fatal(err)
		}
		cache := t.TempDir()
		t.Setenv("CUSTOS_CACHE_DIR", cache)
		args := []string{"analyse", "--fail-on=never", "--rule", "FileReadFailureUnchecked", root}
		if tc.flag {
			args = append(args[:1], append([]string{"--flow-order"}, args[1:]...)...)
		}
		if code, _, stderr := runCLI(t, args...); code != 0 {
			t.Fatalf("exit=%d %s", code, stderr)
		}
		saved, _ := filepath.Glob(filepath.Join(cache, "*-order.gob"))
		if (len(saved) == 1) != tc.want {
			t.Fatalf("%s flag=%v: saved %v", tc.config, tc.flag, saved)
		}
	}
}
