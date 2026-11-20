package format

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCorpus formats every PHP file under $CUSTOS_FORMAT_CORPUS with each
// preset and requires that no file trips the safety check and that
// formatting is idempotent.
func TestCorpus(t *testing.T) {
	root := os.Getenv("CUSTOS_FORMAT_CORPUS")
	if root == "" {
		t.Skip("CUSTOS_FORMAT_CORPUS not set")
	}
	var files []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".php") && !strings.Contains(p, "/vendor/") {
			files = append(files, p)
		}
		return nil
	})
	presets := Presets()
	if p := os.Getenv("CUSTOS_FORMAT_PRESET"); p != "" {
		presets = []string{p}
	}
	risky := os.Getenv("CUSTOS_FORMAT_RISKY") != ""
	for _, name := range presets {
		r, _ := Resolve(name, risky)
		bad, unstable, changed := 0, 0, 0
		for _, p := range files {
			src, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			res := Source(p, src, Options{Settings: &r.Settings})
			if strings.Contains(res.Skipped, "report") {
				bad++
				if bad <= 5 {
					t.Errorf("%s %s: %s", name, p, res.Skipped)
				}
				continue
			}
			if res.Changed() {
				changed++
			}
			if got := apply(src, res.Changes); string(got) != string(res.Output) {
				bad++
				if bad <= 5 {
					t.Errorf("%s %s: the changes do not replay to the output", name, p)
				}
			}
			again := Source(p, res.Output, Options{Settings: &r.Settings})
			if again.Changed() {
				unstable++
				if unstable <= 5 {
					c := again.Changes[0]
					lo := max(int(c.Start)-60, 0)
					hi := min(int(c.End)+60, len(res.Output))
					t.Errorf("%s %s: not idempotent (%s %q)\n%s", name, p, c.Rule, c.New, res.Output[lo:hi])
				}
			}
		}
		t.Logf("%s: %d files, %d changed, %d unsafe, %d unstable", name, len(files), changed, bad, unstable)
	}
}
