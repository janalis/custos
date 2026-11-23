package main

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakePHP writes a script standing for a PHP binary: it prints list for
// the probe's list mode and check for its check mode.
func fakePHP(t *testing.T, dir, name, list, check string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	body := "#!/bin/sh\ncat >/dev/null\ncase \"$8\" in\nlist) printf '%s' '" + list + "';;\n*) printf '%s' '" + check + "';;\nesac\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// releases returns -php flags for every release from 8.0 to 8.6 with bin,
// except those in other.
func releases(bin string, other map[string]string) []string {
	var args []string
	for _, v := range []string{"8.0", "8.1", "8.2", "8.3", "8.4", "8.5", "8.6"} {
		b := bin
		if o, ok := other[v]; ok {
			b = o
		}
		args = append(args, "-php", v+"="+b)
	}
	return args
}

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	list := "f:old_fn\tstandard\nf:new_fn\tboxes\nf:is_double\tstandard\nm:Box::pack\tboxes\nm:Crate::pack\tboxes\nk:Box::OLD\tboxes\nc:LEGACY\tcore\nnoise\n"
	loaded := "x:standard\nx:core\nx:boxes\n"
	all := "f:old_fn\nf:new_fn\nf:is_double\nm:Box::pack\nm:Crate::pack\nk:Box::OLD\nc:LEGACY\n" + loaded
	newest := fakePHP(t, dir, "php86", list, all)
	// The 8.0 to 8.3 binaries lack the boxes extension.
	early := fakePHP(t, dir, "php80", "", "f:old_fn\nc:LEGACY\nx:standard\nx:core\n")
	middle := fakePHP(t, dir, "php84", "", "f:old_fn\nc:LEGACY\nm:Box::pack\nk:Box::OLD\nm:Crate::pack\n"+loaded)
	out := filepath.Join(dir, "data.go")
	args := append(releases(middle, map[string]string{"8.0": early, "8.1": early, "8.2": early, "8.3": early, "8.6": newest}), "-php", "7.4="+early, "-out", out)
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("code %d: %s", code, stderr.String())
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"old_fn": phpversion.PHP80,`, `"new_fn": phpversion.PHP86,`,
		`"pack": {{"Box", phpversion.PHP84}, {"Crate", phpversion.PHP84}},`,
		`"box::OLD": phpversion.PHP84,`, `"LEGACY": phpversion.PHP80,`,
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %s in\n%s", want, got)
		}
	}
	if strings.Contains(string(got), "is_double") || !strings.Contains(stdout.String(), "6 deprecations") {
		t.Errorf("skipped alias or count: %s", stdout.String())
	}
	if w := stderr.String(); strings.Count(w, "lacks boxes") != 4 || !strings.Contains(w, "PHP 8.0 lacks boxes") {
		t.Errorf("warnings: %s", w)
	}
}

func TestGenerateErrors(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "data.go")
	php := fakePHP(t, dir, "php", "f:a\n", "f:a\n")
	silent := fakePHP(t, dir, "silent", "f:a\n", "")
	for name, tc := range map[string]struct {
		args []string
		code int
		msg  string
	}{
		"no binary":      {[]string{"-out", out}, 1, "no PHP 8.0 binary"},
		"missing":        {[]string{"-php", "8.0=" + php, "-php", "8.2=" + php, "-out", out}, 1, "no PHP 8.1 binary"},
		"bad flag":       {[]string{"-php", "8.0"}, 2, "want VERSION=PATH"},
		"bad version":    {[]string{"-php", "9.9=" + php}, 2, "invalid version"},
		"unknown":        {[]string{"-bogus"}, 2, "flag provided but not defined"},
		"failing binary": {releases(filepath.Join(dir, "none"), nil), 1, "none list"},
		"failing check":  {releases(php, map[string]string{"8.3": filepath.Join(dir, "none")}), 1, "none check"},
		"undated":        {append(releases(silent, nil), "-out", out), 1, "does not deprecate it"},
		"unwritable":     {append(releases(php, nil), "-out", filepath.Join(dir, "no", "data.go")), 1, "no such file"},
	} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run(tc.args, &stdout, &stderr); code != tc.code || !strings.Contains(stderr.String(), tc.msg) {
				t.Fatalf("code %d, stderr %q", code, stderr.String())
			}
		})
	}
	formatSource = func([]byte) ([]byte, error) { return nil, os.ErrInvalid }
	defer func() { formatSource = format.Source }()
	if code := run(append(releases(php, nil), "-out", out), &bytes.Buffer{}, &bytes.Buffer{}); code != 1 {
		t.Fatalf("format failure: %d", code)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-h"}, &stdout, &stderr); code != 0 {
		t.Fatalf("help: %d", code)
	}
	if s := (binaries{}).String(); s != "map[]" {
		t.Fatalf("flag string %q", s)
	}
}
