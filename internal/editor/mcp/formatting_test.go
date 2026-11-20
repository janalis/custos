package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/rules/nestednotoperators"
	"github.com/janalis/custos/internal/project/config"
)

const (
	messyPHP     = "<?php\n\nfunction total($a,$b){\n  return $a+$b;\n}\n"
	tidyPHP      = "<?php\n\nfunction total($a, $b)\n{\n    return $a + $b;\n}\n"
	formatConfig = `{"php":"8.2","format":{"exclude":["legacy"]}}`
)

// formatFixture is a workspace with files to format, one already
// formatted, one excluded, one broken and one whose .editorconfig cannot be
// read.
func formatFixture(t *testing.T) (*Server, context.Context, string) {
	t.Helper()
	root, storage := t.TempDir(), t.TempDir()
	if err := os.Chmod(storage, 0o700); err != nil {
		t.Fatal(err)
	}
	var big strings.Builder
	big.WriteString("<?php\n")
	for i := range 300 {
		fmt.Fprintf(&big, "function f%d($a,$b){return $a;}\n", i)
	}
	for name, text := range map[string]string{
		"custos.json":          formatConfig,
		".editorconfig":        "root = true\n",
		"src/a.php":            messyPHP,
		"src/big.php":          big.String(),
		"src/tidy.php":         tidyPHP,
		"src/broken.php":       "<?php\nfunction f($a,$b){ return $a +; }\n",
		"legacy/old.php":       messyPHP,
		"odd/.editorconfig/.x": "",
		"odd/d.php":            messyPHP,
	} {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		writeFile(t, p, text)
	}
	s, err := New(Options{Default: "default", Storage: storage, Workspaces: []Workspace{{ID: "default", Root: root, Readers: []string{"local", "reader"}, Writers: []string{"local"}}}}, []analysis.Rule{nestednotoperators.New()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s, context.WithValue(context.WithValue(context.Background(), identityKey{}, "local"), writeScopeKey{}, true), root
}

func TestFormatCheck(t *testing.T) {
	s, ctx, _ := formatFixture(t)
	paths := []string{"src/a.php", "src/tidy.php", "src/broken.php", "legacy", "odd"}
	out := invoke(ctx, t, s, "format", Input{Paths: paths, Limit: 100})
	summary := out["summary"].(Output)
	if summary["files"] != 4 || summary["changed"] != 1 || summary["unchanged"] != 1 || summary["skipped"] != 2 {
		t.Fatalf("summary: %v", summary)
	}
	var first Output
	skipped := map[string]string{}
	for _, raw := range out["items"].([]any) {
		item := raw.(Output)
		if why, ok := item["skipped"]; ok {
			skipped[item["path"].(string)] = why.(string)
			continue
		}
		if item["path"] == "src/a.php" && first == nil {
			first = item
		}
		if strings.HasPrefix(item["path"].(string), "legacy/") {
			t.Fatal("an excluded file is formatted")
		}
	}
	if first["line"] != 3 || first["column"] != 19 || first["rule"] != "format.spacing" || first["message"] != "Fix the spaces between tokens" || first["fixable"] != true {
		t.Fatalf("first change: %v", first)
	}
	if !strings.Contains(skipped["src/broken.php"], "syntax error") || skipped["odd/d.php"] == "" {
		t.Fatalf("skipped: %v", skipped)
	}
	// Pages follow the cursor of the same selection.
	page := invoke(ctx, t, s, "format", Input{Mode: "check", Limit: 1})
	next := invoke(ctx, t, s, "format", Input{Mode: "check", Limit: 1, Cursor: page["nextCursor"].(string)})
	if len(next["items"].([]any)) != 1 {
		t.Fatal(next)
	}
	rejected(ctx, t, s, "format", Input{Mode: "check", Limit: 2, Cursor: page["nextCursor"].(string)})
	rejected(ctx, t, s, "format", Input{Mode: "rewrite"})
	rejected(ctx, t, s, "format", Input{Paths: []string{"../outside"}})
}

func TestFormatPreview(t *testing.T) {
	s, ctx, _ := formatFixture(t)
	out := invoke(ctx, t, s, "format", Input{Mode: "preview", Paths: []string{"src"}})
	diffs := map[string]string{}
	for _, raw := range out["items"].([]any) {
		item := raw.(Output)
		if d, ok := item["diff"]; ok {
			diffs[item["path"].(string)] = d.(string)
		}
	}
	if len(diffs) != 2 || !strings.Contains(diffs["src/a.php"], "+function total($a, $b)") || !strings.HasSuffix(diffs["src/big.php"], "format narrower paths to see the rest.") || out["untrusted"] == nil {
		t.Fatalf("preview: %v", out)
	}
}

func TestFormatApply(t *testing.T) {
	s, ctx, root := formatFixture(t)
	// Writing needs the write scope and a writer.
	rejected(context.WithValue(context.Background(), identityKey{}, "local"), t, s, "format", Input{Mode: "apply"})
	rejected(context.WithValue(context.WithValue(context.Background(), identityKey{}, "reader"), writeScopeKey{}, true), t, s, "format", Input{Mode: "apply"})
	invoke(context.WithValue(context.Background(), identityKey{}, "reader"), t, s, "format", Input{})
	out := invoke(ctx, t, s, "format", Input{Mode: "apply"})
	if out["applied"] != 2 || out["failedCount"] != 0 || out["skippedCount"] != 2 || out["analysis"] == nil {
		t.Fatalf("apply: %v", out)
	}
	if b, _ := os.ReadFile(filepath.Join(root, "src", "a.php")); string(b) != tidyPHP {
		t.Fatalf("a.php: %q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(root, "legacy", "old.php")); string(b) != messyPHP {
		t.Fatal("an excluded file was written")
	}
	again := invoke(ctx, t, s, "format", Input{Mode: "apply"})
	if again["applied"] != 0 || again["analysis"] != nil {
		t.Fatalf("nothing is left to format: %v", again)
	}
}

func TestFormatApplyFailures(t *testing.T) {
	s, ctx, root := formatFixture(t)
	writeFile(t, filepath.Join(root, "src", "y.php"), messyPHP)
	writeFile(t, filepath.Join(root, "src", "z.php"), messyPHP)
	saved := openTemporary
	defer func() { openTemporary = saved }()
	// The writes fail; the first one takes a later target away and turns
	// another into a link.
	openTemporary = func(*os.Root, string, os.FileMode) (*os.File, error) {
		if os.Remove(filepath.Join(root, "src", "y.php")) == nil {
			z := filepath.Join(root, "src", "z.php")
			_ = os.Remove(z)
			_ = os.Symlink(filepath.Join(root, "src", "a.php"), z)
		}
		return nil, errors.New("disk unavailable")
	}
	out := invoke(ctx, t, s, "format", Input{Mode: "apply"})
	if out["applied"] != 0 || out["failedCount"] != 4 || out["analysis"] != nil {
		t.Fatalf("apply: %v", out)
	}
	openTemporary = saved
	_ = os.Remove(filepath.Join(root, "src", "z.php"))
	// The analysis after writing fails on a broken configuration.
	beforeCommit = func(*os.Root, string) { writeFile(t, filepath.Join(root, "custos.json"), "invalid") }
	defer func() { beforeCommit = nil }()
	out = invoke(ctx, t, s, "format", Input{Mode: "apply", Paths: []string{"src/a.php"}})
	if out["applied"] != 1 || out["analysisError"] == nil {
		t.Fatalf("apply: %v", out)
	}
	rejected(ctx, t, s, "format", Input{})
}

func TestFormatResolverFailure(t *testing.T) {
	s, ctx, _ := formatFixture(t)
	saved := formatResolver
	defer func() { formatResolver = saved }()
	formatResolver = func(*config.Config) (*config.FormatResolver, error) { return nil, errors.New("no settings") }
	rejected(ctx, t, s, "format", Input{})
}

func TestFormatCanceled(t *testing.T) {
	s, ctx, _ := formatFixture(t)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := s.formatTool(canceled, s.workspaces["default"], Input{}); err == nil {
		t.Fatal("a canceled request formatted files")
	}
	// Canceled once the files are read.
	saved := formatResolver
	defer func() { formatResolver = saved }()
	later, stop := context.WithCancel(ctx)
	formatResolver = func(cfg *config.Config) (*config.FormatResolver, error) {
		stop()
		return saved(cfg)
	}
	if _, err := s.formatTool(later, s.workspaces["default"], Input{}); err == nil {
		t.Fatal("a canceled request formatted files")
	}
}
