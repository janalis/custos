package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"custos/internal/platform/selfupdate"
)

// fakeUpdate replaces the self-update seams for one test.
type fakeUpdate struct {
	install    selfupdate.Install
	detectErr  error
	latest     string
	latestErr  error
	replaceErr error
	missing    bool
	runErr     error
	replaced   string
	ran        []string
	ranDir     string
}

func (f *fakeUpdate) use(t *testing.T) {
	t.Helper()
	d, l, r, lp, rc := detectInstall, latestRelease, replaceBinary, lookPath, runCommand
	t.Cleanup(func() { detectInstall, latestRelease, replaceBinary, lookPath, runCommand = d, l, r, lp, rc })
	detectInstall = func(version string) (selfupdate.Install, error) {
		if !selfupdate.Valid(version) {
			return selfupdate.Install{Method: selfupdate.Source}, nil
		}
		return f.install, f.detectErr
	}
	latestRelease = func(context.Context) (string, error) { return f.latest, f.latestErr }
	replaceBinary = func(_ context.Context, exe, version string) error {
		f.replaced = exe + "@" + version
		return f.replaceErr
	}
	lookPath = func(name string) (string, error) {
		if f.missing {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + name, nil
	}
	runCommand = func(dir string, _, _ io.Writer, name string, args ...string) error {
		f.ranDir, f.ran = dir, append([]string{name}, args...)
		return f.runErr
	}
}

func selfUpdate(version string, args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := Run(version, append([]string{"self-update"}, args...), strings.NewReader(""), &out, &errb)
	return code, out.String(), errb.String()
}

func TestSelfUpdate(t *testing.T) {
	standalone := selfupdate.Install{Method: selfupdate.Standalone, Path: "/bin/custos"}
	for _, tc := range []struct {
		name    string
		fake    fakeUpdate
		version string
		args    []string
		code    int
		out     string
		err     string
		ran     string
		ranDir  string
		replace string
	}{
		{name: "help", version: "1.0.0", args: []string{"-h"}, code: 0, err: "Usage of self-update"},
		{name: "argument", version: "1.0.0", args: []string{"x"}, code: 2, err: "takes no arguments"},
		{name: "bad version", version: "1.0.0", args: []string{"--version", "next"}, code: 2, err: "must be X.Y.Z"},
		{name: "source", version: "dev", code: 2, err: "built from source"},
		{name: "detect error", fake: fakeUpdate{detectErr: errors.New("no exe")}, version: "1.0.0", code: 2, err: "no exe"},
		{name: "latest error", fake: fakeUpdate{install: standalone, latestErr: errors.New("offline")}, version: "1.0.0", code: 2, err: "offline"},
		{name: "up to date", fake: fakeUpdate{install: standalone, latest: "1.0.0"}, version: "1.0.0", code: 0, out: "custos 1.0.0 is up to date"},
		{name: "check newer", fake: fakeUpdate{install: standalone, latest: "1.1.0"}, version: "1.0.0", args: []string{"--check"}, code: 1, out: "custos 1.1.0 is available (you have 1.0.0)"},
		{name: "replace", fake: fakeUpdate{install: standalone, latest: "1.1.0"}, version: "1.0.0", code: 0, out: "custos 1.0.0 → 1.1.0", replace: "/bin/custos@1.1.0"},
		{name: "pinned downgrade", fake: fakeUpdate{install: standalone}, version: "1.0.0", args: []string{"--version", "0.9.0"}, code: 0, replace: "/bin/custos@0.9.0"},
		{name: "replace error", fake: fakeUpdate{install: standalone, latest: "1.1.0", replaceErr: errors.New("read-only")}, version: "1.0.0", code: 2, err: "read-only", replace: "/bin/custos@1.1.0"},
		{name: "brew pinned", fake: fakeUpdate{install: selfupdate.Install{Method: selfupdate.Homebrew}}, version: "1.0.0", args: []string{"--version", "1.1.0"}, code: 2, err: "only applies to release binaries"},
		{name: "brew", fake: fakeUpdate{install: selfupdate.Install{Method: selfupdate.Homebrew}, latest: "1.1.0"}, version: "1.0.0", code: 0, err: "running brew upgrade janalis/tap/custos", ran: "brew upgrade janalis/tap/custos"},
		{name: "brew missing", fake: fakeUpdate{install: selfupdate.Install{Method: selfupdate.Homebrew}, latest: "1.1.0", missing: true}, version: "1.0.0", code: 2, err: "brew is not on the PATH; run: brew upgrade janalis/tap/custos"},
		{name: "brew fails", fake: fakeUpdate{install: selfupdate.Install{Method: selfupdate.Homebrew}, latest: "1.1.0", runErr: errors.New("exit status 1")}, version: "1.0.0", code: 2, err: "brew upgrade janalis/tap/custos: exit status 1", ran: "brew upgrade janalis/tap/custos"},
		{name: "composer", fake: fakeUpdate{install: selfupdate.Install{Method: selfupdate.Composer, Root: "/p"}, latest: "1.1.0"}, version: "1.0.0", code: 0, err: "composer require --dev janalis/custos:^1.1.0", ran: "composer update janalis/custos", ranDir: "/p"},
		{name: "composer no root", fake: fakeUpdate{install: selfupdate.Install{Method: selfupdate.Composer}, latest: "1.1.0"}, version: "1.0.0", code: 2, err: "cannot find the Composer project"},
		{name: "composer fails", fake: fakeUpdate{install: selfupdate.Install{Method: selfupdate.Composer, Root: "/p"}, latest: "1.1.0", runErr: errors.New("exit status 2")}, version: "1.0.0", code: 2, err: "composer update janalis/custos: exit status 2", ran: "composer update janalis/custos", ranDir: "/p"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.fake
			f.use(t)
			code, out, stderr := selfUpdate(tc.version, tc.args...)
			if code != tc.code || !strings.Contains(out, tc.out) || !strings.Contains(stderr, tc.err) {
				t.Fatalf("exit %d\nstdout %s\nstderr %s", code, out, stderr)
			}
			if got := strings.Join(f.ran, " "); got != tc.ran || f.ranDir != tc.ranDir {
				t.Fatalf("ran %q in %q", got, f.ranDir)
			}
			if f.replaced != tc.replace {
				t.Fatalf("replaced %q", f.replaced)
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	var out bytes.Buffer
	dir := t.TempDir()
	if err := runCommand(dir, &out, io.Discard, "go", "env", "GOVERSION"); err != nil || !strings.HasPrefix(out.String(), "go") {
		t.Fatal(out.String(), err)
	}
}

// noticeEnv makes the notice apply: a terminal, no CI, an empty cache.
func noticeEnv(t *testing.T) string {
	t.Helper()
	cache := t.TempDir()
	t.Setenv("CUSTOS_CACHE_DIR", cache)
	t.Setenv("CI", "")
	t.Setenv("CUSTOS_NO_UPDATE_CHECK", "")
	old := isTerminal
	isTerminal = func(io.Writer) bool { return true }
	t.Cleanup(func() { isTerminal = old })
	return cache
}

func TestUpdateNotice(t *testing.T) {
	cache := noticeEnv(t)
	f := fakeUpdate{install: selfupdate.Install{Method: selfupdate.Standalone}, latest: "1.1.0"}
	f.use(t)
	var errb bytes.Buffer

	// First run: no saved check, the refresh finds 1.1.0 and saves it.
	n := startUpdateNotice("1.0.0", &errb)
	n.finish(&errb)
	if !strings.Contains(errb.String(), `custos: 1.1.0 is available (you have 1.0.0); run: custos self-update`) {
		t.Fatal(errb.String())
	}
	if c := selfupdate.ReadCheck(cache); c.Latest != "1.1.0" {
		t.Fatal(c)
	}
	// Within a day the saved result is used without a lookup.
	f.latest = "9.9.9"
	n = startUpdateNotice("1.0.0", &errb)
	if n.done != nil {
		t.Fatal("refreshed within a day")
	}
	// The method decides the command shown.
	for method, want := range map[selfupdate.Method]string{selfupdate.Homebrew: "brew upgrade janalis/tap/custos", selfupdate.Composer: "composer update janalis/custos"} {
		f.install.Method = method
		errb.Reset()
		n.finish(&errb)
		if !strings.Contains(errb.String(), "run: "+want) {
			t.Fatal(errb.String())
		}
	}
	// Up to date: nothing printed.
	errb.Reset()
	startUpdateNotice("1.1.0", &errb).finish(&errb)
	if errb.Len() != 0 {
		t.Fatal(errb.String())
	}

	// A day later a failed lookup keeps the saved version and counts as a
	// check.
	oldNow := now
	t.Cleanup(func() { now = oldNow })
	later := time.Now().Add(25 * time.Hour)
	now = func() time.Time { return later }
	f.latestErr = errors.New("offline")
	n = startUpdateNotice("1.0.0", &errb)
	n.finish(&errb)
	if c := selfupdate.ReadCheck(cache); c.Latest != "1.1.0" || !c.Checked.Equal(later) {
		t.Fatal(c)
	}
	// A lookup outlasting the command is awaited until its timeout.
	latestRelease = func(ctx context.Context) (string, error) { <-ctx.Done(); return "", ctx.Err() }
	evenLater := later.Add(25 * time.Hour)
	now = func() time.Time { return evenLater }
	n = startUpdateNotice("1.0.0", &errb)
	n.finish(&errb)
	if c := selfupdate.ReadCheck(cache); c.Latest != "1.1.0" || !c.Checked.Equal(evenLater) {
		t.Fatal(c)
	}
}

func TestUpdateNoticeSkipped(t *testing.T) {
	noticeEnv(t)
	var errb bytes.Buffer
	if startUpdateNotice("dev", &errb) != nil {
		t.Fatal("dev build")
	}
	t.Setenv("CI", "true")
	if startUpdateNotice("1.0.0", &errb) != nil {
		t.Fatal("CI")
	}
	t.Setenv("CI", "")
	t.Setenv("CUSTOS_NO_UPDATE_CHECK", "1")
	if startUpdateNotice("1.0.0", &errb) != nil {
		t.Fatal("opted out")
	}
	t.Setenv("CUSTOS_NO_UPDATE_CHECK", "")
	t.Setenv("CUSTOS_CACHE_DIR", "")
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("LocalAppData", "")
	if startUpdateNotice("1.0.0", &errb) != nil {
		t.Fatal("no cache directory")
	}
	isTerminal = func(io.Writer) bool { return false }
	if startUpdateNotice("1.0.0", &errb) != nil {
		t.Fatal("not a terminal")
	}
	// nil notices finish silently (analyse and fix call it unconditionally).
	var n *updateNotice
	n.finish(&errb)
}

func TestIsTerminal(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), "out"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if isTerminal(file) || isTerminal(&bytes.Buffer{}) {
		t.Fatal("not terminals")
	}
}
