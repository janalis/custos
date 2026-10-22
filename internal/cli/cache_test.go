package cli

import "testing"

func TestCacheDir(t *testing.T) {
	t.Setenv("CUSTOS_CACHE_DIR", "/ci/cache")
	if got := (&common{}).cacheDir(); got != "/ci/cache" {
		t.Fatal(got)
	}
	if got := (&common{noCache: true}).cacheDir(); got != "" {
		t.Fatal(got)
	}
}
