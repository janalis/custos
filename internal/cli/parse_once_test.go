package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCLIParseOnce: parseOnce in custos.json, as --parse-once, keeps the
// output as it is.
func TestCLIParseOnce(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.php"), []byte(`<?php function f(string $p){$v=file_get_contents($p);trim($v);}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var want string
	for i, tc := range []struct {
		config string
		flag   bool
	}{{`{}`, false}, {`{"parseOnce":true}`, false}, {`{}`, true}} {
		if err := os.WriteFile(filepath.Join(root, "custos.json"), []byte(tc.config), 0o644); err != nil {
			t.Fatal(err)
		}
		args := []string{"analyse", "--fail-on=never", "--rule", "FileReadFailureUnchecked", root}
		if tc.flag {
			args = append(args[:1], append([]string{"--parse-once"}, args[1:]...)...)
		}
		code, out, stderr := runCLI(t, args...)
		if code != 0 {
			t.Fatalf("exit=%d %s", code, stderr)
		}
		if i == 0 {
			want = out
		} else if out != want {
			t.Fatalf("%s flag=%v: %q, want %q", tc.config, tc.flag, out, want)
		}
	}
}
