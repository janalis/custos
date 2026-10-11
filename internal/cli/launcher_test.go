package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestComposerMCPWorkspace(t *testing.T) {
	php, err := exec.LookPath("php")
	if err != nil {
		t.Skip("PHP is needed for launcher integration")
	}
	launcher, err := os.ReadFile("../../composer/custos")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, vendor, bin string
		symlink, global   bool
	}{{"standard", "vendor", "vendor/bin", false, false}, {"custom", "deps/packages", "tools/bin", false, false}, {"symlink", "vendor", "vendor/bin", true, false}, {"global", "vendor", "vendor/bin", false, true}} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			root, err = filepath.EvalSymlinks(root)
			if err != nil {
				t.Fatal(err)
			}
			packageDir := filepath.Join(root, tc.vendor, "janalis/custos")
			if err := os.MkdirAll(filepath.Dir(packageDir), 0o700); err != nil {
				t.Fatal(err)
			}
			target := packageDir
			if tc.symlink {
				target = t.TempDir()
				if err := os.Symlink(target, packageDir); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.MkdirAll(filepath.Join(target, "composer"), 0o700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(target, "composer/custos")
			if err := os.WriteFile(path, launcher, 0o600); err != nil {
				t.Fatal(err)
			}
			manifest, _ := json.Marshal(map[string]any{"require-dev": map[string]string{"janalis/custos": "*"}, "config": map[string]string{"vendor-dir": tc.vendor, "bin-dir": tc.bin}})
			if err := os.WriteFile(filepath.Join(root, "composer.json"), manifest, 0o600); err != nil {
				t.Fatal(err)
			}
			binDir := filepath.Join(root, tc.bin)
			if err := os.MkdirAll(binDir, 0o700); err != nil {
				t.Fatal(err)
			}
			proxy := filepath.Join(binDir, "custos")
			proxyCode := "<?php $GLOBALS['_composer_autoload_path']=" + phpQuote(filepath.Join(root, tc.vendor, "autoload.php")) + "; require " + phpQuote(filepath.Join(packageDir, "composer/custos")) + ";"
			if err := os.WriteFile(proxy, []byte(proxyCode), 0o600); err != nil {
				t.Fatal(err)
			}
			fake := filepath.Join(t.TempDir(), "binary")
			if err := os.WriteFile(fake, []byte("#!/usr/bin/env php\n<?php echo json_encode(['root'=>getenv('CUSTOS_COMPOSER_ROOT'),'cwd'=>getcwd(),'args'=>array_slice($argv,1)]);"), 0o700); err != nil {
				t.Fatal(err)
			}
			nested := filepath.Join(root, "src/deep")
			if err := os.MkdirAll(nested, 0o700); err != nil {
				t.Fatal(err)
			}
			for _, cwd := range []string{root, nested, t.TempDir()} {
				cwd, err = filepath.EvalSymlinks(cwd)
				if err != nil {
					t.Fatal(err)
				}
				for _, command := range []string{"mcp", "version"} {
					cmd := exec.Command(php, proxy, command)
					cmd.Dir = cwd
					cmd.Env = append(os.Environ(), "CUSTOS_BINARY="+fake, "CUSTOS_COMPOSER_ROOT=old")
					if tc.global {
						cmd.Env = append(cmd.Env, "COMPOSER_HOME="+root)
					}
					output, err := cmd.Output()
					if err != nil {
						t.Fatal(err)
					}
					var result struct {
						Root any
						Cwd  string
						Args []string
					}
					if err := json.Unmarshal(output, &result); err != nil {
						t.Fatal(string(output), err)
					}
					if result.Cwd != cwd || len(result.Args) != 1 || result.Args[0] != command {
						t.Fatal(result)
					}
					want := any(root)
					if tc.global {
						want = false
					}
					if command == "version" {
						want = "old"
					}
					if result.Root != want {
						t.Fatalf("cwd %s command %s got %+v want root %v", cwd, command, result, want)
					}
				}
			}
		})
	}
}
func phpQuote(s string) string { b, _ := json.Marshal(s); return string(b) }
