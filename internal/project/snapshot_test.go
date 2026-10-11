package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/nestednotoperators"
	"custos/internal/inspection/rules/securityadvisories"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func TestSnapshotConfinedAndReuse(t *testing.T) {
	dir := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("custos.json", `{"php":"8.2","paths":["src"],"exclude":["skip"],"rules":{"NestedNotOperators":{"enabled":true},"SecurityAdvisories":{"enabled":true}}}`)
	write("composer.json", `{"config":{"vendor-dir":"deps"}}`)
	write("src/a.php", "<?php $ok=!!!$x;")
	write("src/b.php", "<?php class Other {}")
	write("src/skip/unread.php", "<?php function excluded() {}")
	write("deps/tests/fixture.php", "<?php function excluded_vendor_test() {}")
	write("deps/v.php", "<?php function vendor_function(): int {return 1;}")
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	rules := []analysis.Rule{nestednotoperators.New(), securityadvisories.New()}
	s, err := BuildSnapshot(context.Background(), root, rules, []string{"src/a.php"})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Files) != 1 || len(s.Results) != 1 || s.Index.Function("vendor_function", s.Config.PHP) == nil || s.Index.Class("Other", s.Config.PHP) == nil || s.Index.Function("excluded", s.Config.PHP) != nil || s.Index.Function("excluded_vendor_test", s.Config.PHP) != nil {
		t.Fatal(s.Files, s.Hashes)
	}
	if s.Hashes["composer.lock"] != "absent" || len(s.Results[0].Findings) != 1 {
		t.Fatal(s.Hashes, s.Results)
	}
	inputs, err := CaptureSnapshot(context.Background(), root, rules, nil)
	if err != nil || len(inputs.Results) != 0 {
		t.Fatal(inputs, err)
	}
	if _, err := BuildSnapshot(context.Background(), root, rules, []string{"src", "../escape"}); err == nil {
		t.Fatal("escaped root")
	}
	if _, err := BuildSnapshot(context.Background(), root, rules, []string{"missing"}); err == nil {
		t.Fatal("missing input")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := BuildSnapshot(ctx, root, rules, nil); err == nil {
		t.Fatal("ignored cancellation")
	}
	if _, err := AnalyzeSnapshot(ctx, inputs); err == nil {
		t.Fatal("ignored semantic cancellation")
	}
	write("custos.json", "bad")
	if _, err := BuildSnapshot(context.Background(), root, rules, nil); err == nil {
		t.Fatal("invalid config")
	}
}

func TestSnapshotRejectsHostileInputs(t *testing.T) {
	for _, name := range []string{"custos.json", "composer.json", "composer.lock", "a.php"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			outside := filepath.Join(t.TempDir(), "secret")
			if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, filepath.Join(dir, name)); err != nil {
				t.Fatal(err)
			}
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			if _, err := BuildSnapshot(context.Background(), root, []analysis.Rule{nestednotoperators.New()}, nil); err == nil {
				t.Fatal("symlink escape allowed")
			}
		})
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "custos.json"), []byte(`{"baseline":"missing.json"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := BuildSnapshot(context.Background(), root, []analysis.Rule{nestednotoperators.New()}, nil); err == nil {
		t.Fatal("missing baseline")
	}
	file, err := os.Create(filepath.Join(dir, "large.php"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(syntax.MaxFileSize + 1); err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	if _, err := ReadConfined(root, "large.php", syntax.MaxFileSize); err == nil {
		t.Fatal("oversized file")
	}
	if _, err := ReadConfined(root, "../outside", syntax.MaxFileSize); err == nil {
		t.Fatal("invalid local path")
	}
}

func TestCapturedExtractionCrash(t *testing.T) {
	original := extract
	t.Cleanup(func() { extract = original })
	extract = func(string, []byte, syntax.Options) *index.FileSymbols { panic("parser crash") }
	if _, err := extractCaptured("a.php", []byte("<?php"), syntax.Options{}); err == nil || !strings.Contains(err.Error(), "a.php") {
		t.Fatal(err)
	}
}

type stepContext struct {
	context.Context
	steps, stop int
	run         func(int)
}

func (c *stepContext) Err() error {
	c.steps++
	if c.run != nil {
		c.run(c.steps)
	}
	if c.stop > 0 && c.steps >= c.stop {
		return context.Canceled
	}
	return c.Context.Err()
}

type missingSnapshotRule struct{}

func (missingSnapshotRule) ID() string                           { return "MissingRule" }
func (missingSnapshotRule) Kinds() []syntax.NodeKind             { return nil }
func (missingSnapshotRule) Check(*analysis.Context, syntax.Node) {}
func TestSnapshotCancellationAndConfigurationChanges(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	writeFile(t, filepath.Join(dir, "a.php"), "<?php $a=!!!$b;")
	rules := []analysis.Rule{nestednotoperators.New()}
	for stop := 1; stop < 15; stop++ {
		ctx := &stepContext{Context: context.Background(), stop: stop}
		_, _ = CaptureSnapshot(ctx, root, rules, []string{""})
	}
	// Configuration races are rejected before publishing mixed evidence.
	for _, replacement := range []string{`{"php":"7.4"}`, "invalid"} {
		writeFile(t, filepath.Join(dir, "custos.json"), `{"php":"8.2"}`)
		ctx := &stepContext{Context: context.Background(), run: func(n int) {
			if n == 1 {
				writeFile(t, filepath.Join(dir, "custos.json"), replacement)
			}
		}}
		if _, err := CaptureSnapshot(ctx, root, rules, nil); err == nil {
			t.Fatal("mixed configuration snapshot")
		}
	}
	_ = root.Remove("custos.json")
	if _, err := CaptureSnapshot(context.Background(), root, []analysis.Rule{missingSnapshotRule{}}, nil); err == nil {
		t.Fatal("unknown inspection catalogue")
	}
	writeFile(t, filepath.Join(dir, "custos.json"), `{"paths":["missing"]}`)
	if _, err := CaptureSnapshot(context.Background(), root, rules, []string{"a.php"}); err == nil {
		t.Fatal("missing semantic source path")
	}
	_ = root.Remove("custos.json")
	if err := root.Mkdir("vendor", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), filepath.Join(dir, "vendor", "link.php")); err != nil {
		t.Fatal(err)
	}
	if _, err := CaptureSnapshot(context.Background(), root, rules, nil); err == nil {
		t.Fatal("vendor symlink input")
	}
	_ = root.Remove("vendor/link.php")
	if err := root.MkdirAll("vendor/janalis", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(dir, "vendor/janalis/custos")); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildSnapshot(context.Background(), root, rules, nil); err != nil {
		t.Fatal("symlinked Composer package failed", err)
	}
	// Closed handles and oversized source files fail before analysis.
	large, err := root.OpenFile("large.php", os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := large.Truncate(syntax.MaxFileSize + 1); err != nil {
		t.Fatal(err)
	}
	_ = large.Close()
	if _, err := CaptureSnapshot(context.Background(), root, rules, nil); err == nil {
		t.Fatal("oversized discovered input")
	}
	_ = root.Remove("large.php")
	snap, err := CaptureSnapshot(context.Background(), root, rules, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AnalyzeSnapshot(&stepContext{Context: context.Background(), stop: 2}, snap); err == nil {
		t.Fatal("analysis cancellation")
	}
	original := extract
	defer func() { extract = original }()
	extract = func(string, []byte, syntax.Options) *index.FileSymbols { panic("broken extraction") }
	if _, err := BuildSnapshot(context.Background(), root, rules, nil); err == nil {
		t.Fatal("extraction fault")
	}
}

func TestSnapshotInputBudgets(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	for i := range MaxSnapshotFiles + 1 {
		if err := root.WriteFile(fmt.Sprintf("%05d.php", i), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := CaptureSnapshot(context.Background(), root, nil, nil); err == nil {
		t.Fatal("file budget not enforced")
	}
}

func TestConfinedFlowSnapshot(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	writeFile(t, filepath.Join(dir, "wrapper.php"), "<?php function invoke($x){system($x);}")
	writeFile(t, filepath.Join(dir, "a.php"), "<?php invoke($_GET['arg']);")
	writeFile(t, filepath.Join(dir, "custos.json"), `{"rules":{"UntrustedShellCommand":{"enabled":true}}}`)
	snapshot, err := BuildSnapshot(context.Background(), root, []analysis.Rule{untrustedshellcommand.New()}, []string{"a.php"})
	if err != nil || len(snapshot.Results) != 1 || len(snapshot.Results[0].Findings) == 0 {
		t.Fatal(snapshot, err)
	}
}

func TestVendorStatFailure(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	// A configured vendor parent that is a regular file yields ENOTDIR, rather
	// than being silently treated as absent semantic inputs.
	writeFile(t, filepath.Join(dir, "composer.json"), `{"config":{"vendor-dir":"file/vendor"}}`)
	writeFile(t, filepath.Join(dir, "file"), "not a directory")
	if _, err := CaptureSnapshot(context.Background(), root, nil, nil); err == nil {
		t.Fatal("invalid vendor parent ignored")
	}
}

func TestBufferSnapshotSemanticPublication(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	writeFile(t, filepath.Join(dir, "a.php"), "<?php class Before {}")
	snapshot, err := BuildSnapshot(context.Background(), root, []analysis.Rule{nestednotoperators.New()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	diskHash := snapshot.Hashes["a.php"]
	snapshot, err = OverlaySnapshot(context.Background(), snapshot, "a.php", []byte("<?php class After {} $x=!!!$a;"))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Index.Class("Before", snapshot.Config.PHP) != nil || snapshot.Index.Class("After", snapshot.Config.PHP) == nil || snapshot.Hashes["a.php"] != diskHash {
		t.Fatal("overlay did not replace declarations while retaining disk manifest")
	}
	snapshot, err = OverlaySnapshot(context.Background(), snapshot, "unsaved.php", []byte("<?php function unsaved(): int { return 1; }"))
	if err != nil || snapshot.Index.Function("unsaved", snapshot.Config.PHP) == nil {
		t.Fatal(snapshot, err)
	}
	oversized := &Snapshot{Sources: map[string][]byte{"other": make([]byte, MaxSnapshotBytes)}}
	if _, err := OverlaySnapshot(context.Background(), oversized, "new.php", []byte("<?php")); err == nil {
		t.Fatal("overlay input budget")
	}
}

func TestSnapshotNormalizesConfiguredPaths(t *testing.T) {
	dir := t.TempDir()
	for name, text := range map[string]string{
		"custos.json":   `{"php":"8.2","paths":["src/"],"baseline":"./baseline.json","rules":{"NestedNotOperators":{"enabled":true}}}`,
		"baseline.json": `{"version":1,"entries":[]}`,
		"src/a.php":     "<?php $ok=!!!$x;",
	} {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	rules := []analysis.Rule{nestednotoperators.New()}
	for _, paths := range [][]string{nil, {"./src/"}, {"src/a.php"}} {
		s, err := BuildSnapshot(context.Background(), root, rules, paths)
		if err != nil {
			t.Fatal(paths, err)
		}
		if len(s.Files) != 1 || s.Files[0] != "src/a.php" || s.Baseline != "baseline.json" || s.Sources[s.Baseline] == nil {
			t.Fatal(paths, s.Files, s.Baseline)
		}
	}
	if _, err := BuildSnapshot(context.Background(), root, rules, []string{"src/../.."}); err == nil {
		t.Fatal("escaped root")
	}
}
