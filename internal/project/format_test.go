package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/format"
	phpversion "github.com/janalis/custos/internal/php/version"
)

func perSettings(string) (*format.Resolved, error) { return format.Resolve("per", false) }

func TestFormatFiles(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string, mode os.FileMode) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), mode); err != nil {
			t.Fatal(err)
		}
		return p
	}
	clean := write("clean.php", "<?php\n\n$a = 1;\n", 0o600)
	messy := write("messy.php", "<?php\n$a=1;\n", 0o644)
	broken := write("broken.php", "<?php\nif (\n", 0o644)
	link := filepath.Join(dir, "link.php")
	if err := os.Symlink(messy, link); err != nil {
		t.Fatal(err)
	}
	files := []string{broken, clean, link, messy, filepath.Join(dir, "missing.php")}
	cacheDir := t.TempDir()
	cache := OpenFormatCache(cacheDir, dir)
	outs := FormatFiles(files, FormatOptions{Settings: perSettings, PHP: phpversion.PHP84, Cache: cache})
	if outs[0].Result.Skipped == "" || outs[0].Err != nil {
		t.Fatalf("broken: %+v", outs[0])
	}
	if outs[1].Result.Changed() || outs[1].Perm != 0o600 || outs[1].Cached {
		t.Fatalf("clean: %+v", outs[1])
	}
	if !errors.Is(outs[2].Err, ErrSkipped) {
		t.Fatalf("link: %v", outs[2].Err)
	}
	if string(outs[3].Result.Output) != "<?php\n\n$a = 1;\n" || outs[3].Perm != 0o644 {
		t.Fatalf("messy: %q", outs[3].Result.Output)
	}
	if outs[4].Err == nil {
		t.Fatal("missing file formatted")
	}
	// The clean file is known formatted from now on; the others are not.
	if err := cache.Save(); err != nil {
		t.Fatal(err)
	}
	cache = OpenFormatCache(cacheDir, dir)
	again := FormatFiles(files[:4], FormatOptions{Settings: perSettings, PHP: phpversion.PHP84, Cache: cache})
	if !again[1].Cached || again[3].Cached || again[0].Cached {
		t.Fatalf("cache: %+v", again)
	}
	// Another target version is another key.
	other := FormatFiles(files[1:2], FormatOptions{Settings: perSettings, PHP: phpversion.PHP83, Cache: cache})
	if other[0].Cached {
		t.Fatal("cached across PHP versions")
	}
	// Settings errors and crashes are per file.
	failing := FormatFiles(files[1:2], FormatOptions{Settings: func(string) (*format.Resolved, error) { return nil, errors.New("bad settings") }})
	if failing[0].Err == nil || !strings.Contains(failing[0].Err.Error(), "bad settings") {
		t.Fatal(failing[0].Err)
	}
	old := formatSource
	t.Cleanup(func() { formatSource = old })
	formatSource = func(string, []byte, format.Options) format.Result { panic("boom") }
	crashed := FormatFiles(files[3:4], FormatOptions{Settings: perSettings})
	if crashed[0].Err == nil || !strings.Contains(crashed[0].Err.Error(), "internal failure") {
		t.Fatal(crashed[0].Err)
	}
}

func TestFormatCache(t *testing.T) {
	dir, root := t.TempDir(), t.TempDir()
	if OpenFormatCache("", root) != nil {
		t.Fatal("cache without a directory")
	}
	var none *FormatCache
	none.Store(FormatKey{})
	if none.Formatted(FormatKey{}) || none.Save() != nil {
		t.Fatal("nil cache keeps entries")
	}
	c := OpenFormatCache(dir, root)
	a, b := c.Key("s", []byte("a")), c.Key("s", []byte("b"))
	if a == b || a == c.Key("t", []byte("a")) || c.Formatted(a) {
		t.Fatal("keys")
	}
	// Nothing added: nothing written.
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(c.path); !os.IsNotExist(err) {
		t.Fatal("empty cache written")
	}
	c.Store(a)
	c.Store(a)
	c.Store(b)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	// A run using only a keeps b too; a run adding nothing writes nothing.
	c = OpenFormatCache(dir, root)
	if !c.Formatted(a) {
		t.Fatal("entry lost")
	}
	if !c.Formatted(a) {
		t.Fatal("entry lost on a second hit")
	}
	c.Store(a)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	k := c.Key("s", []byte("c"))
	c.Store(k)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	c = OpenFormatCache(dir, root)
	if !c.Formatted(a) || !c.Formatted(b) || !c.Formatted(k) {
		t.Fatal("entries lost")
	}
	// Corrupt files and files of another binary are ignored.
	good, err := os.ReadFile(c.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{[]byte("junk"), good[:len(good)-1], append([]byte(formatCacheMagic), 1, 'x')} {
		if err := os.WriteFile(c.path, bad, 0o600); err != nil {
			t.Fatal(err)
		}
		if OpenFormatCache(dir, root).Formatted(a) {
			t.Fatalf("read %q", bad)
		}
	}
	// A cache that cannot be written fails Save.
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	c = OpenFormatCache(filepath.Join(blocked, "sub"), root)
	c.Store(a)
	if c.Save() == nil {
		t.Fatal("unwritable cache saved")
	}
	// Without an identifiable binary there is no cache.
	old := executable
	t.Cleanup(func() { executable = old })
	executable = func() (string, error) { return "", errors.New("unknown") }
	if OpenFormatCache(dir, root) != nil {
		t.Fatal("cache without binary")
	}
	executable = func() (string, error) { return filepath.Join(root, "gone"), nil }
	if OpenFormatCache(dir, root) != nil {
		t.Fatal("cache without binary file")
	}
}
