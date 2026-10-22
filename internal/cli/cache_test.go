package cli

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestCacheDir(t *testing.T) {
	old := userCacheDir
	t.Cleanup(func() { userCacheDir = old })
	userCacheDir = func() (string, error) { return "/home/u/.cache", nil }
	t.Setenv("CUSTOS_CACHE_DIR", "")
	if got := (&common{}).cacheDir(); got != filepath.Join("/home/u/.cache", "custos") {
		t.Fatal(got)
	}
	if got := (&common{noCache: true}).cacheDir(); got != "" {
		t.Fatal(got)
	}
	t.Setenv("CUSTOS_CACHE_DIR", "/ci/cache")
	if got := (&common{}).cacheDir(); got != "/ci/cache" {
		t.Fatal(got)
	}
	t.Setenv("CUSTOS_CACHE_DIR", "")
	userCacheDir = func() (string, error) { return "", errors.New("no home") }
	if got := (&common{}).cacheDir(); got != "" {
		t.Fatal(got)
	}
}
