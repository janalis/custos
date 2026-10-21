package project

import (
	"os"
	"path/filepath"
	"testing"
)

// Project paths do not depend on the working directory or on directories
// above the project.
func TestDiskProjectPaths(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		"tests/app/custos.json", "tests/app/src/Db.php", "tests/app/tests/DbTest.php",
		"tests/app/tests/pkg/composer.json", "tests/app/tests/pkg/src/Db.php",
		"lib/src/Db.php", "lib/tests/fixture/composer.json", "lib/tests/fixture/src/Db.php",
		"lib/composer.json", "lib/pkg/composer.json", "lib/pkg/src/Db.php", "lib/pkg/testdata/Db.php",
		"plain/tests/Db.php", "plain/src/Db.php",
	} {
		full := filepath.Join(base, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, full, "{}")
	}
	for _, tc := range []struct {
		name, cwd, path, want string
	}{
		{"below an outer tests directory, absolute", "", "tests/app/src/Db.php", "src/Db.php"},
		{"below an outer tests directory, relative", "tests/app", "src/Db.php", "src/Db.php"},
		{"from inside the project", "tests/app/tests", "DbTest.php", "tests/DbTest.php"},
		{"custos.json outranks a nested composer.json", "tests/app/tests/pkg/src", "Db.php", "tests/pkg/src/Db.php"},
		{"outermost composer.json", "lib/pkg/src", "Db.php", "pkg/src/Db.php"},
		{"fixture application with its own composer.json", "lib/tests/fixture/src", "Db.php", "tests/fixture/src/Db.php"},
		{"no marker: relative to the fallback", "plain", "tests/Db.php", "tests/Db.php"},
		{"no marker: outside the fallback", "plain/src", "../tests/Db.php", filepath.ToSlash(filepath.Join(base, "plain/tests/Db.php"))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.FromSlash(tc.path)
			if tc.cwd == "" {
				path = filepath.Join(base, path)
			} else {
				t.Chdir(filepath.Join(base, filepath.FromSlash(tc.cwd)))
			}
			if got := DiskProjectPaths(".")(path); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProjectPathsKeepsWhatItCannotRelate(t *testing.T) {
	paths := ProjectPaths(func(string) bool { return false }, "/root")
	if got := paths("relative/a.php"); got != "relative/a.php" {
		t.Fatal(got)
	}
	// Answers are cached per directory.
	asked := 0
	paths = ProjectPaths(func(name string) bool { asked++; return name == "src/custos.json" }, ".")
	if paths("src/a.php") != "a.php" || paths("src/b.php") != "b.php" || asked != 1 {
		t.Fatal(asked)
	}
}
