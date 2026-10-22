package cli

import (
	"os"
	"testing"
)

// TestMain keeps the dependency cache of every test run out of the user's
// cache directory.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "custos-cache-")
	if err != nil {
		panic(err)
	}
	os.Setenv("CUSTOS_CACHE_DIR", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
