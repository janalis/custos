// Package selfupdate finds the latest custos release, tells how the running
// binary was installed, keeps the result of the last release check, and
// replaces a standalone binary with a checksum-verified release binary.
package selfupdate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"custos/internal/platform/safeio"
)

// releasesURL is the GitHub releases page; tests point it at a local server.
var releasesURL = "https://github.com/janalis/custos/releases"

// client downloads release assets; latest checks use a copy that does not
// follow redirects.
var client = &http.Client{Timeout: 2 * time.Minute}

const userAgent = "custos-self-update"

// Latest returns the version (without "v") of the latest release. It reads
// the redirect of the releases/latest page, which, unlike the GitHub API,
// has no anonymous rate limit.
func Latest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, releasesURL+"/latest", http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	noRedirect := *client
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := noRedirect.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	_, tag, _ := strings.Cut(resp.Header.Get("Location"), "/releases/tag/")
	version := strings.TrimPrefix(tag, "v")
	if !Valid(version) {
		return "", fmt.Errorf("cannot find the latest release at %s/latest (%s)", releasesURL, resp.Status)
	}
	return version, nil
}

// parse splits an X.Y.Z version into its numbers.
func parse(v string) ([3]int, bool) {
	var out [3]int
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || p != strconv.Itoa(n) {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

// Valid reports whether v is a release version (X.Y.Z); development builds
// ("dev") and pre-releases are not.
func Valid(v string) bool {
	_, ok := parse(v)
	return ok
}

// Newer reports whether latest is a later release than current. Either one
// not being a release version means no.
func Newer(current, latest string) bool {
	c, ok1 := parse(current)
	l, ok2 := parse(latest)
	if !ok1 || !ok2 {
		return false
	}
	for i := range c {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// Check is the result of the last release check.
type Check struct {
	Latest  string    `json:"latest"`
	Checked time.Time `json:"checked"`
}

// checkFile is the file, in the cache directory, holding the last Check.
const checkFile = "update-check.json"

// ReadCheck returns the last check saved in dir, or the zero Check when
// there is none or it cannot be read.
func ReadCheck(dir string) Check {
	var c Check
	b, err := safeio.ReadFile(filepath.Join(dir, checkFile), 4096)
	if err != nil || json.Unmarshal(b, &c) != nil {
		return Check{}
	}
	return c
}

// SaveCheck records c in dir.
func SaveCheck(dir string, c Check) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	b, _ := json.Marshal(c)
	return os.WriteFile(filepath.Join(dir, checkFile), b, 0o600)
}

// maxAsset bounds a downloaded release asset; tests lower it.
var maxAsset int64 = 256 << 20

// fetch returns the content of url, or of a local file when url is a path
// (CUSTOS_DOWNLOAD_URL may name a directory holding the assets).
func fetch(ctx context.Context, url string) ([]byte, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return safeio.ReadFile(url, maxAsset)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot download %s: %s", url, resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxAsset+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > maxAsset {
		return nil, fmt.Errorf("%s: larger than %d bytes", url, maxAsset)
	}
	return b, nil
}

// errChecksum is returned when a download does not match SHA256SUMS.
var errChecksum = errors.New("checksum mismatch")
