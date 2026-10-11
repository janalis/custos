package hardcodedcredentialatknownsink

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The default path policy is evaluated against the project-relative path, so
// the working directory and directories above the project cannot change it.
func TestReviewPathPolicyIsProjectRelative(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write := func(rel string) {
		t.Helper()
		full := filepath.Join(base, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A project below a directory named tests, with its own tests directory
	// holding a nested package; a composer-only project; no marker at all.
	write("tests/app/custos.json")
	write("tests/app/src/Db.php")
	write("tests/app/tests/DbTest.php")
	write("tests/app/tests/pkg/composer.json")
	write("tests/app/tests/pkg/src/Db.php")
	write("lib/src/Db.php")
	write("lib/tests/fixture/composer.json")
	write("lib/tests/fixture/src/Db.php")
	write("lib/composer.json")
	write("lib/pkg/composer.json")
	write("lib/pkg/src/Db.php")
	write("lib/pkg/testdata/Db.php")
	write("plain/tests/Db.php")
	write("plain/src/Db.php")
	for _, tc := range []struct {
		name, cwd, path string
		want            int
	}{
		{"source below an outer tests directory, absolute", "", "tests/app/src/Db.php", 1},
		{"source below an outer tests directory, relative", "tests/app", "src/Db.php", 1},
		{"test file from the project root", "tests/app", "tests/DbTest.php", 0},
		{"test file from inside tests", "tests/app/tests", "DbTest.php", 0},
		{"test file by absolute path", "", "tests/app/tests/DbTest.php", 0},
		{"custos.json outranks a nested composer.json", "tests/app/tests/pkg/src", "Db.php", 0},
		{"nested package source", "lib/pkg/src", "Db.php", 1},
		{"composer project source", "lib", "src/Db.php", 1},
		{"fixture application with its own composer.json, from the root", "lib", "tests/fixture/src/Db.php", 0},
		{"fixture application with its own composer.json, from inside", "lib/tests/fixture/src", "Db.php", 0},
		{"composer project test data", "lib/pkg/testdata", "Db.php", 0},
		{"no marker: relative to the working directory", "plain", "tests/Db.php", 0},
		{"no marker: source file", "plain", "src/Db.php", 1},
		{"no marker: outside the working directory", "plain/src", "../tests/Db.php", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.FromSlash(tc.path)
			if tc.cwd == "" {
				path = filepath.Join(base, path)
			} else {
				t.Chdir(filepath.Join(base, filepath.FromSlash(tc.cwd)))
			}
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"HardcodedCredentialAtKnownSink"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse(path, []byte(`<?php new PDO($dsn,"service","password");`), syntax.Options{})
			if got := e.Analyze(f); len(got) != tc.want {
				t.Errorf("got %d findings, want %d", len(got), tc.want)
			}
		})
	}
}
