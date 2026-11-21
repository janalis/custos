package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzFormat checks, for every preset, that formatting is safe and stable:
// a result the safety check accepts replays from its changes and formatting
// it again changes nothing.
func FuzzFormat(f *testing.F) {
	inputs, _ := filepath.Glob("testdata/cases/*.php")
	for _, in := range inputs {
		if strings.Count(filepath.Base(in), ".") == 1 {
			if b, err := os.ReadFile(in); err == nil {
				f.Add(b)
			}
		}
	}
	var all []*Settings
	for _, p := range Presets() {
		r, _ := Resolve(p, true)
		all = append(all, &r.Settings)
	}
	f.Fuzz(func(t *testing.T, src []byte) {
		for _, s := range all {
			res := Source("fuzz.php", src, Options{Settings: s})
			if strings.Contains(res.Skipped, "please report") {
				t.Fatalf("refused output for %q: %s", src, res.Skipped)
			}
			if res.Skipped != "" {
				continue
			}
			checkResult(t, "fuzz", src, res, s)
		}
	})
}

// BenchmarkFormat formats the golden inputs with each preset.
func BenchmarkFormat(b *testing.B) {
	inputs, _ := filepath.Glob("testdata/cases/general.php")
	src, err := os.ReadFile(inputs[0])
	if err != nil {
		b.Fatal(err)
	}
	// One open tag: a repeated `<?php` inside code is a syntax error, and
	// a file with errors is left as written without being formatted.
	body := strings.ReplaceAll(strings.TrimPrefix(string(src), "<?php"), "?>\n", "")
	big := []byte("<?php" + strings.Repeat(body+"\n", 20))
	for _, p := range []string{"per", "laravel"} {
		r, _ := Resolve(p, false)
		if res := Source("bench.php", big, Options{Settings: &r.Settings}); res.Skipped != "" {
			b.Fatal(p, ": ", res.Skipped)
		}
		b.Run(p, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(big)))
			for b.Loop() {
				Source("bench.php", big, Options{Settings: &r.Settings})
			}
		})
	}
}
