package runner

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"custos/internal/analysis"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

func writeFile(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestIndexVendor: vendor sources are indexed (tests directories excluded)
// on top of the analysed files, and their symbols resolve.
func TestIndexVendor(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "src", "app.php")
	writeFile(t, app, "<?php\nfunction app_fn() {}\n")
	writeFile(t, filepath.Join(root, "vendor", "lib", "Lib.php"), "<?php\nnamespace Lib;\nclass Thing {}\n")
	writeFile(t, filepath.Join(root, "vendor", "lib", "tests", "T.php"), "<?php\nclass T {}\n")
	files := IndexSources(root, []string{app})
	want := []string{app, filepath.Join(root, "vendor", "lib", "Lib.php")}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("sources: %v", files)
	}
	// a vanished file is skipped, not fatal
	ix := BuildIndex(append(files, filepath.Join(root, "gone.php")), syntax.Options{Version: phpver.Default})
	if ix.Class(`Lib\Thing`, phpver.Default) == nil || ix.Function("app_fn", phpver.Default) == nil {
		t.Fatal("vendor/app symbols missing from the index")
	}
	if got := IndexSources(t.TempDir(), []string{app}); !reflect.DeepEqual(got, []string{app}) {
		t.Fatalf("no vendor: %v", got)
	}
}

// TestIndexVendorDir: composer.json's config.vendor-dir is honoured when it
// stays inside the project; anything else falls back to "vendor".
func TestIndexVendorDir(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "src", "app.php")
	writeFile(t, app, "<?php\n")
	lib := filepath.Join(root, "libraries", "vendor", "lib", "Lib.php")
	writeFile(t, lib, "<?php\nclass Lib {}\n")
	writeFile(t, filepath.Join(root, "composer.json"), `{"config": {"vendor-dir": "libraries/vendor"}}`)
	if got := IndexSources(root, []string{app}); !reflect.DeepEqual(got, []string{app, lib}) {
		t.Fatalf("vendor-dir: %v", got)
	}
	for _, bad := range []string{`{"config": {"vendor-dir": "../outside"}}`, `{"config": {"vendor-dir": "/abs"}}`, `{`, `{}`} {
		writeFile(t, filepath.Join(root, "composer.json"), bad)
		if got := vendorDir(root); got != "vendor" {
			t.Fatalf("%s: %q", bad, got)
		}
	}
}

func TestDiscoverErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := Discover([]string{filepath.Join(dir, "missing")}, nil); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing root: %v", err)
	}
	// a file named explicitly is listed whatever its extension
	f := filepath.Join(dir, "script")
	writeFile(t, f, "<?php\n")
	if got, err := Discover([]string{f}, nil); err != nil || !reflect.DeepEqual(got, []string{f}) {
		t.Fatalf("file root: %v %v", got, err)
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	locked := filepath.Join(dir, "locked")
	writeFile(t, filepath.Join(locked, "a.php"), "<?php\n")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	if _, err := Discover([]string{dir}, nil); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("unreadable directory: %v", err)
	}
}

func TestRunReportsReadErrors(t *testing.T) {
	e, err := analysis.NewEngine(nil, analysis.Config{PHP: phpver.Default})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	ok := filepath.Join(dir, "a.php")
	writeFile(t, ok, "<?php\n$a = ;\n")
	res := Run(e, []string{ok, dir}, syntax.Options{Version: phpver.Default})
	if res[0].Err != nil || len(res[0].Errors) == 0 || res[0].Path != ok {
		t.Fatalf("parsed file: %+v", res[0])
	}
	if res[1].Err == nil || res[1].Src != nil {
		t.Fatalf("directory: %+v", res[1])
	}
}
