package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// release serves releases/latest (redirecting to tag), and the assets of
// v1.2.3: SHA256SUMS and the binary for this platform.
func release(t *testing.T, location string, bin []byte, sums string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			if location != "" {
				w.Header().Set("Location", location)
				w.WriteHeader(http.StatusFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		case "/releases/download/v1.2.3/SHA256SUMS":
			_, _ = w.Write([]byte(sums))
		case "/releases/download/v1.2.3/" + Asset():
			_, _ = w.Write(bin)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	old := releasesURL
	releasesURL = srv.URL + "/releases"
	t.Cleanup(func() { releasesURL = old })
	return srv
}

func sum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func noEnv(t *testing.T, env map[string]string) {
	t.Helper()
	old := getenv
	getenv = func(k string) string { return env[k] }
	t.Cleanup(func() { getenv = old })
}

func TestLatest(t *testing.T) {
	ctx := context.Background()
	release(t, "https://github.com/janalis/custos/releases/tag/v1.2.3", nil, "")
	if v, err := Latest(ctx); err != nil || v != "1.2.3" {
		t.Fatal(v, err)
	}
	release(t, "", nil, "")
	if _, err := Latest(ctx); err == nil || !strings.Contains(err.Error(), "200 OK") {
		t.Fatal(err)
	}
	release(t, "https://example.com/releases/tag/nightly", nil, "")
	if _, err := Latest(ctx); err == nil {
		t.Fatal("nightly accepted")
	}
	releasesURL = "http://127.0.0.1:1/releases"
	if _, err := Latest(ctx); err == nil {
		t.Fatal("unreachable server")
	}
	releasesURL = "://bad"
	if _, err := Latest(ctx); err == nil {
		t.Fatal("bad URL")
	}
}

func TestAsset(t *testing.T) {
	oldOS, oldArch := goos, goarch
	defer func() { goos, goarch = oldOS, oldArch }()
	goos, goarch = "windows", "arm64"
	if a := Asset(); a != "custos_windows_arm64.exe" {
		t.Fatal(a)
	}
	goos = "linux"
	if a := Asset(); a != "custos_linux_arm64" {
		t.Fatal(a)
	}
}

func TestVersions(t *testing.T) {
	for _, v := range []string{"dev", "1.2", "1.2.3.4", "1.x.3", "1.02.3", "-1.2.3", "1.2.3-rc1", ""} {
		if Valid(v) {
			t.Errorf("%q valid", v)
		}
	}
	for _, tc := range []struct {
		current, latest string
		want            bool
	}{
		{"0.1.9", "0.1.10", true},
		{"0.1.10", "0.1.10", false},
		{"0.2.0", "0.1.10", false},
		{"0.9.9", "1.0.0", true},
		{"dev", "1.0.0", false},
		{"1.0.0", "", false},
	} {
		if got := Newer(tc.current, tc.latest); got != tc.want {
			t.Errorf("Newer(%q, %q) = %v", tc.current, tc.latest, got)
		}
	}
}

func TestCheck(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	if c := ReadCheck(dir); c != (Check{}) {
		t.Fatal(c)
	}
	want := Check{Latest: "1.2.3", Checked: time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)}
	if err := SaveCheck(dir, want); err != nil {
		t.Fatal(err)
	}
	if c := ReadCheck(dir); !c.Checked.Equal(want.Checked) || c.Latest != want.Latest {
		t.Fatal(c)
	}
	if err := os.WriteFile(filepath.Join(dir, checkFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if c := ReadCheck(dir); c != (Check{}) {
		t.Fatal(c)
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := SaveCheck(filepath.Join(file, "dir"), want); err == nil {
		t.Fatal("cache dir under a file")
	}
}

func TestDetect(t *testing.T) {
	oldExe := executable
	t.Cleanup(func() { executable = oldExe })
	if in, err := Detect("dev"); err != nil || in.Method != Source {
		t.Fatal(in, err)
	}
	executable = func() (string, error) { return "", errors.New("no exe") }
	if _, err := Detect("1.0.0"); err == nil {
		t.Fatal("executable error ignored")
	}
	dir := t.TempDir()
	real := filepath.Join(dir, "custos")
	if err := os.WriteFile(real, nil, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	resolved, _ := filepath.EvalSymlinks(real)
	for _, tc := range []struct {
		exe  string
		env  map[string]string
		want Install
	}{
		{link, nil, Install{Method: Standalone, Path: resolved}},
		{"/opt/homebrew/Caskroom/custos/1.0.0/custos", nil, Install{Method: Homebrew, Path: "/opt/homebrew/Caskroom/custos/1.0.0/custos"}},
		{"/p/vendor/janalis/custos/composer/.bin/custos-1.0.0", nil, Install{Method: Composer, Path: "/p/vendor/janalis/custos/composer/.bin/custos-1.0.0"}},
		{"/bin/custos", map[string]string{"CUSTOS_INSTALL": "composer", "CUSTOS_COMPOSER_ROOT": "/p"}, Install{Method: Composer, Path: "/bin/custos", Root: "/p"}},
	} {
		executable = func() (string, error) { return tc.exe, nil }
		noEnv(t, tc.env)
		if in, err := Detect("1.0.0"); err != nil || in != tc.want {
			t.Errorf("%s: %+v %v", tc.exe, in, err)
		}
	}
}

func TestReplace(t *testing.T) {
	ctx := context.Background()
	noEnv(t, nil)
	bin := []byte("new binary")
	exe := filepath.Join(t.TempDir(), "custos")
	write := func() {
		if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write()
	release(t, "", bin, "deadbeef  other\n"+sum(bin)+" *"+Asset()+"\n")
	if err := Replace(ctx, exe, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(exe); string(got) != string(bin) {
		t.Fatalf("%q", got)
	}
	// A missing executable, a directory in its place, and a read-only
	// directory.
	if err := Replace(ctx, filepath.Join(t.TempDir(), "gone"), "1.2.3"); err == nil {
		t.Fatal("missing exe")
	}
	dirExe := filepath.Join(t.TempDir(), "custos")
	if err := os.MkdirAll(filepath.Join(dirExe, "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := Replace(ctx, dirExe, "1.2.3"); err == nil || !strings.Contains(err.Error(), "cannot replace") {
		t.Fatal(err)
	}
	if matches, _ := filepath.Glob(dirExe + ".*.tmp"); len(matches) != 0 {
		t.Fatal("temporary file left", matches)
	}
	ro := t.TempDir()
	roExe := filepath.Join(ro, "custos")
	if err := os.WriteFile(roExe, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(ro, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ro, 0o700) })
	if err := Replace(ctx, roExe, "1.2.3"); err == nil || !strings.Contains(err.Error(), "cannot write") {
		t.Fatal(err)
	}

	// Checksum mismatch, binary not listed, asset missing, sums missing.
	release(t, "", bin, sum([]byte("other"))+"  "+Asset()+"\n")
	if err := Replace(ctx, exe, "1.2.3"); !errors.Is(err, errChecksum) {
		t.Fatal(err)
	}
	release(t, "", bin, sum(bin)+"  custos_plan9_mips\n")
	if err := Replace(ctx, exe, "1.2.3"); err == nil || !strings.Contains(err.Error(), "not listed") {
		t.Fatal(err)
	}
	if err := Replace(ctx, exe, "9.9.9"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatal(err)
	}
	srv := release(t, "", bin, sum(bin)+"  "+Asset()+"\n")
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "SHA256SUMS") {
			_, _ = w.Write([]byte(sum(bin) + "  " + Asset() + "\n"))
			return
		}
		http.NotFound(w, r)
	})
	if err := Replace(ctx, exe, "1.2.3"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatal(err)
	}

	// A mirror directory named by CUSTOS_DOWNLOAD_URL.
	mirror := t.TempDir()
	if err := os.WriteFile(filepath.Join(mirror, "SHA256SUMS"), []byte(sum(bin)+"  "+Asset()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mirror, Asset()), bin, 0o600); err != nil {
		t.Fatal(err)
	}
	write()
	noEnv(t, map[string]string{"CUSTOS_DOWNLOAD_URL": mirror + "/"})
	if err := Replace(ctx, exe, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(exe); string(got) != string(bin) {
		t.Fatalf("%q", got)
	}
}

func TestFetch(t *testing.T) {
	ctx := context.Background()
	old := maxAsset
	maxAsset = 4
	defer func() { maxAsset = old }()
	if _, err := fetch(ctx, "http://127.0.0.1:1/x"); err == nil {
		t.Fatal("unreachable")
	}
	if _, err := fetch(ctx, "http://bad host/x"); err == nil {
		t.Fatal("bad URL")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/big" {
			_, _ = w.Write([]byte("12345"))
			return
		}
		w.Header().Set("Content-Length", "10")
		_, _ = w.Write([]byte("ab"))
	}))
	defer srv.Close()
	if _, err := fetch(ctx, srv.URL+"/big"); err == nil || !strings.Contains(err.Error(), "larger") {
		t.Fatal(err)
	}
	if _, err := fetch(ctx, srv.URL+"/truncated"); err == nil {
		t.Fatal("truncated body accepted")
	}
}
